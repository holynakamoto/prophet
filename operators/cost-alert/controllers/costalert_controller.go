package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/cost-alert/api/v1alpha1"
)

// CostAlertReconciler reconciles a CostAlert object
type CostAlertReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=costalerts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=costalerts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=costalerts/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *CostAlertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var costAlert aiopsv1alpha1.CostAlert
	if err := r.Get(ctx, req.NamespacedName, &costAlert); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling CostAlert", "name", req.Name, "scope", costAlert.Spec.Scope)

	// Fetch current cost
	currentCost, err := r.fetchCostData(ctx, &costAlert)
	if err != nil {
		logger.Error(err, "Failed to fetch cost data")
		costAlert.Status.ErrorMessage = err.Error()
		if err := r.Status().Update(ctx, &costAlert); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	now := metav1.Now()
	costAlert.Status.LastCheckTime = &now
	costAlert.Status.CurrentCost = currentCost

	// Check threshold
	triggered := false
	thresholdValue := costAlert.Spec.Threshold.Value

	switch costAlert.Spec.Threshold.Type {
	case "percentage_increase":
		// Compare with previous period
		if costAlert.Status.PreviousCost > 0 {
			increase := ((currentCost - costAlert.Status.PreviousCost) / costAlert.Status.PreviousCost) * 100
			if increase >= thresholdValue {
				triggered = true
				thresholdValue = increase
			}
		} else {
			// First check - store as baseline
			costAlert.Status.PreviousCost = currentCost
		}

	case "absolute":
		if currentCost >= thresholdValue {
			triggered = true
		}
	}

	// Update triggered status
	if triggered && !costAlert.Status.Triggered {
		// Alert just triggered
		costAlert.Status.Triggered = true
		costAlert.Status.LastTriggeredTime = &now
		costAlert.Status.TriggerCount++
		costAlert.Status.ThresholdValue = thresholdValue

		// Send notifications
		if err := r.sendAlert(ctx, &costAlert); err != nil {
			logger.Error(err, "Failed to send alert")
		}
	} else if !triggered {
		costAlert.Status.Triggered = false
	}

	// Update conditions
	condition := metav1.Condition{
		Type:               "AlertStatus",
		Status:             metav1.ConditionFalse,
		Reason:             "WithinThreshold",
		Message:            fmt.Sprintf("Current cost: %.2f %s", currentCost, costAlert.Spec.Threshold.Currency),
		LastTransitionTime: now,
	}
	if triggered {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "ThresholdExceeded"
		condition.Message = fmt.Sprintf("Cost threshold exceeded! Current: %.2f %s, Threshold: %.2f", currentCost, costAlert.Spec.Threshold.Currency, thresholdValue)
	}
	costAlert.Status.Conditions = []metav1.Condition{condition}

	// Update status
	if err := r.Status().Update(ctx, &costAlert); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue after check interval
	checkInterval := time.Duration(costAlert.Spec.CheckIntervalSeconds) * time.Second
	if checkInterval == 0 {
		checkInterval = 1 * time.Hour
	}
	return ctrl.Result{RequeueAfter: checkInterval}, nil
}

// fetchCostData fetches cost data from OpenCost/Kubecost API
func (r *CostAlertReconciler) fetchCostData(ctx context.Context, costAlert *aiopsv1alpha1.CostAlert) (float64, error) {
	endpoint := costAlert.Spec.OpenCostEndpoint
	if endpoint == "" {
		endpoint = "http://opencost.opencost.svc.cluster.local:9003"
	}

	// Build query based on scope
	var url string
	switch costAlert.Spec.Scope {
	case "workload":
		if costAlert.Spec.WorkloadRef == nil {
			return 0, fmt.Errorf("workloadRef is required for workload-scoped alert")
		}
		ref := costAlert.Spec.WorkloadRef
		url = fmt.Sprintf("%s/allocation?window=1d&aggregate=controller&controller=%s&namespace=%s",
			endpoint, ref.Name, ref.Namespace)
	case "namespace":
		if costAlert.Spec.Namespace == "" {
			return 0, fmt.Errorf("namespace is required for namespace-scoped alert")
		}
		url = fmt.Sprintf("%s/allocation?window=1d&aggregate=namespace&namespace=%s",
			endpoint, costAlert.Spec.Namespace)
	case "cluster":
		url = fmt.Sprintf("%s/allocation?window=1d&aggregate=cluster", endpoint)
	default:
		return 0, fmt.Errorf("unsupported scope: %s", costAlert.Spec.Scope)
	}

	// Make HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		// If OpenCost is not available, return a mock value for testing
		return 0, fmt.Errorf("failed to fetch cost data (OpenCost may not be deployed): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("OpenCost API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (simplified)
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Extract total cost
	totalCost := 0.0
	if allocations, ok := data["data"].(map[string]interface{}); ok {
		for _, allocation := range allocations {
			if alloc, ok := allocation.(map[string]interface{}); ok {
				if cost, ok := alloc["totalCost"].(float64); ok {
					totalCost += cost
				}
			}
		}
	}

	return totalCost, nil
}

// sendAlert sends cost alert notifications
func (r *CostAlertReconciler) sendAlert(ctx context.Context, costAlert *aiopsv1alpha1.CostAlert) error {
	logger := log.FromContext(ctx)

	// Send webhook notification
	if costAlert.Spec.Notify.WebhookURL != "" {
		// In production, send HTTP POST to webhook URL
		logger.Info("Sending cost alert webhook", "url", costAlert.Spec.Notify.WebhookURL)
	}

	// Create Kubernetes event
	r.recordEvent(ctx, costAlert, "Warning", "CostThresholdExceeded",
		fmt.Sprintf("Cost threshold exceeded! Current: %.2f %s, Threshold: %.2f",
			costAlert.Status.CurrentCost, costAlert.Spec.Threshold.Currency, costAlert.Status.ThresholdValue))

	// In production, also trigger PrometheusRule if AlertRuleRef is set
	if costAlert.Spec.AlertRuleRef != nil {
		logger.Info("Cost alert would trigger PrometheusRule", "name", costAlert.Spec.AlertRuleRef.Name)
	}

	return nil
}

// recordEvent records a Kubernetes event
func (r *CostAlertReconciler) recordEvent(ctx context.Context, costAlert *aiopsv1alpha1.CostAlert, eventType, reason, message string) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", costAlert.Name),
			Namespace:    costAlert.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: costAlert.APIVersion,
			Kind:       costAlert.Kind,
			Name:       costAlert.Name,
			Namespace:  costAlert.Namespace,
			UID:        costAlert.UID,
		},
		Type:    eventType,
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "cost-alert-controller",
		},
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
		Count:          1,
	}

	_ = r.Create(ctx, event)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CostAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.CostAlert{}).
		Complete(r)
}
