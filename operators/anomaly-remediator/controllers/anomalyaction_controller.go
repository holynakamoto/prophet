package controllers

import (
	"context"
	"fmt"
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

	aiopsv1alpha1 "github.com/prophet-aiops/anomaly-remediator/api/v1alpha1"
)

// AnomalyActionReconciler reconciles a AnomalyAction object
type AnomalyActionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=anomalyactions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=anomalyactions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=anomalyactions/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AnomalyActionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var anomalyAction aiopsv1alpha1.AnomalyAction
	if err := r.Get(ctx, req.NamespacedName, &anomalyAction); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling AnomalyAction", "name", req.Name, "phase", anomalyAction.Status.Phase)

	// Check for anomalies (simplified - in production, query Prometheus/Grafana ML)
	anomalyDetected := r.detectAnomaly(ctx, &anomalyAction)

	if anomalyDetected {
		now := metav1.Now()
		anomalyAction.Status.LastDetected = &now
		anomalyAction.Status.Phase = "Detected"

		// Check cooldown period
		if anomalyAction.Status.LastRemediated != nil {
			cooldown := time.Duration(anomalyAction.Spec.Remediation.CooldownSeconds) * time.Second
			if time.Since(anomalyAction.Status.LastRemediated.Time) < cooldown {
				logger.Info("In cooldown period, skipping remediation", "remaining", cooldown-time.Since(anomalyAction.Status.LastRemediated.Time))
				return ctrl.Result{RequeueAfter: cooldown - time.Since(anomalyAction.Status.LastRemediated.Time)}, nil
			}
		}

		// Check if approval is required
		if anomalyAction.Spec.Remediation.RequireApproval {
			anomalyAction.Status.Phase = "PendingApproval"
			logger.Info("Remediation requires approval, waiting")
			if err := r.Status().Update(ctx, &anomalyAction); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}

		// Perform remediation
		if err := r.remediate(ctx, &anomalyAction); err != nil {
			logger.Error(err, "Failed to remediate")
			anomalyAction.Status.Phase = "Failed"
			anomalyAction.Status.ErrorMessage = err.Error()
			if err := r.Status().Update(ctx, &anomalyAction); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}

		// Update status
		now = metav1.Now()
		anomalyAction.Status.LastRemediated = &now
		anomalyAction.Status.RemediationCount++
		anomalyAction.Status.Phase = "Resolved"
	} else {
		if anomalyAction.Status.Phase == "Detected" || anomalyAction.Status.Phase == "Resolved" {
			anomalyAction.Status.Phase = "Pending"
		}
	}

	// Trigger K8sGPT analysis if enabled
	if anomalyDetected && anomalyAction.Spec.K8sGPT.Enabled {
		if analysis, err := r.triggerK8sGPTAnalysis(ctx, &anomalyAction); err == nil {
			anomalyAction.Status.K8sGPTAnalysis = analysis
		}
	}

	// Send webhook notification if configured
	if anomalyDetected && anomalyAction.Spec.WebhookURL != "" {
		r.sendWebhookNotification(ctx, &anomalyAction)
	}

	if err := r.Status().Update(ctx, &anomalyAction); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// detectAnomaly checks if an anomaly is present (simplified implementation)
func (r *AnomalyActionReconciler) detectAnomaly(ctx context.Context, action *aiopsv1alpha1.AnomalyAction) bool {
	// In production, this would:
	// 1. Query Prometheus for the metric
	// 2. Query Grafana ML for anomaly detection
	// 3. Check OpenTelemetry traces
	// For now, we'll simulate by checking pod status
	logger := log.FromContext(ctx)

	if action.Spec.Source == "prometheus" || action.Spec.Source == "grafana-ml" {
		// Placeholder: In production, query Prometheus client
		logger.Info("Checking for anomalies via Prometheus/Grafana ML", "metric", action.Spec.Metric)
		// Return true for demo purposes - in production, implement actual query
		return false // Change to true to trigger remediation
	}

	// Fallback: Check pod status
	pods := &corev1.PodList{}
	selector := client.MatchingLabels(action.Spec.Target.Labels)
	if err := r.List(ctx, pods, client.InNamespace(action.Spec.Target.Namespace), selector); err != nil {
		logger.Error(err, "Failed to list pods")
		return false
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
			logger.Info("Anomaly detected: pod in failed/unknown state", "pod", pod.Name)
			return true
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				logger.Info("Anomaly detected: pod not ready", "pod", pod.Name)
				return true
			}
		}
	}

	return false
}

