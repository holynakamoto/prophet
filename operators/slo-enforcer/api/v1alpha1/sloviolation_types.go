package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SLOViolationSpec defines the desired state of SLOViolation
// +kubebuilder:object:generate=true
type SLOViolationSpec struct {
	// SLO name
	SLOName string `json:"sloName"`

	// SLO target (e.g., "99.9%")
	SLOTarget string `json:"sloTarget"`

	// Error budget remaining threshold
	ErrorBudgetThreshold float64 `json:"errorBudgetThreshold"`

	// Actions to take on violation
	Actions []SLOAction `json:"actions"`

	// HPA to adjust
	HPARef HPARef `json:"hpaRef,omitempty"`

	// Enable chaos testing
	EnableChaos bool `json:"enableChaos,omitempty"`
}

// SLOAction defines actions to take
type SLOAction struct {
	Type  string `json:"type"` // "scale", "rollback", "alert", "chaos"
	Value string `json:"value,omitempty"`
}

// HPARef references an HPA
type HPARef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// SLOViolationStatus defines the observed state of SLOViolation
// +kubebuilder:object:generate=true
type SLOViolationStatus struct {
	// Phase: "Monitoring", "Violated", "Remediating", "Resolved"
	Phase string `json:"phase,omitempty"`

	// Current error budget remaining
	ErrorBudgetRemaining float64 `json:"errorBudgetRemaining,omitempty"`

	// Time to exhaustion (days)
	TimeToExhaustion float64 `json:"timeToExhaustion,omitempty"`

	// Last violation time
	LastViolated *metav1.Time `json:"lastViolated,omitempty"`

	// Violation count
	ViolationCount int32 `json:"violationCount,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SLOViolation is the Schema for the sloviolations API
type SLOViolation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SLOViolationSpec   `json:"spec,omitempty"`
	Status SLOViolationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SLOViolationList contains a list of SLOViolation
type SLOViolationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SLOViolation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SLOViolation{}, &SLOViolationList{})
}
