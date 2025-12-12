package controllers

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/predictive-scaler/api/v1alpha1"
)

type PredictiveScaleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *PredictiveScaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var ps aiopsv1alpha1.PredictiveScale
	if err := r.Get(ctx, req.NamespacedName, &ps); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling PredictiveScale", "name", req.Name)

	// Query Grafana ML for forecast
	forecast, err := r.queryForecast(ctx, &ps)
	if err != nil {
		logger.Error(err, "Failed to query forecast")
		return ctrl.Result{RequeueAfter: 1 * 60}, nil
	}

	ps.Status.LastForecast = forecast
	ps.Status.Phase = "Monitoring"

	// Check if scaling is needed
	if forecast > ps.Spec.ThresholdPercent {
		if err := r.scaleNodePool(ctx, &ps); err != nil {
			logger.Error(err, "Failed to scale NodePool")
			ps.Status.Phase = "Failed"
		} else {
			now := metav1.Now()
			ps.Status.LastScaled = &now
			ps.Status.ScalingCount++
			ps.Status.Phase = "Complete"
		}
	}

	if err := r.Status().Update(ctx, &ps); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * 60}, nil
}

func (r *PredictiveScaleReconciler) queryForecast(ctx context.Context, ps *aiopsv1alpha1.PredictiveScale) (float64, error) {
	// In production: Query Grafana ML API
	// For now, return placeholder
	return 0.0, nil
}

func (r *PredictiveScaleReconciler) scaleNodePool(ctx context.Context, ps *aiopsv1alpha1.PredictiveScale) error {
	// In production: Patch Karpenter NodePool
	// For now, log action
	return nil
}

func (r *PredictiveScaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.PredictiveScale{}).
		Complete(r)
}