// remediate performs the remediation action
func (r *AnomalyActionReconciler) remediate(ctx context.Context, action *aiopsv1alpha1.AnomalyAction) error {
	logger := log.FromContext(ctx)
	remediationType := action.Spec.Remediation.Type

	switch remediationType {
	case "restart":
		return r.restartPods(ctx, action)
	case "scale":
		return r.scaleResource(ctx, action)
	case "alert":
		logger.Info("Alert-only remediation, no action taken")
		return nil
	default:
		return fmt.Errorf("unknown remediation type: %s", remediationType)
	}
}

// restartPods restarts pods matching the selector
func (r *AnomalyActionReconciler) restartPods(ctx context.Context, action *aiopsv1alpha1.AnomalyAction) error {
	logger := log.FromContext(ctx)
	pods := &corev1.PodList{}
	selector := client.MatchingLabels(action.Spec.Target.Labels)
	if err := r.List(ctx, pods, client.InNamespace(action.Spec.Target.Namespace), selector); err != nil {
		return err
	}

	for _, pod := range pods.Items {
		logger.Info("Restarting pod", "pod", pod.Name, "namespace", pod.Namespace)
		if err := r.Delete(ctx, &pod); err != nil {
			return err
		}
	}

	return nil
}

// scaleResource scales a Deployment or StatefulSet
func (r *AnomalyActionReconciler) scaleResource(ctx context.Context, action *aiopsv1alpha1.AnomalyAction) error {
	logger := log.FromContext(ctx)
	resourceType := action.Spec.Target.ResourceType

	if action.Spec.Remediation.Replicas == nil {
		return fmt.Errorf("replicas not specified for scale action")
	}

	replicas := *action.Spec.Remediation.Replicas

	switch resourceType {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		key := types.NamespacedName{
			Namespace: action.Spec.Target.Namespace,
			Name:      action.Spec.Target.Labels["app"], // Simplified - use proper selector
		}
		if err := r.Get(ctx, key, deployment); err != nil {
			return err
		}
		logger.Info("Scaling deployment", "name", deployment.Name, "replicas", replicas)
		deployment.Spec.Replicas = &replicas
		return r.Update(ctx, deployment)

	case "StatefulSet":
		statefulSet := &appsv1.StatefulSet{}
		key := types.NamespacedName{
			Namespace: action.Spec.Target.Namespace,
			Name:      action.Spec.Target.Labels["app"],
		}
		if err := r.Get(ctx, key, statefulSet); err != nil {
			return err
		}
		logger.Info("Scaling statefulset", "name", statefulSet.Name, "replicas", replicas)
		statefulSet.Spec.Replicas = &replicas
		return r.Update(ctx, statefulSet)

	default:
		return fmt.Errorf("unsupported resource type for scaling: %s", resourceType)
	}
}

// triggerK8sGPTAnalysis triggers K8sGPT analysis
func (r *AnomalyActionReconciler) triggerK8sGPTAnalysis(ctx context.Context, action *aiopsv1alpha1.AnomalyAction) (string, error) {
	logger := log.FromContext(ctx)
	endpoint := action.Spec.K8sGPT.Endpoint
	if endpoint == "" {
		endpoint = "http://k8sgpt-operator.k8sgpt.svc.cluster.local:8080"
	}

	logger.Info("Triggering K8sGPT analysis", "endpoint", endpoint)
	// In production, make HTTP request to K8sGPT API
	// For now, return placeholder
	return "K8sGPT analysis: Anomaly detected in " + action.Spec.Metric + ". Suggested action: " + action.Spec.Remediation.Type, nil
}

// sendWebhookNotification sends webhook notification
func (r *AnomalyActionReconciler) sendWebhookNotification(ctx context.Context, action *aiopsv1alpha1.AnomalyAction) {
	logger := log.FromContext(ctx)
	logger.Info("Sending webhook notification", "url", action.Spec.WebhookURL)
	// In production, make HTTP POST to webhook URL
}

// SetupWithManager sets up the controller with the Manager.
func (r *AnomalyActionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.AnomalyAction{}).
		Complete(r)
}
