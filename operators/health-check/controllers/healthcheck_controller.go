package controllers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/health-check/api/v1alpha1"
)

// HealthCheckReconciler reconciles a HealthCheck object
type HealthCheckReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=healthchecks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=healthchecks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=healthchecks/finalizers,verbs=update
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=anomalyactions,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *HealthCheckReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var healthCheck aiopsv1alpha1.HealthCheck
	if err := r.Get(ctx, req.NamespacedName, &healthCheck); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling HealthCheck", "name", req.Name, "healthy", healthCheck.Status.Healthy)

	// Check if initial delay has passed
	if healthCheck.Status.LastCheckTime == nil {
		initialDelay := time.Duration(healthCheck.Spec.InitialDelaySeconds) * time.Second
		if initialDelay > 0 {
			logger.Info("Waiting for initial delay", "delay", initialDelay)
			return ctrl.Result{RequeueAfter: initialDelay}, nil
		}
	}

	// Execute all probes
	allHealthy := true
	probeResults := make([]aiopsv1alpha1.ProbeResult, 0, len(healthCheck.Spec.Probes))

	for _, probe := range healthCheck.Spec.Probes {
		result := r.executeProbe(ctx, &healthCheck, &probe)
		probeResults = append(probeResults, result)
		if !result.Success {
			allHealthy = false
		}
	}

	now := metav1.Now()
	healthCheck.Status.LastCheckTime = &now
	healthCheck.Status.ProbeResults = probeResults

	// Update failure count
	if !allHealthy {
		healthCheck.Status.FailureCount++
		if healthCheck.Status.LastFailureTime == nil {
			healthCheck.Status.LastFailureTime = &now
		}
	} else {
		healthCheck.Status.FailureCount = 0
		healthCheck.Status.LastFailureTime = nil
	}

	// Determine if workload is unhealthy based on failure threshold
	unhealthy := healthCheck.Status.FailureCount >= healthCheck.Spec.FailureThreshold

	// Update healthy status
	if unhealthy {
		healthCheck.Status.Healthy = false
		logger.Info("Health check failed", "failureCount", healthCheck.Status.FailureCount, "threshold", healthCheck.Spec.FailureThreshold)

		// Trigger remediation if configured
		if healthCheck.Spec.Remediation.Action != "" && healthCheck.Spec.Remediation.Action != "none" {
			if err := r.triggerRemediation(ctx, &healthCheck); err != nil {
				logger.Error(err, "Failed to trigger remediation")
				healthCheck.Status.ErrorMessage = err.Error()
			}
		}
	} else {
		healthCheck.Status.Healthy = true
	}

	// Update conditions
	condition := metav1.Condition{
		Type:               "Healthy",
		Status:             metav1.ConditionTrue,
		Reason:             "AllProbesPassed",
		Message:            "All health check probes are passing",
		LastTransitionTime: now,
	}
	if !healthCheck.Status.Healthy {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "ProbesFailing"
		condition.Message = fmt.Sprintf("%d consecutive failures (threshold: %d)", healthCheck.Status.FailureCount, healthCheck.Spec.FailureThreshold)
	}
	healthCheck.Status.Conditions = []metav1.Condition{condition}

	// Update status
	if err := r.Status().Update(ctx, &healthCheck); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue after period
	period := time.Duration(healthCheck.Spec.PeriodSeconds) * time.Second
	return ctrl.Result{RequeueAfter: period}, nil
}

