package controllers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/diagnostic-remediator/api/v1alpha1"
)

// DiagnosticRemediationReconciler reconciles a DiagnosticRemediation object
type DiagnosticRemediationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=diagnosticremediations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=diagnosticremediations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile performs diagnostic checks and remediation
func (r *DiagnosticRemediationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dr aiopsv1alpha1.DiagnosticRemediation
	if err := r.Get(ctx, req.NamespacedName, &dr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling DiagnosticRemediation", "name", req.Name, "phase", dr.Status.Phase)

	// Update phase to Diagnosing
	dr.Status.Phase = "Diagnosing"
	now := metav1.Now()
	dr.Status.LastDiagnosed = &now

	// Perform diagnostics
	issues := r.runDiagnostics(ctx, &dr, logger)
	dr.Status.Issues = issues

	if len(issues) > 0 {
		dr.Status.Phase = "IssuesFound"
		logger.Info("Issues found", "count", len(issues))

		// Check cooldown
		if dr.Status.LastRemediated != nil {
			cooldown := time.Duration(dr.Spec.CooldownSeconds) * time.Second
			if time.Since(dr.Status.LastRemediated.Time) < cooldown {
				logger.Info("In cooldown period, skipping remediation", "remaining", cooldown-time.Since(dr.Status.LastRemediated.Time))
				if err := r.Status().Update(ctx, &dr); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: cooldown - time.Since(dr.Status.LastRemediated.Time)}, nil
			}
		}

		// Guardrail: Check max remediations per hour (default: 6, configurable via annotation)
		maxRemediationsPerHour := 6
		if dr.Annotations != nil {
			if maxStr, ok := dr.Annotations["prophet.aiops.io/maxRemediationsPerHour"]; ok {
				if parsed, err := fmt.Sscanf(maxStr, "%d", &maxRemediationsPerHour); err == nil && parsed == 1 {
					// Parsed successfully
				} else {
					logger.Info("Failed to parse maxRemediationsPerHour annotation, using default", "value", maxStr, "default", 6)
					maxRemediationsPerHour = 6
				}
			}
		}

		// Count remediations in the last hour
		oneHourAgo := time.Now().Add(-1 * time.Hour)
		recentRemediations := 0
		for _, rem := range dr.Status.Remediations {
			if rem.Timestamp.After(oneHourAgo) && rem.Success {
				recentRemediations++
			}
		}

		if recentRemediations >= maxRemediationsPerHour {
			logger.Info("Max remediations per hour reached, skipping",
				"count", recentRemediations,
				"max", maxRemediationsPerHour,
				"nextWindow", oneHourAgo.Add(1*time.Hour))
			dr.Status.Phase = "IssuesFound" // Keep in IssuesFound, don't fail
			if err := r.Status().Update(ctx, &dr); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: time.Until(oneHourAgo.Add(1 * time.Hour))}, nil
		}

		// Perform remediation if auto-fix enabled
		if dr.Spec.AutoFix {
			dr.Status.Phase = "Remediating"
			remediations := r.performRemediation(ctx, &dr, issues, logger)
			dr.Status.Remediations = append(dr.Status.Remediations, remediations...)
			dr.Status.RemediationCount += int32(len(remediations))

			// Check if all remediations succeeded
			allSucceeded := true
			for _, rem := range remediations {
				if !rem.Success {
					allSucceeded = false
					break
				}
			}

			if allSucceeded && len(remediations) > 0 {
				dr.Status.Phase = "Resolved"
				now = metav1.Now()
				dr.Status.LastRemediated = &now
			} else if len(remediations) > 0 {
				dr.Status.Phase = "IssuesFound" // Some fixes failed, keep trying
			}
		}
	} else {
		dr.Status.Phase = "Resolved"
		logger.Info("No issues found")
	}

	if err := r.Status().Update(ctx, &dr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

// runDiagnostics performs all diagnostic checks
func (r *DiagnosticRemediationReconciler) runDiagnostics(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation, logger logr.Logger) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	// Get the target workload
	workload, err := r.getTargetWorkload(ctx, dr)
	if err != nil {
		logger.Error(err, "Failed to get target workload")
		issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
			Type:        "WorkloadNotFound",
			Severity:    "Critical",
			Description: fmt.Sprintf("Failed to find target workload: %v", err),
		})
		return issues
	}

	// Check resources
	if dr.Spec.Diagnostics.Resources {
		issues = append(issues, r.checkResources(ctx, workload, dr)...)
	}

	// Check environment variables
	if dr.Spec.Diagnostics.Environment {
		issues = append(issues, r.checkEnvironment(ctx, workload, dr)...)
	}

	// Check ConfigMap/Secret references
	if dr.Spec.Diagnostics.ConfigReferences {
		issues = append(issues, r.checkConfigReferences(ctx, workload, dr)...)
	}

	// Check service dependencies
	if len(dr.Spec.Diagnostics.ServiceDependencies) > 0 {
		issues = append(issues, r.checkServiceDependencies(ctx, dr)...)
	}

	// Check image pull policy
	if dr.Spec.Diagnostics.ImagePull {
		issues = append(issues, r.checkImagePullPolicy(ctx, workload)...)
	}

	// Check pod health (CrashLoopBackOff, high restart counts, stuck states)
	issues = append(issues, r.checkPodHealth(ctx, dr, logger)...)

	return issues
}

