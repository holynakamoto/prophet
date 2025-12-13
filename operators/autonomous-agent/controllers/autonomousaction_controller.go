package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/autonomous-agent/api/v1alpha1"
	"github.com/prophet-aiops/autonomous-agent/llm-inference"
	"github.com/prophet-aiops/autonomous-agent/mcp-server"
)

// AutonomousActionReconciler reconciles a AutonomousAction object
type AutonomousActionReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Log            logr.Logger
	LLMClient      llminference.LLMClient
	MCPServer      *mcpserver.MCPServer
	ActionExecutor *ActionExecutor
}

//+kubebuilder:rbac:groups=aiops.prophet.io,resources=autonomousactions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=autonomousactions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiops.prophet.io,resources=autonomousactions/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AutonomousActionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var action aiopsv1alpha1.AutonomousAction
	if err := r.Get(ctx, req.NamespacedName, &action); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling AutonomousAction", "name", req.Name, "phase", action.Status.Phase)

	// Check trigger conditions
	triggered, err := r.checkTrigger(ctx, &action)
	if err != nil {
		logger.Error(err, "Failed to check trigger")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	if triggered {
		now := metav1.Now()
		action.Status.LastTriggered = &now
		action.Status.Phase = "Triggered"

		// Gather context for LLM
		context, err := r.gatherContext(ctx, &action)
		if err != nil {
			logger.Error(err, "Failed to gather context")
			action.Status.Phase = "Failed"
			action.Status.ErrorMessage = err.Error()
			if err := r.Status().Update(ctx, &action); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
		}

		// LLM reasoning phase
		action.Status.Phase = "Reasoning"
		if err := r.Status().Update(ctx, &action); err != nil {
			return ctrl.Result{}, err
		}

		reasoning, proposedAction, err := r.reasonWithLLM(ctx, &action, context)
		if err != nil {
			logger.Error(err, "LLM reasoning failed")
			action.Status.Phase = "Failed"
			action.Status.ErrorMessage = err.Error()
			if err := r.Status().Update(ctx, &action); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
		}

		action.Status.Reasoning = reasoning
		action.Status.ProposedAction = proposedAction

		// Approval phase
		if action.Spec.ApprovalMode == "human-in-loop" {
			action.Status.Phase = "PendingApproval"
			logger.Info("Human approval required", "action", proposedAction)
		} else if action.Spec.ApprovalMode == "autonomous" {
			// Execute action
			action.Status.Phase = "Executing"
			if err := r.Status().Update(ctx, &action); err != nil {
				return ctrl.Result{}, err
			}

			// Use ActionExecutor for safe execution
			if r.ActionExecutor == nil {
				r.ActionExecutor = NewActionExecutor(r.Client, logger)
			}

			result, err := r.ActionExecutor.ExecuteAction(ctx, &action, proposedAction)
			if err != nil {
				logger.Error(err, "Action execution failed")
				action.Status.Phase = "Failed"
				action.Status.ErrorMessage = err.Error()
			} else {
				action.Status.Phase = "Completed"
				action.Status.ExecutionResult = result
				action.Status.ActionCount++
			}
		} else if action.Spec.ApprovalMode == "dry-run" {
			// Execute in dry-run mode
			action.Status.Phase = "Executing"
			if err := r.Status().Update(ctx, &action); err != nil {
				return ctrl.Result{}, err
			}

			if r.ActionExecutor == nil {
				r.ActionExecutor = NewActionExecutor(r.Client, logger)
			}

			result, err := r.ActionExecutor.ExecuteAction(ctx, &action, proposedAction)
			if err != nil {
				logger.Error(err, "Dry-run execution failed")
				action.Status.Phase = "Failed"
				action.Status.ErrorMessage = err.Error()
			} else {
				action.Status.Phase = "Completed"
				action.Status.ExecutionResult = result
				logger.Info("Dry-run completed", "result", result.Output)
			}
		}
	} else {
		if action.Status.Phase == "" || action.Status.Phase == "Completed" {
			action.Status.Phase = "Monitoring"
		}
	}

	if err := r.Status().Update(ctx, &action); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// checkTrigger checks if trigger conditions are met
func (r *AutonomousActionReconciler) checkTrigger(ctx context.Context, action *aiopsv1alpha1.AutonomousAction) (bool, error) {
	trigger := action.Spec.Trigger

	switch trigger.Type {
	case "anomaly":
		// Query anomaly detection (Grafana ML, Prometheus)
		// Placeholder: In production, query actual anomaly scores
		return false, nil
	case "slo-violation":
		// Query SLO error budget
		// Placeholder: In production, query SLO metrics
		return false, nil
	case "forecast":
		// Query forecast
		// Placeholder: In production, query Grafana ML forecasts
		return false, nil
	case "event":
		// Match event pattern
		// Placeholder: In production, watch events
		return false, nil
	default:
		return false, fmt.Errorf("unknown trigger type: %s", trigger.Type)
	}
}

// gatherContext gathers context for LLM reasoning using MCP tools
func (r *AutonomousActionReconciler) gatherContext(ctx context.Context, action *aiopsv1alpha1.AutonomousAction) (map[string]interface{}, error) {
	logger := log.FromContext(ctx)
	context := make(map[string]interface{})

	// Use MCP tools to gather context if available
	if r.MCPServer != nil && r.MCPServer.ToolExecutor() != nil {
		// Gather pods in target namespaces
		if len(action.Spec.Context.Namespaces) > 0 {
			for _, ns := range action.Spec.Context.Namespaces {
				pods, err := r.MCPServer.ToolExecutor().ExecuteTool(ctx, "k8s_get_pods", map[string]interface{}{
					"namespace": ns,
				})
				if err == nil {
					context[fmt.Sprintf("pods_%s", ns)] = pods
				}
			}
		}

		// Get nodes
		nodes, err := r.MCPServer.ToolExecutor().ExecuteTool(ctx, "k8s_get_nodes", map[string]interface{}{})
		if err == nil {
			context["nodes"] = nodes
		}
	}

	if action.Spec.Context.IncludeK8sGPT {
		// Query K8sGPT analysis via MCP tool
		if r.MCPServer != nil && r.MCPServer.ToolExecutor() != nil {
			analysis, err := r.MCPServer.ToolExecutor().ExecuteTool(ctx, "k8s_get_k8sgpt_analysis", map[string]interface{}{})
			if err == nil {
				context["k8sgpt"] = analysis
			} else {
				logger.Info("K8sGPT analysis not available", "error", err)
				context["k8sgpt"] = "K8sGPT analysis unavailable"
			}
		} else {
			context["k8sgpt"] = "K8sGPT analysis placeholder"
		}
	}

	if action.Spec.Context.IncludeMetrics {
		// Query metrics via MCP tool
		if r.MCPServer != nil && r.MCPServer.ToolExecutor() != nil {
			metrics, err := r.MCPServer.ToolExecutor().ExecuteTool(ctx, "k8s_get_metrics", map[string]interface{}{})
			if err == nil {
				context["metrics"] = metrics
			} else {
				context["metrics"] = "Metrics unavailable"
			}
		} else {
			context["metrics"] = "Metrics placeholder"
		}
	}

	if action.Spec.Context.IncludeEvents {
		// Get recent events via MCP tool
		if len(action.Spec.Context.Namespaces) > 0 {
			allEvents := make([]interface{}, 0)
			for _, ns := range action.Spec.Context.Namespaces {
				if r.MCPServer != nil && r.MCPServer.ToolExecutor() != nil {
					events, err := r.MCPServer.ToolExecutor().ExecuteTool(ctx, "k8s_get_events", map[string]interface{}{
						"namespace": ns,
					})
					if err == nil {
						allEvents = append(allEvents, events)
					}
				}
			}
			if len(allEvents) > 0 {
				context["events"] = allEvents
			} else {
				context["events"] = "Events placeholder"
			}
		} else {
			context["events"] = "Events placeholder"
		}
	}

	if action.Spec.Context.IncludeHubble {
		// Query Hubble flows (would need Hubble-specific tool)
		context["hubble"] = "Hubble flows placeholder"
	}

	return context, nil
}

// reasonWithLLM uses LLM to reason about the situation and propose action
func (r *AutonomousActionReconciler) reasonWithLLM(ctx context.Context, action *aiopsv1alpha1.AutonomousAction, context map[string]interface{}) (string, aiopsv1alpha1.ProposedAction, error) {
	systemPrompt := action.Spec.LLM.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are an expert Kubernetes SRE. Analyze the cluster state and propose remediation actions.
Consider: safety, impact, and best practices. Return your reasoning and a proposed action in JSON format.`
	}

	userPrompt := fmt.Sprintf(`Cluster Context:
%v

Analyze the situation and propose a remediation action. Consider the constraints: %v`,
		context, action.Spec.Constraints)

	var llmClient llminference.LLMClient
	if r.LLMClient != nil {
		llmClient = r.LLMClient
	} else {
		// Initialize LLM client based on spec
		if action.Spec.LLM.Provider == "ollama" {
			endpoint := action.Spec.LLM.Endpoint
			if endpoint == "" {
				endpoint = "http://ollama:11434"
			}
			llmClient = llminference.NewOllamaClient(endpoint, action.Spec.LLM.Model)
		} else if action.Spec.LLM.Provider == "openai" {
			// In production, get API key from Secret
			llmClient = llminference.NewOpenAIClient("", action.Spec.LLM.Model)
		}
	}

	response, err := llmClient.GenerateWithContext(ctx, userPrompt, context)
	if err != nil {
		return "", aiopsv1alpha1.ProposedAction{}, err
	}

	// Parse LLM response (in production, use structured output)
	var proposedAction aiopsv1alpha1.ProposedAction
	if err := json.Unmarshal([]byte(response), &proposedAction); err != nil {
		// Fallback: create action from response
		proposedAction = aiopsv1alpha1.ProposedAction{
			Type:        "restart",
			Description: response,
			Confidence:  0.8,
			RiskLevel:   "medium",
		}
	}

	return response, proposedAction, nil
}

// executeAction is deprecated - use ActionExecutor.ExecuteAction instead
// Kept for backward compatibility
func (r *AutonomousActionReconciler) executeAction(ctx context.Context, action *aiopsv1alpha1.AutonomousAction, proposedAction aiopsv1alpha1.ProposedAction) (aiopsv1alpha1.ExecutionResult, error) {
	if r.ActionExecutor == nil {
		r.ActionExecutor = NewActionExecutor(r.Client, r.Log)
	}
	return r.ActionExecutor.ExecuteAction(ctx, action, proposedAction)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutonomousActionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopsv1alpha1.AutonomousAction{}).
		Complete(r)
}
