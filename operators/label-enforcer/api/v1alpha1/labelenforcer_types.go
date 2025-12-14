package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelEnforcerSpec defines the desired state of LabelEnforcer
type LabelEnforcerSpec struct {
	// Target resources to enforce labels/annotations on
	// +kubebuilder:validation:Enum=pods;deployments;statefulsets;daemonsets;services;configmaps;secrets
	TargetResource string `json:"targetResource"`

	// Namespace to watch (empty means all namespaces)
	Namespace string `json:"namespace,omitempty"`

	// Label selector to match target resources
	LabelSelector map[string]string `json:"labelSelector,omitempty"`

	// Required labels that must be present on resources
	RequiredLabels map[string]string `json:"requiredLabels,omitempty"`

	// Required annotations that must be present on resources
	RequiredAnnotations map[string]string `json:"requiredAnnotations,omitempty"`

	// Whether to enforce on existing resources (default: true)
	EnforceExisting bool `json:"enforceExisting,omitempty"`
}

// LabelEnforcerStatus defines the observed state of LabelEnforcer
type LabelEnforcerStatus struct {
	// Number of resources that were corrected
	CorrectedResources int32 `json:"correctedResources,omitempty"`

	// Last time a correction was made
	LastCorrected *metav1.Time `json:"lastCorrected,omitempty"`

	// Conditions for the enforcer
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Target",type="string",JSONPath=".spec.targetResource"
//+kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.namespace"
//+kubebuilder:printcolumn:name="Corrected",type="integer",JSONPath=".status.correctedResources"

// LabelEnforcer is the Schema for the labelenforcers API
type LabelEnforcer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabelEnforcerSpec   `json:"spec,omitempty"`
	Status LabelEnforcerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LabelEnforcerList contains a list of LabelEnforcer
type LabelEnforcerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabelEnforcer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabelEnforcer{}, &LabelEnforcerList{})
}