// executeProbe executes a single health check probe
func (r *HealthCheckReconciler) executeProbe(ctx context.Context, healthCheck *aiopsv1alpha1.HealthCheck, probe *aiopsv1alpha1.ProbeSpec) aiopsv1alpha1.ProbeResult {
	logger := log.FromContext(ctx)
	result := aiopsv1alpha1.ProbeResult{
		Name:          probe.Name,
		LastCheckTime: &metav1.Time{Time: time.Now()},
	}

	// Get target pods to check
	pods, err := r.getTargetPods(ctx, healthCheck)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to get target pods: %v", err)
		return result
	}

	if len(pods) == 0 {
		result.Success = false
		result.Message = "No target pods found"
		return result
	}

	// Execute probe against first pod (or all pods for composite checks)
	timeout := time.Duration(healthCheck.Spec.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	switch probe.Type {
	case "http":
		result.Success = r.executeHTTPProbe(ctx, pods[0], probe.HTTPGet, timeout)
	case "tcp":
		result.Success = r.executeTCPProbe(ctx, pods[0], probe.TCPSocket, timeout)
	case "command":
		result.Success = r.executeCommandProbe(ctx, pods[0], probe.Exec, timeout)
	case "custom":
		result.Success = r.executeCustomProbe(ctx, pods[0], probe.Custom, timeout)
	default:
		result.Success = false
		result.Message = fmt.Sprintf("Unknown probe type: %s", probe.Type)
	}

	if !result.Success && result.Message == "" {
		result.Message = fmt.Sprintf("Probe %s failed", probe.Name)
	}

	return result
}

// getTargetPods retrieves pods for the target workload
func (r *HealthCheckReconciler) getTargetPods(ctx context.Context, healthCheck *aiopsv1alpha1.HealthCheck) ([]corev1.Pod, error) {
	namespace := healthCheck.Spec.TargetRef.Namespace
	if namespace == "" {
		namespace = healthCheck.Namespace
	}

	switch healthCheck.Spec.TargetRef.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: healthCheck.Spec.TargetRef.Name}, &pod); err != nil {
			return nil, err
		}
		return []corev1.Pod{pod}, nil

	case "Deployment":
		var deployment appsv1.Deployment
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: healthCheck.Spec.TargetRef.Name}, &deployment); err != nil {
			return nil, err
		}
		// Get pods matching deployment labels
		pods := &corev1.PodList{}
		selector := client.MatchingLabels(deployment.Spec.Selector.MatchLabels)
		if err := r.List(ctx, pods, client.InNamespace(namespace), selector); err != nil {
			return nil, err
		}
		return pods.Items, nil

	case "StatefulSet":
		var statefulSet appsv1.StatefulSet
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: healthCheck.Spec.TargetRef.Name}, &statefulSet); err != nil {
			return nil, err
		}
		// Get pods matching statefulset labels
		pods := &corev1.PodList{}
		selector := client.MatchingLabels(statefulSet.Spec.Selector.MatchLabels)
		if err := r.List(ctx, pods, client.InNamespace(namespace), selector); err != nil {
			return nil, err
		}
		return pods.Items, nil

	default:
		return nil, fmt.Errorf("unsupported target kind: %s", healthCheck.Spec.TargetRef.Kind)
	}
}