// getTargetWorkload retrieves the target Deployment/StatefulSet/DaemonSet
func (r *DiagnosticRemediationReconciler) getTargetWorkload(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation) (client.Object, error) {
	namespace := dr.Spec.Target.Namespace
	name := dr.Spec.Target.Name

	switch dr.Spec.Target.Kind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, deployment); err != nil {
			return nil, err
		}
		return deployment, nil
	case "StatefulSet":
		statefulSet := &appsv1.StatefulSet{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, statefulSet); err != nil {
			return nil, err
		}
		return statefulSet, nil
	case "DaemonSet":
		daemonSet := &appsv1.DaemonSet{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, daemonSet); err != nil {
			return nil, err
		}
		return daemonSet, nil
	default:
		return nil, fmt.Errorf("unsupported workload kind: %s", dr.Spec.Target.Kind)
	}
}

// checkResources checks if resource limits/requests are set
func (r *DiagnosticRemediationReconciler) checkResources(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	var containers []corev1.Container
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = w.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		containers = w.Spec.Template.Spec.Containers
	case *appsv1.DaemonSet:
		containers = w.Spec.Template.Spec.Containers
	}

	for i, container := range containers {
		if container.Resources.Requests == nil || len(container.Resources.Requests) == 0 {
			issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
				Type:         "MissingResources",
				Severity:     "Warning",
				Description:  fmt.Sprintf("Container %s has no resource requests", container.Name),
				Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
				SuggestedFix: "Add resource requests for CPU and memory",
			})
		}
		if container.Resources.Limits == nil || len(container.Resources.Limits) == 0 {
			issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
				Type:         "MissingResourceLimits",
				Severity:     "Warning",
				Description:  fmt.Sprintf("Container %s has no resource limits", container.Name),
				Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
				SuggestedFix: "Add resource limits for CPU and memory",
			})
		}
	}

	return issues
}

// checkEnvironment checks for required environment variables
func (r *DiagnosticRemediationReconciler) checkEnvironment(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	if len(dr.Spec.Remediation.RequiredEnvVars) == 0 {
		return issues
	}

	var containers []corev1.Container
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = w.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		containers = w.Spec.Template.Spec.Containers
	case *appsv1.DaemonSet:
		containers = w.Spec.Template.Spec.Containers
	}

	requiredVars := make(map[string]bool)
	for _, envVar := range dr.Spec.Remediation.RequiredEnvVars {
		requiredVars[envVar.Name] = true
	}

	for i, container := range containers {
		existingVars := make(map[string]bool)
		for _, envVar := range container.Env {
			existingVars[envVar.Name] = true
		}

		for varName := range requiredVars {
			if !existingVars[varName] {
				issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
					Type:         "MissingEnvVar",
					Severity:     "Critical",
					Description:  fmt.Sprintf("Container %s missing required environment variable: %s", container.Name, varName),
					Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
					SuggestedFix: fmt.Sprintf("Add environment variable %s", varName),
				})
			}
		}
	}

	return issues
}

