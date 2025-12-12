package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutonomousActionSpec defines the desired state of AutonomousAction
// +kubebuilder:object:generate=true
type AutonomousActionSpec struct {
	// Trigger conditions for autonomous action
	Trigger TriggerSpec `json:"trigger"`

	// LLM configuration
	LLM LLMSpec `json:"llm"`

	// Context sources for LLM decision-making
	Context ContextSpec `json:"context"`

	// Approval mode: "autonomous", "human-in-loop", "dry-run"
	ApprovalMode string `json:"approvalMode"`

	// Action constraints
	Constraints ConstraintsSpec `json:"constraints,omitempty"`

	// MCP server endpoint for external agent integration
	MCPServer string `json:"mcpServer,omitempty"`
}

// TriggerSpec defines when to trigger autonomous action
// +kubebuilder:object:generate=true
type TriggerSpec struct {
	// Trigger type: "anomaly", "slo-violation", "forecast", "event"
	Type string `json:"type"`

	// Anomaly score threshold (0-1)
	AnomalyScoreThreshold float64 `json:"anomalyScoreThreshold,omitempty"`

	// SLO error budget threshold
	ErrorBudgetThreshold float64 `json:"errorBudgetThreshold,omitempty"`

	// Forecast threshold (percentage change)
	ForecastThreshold float64 `json:"forecastThreshold,omitempty"`

	// Event pattern to match
	EventPattern string `json:"eventPattern,omitempty"`
}

// LLMSpec defines LLM configuration
// +kubebuilder:object:generate=true
type LLMSpec struct {
	// Provider: "ollama", "vllm", "openai", "anthropic"
	Provider string `json:"provider"`

	// Model name (e.g., "phi-3", "llama-3.2", "gpt-4")
	Model string `json:"model"`

	// Endpoint for local inference (Ollama/vLLM)
	Endpoint string `json:"endpoint,omitempty"`

	// API key for cloud providers (stored in Secret)
	APIKeySecret string `json:"apiKeySecret,omitempty"`

	// Temperature for generation (0-1)
	Temperature float64 `json:"temperature,omitempty"`

	// Max tokens
	MaxTokens int `json:"maxTokens,omitempty"`

	// System prompt template
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

// ContextSpec defines context sources for LLM
// +kubebuilder:object:generate=true
type ContextSpec struct {
	// Include K8sGPT analysis
	IncludeK8sGPT bool `json:"includeK8sGPT,omitempty"`

	// Include Prometheus metrics
	IncludeMetrics bool `json:"includeMetrics,omitempty"`

	// Include Hubble flows
	IncludeHubble bool `json:"includeHubble,omitempty"`

	// Include recent events
	IncludeEvents bool `json:"includeEvents,omitempty"`

	// Time window for context (e.g., "5m", "1h")
	TimeWindow string `json:"timeWindow,omitempty"`

	// Namespace scope
	Namespaces []string `json:"namespaces,omitempty"`
}

// ConstraintsSpec defines action constraints
// +kubebuilder:object:generate=true
type ConstraintsSpec struct {
	// Allowed action types: "scale", "restart", "rollback", "drain", "custom"
	AllowedActions []string `json:"allowedActions,omitempty"`

	// Forbidden namespaces
	ForbiddenNamespaces []string `json:"forbiddenNamespaces,omitempty"`

	// Max concurrent actions
	MaxConcurrent int `json:"maxConcurrent,omitempty"`

	// Cooldown period between actions
	CooldownSeconds int32 `json:"cooldownSeconds,omitempty"`
}

// AutonomousActionStatus defines the observed state of AutonomousAction
// +kubebuilder:object:generate=true
type AutonomousActionStatus struct {
	// Phase: "Monitoring", "Triggered", "Reasoning", "Approved", "Executing", "Completed", "Failed", "Rejected"
	Phase string `json:"phase,omitempty"`

	// Last trigger time
	LastTriggered *metav1.Time `json:"lastTriggered,omitempty"`

	// LLM reasoning output
	Reasoning string `json:"reasoning,omitempty"`

	// Proposed action
	ProposedAction ProposedAction `json:"proposedAction,omitempty"`

	// Execution result
	ExecutionResult ExecutionResult `json:"executionResult,omitempty"`

	// Action count
	ActionCount int32 `json:"actionCount,omitempty"`

	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Error message
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// ProposedAction represents LLM-proposed action
// +kubebuilder:object:generate=true
type ProposedAction struct {
	// Action type
	Type string `json:"type"`

	// Action description
	Description string `json:"description"`

	// Action parameters (JSON)
	Parameters string `json:"parameters,omitempty"`

	// Confidence score (0-1)
	Confidence float64 `json:"confidence,omitempty"`

	// Risk level: "low", "medium", "high"
	RiskLevel string `json:"riskLevel,omitempty"`
}

// ExecutionResult represents action execution result
// +kubebuilder:object:generate=true
type ExecutionResult struct {
	// Success status
	Success bool `json:"success,omitempty"`

	// Execution time
	ExecutedAt *metav1.Time `json:"executedAt,omitempty"`

	// Output/result message
	Output string `json:"output,omitempty"`

	// Duration in seconds
	DurationSeconds float64 `json:"durationSeconds,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Actions",type="integer",JSONPath=".status.actionCount"
//+kubebuilder:printcolumn:name="Last Triggered",type="date",JSONPath=".status.lastTriggered"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AutonomousAction is the Schema for the autonomousactions API
type AutonomousAction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutonomousActionSpec   `json:"spec,omitempty"`
	Status AutonomousActionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutonomousActionList contains a list of AutonomousAction
type AutonomousActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutonomousAction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutonomousAction{}, &AutonomousActionList{})
}
