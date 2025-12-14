package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/budget-guard/api/v1alpha1"
)

// BudgetGuardReconciler reconciles a BudgetGuard object
type BudgetGuardReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=budgetguards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=budgetguards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=budgetguards/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete;evict
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *BudgetGuardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var budgetGuard aiopsv1alpha1.BudgetGuard
	if err := r.Get(ctx, req.NamespacedName, &budgetGuard); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling BudgetGuard", "name", req.Name, "scope", budgetGuard.Spec.Scope)

	// Fetch cost data from OpenCost/Kubecost
	currentSpend, err := r.fetchCostData(ctx, &budgetGuard)
	if err != nil {
		logger.Error(err, "Failed to fetch cost data")
		budgetGuard.Status.ErrorMessage = err.Error()
		if err := r.Status().Update(ctx, &budgetGuard); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// Update status
	now := metav1.Now()
	budgetGuard.Status.LastRefreshTime = &now
	budgetGuard.Status.CurrentSpend = currentSpend
	budgetGuard.Status.BudgetLimit = budgetGuard.Spec.Budget.Amount
	budgetGuard.Status.PercentageUsed = (currentSpend / budgetGuard.Spec.Budget.Amount) * 100

	// Check if budget is exceeded
	exceeded := currentSpend >= budgetGuard.Spec.Budget.Amount
	budgetGuard.Status.Exceeded = exceeded

	// Project exceed time (simplified calculation)
	if !exceeded && budgetGuard.Status.PercentageUsed > 0 {
		// Estimate time to exceed based on current spend rate
		// This is a simplified calculation - in production, use historical data
		daysRemaining := (100 - budgetGuard.Status.PercentageUsed) / (budgetGuard.Status.PercentageUsed / float64(time.Since(budgetGuard.CreationTimestamp.Time).Hours()/24))
		if daysRemaining > 0 {
			projectedTime := metav1.NewTime(time.Now().Add(time.Duration(daysRemaining*24) * time.Hour))
			budgetGuard.Status.ProjectedExceedTime = &projectedTime
		}
	}

	// Take actions if budget is exceeded
	if exceeded {
		actionsTaken := []string{}
		if err := r.enforceBudget(ctx, &budgetGuard, &actionsTaken); err != nil {
			logger.Error(err, "Failed to enforce budget")
			budgetGuard.Status.ErrorMessage = err.Error()
		} else {
			budgetGuard.Status.ActionsTaken = actionsTaken
		}
	} else {
		budgetGuard.Status.ActionsTaken = []string{}
	}

	// Update conditions
	condition := metav1.Condition{
		Type:               "BudgetStatus",
		Status:             metav1.ConditionTrue,
		Reason:             "WithinBudget",
		Message:            fmt.Sprintf("Current spend: %.2f %s (%.1f%% of budget)", currentSpend, budgetGuard.Spec.Budget.Currency, budgetGuard.Status.PercentageUsed),
		LastTransitionTime: now,
	}
	if exceeded {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "BudgetExceeded"
		condition.Message = fmt.Sprintf("Budget exceeded! Current spend: %.2f %s (%.1f%% of budget)", currentSpend, budgetGuard.Spec.Budget.Currency, budgetGuard.Status.PercentageUsed)
	}
	budgetGuard.Status.Conditions = []metav1.Condition{condition}

	// Update status
	if err := r.Status().Update(ctx, &budgetGuard); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue after refresh interval
	refreshInterval := time.Duration(budgetGuard.Spec.RefreshIntervalSeconds) * time.Second
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Minute
	}
	return ctrl.Result{RequeueAfter: refreshInterval}, nil
}