// checkConfigReferences verifies ConfigMap/Secret references exist
func (r *DiagnosticRemediationReconciler) checkConfigReferences(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	var containers []corev1.Container
	var namespace string
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = w.Spec.Template.Spec.Containers
		namespace = w.Namespace
	case *appsv1.StatefulSet:
		containers = w.Spec.Template.Spec.Containers
		namespace = w.Namespace
	case *appsv1.DaemonSet:
		containers = w.Spec.Template.Spec.Containers
		namespace = w.Namespace
	}

	for i, container := range containers {
		// Check envFrom (ConfigMap/Secret references)
		for _, envFrom := range container.EnvFrom {
			if envFrom.ConfigMapRef != nil {
				cm := &corev1.ConfigMap{}
				if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: envFrom.ConfigMapRef.Name}, cm); err != nil {
					issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
						Type:         "MissingConfigMap",
						Severity:     "Critical",
						Description:  fmt.Sprintf("Container %s references non-existent ConfigMap: %s", container.Name, envFrom.ConfigMapRef.Name),
						Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
						SuggestedFix: fmt.Sprintf("Create ConfigMap %s in namespace %s", envFrom.ConfigMapRef.Name, namespace),
					})
				}
			}
			if envFrom.SecretRef != nil {
				secret := &corev1.Secret{}
				if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: envFrom.SecretRef.Name}, secret); err != nil {
					issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
						Type:         "MissingSecret",
						Severity:     "Critical",
						Description:  fmt.Sprintf("Container %s references non-existent Secret: %s", container.Name, envFrom.SecretRef.Name),
						Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
						SuggestedFix: fmt.Sprintf("Create Secret %s in namespace %s", envFrom.SecretRef.Name, namespace),
					})
				}
			}
		}

		// Check env valueFrom references
		for _, env := range container.Env {
			if env.ValueFrom != nil {
				if env.ValueFrom.ConfigMapKeyRef != nil {
					cm := &corev1.ConfigMap{}
					if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: env.ValueFrom.ConfigMapKeyRef.Name}, cm); err != nil {
						issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
							Type:         "MissingConfigMap",
							Severity:     "Critical",
							Description:  fmt.Sprintf("Container %s references non-existent ConfigMap key: %s/%s", container.Name, env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key),
							Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
							SuggestedFix: fmt.Sprintf("Create ConfigMap %s with key %s", env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key),
						})
					}
				}
				if env.ValueFrom.SecretKeyRef != nil {
					secret := &corev1.Secret{}
					if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: env.ValueFrom.SecretKeyRef.Name}, secret); err != nil {
						issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
							Type:         "MissingSecret",
							Severity:     "Critical",
							Description:  fmt.Sprintf("Container %s references non-existent Secret key: %s/%s", container.Name, env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key),
							Resource:     fmt.Sprintf("%s/%s/container[%d]", dr.Spec.Target.Kind, dr.Spec.Target.Name, i),
							SuggestedFix: fmt.Sprintf("Create Secret %s with key %s", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key),
						})
					}
				}
			}
		}
	}

	return issues
}

// checkServiceDependencies verifies service dependencies are available
func (r *DiagnosticRemediationReconciler) checkServiceDependencies(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	for _, dep := range dr.Spec.Diagnostics.ServiceDependencies {
		namespace := dep.Namespace
		if namespace == "" {
			namespace = dr.Spec.Target.Namespace
		}

		// Check if service exists
		svc := &corev1.Service{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: dep.Name}, svc); err != nil {
			issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
				Type:         "ServiceUnavailable",
				Severity:     "Critical",
				Description:  fmt.Sprintf("Service dependency %s/%s not found", namespace, dep.Name),
				Resource:     fmt.Sprintf("Service/%s", dep.Name),
				SuggestedFix: fmt.Sprintf("Create Service %s in namespace %s", dep.Name, namespace),
			})
			continue
		}

		// Check connectivity
		if dep.Protocol == "HTTP" || dep.Protocol == "HTTPS" {
			url := fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d%s", strings.ToLower(dep.Protocol), dep.Name, namespace, dep.Port, dep.Path)
			if !r.checkHTTPEndpoint(url) {
				issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
					Type:         "ServiceUnreachable",
					Severity:     "Warning",
					Description:  fmt.Sprintf("Service %s/%s endpoint not reachable: %s", namespace, dep.Name, url),
					Resource:     fmt.Sprintf("Service/%s", dep.Name),
					SuggestedFix: "Check service endpoints and pod readiness",
				})
			}
		} else if dep.Protocol == "TCP" || dep.Protocol == "" {
			address := fmt.Sprintf("%s.%s.svc.cluster.local:%d", dep.Name, namespace, dep.Port)
			if !r.checkTCPEndpoint(address) {
				issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
					Type:         "ServiceUnreachable",
					Severity:     "Warning",
					Description:  fmt.Sprintf("Service %s/%s TCP port %d not reachable", namespace, dep.Name, dep.Port),
					Resource:     fmt.Sprintf("Service/%s", dep.Name),
					SuggestedFix: "Check service endpoints and pod readiness",
				})
			}
		}
	}

	return issues
}

