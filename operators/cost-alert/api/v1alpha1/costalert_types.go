package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CostAlertSpec defines the desired state of CostAlert
type CostAlertSpec struct {
	// Threshold defines the cost threshold that triggers an alert
	Threshold ThresholdSpec `json:"threshold"`

	// Scope defines the scope of the alert: "workload", "namespace", or "cluster"
	// +kubebuilder:validation:Enum=workload;namespace;cluster
	Scope string `json:"scope"`

	// WorkloadRef references a specific workload (required if scope is "workload")
	WorkloadRef *WorkloadRef `json:"workloadRef,omitempty"`

	// Namespace is the namespace to monitor (required if scope is "namespace")
	Namespace string `json:"namespace,omitempty"`

	// Period is the time period for cost calculation: "hourly", "daily", "weekly", "monthly"
	// +kubebuilder:validation:Enum=hourly;daily;weekly;monthly
	// +kubebuilder:default=daily
	Period string `json:"period,omitempty"`

	// Notify defines notification settings
	Notify NotifySpec `json:"notify,omitempty"`

	// AlertRuleRef references a PrometheusRule for alerting
	AlertRuleRef *AlertRuleRef `json:"alertRuleRef,omitempty"`

	// OpenCostEndpoint is the OpenCost/Kubecost API endpoint
	// Default: http://opencost.opencost.svc.cluster.local:9003
	OpenCostEndpoint string `json:"openCostEndpoint,omitempty"`

	// CheckIntervalSeconds is how often to check costs (in seconds)
	// Default: 3600 (1 hour)
	// +kubebuilder:default=3600
	CheckIntervalSeconds int32 `json:"checkIntervalSeconds,omitempty"`
}

// ThresholdSpec defines the cost threshold
type ThresholdSpec struct {
	// Type is the threshold type: "percentage_increase" or "absolute"
	// +kubebuilder:validation:Enum=percentage_increase;absolute
	Type string `json:"type"`

	// Value is the threshold value
	// For percentage_increase: percentage increase (e.g., 50 means 50% increase)
	// For absolute: absolute cost amount (e.g., 100.50 means $100.50)
	Value float64 `json:"value"`

	// Currency is the currency unit (USD, EUR, etc.)
	// Default: USD
	// +kubebuilder:default=USD
	Currency string `json:"currency,omitempty"`

	// BaselinePeriod is the period to compare against for percentage_increase
	// Default: previous period (e.g., previous day for daily, previous month for monthly)
	BaselinePeriod string `json:"baselinePeriod,omitempty"`
}

// WorkloadRef references a Kubernetes workload
type WorkloadRef struct {
	// APIVersion of the workload (e.g., "apps/v1")
	APIVersion string `json:"apiVersion"`

	// Kind of the workload (e.g., "Deployment", "StatefulSet")
	Kind string `json:"kind"`

	// Name of the workload
	Name string `json:"name"`

	// Namespace of the workload
	Namespace string `json:"namespace"`
}

// AlertRuleRef references a PrometheusRule
type AlertRuleRef struct {
	// Name of the PrometheusRule
	Name string `json:"name"`

	// Namespace of the PrometheusRule
	Namespace string `json:"namespace"`
}

// NotifySpec defines notification settings
type NotifySpec struct {
	// Enabled enables notifications
	Enabled bool `json:"enabled,omitempty"`

	// WebhookURL is the webhook URL for notifications
	WebhookURL string `json:"webhookUrl,omitempty"`

	// EmailRecipients is a list of email addresses to notify
	EmailRecipients []string `json:"emailRecipients,omitempty"`
}

// CostAlertStatus defines the observed state of CostAlert
type CostAlertStatus struct {
	// Triggered indicates if the alert has been triggered
	Triggered bool `json:"triggered"`

	// CurrentCost is the current cost for the period
	CurrentCost float64 `json:"currentCost"`

	// PreviousCost is the previous period's cost (for percentage_increase comparison)
	PreviousCost float64 `json:"previousCost,omitempty"`

	// ThresholdValue is the threshold value that triggered the alert
	ThresholdValue float64 `json:"thresholdValue,omitempty"`

	// LastTriggeredTime is when the alert was last triggered
	LastTriggeredTime *metav1.Time `json:"lastTriggeredTime,omitempty"`

	// TriggerCount is the number of times the alert has been triggered
	TriggerCount int32 `json:"triggerCount"`

	// LastCheckTime is when the cost was last checked
	LastCheckTime *metav1.Time `json:"lastCheckTime,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ErrorMessage contains any error message from the last check
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope"
//+kubebuilder:printcolumn:name="Threshold",type="string",JSONPath=".spec.threshold.type + ': ' + .spec.threshold.value"
//+kubebuilder:printcolumn:name="Current Cost",type="number",JSONPath=".status.currentCost"
//+kubebuilder:printcolumn:name="Triggered",type="boolean",JSONPath=".status.triggered"
//+kubebuilder:printcolumn:name="Last Triggered",type="date",JSONPath=".status.lastTriggeredTime"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CostAlert is the Schema for the costalerts API
type CostAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CostAlertSpec   `json:"spec,omitempty"`
	Status CostAlertStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CostAlertList contains a list of CostAlert
type CostAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CostAlert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CostAlert{}, &CostAlertList{})
}
