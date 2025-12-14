package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosticRemediationSpec defines the desired state of DiagnosticRemediation
type DiagnosticRemediationSpec struct {
	// Target workload to diagnose and remediate
	Target TargetSpec `json:"target"`

	// Diagnostic checks to perform
	Diagnostics DiagnosticChecks `json:"diagnostics"`

	// Remediation actions to take when issues are found
	Remediation RemediationActions `json:"remediation"`

	// Auto-fix enabled (default: true)
	AutoFix bool `json:"autoFix,omitempty"`

	// Cooldown period in seconds before allowing another remediation
	CooldownSeconds int32 `json:"cooldownSeconds,omitempty"`
}

// TargetSpec defines the target workload
type TargetSpec struct {
	// Namespace
	Namespace string `json:"namespace"`

	// Resource type: Deployment, StatefulSet, DaemonSet
	Kind string `json:"kind"`

	// Resource name
	Name string `json:"name"`

	// Label selector (alternative to name)
	Labels map[string]string `json:"labels,omitempty"`
}

// DiagnosticChecks defines what to check
type DiagnosticChecks struct {
	// Check resource limits/requests
	Resources bool `json:"resources,omitempty"`

	// Check environment variables
	Environment bool `json:"environment,omitempty"`

	// Check ConfigMaps/Secrets references
	ConfigReferences bool `json:"configReferences,omitempty"`

	// Check service dependencies
	ServiceDependencies []ServiceDependency `json:"serviceDependencies,omitempty"`

	// Check image pull policy and availability
	ImagePull bool `json:"imagePull,omitempty"`

	// Check pod disruption budget
	PodDisruptionBudget bool `json:"podDisruptionBudget,omitempty"`

	// Check persistent volume claims
	PersistentVolumes bool `json:"persistentVolumes,omitempty"`

	// Check network policies
	NetworkPolicies bool `json:"networkPolicies,omitempty"`

	// Custom diagnostic script
	CustomScript string `json:"customScript,omitempty"`
}

// ServiceDependency defines a service that must be available
type ServiceDependency struct {
	// Service name
	Name string `json:"name"`

	// Service namespace (defaults to target namespace)
	Namespace string `json:"namespace,omitempty"`

	// Port to check
	Port int32 `json:"port"`

	// Protocol: TCP, HTTP, HTTPS
	Protocol string `json:"protocol,omitempty"`

	// HTTP path to check (for HTTP/HTTPS)
	Path string `json:"path,omitempty"`
}

// RemediationActions defines what fixes to apply
type RemediationActions struct {
	// Fix resource limits (add defaults if missing)
	FixResources bool `json:"fixResources,omitempty"`

	// Fix environment variables (add required env vars)
	FixEnvironment bool `json:"fixEnvironment,omitempty"`

	// Fix image pull policy
	FixImagePullPolicy bool `json:"fixImagePullPolicy,omitempty"`

	// Scale up if resources insufficient
	ScaleUp bool `json:"scaleUp,omitempty"`

	// Restart pods if configuration changed
	RestartOnConfigChange bool `json:"restartOnConfigChange,omitempty"`

	// Create missing ConfigMaps/Secrets
	CreateMissingConfigs bool `json:"createMissingConfigs,omitempty"`

	// Default resource limits to apply
	DefaultResources ResourceSpec `json:"defaultResources,omitempty"`

	// Required environment variables
	RequiredEnvVars []EnvVarSpec `json:"requiredEnvVars,omitempty"`

	// Default image pull policy
	DefaultImagePullPolicy string `json:"defaultImagePullPolicy,omitempty"`
}

// ResourceSpec defines resource limits and requests
type ResourceSpec struct {
	// CPU request
	CPURequest string `json:"cpuRequest,omitempty"`

	// CPU limit
	CPULimit string `json:"cpuLimit,omitempty"`

	// Memory request
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// Memory limit
	MemoryLimit string `json:"memoryLimit,omitempty"`
}

// EnvVarSpec defines an environment variable
type EnvVarSpec struct {
	// Variable name
	Name string `json:"name"`

	// Variable value (or valueFrom)
	Value string `json:"value,omitempty"`

	// Value from ConfigMap/Secret
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// EnvVarSource defines where to get the value
type EnvVarSource struct {
	// ConfigMap key reference
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`

	// Secret key reference
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// ConfigMapKeySelector selects a key from a ConfigMap
type ConfigMapKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// SecretKeySelector selects a key from a Secret
type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// DiagnosticRemediationStatus defines the observed state of DiagnosticRemediation
type DiagnosticRemediationStatus struct {
	// Phase: Pending, Diagnosing, IssuesFound, Remediating, Resolved, Failed
	Phase string `json:"phase,omitempty"`

	// Last diagnostic time
	LastDiagnosed *metav1.Time `json:"lastDiagnosed,omitempty"`

	// Last remediation time
	LastRemediated *metav1.Time `json:"lastRemediated,omitempty"`

	// Issues found
	Issues []DiagnosticIssue `json:"issues,omitempty"`

	// Remediations applied
	Remediations []RemediationAction `json:"remediations,omitempty"`

	// Remediation count
	RemediationCount int32 `json:"remediationCount,omitempty"`

	// Error message if failed
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// DiagnosticIssue represents a found issue
type DiagnosticIssue struct {
	// Issue type: MissingResources, MissingEnvVar, MissingConfig, ServiceUnavailable, etc.
	Type string `json:"type"`

	// Severity: Critical, Warning, Info
	Severity string `json:"severity"`

	// Description
	Description string `json:"description"`

	// Affected resource
	Resource string `json:"resource,omitempty"`

	// Suggested fix
	SuggestedFix string `json:"suggestedFix,omitempty"`
}

// RemediationAction represents an applied fix
type RemediationAction struct {
	// Action type: AddedResources, AddedEnvVar, UpdatedConfig, ScaledUp, etc.
	Type string `json:"type"`

	// Description
	Description string `json:"description"`

	// Timestamp
	Timestamp metav1.Time `json:"timestamp"`

	// Success
	Success bool `json:"success"`

	// Error message if failed
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Issues",type="integer",JSONPath=".status.issues[*]"
//+kubebuilder:printcolumn:name="Remediations",type="integer",JSONPath=".status.remediationCount"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// DiagnosticRemediation is the Schema for the diagnosticremediations API
type DiagnosticRemediation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiagnosticRemediationSpec   `json:"spec,omitempty"`
	Status DiagnosticRemediationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiagnosticRemediationList contains a list of DiagnosticRemediation
type DiagnosticRemediationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiagnosticRemediation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiagnosticRemediation{}, &DiagnosticRemediationList{})
}