// checkImagePullPolicy checks if image pull policy is set appropriately
func (r *DiagnosticRemediationReconciler) checkImagePullPolicy(ctx context.Context, workload client.Object) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	var containers []corev1.Container
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = w.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		containers = w.Spec.Template.Spec.Containers
	case *appsv1.DaemonSet:
		containers = w.Spec.Template.Spec.Containers
	}

	for i, container := range containers {
		if container.ImagePullPolicy == "" || container.ImagePullPolicy == corev1.PullAlways {
			// Check if image tag is "latest" - should use PullAlways or specific tag
			if strings.Contains(container.Image, ":latest") || !strings.Contains(container.Image, ":") {
				issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
					Type:         "ImagePullPolicy",
					Severity:     "Warning",
					Description:  fmt.Sprintf("Container %s uses 'latest' tag without explicit pull policy", container.Name),
					Resource:     fmt.Sprintf("container[%d]", i),
					SuggestedFix: "Use specific image tags or set ImagePullPolicy: Always",
				})
			}
		}
	}

	return issues
}

// performRemediation applies fixes based on found issues
func (r *DiagnosticRemediationReconciler) performRemediation(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation, issues []aiopsv1alpha1.DiagnosticIssue, logger logr.Logger) []aiopsv1alpha1.RemediationAction {
	var remediations []aiopsv1alpha1.RemediationAction

	workload, err := r.getTargetWorkload(ctx, dr)
	if err != nil {
		logger.Error(err, "Failed to get workload for remediation")
		return remediations
	}

	needsUpdate := false

	// Fix resources
	if dr.Spec.Remediation.FixResources {
		for _, issue := range issues {
			if issue.Type == "MissingResources" || issue.Type == "MissingResourceLimits" {
				if fixed := r.fixResources(ctx, workload, dr); fixed {
					needsUpdate = true
					remediations = append(remediations, aiopsv1alpha1.RemediationAction{
						Type:        "AddedResources",
						Description: "Added default resource requests and limits",
						Timestamp:   metav1.Now(),
						Success:     true,
					})
				}
			}
		}
	}

	// Fix environment variables
	if dr.Spec.Remediation.FixEnvironment {
		for _, issue := range issues {
			if issue.Type == "MissingEnvVar" {
				if fixed := r.fixEnvironment(ctx, workload, dr); fixed {
					needsUpdate = true
					remediations = append(remediations, aiopsv1alpha1.RemediationAction{
						Type:        "AddedEnvVar",
						Description: "Added required environment variables",
						Timestamp:   metav1.Now(),
						Success:     true,
					})
				}
			}
		}
	}

	// Fix image pull policy
	if dr.Spec.Remediation.FixImagePullPolicy {
		for _, issue := range issues {
			if issue.Type == "ImagePullPolicy" {
				if fixed := r.fixImagePullPolicy(ctx, workload, dr); fixed {
					needsUpdate = true
					remediations = append(remediations, aiopsv1alpha1.RemediationAction{
						Type:        "UpdatedImagePullPolicy",
						Description: "Updated image pull policy",
						Timestamp:   metav1.Now(),
						Success:     true,
					})
				}
			}
		}
	}

	// Create missing ConfigMaps/Secrets
	if dr.Spec.Remediation.CreateMissingConfigs {
		for _, issue := range issues {
			if issue.Type == "MissingConfigMap" {
				if created := r.createMissingConfigMap(ctx, dr, issue); created {
					remediations = append(remediations, aiopsv1alpha1.RemediationAction{
						Type:        "CreatedConfigMap",
						Description: fmt.Sprintf("Created missing ConfigMap: %s", issue.Resource),
						Timestamp:   metav1.Now(),
						Success:     true,
					})
				}
			}
			if issue.Type == "MissingSecret" {
				if created := r.createMissingSecret(ctx, dr, issue); created {
					remediations = append(remediations, aiopsv1alpha1.RemediationAction{
						Type:        "CreatedSecret",
						Description: fmt.Sprintf("Created missing Secret: %s", issue.Resource),
						Timestamp:   metav1.Now(),
						Success:     true,
					})
				}
			}
		}
	}

	// Update workload if changes were made
	if needsUpdate {
		if err := r.Update(ctx, workload); err != nil {
			logger.Error(err, "Failed to update workload")
			remediations = append(remediations, aiopsv1alpha1.RemediationAction{
				Type:         "UpdateWorkload",
				Description:  "Failed to update workload with fixes",
				Timestamp:    metav1.Now(),
				Success:      false,
				ErrorMessage: err.Error(),
			})
		} else {
			// Restart pods if configured
			if dr.Spec.Remediation.RestartOnConfigChange {
				if err := r.restartPods(ctx, dr); err != nil {
					logger.Error(err, "Failed to restart pods")
				} else {
					remediations = append(remediations, aiopsv1alpha1.RemediationAction{
						Type:        "RestartedPods",
						Description: "Restarted pods after configuration changes",
						Timestamp:   metav1.Now(),
						Success:     true,
					})
				}
			}
		}
	}

	return remediations
}

