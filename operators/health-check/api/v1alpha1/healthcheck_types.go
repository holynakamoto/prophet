package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HealthCheckSpec defines the desired state of HealthCheck
type HealthCheckSpec struct {
	// TargetRef references the workload to check (Deployment, StatefulSet, Pod, etc.)
	TargetRef TargetRef `json:"targetRef"`

	// Probes defines the health check probes to execute
	Probes []ProbeSpec `json:"probes"`

	// FailureThreshold is the number of consecutive failures before marking unhealthy
	// Default: 3
	// +kubebuilder:default=3
	FailureThreshold int32 `json:"failureThreshold,omitempty"`

	// PeriodSeconds is the interval between health checks in seconds
	// Default: 10
	// +kubebuilder:default=10
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`

	// InitialDelaySeconds is the delay before starting health checks
	// Default: 0
	// +kubebuilder:default=0
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`

	// TimeoutSeconds is the timeout for each probe execution
	// Default: 5
	// +kubebuilder:default=5
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`

	// Remediation defines what action to take when health check fails
	Remediation RemediationSpec `json:"remediation,omitempty"`
}

// TargetRef references a Kubernetes workload
type TargetRef struct {
	// APIVersion of the target resource (e.g., "apps/v1")
	APIVersion string `json:"apiVersion"`

	// Kind of the target resource (e.g., "Deployment", "StatefulSet", "Pod")
	Kind string `json:"kind"`

	// Name of the target resource
	Name string `json:"name"`

	// Namespace of the target resource (optional, defaults to HealthCheck namespace)
	Namespace string `json:"namespace,omitempty"`
}

// ProbeSpec defines a single health check probe
type ProbeSpec struct {
	// Name is a unique identifier for this probe
	Name string `json:"name"`

	// Type of probe: "http", "tcp", "command", or "custom"
	// +kubebuilder:validation:Enum=http;tcp;command;custom
	Type string `json:"type"`

	// HTTPGet defines an HTTP health check (used when type is "http")
	HTTPGet *corev1.HTTPGetAction `json:"httpGet,omitempty"`

	// TCPSocket defines a TCP health check (used when type is "tcp")
	TCPSocket *corev1.TCPSocketAction `json:"tcpSocket,omitempty"`

	// Exec defines a command-based health check (used when type is "command")
	Exec *corev1.ExecAction `json:"exec,omitempty"`

	// Custom defines a custom health check (e.g., database connectivity)
	// Used when type is "custom"
	Custom *CustomProbe `json:"custom,omitempty"`
}

// CustomProbe defines a custom health check (e.g., database connectivity, external API)
type CustomProbe struct {
	// Script is a shell script or command to execute for the custom check
	Script string `json:"script,omitempty"`

	// Image is the container image to use for executing the custom probe
	// If not specified, uses the target workload's container image
	Image string `json:"image,omitempty"`

	// Env defines environment variables for the custom probe
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Description of what this custom probe checks
	Description string `json:"description,omitempty"`
}

// RemediationSpec defines remediation actions when health check fails
type RemediationSpec struct {
	// Action to take: "restart", "trigger-recovery-plan", "alert", or "none"
	// +kubebuilder:validation:Enum=restart;trigger-recovery-plan;alert;none
	Action string `json:"action"`

	// RecoveryPlanRef references an AnomalyAction to trigger for recovery
	// Used when action is "trigger-recovery-plan"
	RecoveryPlanRef *RecoveryPlanRef `json:"recoveryPlanRef,omitempty"`

	// RequireApproval requires manual approval before executing remediation
	// Default: false
	RequireApproval bool `json:"requireApproval,omitempty"`

	// CooldownSeconds is the minimum time between remediation actions
	// Default: 300 (5 minutes)
	// +kubebuilder:default=300
	CooldownSeconds int32 `json:"cooldownSeconds,omitempty"`
}

// RecoveryPlanRef references an AnomalyAction for recovery
type RecoveryPlanRef struct {
	// Name of the AnomalyAction resource
	Name string `json:"name"`

	// Namespace of the AnomalyAction (optional, defaults to HealthCheck namespace)
	Namespace string `json:"namespace,omitempty"`
}

// HealthCheckStatus defines the observed state of HealthCheck
type HealthCheckStatus struct {
	// Healthy indicates whether the target workload is currently healthy
	Healthy bool `json:"healthy"`

	// LastCheckTime is the timestamp of the last health check
	LastCheckTime *metav1.Time `json:"lastCheckTime,omitempty"`

	// FailureCount is the number of consecutive failures
	FailureCount int32 `json:"failureCount"`

	// LastFailureTime is the timestamp of the last failure
	LastFailureTime *metav1.Time `json:"lastFailureTime,omitempty"`

	// ProbeResults contains the results of each probe
	ProbeResults []ProbeResult `json:"probeResults,omitempty"`

	// LastRemediationTime is the timestamp of the last remediation action
	LastRemediationTime *metav1.Time `json:"lastRemediationTime,omitempty"`

	// RemediationCount is the number of remediation actions performed
	RemediationCount int32 `json:"remediationCount"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ErrorMessage contains any error message from the last check
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// ProbeResult contains the result of a single probe execution
type ProbeResult struct {
	// Name of the probe
	Name string `json:"name"`

	// Success indicates whether the probe succeeded
	Success bool `json:"success"`

	// LastCheckTime is when this probe was last executed
	LastCheckTime *metav1.Time `json:"lastCheckTime,omitempty"`

	// Message contains additional information about the probe result
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Healthy",type="boolean",JSONPath=".status.healthy"
//+kubebuilder:printcolumn:name="Target",type="string",JSONPath=".spec.targetRef.kind + '/' + .spec.targetRef.name"
//+kubebuilder:printcolumn:name="Failure Count",type="integer",JSONPath=".status.failureCount"
//+kubebuilder:printcolumn:name="Last Check",type="date",JSONPath=".status.lastCheckTime"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// HealthCheck is the Schema for the healthchecks API
type HealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HealthCheckSpec   `json:"spec,omitempty"`
	Status HealthCheckStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HealthCheckList contains a list of HealthCheck
type HealthCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealthCheck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HealthCheck{}, &HealthCheckList{})
}
