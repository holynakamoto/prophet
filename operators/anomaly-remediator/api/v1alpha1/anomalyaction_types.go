package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AnomalyActionSpec defines the desired state of AnomalyAction
type AnomalyActionSpec struct {
	// Anomaly detection source (e.g., "grafana-ml", "prometheus", "otel")
	Source string `json:"source"`

	// Metric name or query that triggered the anomaly
	Metric string `json:"metric"`

	// Threshold or condition that defines an anomaly
	Threshold string `json:"threshold,omitempty"`

	// Remediation actions to take
	Remediation RemediationSpec `json:"remediation"`

	// Target resources (namespace, label selectors)
	Target TargetSpec `json:"target"`

	// Integration with K8sGPT for diagnostics
	K8sGPT K8sGPTSpec `json:"k8sgpt,omitempty"`

	// Webhook URL for external notifications (e.g., PagerDuty)
	WebhookURL string `json:"webhookUrl,omitempty"`
}

// RemediationSpec defines remediation actions
type RemediationSpec struct {
	// Action type: "restart", "scale", "rollback", "alert"
	Type string `json:"type"`

	// For scale action: target replica count
	Replicas *int32 `json:"replicas,omitempty"`

	// For restart action: pod selector
	PodSelector map[string]string `json:"podSelector,omitempty"`

	// Require approval before executing (default: false)
	RequireApproval bool `json:"requireApproval,omitempty"`

	// Cooldown period in seconds before allowing another remediation
	CooldownSeconds int32 `json:"cooldownSeconds,omitempty"`
}

// TargetSpec defines target resources
type TargetSpec struct {
	// Namespace to target
	Namespace string `json:"namespace"`

	// Label selectors
	Labels map[string]string `json:"labels,omitempty"`

	// Resource type: "Deployment", "StatefulSet", "Pod"
	ResourceType string `json:"resourceType"`
}

// K8sGPTSpec defines K8sGPT integration
type K8sGPTSpec struct {
	// Enable automatic K8sGPT analysis on anomaly
	Enabled bool `json:"enabled,omitempty"`

	// K8sGPT service endpoint
	Endpoint string `json:"endpoint,omitempty"`
}

// AnomalyActionStatus defines the observed state of AnomalyAction
type AnomalyActionStatus struct {
	// Phase: "Pending", "Detected", "Remediating", "Resolved", "Failed"
	Phase string `json:"phase,omitempty"`

	// Last anomaly detection time
	LastDetected *metav1.Time `json:"lastDetected,omitempty"`

	// Last remediation time
	LastRemediated *metav1.Time `json:"lastRemediated,omitempty"`

	// Number of remediations performed
	RemediationCount int32 `json:"remediationCount,omitempty"`

	// K8sGPT analysis result
	K8sGPTAnalysis string `json:"k8sgptAnalysis,omitempty"`

	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Error message if remediation failed
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Last Detected",type="date",JSONPath=".status.lastDetected"
//+kubebuilder:printcolumn:name="Remediations",type="integer",JSONPath=".status.remediationCount"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AnomalyAction is the Schema for the anomalyactions API
type AnomalyAction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AnomalyActionSpec   `json:"spec,omitempty"`
	Status AnomalyActionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AnomalyActionList contains a list of AnomalyAction
type AnomalyActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AnomalyAction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AnomalyAction{}, &AnomalyActionList{})
}