// fixResources adds default resource requests/limits
func (r *DiagnosticRemediationReconciler) fixResources(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation) bool {
	changed := false
	defaultRes := dr.Spec.Remediation.DefaultResources

	var containers *[]corev1.Container
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = &w.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		containers = &w.Spec.Template.Spec.Containers
	case *appsv1.DaemonSet:
		containers = &w.Spec.Template.Spec.Containers
	}

	for i := range *containers {
		container := &(*containers)[i]
		if container.Resources.Requests == nil {
			container.Resources.Requests = make(corev1.ResourceList)
			changed = true
		}
		if container.Resources.Limits == nil {
			container.Resources.Limits = make(corev1.ResourceList)
			changed = true
		}

		if defaultRes.CPURequest != "" && container.Resources.Requests[corev1.ResourceCPU] == (resource.Quantity{}) {
			qty, _ := resource.ParseQuantity(defaultRes.CPURequest)
			container.Resources.Requests[corev1.ResourceCPU] = qty
			changed = true
		}
		if defaultRes.MemoryRequest != "" && container.Resources.Requests[corev1.ResourceMemory] == (resource.Quantity{}) {
			qty, _ := resource.ParseQuantity(defaultRes.MemoryRequest)
			container.Resources.Requests[corev1.ResourceMemory] = qty
			changed = true
		}
		if defaultRes.CPULimit != "" && container.Resources.Limits[corev1.ResourceCPU] == (resource.Quantity{}) {
			qty, _ := resource.ParseQuantity(defaultRes.CPULimit)
			container.Resources.Limits[corev1.ResourceCPU] = qty
			changed = true
		}
		if defaultRes.MemoryLimit != "" && container.Resources.Limits[corev1.ResourceMemory] == (resource.Quantity{}) {
			qty, _ := resource.ParseQuantity(defaultRes.MemoryLimit)
			container.Resources.Limits[corev1.ResourceMemory] = qty
			changed = true
		}
	}

	return changed
}

// fixEnvironment adds required environment variables
func (r *DiagnosticRemediationReconciler) fixEnvironment(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation) bool {
	changed := false

	var containers *[]corev1.Container
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = &w.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		containers = &w.Spec.Template.Spec.Containers
	case *appsv1.DaemonSet:
		containers = &w.Spec.Template.Spec.Containers
	}

	for i := range *containers {
		container := &(*containers)[i]
		existingVars := make(map[string]bool)
		for _, env := range container.Env {
			existingVars[env.Name] = true
		}

		for _, requiredVar := range dr.Spec.Remediation.RequiredEnvVars {
			if !existingVars[requiredVar.Name] {
				envVar := corev1.EnvVar{
					Name: requiredVar.Name,
				}
				if requiredVar.Value != "" {
					envVar.Value = requiredVar.Value
				} else if requiredVar.ValueFrom != nil {
					envVar.ValueFrom = &corev1.EnvVarSource{}
					if requiredVar.ValueFrom.ConfigMapKeyRef != nil {
						envVar.ValueFrom.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: requiredVar.ValueFrom.ConfigMapKeyRef.Name,
							},
							Key: requiredVar.ValueFrom.ConfigMapKeyRef.Key,
						}
					}
					if requiredVar.ValueFrom.SecretKeyRef != nil {
						envVar.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: requiredVar.ValueFrom.SecretKeyRef.Name,
							},
							Key: requiredVar.ValueFrom.SecretKeyRef.Key,
						}
					}
				}
				container.Env = append(container.Env, envVar)
				changed = true
			}
		}
	}

	return changed
}