// fetchCostData fetches cost data from OpenCost/Kubecost API
func (r *BudgetGuardReconciler) fetchCostData(ctx context.Context, budgetGuard *aiopsv1alpha1.BudgetGuard) (float64, error) {
	endpoint := budgetGuard.Spec.OpenCostEndpoint
	if endpoint == "" {
		endpoint = "http://opencost.opencost.svc.cluster.local:9003"
	}

	// Build query based on scope
	var url string
	switch budgetGuard.Spec.Scope {
	case "namespace":
		if budgetGuard.Spec.Namespace == "" {
			return 0, fmt.Errorf("namespace is required for namespace-scoped budget")
		}
		url = fmt.Sprintf("%s/allocation?window=7d&aggregate=namespace&namespace=%s", endpoint, budgetGuard.Spec.Namespace)
	case "cluster":
		url = fmt.Sprintf("%s/allocation?window=7d&aggregate=cluster", endpoint)
	default:
		return 0, fmt.Errorf("unsupported scope: %s", budgetGuard.Spec.Scope)
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
		// In production, this should be an error
		return 0, fmt.Errorf("failed to fetch cost data (OpenCost may not be deployed): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("OpenCost API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (simplified - OpenCost returns complex JSON)
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Extract total cost (simplified parsing)
	// In production, properly parse OpenCost's allocation response
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

// enforceBudget enforces budget limits by taking configured actions
func (r *BudgetGuardReconciler) enforceBudget(ctx context.Context, budgetGuard *aiopsv1alpha1.BudgetGuard, actionsTaken *[]string) error {
	logger := log.FromContext(ctx)
	actions := budgetGuard.Spec.ActionsOnExceed

	// Throttle scaling
	if actions.ThrottleScaling {
		// In production, this would patch HPA/PredictiveScale to set min/max replicas
		// For now, just log
		logger.Info("Throttling scaling due to budget exceed")
		*actionsTaken = append(*actionsTaken, "throttle-scaling")
	}

	// Evict low priority workloads
	if actions.EvictLowPriorityWorkloads {
		if err := r.evictLowPriorityPods(ctx, budgetGuard); err != nil {
			return err
		}
		*actionsTaken = append(*actionsTaken, "evict-low-priority-workloads")
	}

	// Send notifications
	if actions.Notify.Enabled {
		if err := r.sendNotification(ctx, budgetGuard); err != nil {
			logger.Error(err, "Failed to send notification")
		} else {
			*actionsTaken = append(*actionsTaken, "notify")
		}
	}

	return nil
}

// evictLowPriorityPods evicts pods with low priority classes
func (r *BudgetGuardReconciler) evictLowPriorityPods(ctx context.Context, budgetGuard *aiopsv1alpha1.BudgetGuard) error {
	logger := log.FromContext(ctx)

	// Get pods in scope
	var pods corev1.PodList
	opts := []client.ListOption{}
	if budgetGuard.Spec.Scope == "namespace" && budgetGuard.Spec.Namespace != "" {
		opts = append(opts, client.InNamespace(budgetGuard.Spec.Namespace))
	}

	if err := r.List(ctx, &pods, opts...); err != nil {
		return err
	}

	// Evict pods with low priority (priorityClassName < 1000 or no priority class)
	evictedCount := 0
	for _, pod := range pods.Items {
		// Check priority (simplified - in production, check PriorityClass resource)
		priority := int32(0)
		if pod.Spec.PriorityClassName != "" {
			// In production, fetch PriorityClass to get actual priority value
			// For now, assume pods without explicit priority are low priority
		}

		if priority < 1000 || pod.Spec.PriorityClassName == "" {
			logger.Info("Evicting low priority pod due to budget exceed", "pod", pod.Name, "namespace", pod.Namespace)
			if err := r.Delete(ctx, &pod); err != nil {
				logger.Error(err, "Failed to evict pod", "pod", pod.Name)
			} else {
				evictedCount++
			}
		}
	}

	logger.Info("Evicted pods due to budget exceed", "count", evictedCount)
	return nil
}

// sendNotification sends budget exceed notifications
func (r *BudgetGuardReconciler) sendNotification(ctx context.Context, budgetGuard *aiopsv1alpha1.BudgetGuard) error {
	notify := budgetGuard.Spec.ActionsOnExceed.Notify

	// Send webhook notification
	if notify.WebhookURL != "" {
		// In production, send HTTP POST to webhook URL
		// For now, just create a Kubernetes event
		r.recordEvent(ctx, budgetGuard, "Warning", "BudgetExceeded",
			fmt.Sprintf("Budget exceeded! Current spend: %.2f %s (%.1f%% of budget)",
				budgetGuard.Status.CurrentSpend, budgetGuard.Spec.Budget.Currency, budgetGuard.Status.PercentageUsed))
	}

	return nil
}

// recordEvent records a Kubernetes event
func (r *BudgetGuardReconciler) recordEvent(ctx context.Context, budgetGuard *aiopsv1alpha1.BudgetGuard, eventType, reason, message string) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", budgetGuard.Name),
			Namespace:    budgetGuard.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: budgetGuard.APIVersion,
			Kind:       budgetGuard.Kind,
			Name:       budgetGuard.Name,
			Namespace:  budgetGuard.Namespace,
			UID:        budgetGuard.UID,
		},
		Type:    eventType,
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "budget-guard-controller",
		},
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
		Count:          1,
	}

	_ = r.Create(ctx, event)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BudgetGuardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.BudgetGuard{}).
		Complete(r)
}
