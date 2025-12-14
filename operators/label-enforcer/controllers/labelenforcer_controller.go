package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/prophet/operators/label-enforcer/api/v1alpha1"
)

// LabelEnforcerReconciler reconciles a LabelEnforcer object
type LabelEnforcerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=labelenforcers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=labelenforcers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=labelenforcers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *LabelEnforcerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var labelEnforcer aiopsv1alpha1.LabelEnforcer
	if err := r.Get(ctx, req.NamespacedName, &labelEnforcer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling LabelEnforcer", "name", req.Name, "target", labelEnforcer.Spec.TargetResource)

	// Find and correct resources that need enforcement
	correctedCount, err := r.enforceLabelsAndAnnotations(ctx, &labelEnforcer)
	if err != nil {
		logger.Error(err, "Failed to enforce labels/annotations")
		return ctrl.Result{}, err
	}

	// Update status if corrections were made
	if correctedCount > 0 {
		labelEnforcer.Status.CorrectedResources = int32(correctedCount)
		labelEnforcer.Status.LastCorrected = &metav1.Time{Time: metav1.Now().Time}
		if err := r.Status().Update(ctx, &labelEnforcer); err != nil {
			logger.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
		logger.Info("Corrected resources", "count", correctedCount)
	}

	return ctrl.Result{}, nil
}

// enforceLabelsAndAnnotations finds resources and ensures they have required labels/annotations
func (r *LabelEnforcerReconciler) enforceLabelsAndAnnotations(ctx context.Context, enforcer *aiopsv1alpha1.LabelEnforcer) (int, error) {
	logger := log.FromContext(ctx)
	correctedCount := 0

	switch enforcer.Spec.TargetResource {
	case "pods":
		count, err := r.enforceOnPods(ctx, enforcer)
		if err != nil {
			return correctedCount, err
		}
		correctedCount += count
	case "deployments":
		count, err := r.enforceOnDeployments(ctx, enforcer)
		if err != nil {
			return correctedCount, err
		}
		correctedCount += count
	case "services":
		count, err := r.enforceOnServices(ctx, enforcer)
		if err != nil {
			return correctedCount, err
		}
		correctedCount += count
	case "configmaps":
		count, err := r.enforceOnConfigMaps(ctx, enforcer)
		if err != nil {
			return correctedCount, err
		}
		correctedCount += count
	case "secrets":
		count, err := r.enforceOnSecrets(ctx, enforcer)
		if err != nil {
			return correctedCount, err
		}
		correctedCount += count
	default:
		logger.Info("Unsupported target resource", "resource", enforcer.Spec.TargetResource)
	}

	return correctedCount, nil
}

// enforceOnPods ensures pods have required labels/annotations
func (r *LabelEnforcerReconciler) enforceOnPods(ctx context.Context, enforcer *aiopsv1alpha1.LabelEnforcer) (int, error) {
	logger := log.FromContext(ctx)
	correctedCount := 0

	var podList corev1.PodList
	listOpts := []client.ListOption{
		client.InNamespace(enforceNamespace(enforcer)),
	}

	if enforcer.Spec.LabelSelector != nil {
		selector := client.MatchingLabels(enforcer.Spec.LabelSelector)
		listOpts = append(listOpts, selector)
	}

	if err := r.List(ctx, &podList, listOpts...); err != nil {
		return correctedCount, err
	}

	for _, pod := range podList.Items {
		needsUpdate := false

		// Check and add required labels
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredLabels {
			if currentValue, exists := pod.Labels[key]; !exists || currentValue != value {
				pod.Labels[key] = value
				needsUpdate = true
			}
		}

		// Check and add required annotations
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredAnnotations {
			if currentValue, exists := pod.Annotations[key]; !exists || currentValue != value {
				pod.Annotations[key] = value
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := r.Update(ctx, &pod); err != nil {
				logger.Error(err, "Failed to update pod", "name", pod.Name)
				continue
			}
			correctedCount++
			logger.Info("Corrected pod labels/annotations", "name", pod.Name)
		}
	}

	return correctedCount, nil
}

// enforceOnDeployments ensures deployments have required labels/annotations
func (r *LabelEnforcerReconciler) enforceOnDeployments(ctx context.Context, enforcer *aiopsv1alpha1.LabelEnforcer) (int, error) {
	logger := log.FromContext(ctx)
	correctedCount := 0

	var deploymentList appsv1.DeploymentList
	listOpts := []client.ListOption{
		client.InNamespace(enforceNamespace(enforcer)),
	}

	if enforcer.Spec.LabelSelector != nil {
		selector := client.MatchingLabels(enforcer.Spec.LabelSelector)
		listOpts = append(listOpts, selector)
	}

	if err := r.List(ctx, &deploymentList, listOpts...); err != nil {
		return correctedCount, err
	}

	for _, deployment := range deploymentList.Items {
		needsUpdate := false

		// Check and add required labels
		if deployment.Labels == nil {
			deployment.Labels = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredLabels {
			if currentValue, exists := deployment.Labels[key]; !exists || currentValue != value {
				deployment.Labels[key] = value
				needsUpdate = true
			}
		}

		// Check and add required annotations
		if deployment.Annotations == nil {
			deployment.Annotations = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredAnnotations {
			if currentValue, exists := deployment.Annotations[key]; !exists || currentValue != value {
				deployment.Annotations[key] = value
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := r.Update(ctx, &deployment); err != nil {
				logger.Error(err, "Failed to update deployment", "name", deployment.Name)
				continue
			}
			correctedCount++
			logger.Info("Corrected deployment labels/annotations", "name", deployment.Name)
		}
	}

	return correctedCount, nil
}

// enforceOnServices ensures services have required labels/annotations
func (r *LabelEnforcerReconciler) enforceOnServices(ctx context.Context, enforcer *aiopsv1alpha1.LabelEnforcer) (int, error) {
	logger := log.FromContext(ctx)
	correctedCount := 0

	var serviceList corev1.ServiceList
	listOpts := []client.ListOption{
		client.InNamespace(enforceNamespace(enforcer)),
	}

	if enforcer.Spec.LabelSelector != nil {
		selector := client.MatchingLabels(enforcer.Spec.LabelSelector)
		listOpts = append(listOpts, selector)
	}

	if err := r.List(ctx, &serviceList, listOpts...); err != nil {
		return correctedCount, err
	}

	for _, service := range serviceList.Items {
		needsUpdate := false

		// Check and add required labels
		if service.Labels == nil {
			service.Labels = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredLabels {
			if currentValue, exists := service.Labels[key]; !exists || currentValue != value {
				service.Labels[key] = value
				needsUpdate = true
			}
		}

		// Check and add required annotations
		if service.Annotations == nil {
			service.Annotations = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredAnnotations {
			if currentValue, exists := service.Annotations[key]; !exists || currentValue != value {
				service.Annotations[key] = value
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := r.Update(ctx, &service); err != nil {
				logger.Error(err, "Failed to update service", "name", service.Name)
				continue
			}
			correctedCount++
			logger.Info("Corrected service labels/annotations", "name", service.Name)
		}
	}

	return correctedCount, nil
}

// enforceOnConfigMaps ensures configmaps have required labels/annotations
func (r *LabelEnforcerReconciler) enforceOnConfigMaps(ctx context.Context, enforcer *aiopsv1alpha1.LabelEnforcer) (int, error) {
	logger := log.FromContext(ctx)
	correctedCount := 0

	var configMapList corev1.ConfigMapList
	listOpts := []client.ListOption{
		client.InNamespace(enforceNamespace(enforcer)),
	}

	if enforcer.Spec.LabelSelector != nil {
		selector := client.MatchingLabels(enforcer.Spec.LabelSelector)
		listOpts = append(listOpts, selector)
	}

	if err := r.List(ctx, &configMapList, listOpts...); err != nil {
		return correctedCount, err
	}

	for _, configMap := range configMapList.Items {
		needsUpdate := false

		// Check and add required labels
		if configMap.Labels == nil {
			configMap.Labels = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredLabels {
			if currentValue, exists := configMap.Labels[key]; !exists || currentValue != value {
				configMap.Labels[key] = value
				needsUpdate = true
			}
		}

		// Check and add required annotations
		if configMap.Annotations == nil {
			configMap.Annotations = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredAnnotations {
			if currentValue, exists := configMap.Annotations[key]; !exists || currentValue != value {
				configMap.Annotations[key] = value
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := r.Update(ctx, &configMap); err != nil {
				logger.Error(err, "Failed to update configmap", "name", configMap.Name)
				continue
			}
			correctedCount++
			logger.Info("Corrected configmap labels/annotations", "name", configMap.Name)
		}
	}

	return correctedCount, nil
}

// enforceOnSecrets ensures secrets have required labels/annotations
func (r *LabelEnforcerReconciler) enforceOnSecrets(ctx context.Context, enforcer *aiopsv1alpha1.LabelEnforcer) (int, error) {
	logger := log.FromContext(ctx)
	correctedCount := 0

	var secretList corev1.SecretList
	listOpts := []client.ListOption{
		client.InNamespace(enforceNamespace(enforcer)),
	}

	if enforcer.Spec.LabelSelector != nil {
		selector := client.MatchingLabels(enforcer.Spec.LabelSelector)
		listOpts = append(listOpts, selector)
	}

	if err := r.List(ctx, &secretList, listOpts...); err != nil {
		return correctedCount, err
	}

	for _, secret := range secretList.Items {
		needsUpdate := false

		// Check and add required labels
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredLabels {
			if currentValue, exists := secret.Labels[key]; !exists || currentValue != value {
				secret.Labels[key] = value
				needsUpdate = true
			}
		}

		// Check and add required annotations
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}
		for key, value := range enforcer.Spec.RequiredAnnotations {
			if currentValue, exists := secret.Annotations[key]; !exists || currentValue != value {
				secret.Annotations[key] = value
				needsUpdate = true
			}
		}

		if needsUpdate {
			if err := r.Update(ctx, &secret); err != nil {
				logger.Error(err, "Failed to update secret", "name", secret.Name)
				continue
			}
			correctedCount++
			logger.Info("Corrected secret labels/annotations", "name", secret.Name)
		}
	}

	return correctedCount, nil
}

// enforceNamespace returns the namespace to enforce in, defaulting to all namespaces if empty
func enforceNamespace(enforcer *aiopsv1alpha1.LabelEnforcer) string {
	if enforcer.Spec.Namespace != "" {
		return enforcer.Spec.Namespace
	}
	return ""
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabelEnforcerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.LabelEnforcer{}).
		Complete(r)
}