// fixImagePullPolicy updates image pull policy
func (r *DiagnosticRemediationReconciler) fixImagePullPolicy(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation) bool {
	changed := false
	policy := dr.Spec.Remediation.DefaultImagePullPolicy
	if policy == "" {
		policy = string(corev1.PullIfNotPresent)
	}

	var containers *[]corev1.Container
	switch w := workload.(type) {
	case *appsv1.Deployment:
		containers = &w.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		containers = &w.Spec.Template.Spec.Containers
	case *appsv1.DaemonSet:
		containers = &w.Spec.Template.Spec.Containers
	}

	for i := range *containers {
		container := &(*containers)[i]
		if container.ImagePullPolicy != corev1.PullPolicy(policy) {
			container.ImagePullPolicy = corev1.PullPolicy(policy)
			changed = true
		}
	}

	return changed
}

// createMissingConfigMap creates a ConfigMap if it doesn't exist
func (r *DiagnosticRemediationReconciler) createMissingConfigMap(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation, issue aiopsv1alpha1.DiagnosticIssue) bool {
	// Extract ConfigMap name from issue description
	// This is a simplified implementation - in production, parse the issue more carefully
	namespace := dr.Spec.Target.Namespace
	cmName := extractResourceName(issue.Description, "ConfigMap")

	if cmName == "" {
		return false
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"placeholder": "created-by-diagnostic-remediator",
		},
	}

	if err := r.Create(ctx, cm); err != nil {
		// ConfigMap might already exist, which is fine
		return false
	}

	return true
}

// createMissingSecret creates a Secret if it doesn't exist
func (r *DiagnosticRemediationReconciler) createMissingSecret(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation, issue aiopsv1alpha1.DiagnosticIssue) bool {
	namespace := dr.Spec.Target.Namespace
	secretName := extractResourceName(issue.Description, "Secret")

	if secretName == "" {
		return false
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"placeholder": []byte("created-by-diagnostic-remediator"),
		},
	}

	if err := r.Create(ctx, secret); err != nil {
		return false
	}

	return true
}

