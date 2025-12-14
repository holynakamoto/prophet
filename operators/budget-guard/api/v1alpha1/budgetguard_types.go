package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BudgetGuardSpec defines the desired state of BudgetGuard
type BudgetGuardSpec struct {
	// Budget is the cost limit (in USD or resource units)
	Budget BudgetLimit `json:"budget"`

	// Scope defines the scope of the budget: "namespace" or "cluster"
	// +kubebuilder:validation:Enum=namespace;cluster
	Scope string `json:"scope"`

	// Namespace is the namespace to apply the budget to (required if scope is "namespace")
	Namespace string `json:"namespace,omitempty"`

	// Period is the time period for the budget: "daily", "weekly", "monthly", "yearly"
	// +kubebuilder:validation:Enum=daily;weekly;monthly;yearly
	// +kubebuilder:default=monthly
	Period string `json:"period,omitempty"`

	// ActionsOnExceed defines what actions to take when budget is exceeded
	ActionsOnExceed ActionsOnExceedSpec `json:"actionsOnExceed"`

	// OpenCostEndpoint is the OpenCost/Kubecost API endpoint
	// Default: http://opencost.opencost.svc.cluster.local:9003
	OpenCostEndpoint string `json:"openCostEndpoint,omitempty"`

	// RefreshIntervalSeconds is how often to check budget status (in seconds)
	// Default: 300 (5 minutes)
	// +kubebuilder:default=300
	RefreshIntervalSeconds int32 `json:"refreshIntervalSeconds,omitempty"`
}

// BudgetLimit defines the budget limit
type BudgetLimit struct {
	// Amount is the budget amount
	Amount float64 `json:"amount"`

	// Currency is the currency unit (USD, EUR, etc.) or resource unit (CPU-hours, Memory-GB-hours)
	// Default: USD
	// +kubebuilder:default=USD
	Currency string `json:"currency,omitempty"`
}

// ActionsOnExceedSpec defines actions to take when budget is exceeded
type ActionsOnExceedSpec struct {
	// ThrottleScaling prevents new scaling operations when budget is exceeded
	ThrottleScaling bool `json:"throttleScaling,omitempty"`

	// EvictLowPriorityWorkloads evicts pods with low priority when budget is exceeded
	EvictLowPriorityWorkloads bool `json:"evictLowPriorityWorkloads,omitempty"`

	// Notify sends notifications when budget is exceeded
	Notify NotifySpec `json:"notify,omitempty"`

	// BlockNewResources prevents creation of new resources when budget is exceeded
	BlockNewResources bool `json:"blockNewResources,omitempty"`
}

// NotifySpec defines notification settings
type NotifySpec struct {
	// Enabled enables notifications
	Enabled bool `json:"enabled,omitempty"`

	// WebhookURL is the webhook URL for notifications (e.g., Slack, PagerDuty)
	WebhookURL string `json:"webhookUrl,omitempty"`

	// EmailRecipients is a list of email addresses to notify
	EmailRecipients []string `json:"emailRecipients,omitempty"`
}

// BudgetGuardStatus defines the observed state of BudgetGuard
type BudgetGuardStatus struct {
	// CurrentSpend is the current spend for the period
	CurrentSpend float64 `json:"currentSpend"`

	// BudgetLimit is the budget limit
	BudgetLimit float64 `json:"budgetLimit"`

	// PercentageUsed is the percentage of budget used (0-100)
	PercentageUsed float64 `json:"percentageUsed"`

	// Exceeded indicates if the budget has been exceeded
	Exceeded bool `json:"exceeded"`

	// ProjectedExceedTime is the projected time when budget will be exceeded (if current trend continues)
	ProjectedExceedTime *metav1.Time `json:"projectedExceedTime,omitempty"`

	// LastRefreshTime is when the budget was last refreshed
	LastRefreshTime *metav1.Time `json:"lastRefreshTime,omitempty"`

	// ActionsTaken is a list of actions that have been taken due to budget exceed
	ActionsTaken []string `json:"actionsTaken,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ErrorMessage contains any error message from the last refresh
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope"
//+kubebuilder:printcolumn:name="Budget",type="string",JSONPath=".spec.budget.amount + ' ' + .spec.budget.currency"
//+kubebuilder:printcolumn:name="Spend",type="string",JSONPath=".status.currentSpend"
//+kubebuilder:printcolumn:name="% Used",type="number",JSONPath=".status.percentageUsed"
//+kubebuilder:printcolumn:name="Exceeded",type="boolean",JSONPath=".status.exceeded"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// BudgetGuard is the Schema for the budgetguards API
// +kubebuilder:resource:scope=Cluster
type BudgetGuard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BudgetGuardSpec   `json:"spec,omitempty"`
	Status BudgetGuardStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BudgetGuardList contains a list of BudgetGuard
type BudgetGuardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BudgetGuard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BudgetGuard{}, &BudgetGuardList{})
}
