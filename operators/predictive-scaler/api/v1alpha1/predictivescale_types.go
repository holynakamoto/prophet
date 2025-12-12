package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PredictiveScaleSpec defines the desired state of PredictiveScale
// +kubebuilder:object:generate=true
type PredictiveScaleSpec struct {
	// Grafana ML forecast query
	ForecastQuery string `json:"forecastQuery"`

	// Karpenter NodePool to adjust
	NodePoolRef NodePoolRef `json:"nodePoolRef"`

	// Forecast horizon (e.g., "1h", "15m")
	Horizon string `json:"horizon"`

	// Threshold for triggering scaling (percentage increase)
	ThresholdPercent float64 `json:"thresholdPercent"`

	// Scaling action: "provision", "consolidate", "adjust"
	Action string `json:"action"`

	// Grafana API endpoint
	GrafanaEndpoint string `json:"grafanaEndpoint,omitempty"`
}

// NodePoolRef references a Karpenter NodePool
type NodePoolRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// PredictiveScaleStatus defines the observed state of PredictiveScale
// +kubebuilder:object:generate=true
type PredictiveScaleStatus struct {
	// Phase: "Monitoring", "Scaling", "Complete", "Failed"
	Phase string `json:"phase,omitempty"`

	// Last forecast value
	LastForecast float64 `json:"lastForecast,omitempty"`

	// Last scaling action time
	LastScaled *metav1.Time `json:"lastScaled,omitempty"`

	// Scaling actions performed
	ScalingCount int32 `json:"scalingCount,omitempty"`

	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Last Forecast",type="number",JSONPath=".status.lastForecast"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PredictiveScale is the Schema for the predictivescales API
type PredictiveScale struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PredictiveScaleSpec   `json:"spec,omitempty"`
	Status PredictiveScaleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PredictiveScaleList contains a list of PredictiveScale
type PredictiveScaleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PredictiveScale `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PredictiveScale{}, &PredictiveScaleList{})
}