// executeHTTPProbe executes an HTTP health check
func (r *HealthCheckReconciler) executeHTTPProbe(ctx context.Context, pod corev1.Pod, httpGet *corev1.HTTPGetAction, timeout time.Duration) bool {
	if httpGet == nil {
		return false
	}

	// Build URL (simplified - in production, use pod IP or service endpoint)
	// For now, we'll check if the pod is ready (simplified implementation)
	// In production, this should make an actual HTTP request to the pod
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// executeTCPProbe executes a TCP health check
func (r *HealthCheckReconciler) executeTCPProbe(ctx context.Context, pod corev1.Pod, tcpSocket *corev1.TCPSocketAction, timeout time.Duration) bool {
	if tcpSocket == nil {
		return false
	}

	// Simplified: check if pod is running
	// In production, this should attempt a TCP connection to the pod IP:port
	if pod.Status.Phase == corev1.PodRunning {
		// Try to connect (simplified - in production, use pod IP)
		address := fmt.Sprintf("%s:%d", pod.Status.PodIP, tcpSocket.Port.IntVal)
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}
	return false
}

// executeCommandProbe executes a command-based health check
func (r *HealthCheckReconciler) executeCommandProbe(ctx context.Context, pod corev1.Pod, exec *corev1.ExecAction, timeout time.Duration) bool {
	if exec == nil {
		return false
	}

	// Simplified: check if pod is running
	// In production, this should exec into the pod and run the command
	// This requires pod exec permissions and a more complex implementation
	return pod.Status.Phase == corev1.PodRunning
}

// executeCustomProbe executes a custom health check
func (r *HealthCheckReconciler) executeCustomProbe(ctx context.Context, pod corev1.Pod, custom *aiopsv1alpha1.CustomProbe, timeout time.Duration) bool {
	if custom == nil {
		return false
	}

	// Simplified: for now, just check pod is running
	// In production, this would create a Job or Pod to execute the custom script
	// and check the result
	return pod.Status.Phase == corev1.PodRunning
}

// triggerRemediation triggers remediation actions when health check fails
func (r *HealthCheckReconciler) triggerRemediation(ctx context.Context, healthCheck *aiopsv1alpha1.HealthCheck) error {
	logger := log.FromContext(ctx)
	remediation := healthCheck.Spec.Remediation

	// Check cooldown
	if healthCheck.Status.LastRemediationTime != nil {
		cooldown := time.Duration(remediation.CooldownSeconds) * time.Second
		if time.Since(healthCheck.Status.LastRemediationTime.Time) < cooldown {
			logger.Info("In cooldown period, skipping remediation")
			return nil
		}
	}

	// Check if approval required
	if remediation.RequireApproval {
		logger.Info("Remediation requires approval, skipping")
		return nil
	}

	switch remediation.Action {
	case "restart":
		return r.restartTarget(ctx, healthCheck)

	case "trigger-recovery-plan":
		return r.triggerRecoveryPlan(ctx, healthCheck)

	case "alert":
		// Create event for alerting
		r.recordEvent(ctx, healthCheck, "Warning", "HealthCheckFailed", "Health check failed, alerting")
		return nil

	default:
		return fmt.Errorf("unknown remediation action: %s", remediation.Action)
	}
}

// restartTarget restarts the target workload
func (r *HealthCheckReconciler) restartTarget(ctx context.Context, healthCheck *aiopsv1alpha1.HealthCheck) error {
	logger := log.FromContext(ctx)
	pods, err := r.getTargetPods(ctx, healthCheck)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		logger.Info("Restarting pod due to health check failure", "pod", pod.Name)
		if err := r.Delete(ctx, &pod); err != nil {
			return err
		}
	}

	now := metav1.Now()
	healthCheck.Status.LastRemediationTime = &now
	healthCheck.Status.RemediationCount++

	return nil
}

// triggerRecoveryPlan triggers an AnomalyAction for recovery
func (r *HealthCheckReconciler) triggerRecoveryPlan(ctx context.Context, healthCheck *aiopsv1alpha1.HealthCheck) error {
	if healthCheck.Spec.Remediation.RecoveryPlanRef == nil {
		return fmt.Errorf("recoveryPlanRef not specified")
	}

	ref := healthCheck.Spec.Remediation.RecoveryPlanRef
	namespace := ref.Namespace
	if namespace == "" {
		namespace = healthCheck.Namespace
	}

	// Get or create AnomalyAction
	// For now, we'll just log - in production, create/update AnomalyAction
	logger := log.FromContext(ctx)
	logger.Info("Triggering recovery plan", "name", ref.Name, "namespace", namespace)

	// TODO: Create or update AnomalyAction to trigger recovery
	// This would involve creating an AnomalyAction with appropriate spec
	// to handle the health check failure

	return nil
}

// recordEvent records a Kubernetes event
func (r *HealthCheckReconciler) recordEvent(ctx context.Context, healthCheck *aiopsv1alpha1.HealthCheck, eventType, reason, message string) {
	// Create event object
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", healthCheck.Name),
			Namespace:    healthCheck.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: healthCheck.APIVersion,
			Kind:       healthCheck.Kind,
			Name:       healthCheck.Name,
			Namespace:  healthCheck.Namespace,
			UID:        healthCheck.UID,
		},
		Type:    eventType,
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "health-check-controller",
		},
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
		Count:          1,
	}

	_ = r.Create(ctx, event)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HealthCheckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.HealthCheck{}).
		Complete(r)
}

