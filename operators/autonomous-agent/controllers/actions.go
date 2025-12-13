package controllers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopsv1alpha1 "github.com/prophet-aiops/autonomous-agent/api/v1alpha1"
)

// ActionExecutor executes remediation actions with safety gates
type ActionExecutor struct {
	Client      client.Client
	Logger      logr.Logger
	rateLimiter *RateLimiter
	auditLog    *AuditLogger
}

// RateLimiter prevents runaway action loops
type RateLimiter struct {
	mu           sync.Mutex
	actionCounts map[string]int
	windowStart  time.Time
	windowSize   time.Duration
	maxActions   int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(windowSize time.Duration, maxActions int) *RateLimiter {
	return &RateLimiter{
		actionCounts: make(map[string]int),
		windowStart:  time.Now(),
		windowSize:   windowSize,
		maxActions:   maxActions,
	}
}

// Allow checks if an action is allowed
func (rl *RateLimiter) Allow(actionKey string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.windowStart) > rl.windowSize {
		// Reset window
		rl.actionCounts = make(map[string]int)
		rl.windowStart = now
	}

	count := rl.actionCounts[actionKey]
	if count >= rl.maxActions {
		return false
	}

	rl.actionCounts[actionKey]++
	return true
}

// AuditLogger logs all agent decisions and executions
type AuditLogger struct {
	mu     sync.Mutex
	events []AuditEvent
	maxLen int
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	Timestamp  time.Time
	ActionType string
	Action     string
	Namespace  string
	Resource   string
	User       string
	Approved   bool
	DryRun     bool
	Result     string
	Error      string
	Reasoning  string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(maxLen int) *AuditLogger {
	return &AuditLogger{
		events: make([]AuditEvent, 0, maxLen),
		maxLen: maxLen,
	}
}

// Log logs an audit event
func (al *AuditLogger) Log(event AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	event.Timestamp = time.Now()
	al.events = append(al.events, event)

	// Keep only last maxLen events
	if len(al.events) > al.maxLen {
		al.events = al.events[len(al.events)-al.maxLen:]
	}
}

// GetEvents returns recent audit events
func (al *AuditLogger) GetEvents(limit int) []AuditEvent {
	al.mu.Lock()
	defer al.mu.Unlock()

	if limit > len(al.events) {
		limit = len(al.events)
	}

	start := len(al.events) - limit
	if start < 0 {
		start = 0
	}

	result := make([]AuditEvent, limit)
	copy(result, al.events[start:])
	return result
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(client client.Client, logger logr.Logger) *ActionExecutor {
	return &ActionExecutor{
		Client:      client,
		Logger:      logger,
		rateLimiter: NewRateLimiter(5*time.Minute, 10), // Max 10 actions per 5 minutes
		auditLog:    NewAuditLogger(1000),
	}
}

// ExecuteAction executes a proposed action with safety checks
func (e *ActionExecutor) ExecuteAction(ctx context.Context, action *aiopsv1alpha1.AutonomousAction, proposedAction aiopsv1alpha1.ProposedAction) (aiopsv1alpha1.ExecutionResult, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	result := aiopsv1alpha1.ExecutionResult{}

	// Check rate limit
	actionKey := fmt.Sprintf("%s/%s", action.Namespace, action.Name)
	if !e.rateLimiter.Allow(actionKey) {
		err := fmt.Errorf("rate limit exceeded for action %s", actionKey)
		logger.Error(err, "Action blocked by rate limiter")

		e.auditLog.Log(AuditEvent{
			ActionType: proposedAction.Type,
			Action:     proposedAction.Description,
			Namespace:  action.Namespace,
			Approved:   false,
			Result:     "blocked",
			Error:      err.Error(),
			Reasoning:  action.Status.Reasoning,
		})

		return result, err
	}

	// Check constraints
	if err := e.checkConstraints(ctx, action, proposedAction); err != nil {
		logger.Error(err, "Action violates constraints")

		e.auditLog.Log(AuditEvent{
			ActionType: proposedAction.Type,
			Action:     proposedAction.Description,
			Namespace:  action.Namespace,
			Approved:   false,
			Result:     "blocked",
			Error:      err.Error(),
			Reasoning:  action.Status.Reasoning,
		})

		return result, err
	}

	// Execute based on action type
	var execErr error
	var output string

	dryRun := action.Spec.ApprovalMode == "dry-run"

	switch proposedAction.Type {
	case "scale":
		output, execErr = e.scaleDeployment(ctx, proposedAction, dryRun)
	case "restart":
		output, execErr = e.restartPods(ctx, proposedAction, dryRun)
	case "rollback":
		output, execErr = e.rollbackDeployment(ctx, proposedAction, dryRun)
	case "drain":
		output, execErr = e.drainNode(ctx, proposedAction, dryRun)
	case "cordon":
		output, execErr = e.cordonNode(ctx, proposedAction, dryRun)
	case "network_policy":
		output, execErr = e.applyNetworkPolicy(ctx, proposedAction, dryRun)
	default:
		execErr = fmt.Errorf("unknown action type: %s", proposedAction.Type)
	}

	// Log audit event
	e.auditLog.Log(AuditEvent{
		ActionType: proposedAction.Type,
		Action:     proposedAction.Description,
		Namespace:  action.Namespace,
		Approved:   true,
		DryRun:     dryRun,
		Result:     "success",
		Error:      "",
		Reasoning:  action.Status.Reasoning,
	})

	if execErr != nil {
		result.Success = false
		result.Output = execErr.Error()
		e.auditLog.Log(AuditEvent{
			ActionType: proposedAction.Type,
			Action:     proposedAction.Description,
			Namespace:  action.Namespace,
			Approved:   true,
			DryRun:     dryRun,
			Result:     "failed",
			Error:      execErr.Error(),
			Reasoning:  action.Status.Reasoning,
		})
	} else {
		result.Success = true
		result.Output = output
	}

	now := metav1.Now()
	result.ExecutedAt = &now
	result.DurationSeconds = time.Since(startTime).Seconds()

	// Create Kubernetes Event for audit trail
	e.createK8sEvent(ctx, action, proposedAction, result)

	return result, execErr
}

// checkConstraints validates action against constraints
func (e *ActionExecutor) checkConstraints(ctx context.Context, action *aiopsv1alpha1.AutonomousAction, proposedAction aiopsv1alpha1.ProposedAction) error {
	constraints := action.Spec.Constraints

	// Check allowed actions
	if len(constraints.AllowedActions) > 0 {
		allowed := false
		for _, allowedAction := range constraints.AllowedActions {
			if allowedAction == proposedAction.Type {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("action type %s not in allowed actions", proposedAction.Type)
		}
	}

	// Check forbidden namespaces
	for _, forbiddenNS := range constraints.ForbiddenNamespaces {
		// Extract namespace from proposed action (would need to parse parameters)
		// For now, check action namespace
		if action.Namespace == forbiddenNS {
			return fmt.Errorf("namespace %s is forbidden", forbiddenNS)
		}
	}

	return nil
}

// scaleDeployment scales a deployment
func (e *ActionExecutor) scaleDeployment(ctx context.Context, action aiopsv1alpha1.ProposedAction, dryRun bool) (string, error) {
	// Parse parameters (in production, use structured parameters)
	// For now, assume parameters contain namespace, name, replicas
	logger := e.Logger.WithValues("action", "scale_deployment")

	if dryRun {
		return fmt.Sprintf("DRY-RUN: Would scale deployment as described: %s", action.Description), nil
	}

	// In production, parse action.Parameters JSON to get namespace, name, replicas
	// For now, placeholder implementation
	deployment := &appsv1.Deployment{}
	// Would get from action.Parameters
	key := types.NamespacedName{
		Namespace: "default", // Would parse from parameters
		Name:      "example", // Would parse from parameters
	}

	if err := e.Client.Get(ctx, key, deployment); err != nil {
		return "", fmt.Errorf("failed to get deployment: %w", err)
	}

	replicas := int32(3)    // Would parse from parameters
	oldReplicas := int32(1) // Kubernetes default when .spec.replicas is omitted
	if deployment.Spec.Replicas != nil {
		oldReplicas = *deployment.Spec.Replicas
	}
	desiredReplicas := replicas
	deployment.Spec.Replicas = &desiredReplicas

	if err := e.Client.Update(ctx, deployment); err != nil {
		return "", fmt.Errorf("failed to scale deployment: %w", err)
	}

	logger.Info("Scaled deployment", "name", deployment.Name, "old", oldReplicas, "new", replicas)
	return fmt.Sprintf("Scaled deployment %s/%s from %d to %d replicas", deployment.Namespace, deployment.Name, oldReplicas, replicas), nil
}

// restartPods restarts pods
func (e *ActionExecutor) restartPods(ctx context.Context, action aiopsv1alpha1.ProposedAction, dryRun bool) (string, error) {
	logger := e.Logger.WithValues("action", "restart_pods")

	if dryRun {
		return fmt.Sprintf("DRY-RUN: Would restart pods as described: %s", action.Description), nil
	}

	// FIXME: Parse namespace and label selectors from parameters.
	// Refusing to proceed without selectors to prevent cluster-wide pod deletion.
	//
	// Example of the *minimum* safety bar:
	// pods := &corev1.PodList{}
	// if err := e.Client.List(ctx, pods,
	// 	client.InNamespace("target-ns"),
	// 	client.MatchingLabels{"app": "target"},
	// ); err != nil {
	// 	return "", fmt.Errorf("failed to list pods: %w", err)
	// }
	logger.Info("Refusing to restart pods without selectors", "reason", "selectors required to prevent cluster-wide pod deletion")
	return "", fmt.Errorf("restart not implemented: selectors required to prevent cluster-wide pod deletion")
}

// rollbackDeployment rolls back a deployment
func (e *ActionExecutor) rollbackDeployment(ctx context.Context, action aiopsv1alpha1.ProposedAction, dryRun bool) (string, error) {
	if dryRun {
		return fmt.Sprintf("DRY-RUN: Would rollback deployment as described: %s", action.Description), nil
	}

	// In production, implement rollback logic
	return "Rollback not yet implemented", fmt.Errorf("rollback not implemented")
}

// drainNode drains a node
func (e *ActionExecutor) drainNode(ctx context.Context, action aiopsv1alpha1.ProposedAction, dryRun bool) (string, error) {
	if dryRun {
		return fmt.Sprintf("DRY-RUN: Would drain node as described: %s (HIGH-IMPACT ACTION)", action.Description), nil
	}

	// In production, implement node draining logic
	return "Node drain not yet implemented", fmt.Errorf("node drain not implemented")
}

// cordonNode cordons a node
func (e *ActionExecutor) cordonNode(ctx context.Context, action aiopsv1alpha1.ProposedAction, dryRun bool) (string, error) {
	if dryRun {
		return fmt.Sprintf("DRY-RUN: Would cordon node as described: %s", action.Description), nil
	}

	// In production, implement node cordoning logic
	return "Node cordon not yet implemented", fmt.Errorf("node cordon not implemented")
}

// applyNetworkPolicy applies a network policy
func (e *ActionExecutor) applyNetworkPolicy(ctx context.Context, action aiopsv1alpha1.ProposedAction, dryRun bool) (string, error) {
	if dryRun {
		return fmt.Sprintf("DRY-RUN: Would apply network policy as described: %s", action.Description), nil
	}

	// In production, implement network policy creation
	return "Network policy not yet implemented", fmt.Errorf("network policy not implemented")
}

// createK8sEvent creates a Kubernetes Event for audit trail
func (e *ActionExecutor) createK8sEvent(ctx context.Context, action *aiopsv1alpha1.AutonomousAction, proposedAction aiopsv1alpha1.ProposedAction, result aiopsv1alpha1.ExecutionResult) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%d", action.Name, time.Now().Unix()),
			Namespace: action.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "AutonomousAction",
			Namespace: action.Namespace,
			Name:      action.Name,
		},
		Reason:  "AgentActionExecuted",
		Message: fmt.Sprintf("Agent executed action: %s. Result: %s", proposedAction.Description, result.Output),
		Type:    "Normal",
		Source: corev1.EventSource{
			Component: "autonomous-agent",
		},
		FirstTimestamp: metav1.Now(),
		LastTimestamp:  metav1.Now(),
		Count:          1,
	}

	if !result.Success {
		event.Type = "Warning"
		event.Reason = "AgentActionFailed"
	}

	if err := e.Client.Create(ctx, event); err != nil {
		// Non-fatal: audit event creation should not block the action result, but failures must be observable.
		log.FromContext(ctx).WithValues(
			"event_kind", event.InvolvedObject.Kind,
			"event_namespace", event.Namespace,
			"event_name", event.Name,
			"event_type", event.Type,
			"event_reason", event.Reason,
			"action_namespace", action.Namespace,
			"action_name", action.Name,
			"proposed_action_type", proposedAction.Type,
		).Error(err, "Failed to create Kubernetes audit event (non-fatal)")
	}
}