// restartPods restarts pods by deleting them (ReplicaSet will recreate)
func (r *DiagnosticRemediationReconciler) restartPods(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation) error {
	pods := &corev1.PodList{}
	selector := client.MatchingLabels(dr.Spec.Target.Labels)
	if err := r.List(ctx, pods, client.InNamespace(dr.Spec.Target.Namespace), selector); err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if err := r.Delete(ctx, &pod); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions
func (r *DiagnosticRemediationReconciler) checkHTTPEndpoint(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

func (r *DiagnosticRemediationReconciler) checkTCPEndpoint(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func extractResourceName(description, resourceType string) string {
	// Simple extraction - in production, use regex or better parsing
	parts := strings.Split(description, resourceType+":")
	if len(parts) > 1 {
		name := strings.TrimSpace(parts[1])
		// Remove any trailing punctuation
		name = strings.TrimRight(name, ".,;")
		return name
	}
	return ""
}

// SetupWithManager sets up the controller with the Manager
// checkPodHealth checks for pod health issues: CrashLoopBackOff, high restart counts, stuck states
func (r *DiagnosticRemediationReconciler) checkPodHealth(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation, logger logr.Logger) []aiopsv1alpha1.DiagnosticIssue {
	var issues []aiopsv1alpha1.DiagnosticIssue

	// Get pods for the target workload
	pods := &corev1.PodList{}
	selector := client.MatchingLabels(dr.Spec.Target.Labels)
	if len(dr.Spec.Target.Labels) == 0 {
		// If no labels specified, try to find pods by owner reference
		workload, err := r.getTargetWorkload(ctx, dr)
		if err == nil {
			switch w := workload.(type) {
			case *appsv1.Deployment:
				selector = client.MatchingLabels(w.Spec.Selector.MatchLabels)
			case *appsv1.StatefulSet:
				selector = client.MatchingLabels(w.Spec.Selector.MatchLabels)
			case *appsv1.DaemonSet:
				selector = client.MatchingLabels(w.Spec.Selector.MatchLabels)
			}
		}
	}

	if err := r.List(ctx, pods, client.InNamespace(dr.Spec.Target.Namespace), selector); err != nil {
		logger.Error(err, "Failed to list pods")
		return issues
	}

	for _, pod := range pods.Items {
		// Check for CrashLoopBackOff
		if pod.Status.Phase == corev1.PodFailed {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					if containerStatus.State.Waiting.Reason == "CrashLoopBackOff" ||
						containerStatus.State.Waiting.Reason == "ImagePullBackOff" ||
						containerStatus.State.Waiting.Reason == "ErrImagePull" {
						issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
							Type:        "PodCrashLoopBackOff",
							Severity:    "Critical",
							Description: fmt.Sprintf("Pod %s is in %s state: %s", pod.Name, containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message),
							Resource:    fmt.Sprintf("pod/%s", pod.Name),
						})
					}
				}
			}
		}

		// Check for high restart counts (>3)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 3 {
				issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
					Type:        "PodHighRestartCount",
					Severity:    "Warning",
					Description: fmt.Sprintf("Pod %s container %s has %d restarts", pod.Name, containerStatus.Name, containerStatus.RestartCount),
					Resource:    fmt.Sprintf("pod/%s", pod.Name),
				})
			}
		}

		// Check for stuck states (ContainerCreating for >5 minutes, Pending for >5 minutes)
		if pod.Status.Phase == corev1.PodPending {
			age := time.Since(pod.CreationTimestamp.Time)
			if age > 5*time.Minute {
				issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
					Type:        "PodStuck",
					Severity:    "Warning",
					Description: fmt.Sprintf("Pod %s has been in Pending state for %v", pod.Name, age),
					Resource:    fmt.Sprintf("pod/%s", pod.Name),
				})
			}
		}

		// Check for containers stuck in ContainerCreating
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "ContainerCreating" {
				// Check if it's been creating for too long
				age := time.Since(pod.CreationTimestamp.Time)
				if age > 5*time.Minute {
					issues = append(issues, aiopsv1alpha1.DiagnosticIssue{
						Type:        "PodStuck",
						Severity:    "Warning",
						Description: fmt.Sprintf("Pod %s container %s stuck in ContainerCreating for %v", pod.Name, containerStatus.Name, age),
						Resource:    fmt.Sprintf("pod/%s", pod.Name),
					})
				}
			}
		}
	}

	return issues
}

// remediatePodHealth remediates pod health issues
// For Helm-managed resources, prefers rollout restart over pod deletion
func (r *DiagnosticRemediationReconciler) remediatePodHealth(ctx context.Context, dr *aiopsv1alpha1.DiagnosticRemediation, issue aiopsv1alpha1.DiagnosticIssue, logger logr.Logger) bool {
	// Get the target workload first to check if it's Helm-managed
	workload, err := r.getTargetWorkload(ctx, dr)
	if err != nil {
		logger.Error(err, "Failed to get workload for remediation")
		return false
	}

	// Check if workload is Helm-managed
	isHelmManaged := false
	switch w := workload.(type) {
	case *appsv1.Deployment:
		if w.Labels != nil {
			if w.Labels["app.kubernetes.io/managed-by"] == "Helm" || w.Labels["heritage"] == "Helm" {
				isHelmManaged = true
				logger.Info("Detected Helm-managed Deployment", "release", w.Labels["release"], "chart", w.Labels["chart"])
			}
		}
	}

	// For Helm-managed resources, always use rollout restart (safer)
	// For non-Helm resources, use rollout restart for stuck pods, delete for crash loops
	if isHelmManaged {
		return r.triggerRolloutRestart(ctx, workload, dr, logger)
	}

	// For non-Helm resources, extract pod name for potential deletion
	parts := strings.Split(issue.Resource, "/")
	if len(parts) != 2 || parts[0] != "pod" {
		logger.Info("Invalid pod resource format, using rollout restart", "resource", issue.Resource)
		return r.triggerRolloutRestart(ctx, workload, dr, logger)
	}
	podName := parts[1]

	// For CrashLoopBackOff or high restart counts on non-Helm resources, delete pod
	if issue.Type == "PodCrashLoopBackOff" || issue.Type == "PodHighRestartCount" {
		pod := &corev1.Pod{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: dr.Spec.Target.Namespace, Name: podName}, pod); err != nil {
			logger.Error(err, "Failed to get pod, falling back to rollout restart", "pod", podName)
			return r.triggerRolloutRestart(ctx, workload, dr, logger)
		}
		logger.Info("Deleting failing pod to trigger recreation", "pod", podName, "reason", issue.Type)
		if err := r.Delete(ctx, pod); err != nil {
			logger.Error(err, "Failed to delete pod, falling back to rollout restart", "pod", podName)
			return r.triggerRolloutRestart(ctx, workload, dr, logger)
		}
		return true
	}

	// For stuck pods, use rollout restart
	if issue.Type == "PodStuck" {
		return r.triggerRolloutRestart(ctx, workload, dr, logger)
	}

	return false
}

// triggerRolloutRestart triggers a rollout restart by updating deployment annotation
// This is equivalent to `kubectl rollout restart deployment/name -n namespace`
// Includes idempotency check to avoid unnecessary restarts
func (r *DiagnosticRemediationReconciler) triggerRolloutRestart(ctx context.Context, workload client.Object, dr *aiopsv1alpha1.DiagnosticRemediation, logger logr.Logger) bool {
	switch w := workload.(type) {
	case *appsv1.Deployment:
		// Idempotency check: Don't restart if we just restarted recently (within last 2 minutes)
		if w.Spec.Template.Annotations != nil {
			if lastRestart, ok := w.Spec.Template.Annotations["prophet.aiops.io/restartedAt"]; ok {
				if lastRestartTime, err := time.Parse(time.RFC3339, lastRestart); err == nil {
					if time.Since(lastRestartTime) < 2*time.Minute {
						logger.Info("Skipping rollout restart - recent restart detected",
							"lastRestart", lastRestartTime,
							"deployment", w.Name)
						return true // Return true because we're in the desired state
					}
				}
			}
		}

		if w.Spec.Template.Annotations == nil {
			w.Spec.Template.Annotations = make(map[string]string)
		}
		// Use kubectl.kubernetes.io/restartedAt annotation (standard Kubernetes pattern)
		// This triggers a rollout restart without changing the deployment spec
		restartTime := time.Now().Format(time.RFC3339)
		w.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

		// Also track our own annotation for monitoring
		w.Spec.Template.Annotations["prophet.aiops.io/restartedAt"] = restartTime
		w.Spec.Template.Annotations["prophet.aiops.io/restartReason"] = "pod-health-remediation"
		w.Spec.Template.Annotations["prophet.aiops.io/restartedBy"] = "diagnostic-remediator"

		logger.Info("Triggering rollout restart for Helm-managed Deployment",
			"deployment", w.Name,
			"namespace", w.Namespace,
			"release", w.Labels["release"],
			"restartTime", restartTime)

		if err := r.Update(ctx, w); err != nil {
			logger.Error(err, "Failed to trigger rollout restart")
			return false
		}
		return true
	case *appsv1.StatefulSet:
		// StatefulSets also support rollout restart via annotation
		if w.Spec.Template.Annotations == nil {
			w.Spec.Template.Annotations = make(map[string]string)
		}
		w.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
		w.Spec.Template.Annotations["prophet.aiops.io/restartedAt"] = time.Now().Format(time.RFC3339)
		logger.Info("Triggering rollout restart for StatefulSet", "statefulset", w.Name, "namespace", w.Namespace)
		if err := r.Update(ctx, w); err != nil {
			logger.Error(err, "Failed to trigger rollout restart")
			return false
		}
		return true
	default:
		logger.Info("Workload type does not support rollout restart", "type", fmt.Sprintf("%T", w))
		return false
	}
}

func (r *DiagnosticRemediationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.DiagnosticRemediation{}).
		Complete(r)
}
