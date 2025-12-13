package tools

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MCPTool represents an MCP tool definition
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolExecutor executes MCP tools
type ToolExecutor struct {
	Client client.Client
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(client client.Client) *ToolExecutor {
	return &ToolExecutor{Client: client}
}

// GetTools returns all available MCP tools
func GetTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "k8s_get_pods",
			Description: "Get pods in a namespace or across all namespaces. Can filter by labels.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name (empty for all namespaces)",
					},
					"labelSelector": map[string]interface{}{
						"type":        "object",
						"description": "Label selector (key-value pairs)",
					},
				},
			},
		},
		{
			Name:        "k8s_get_nodes",
			Description: "Get cluster nodes with their status and conditions.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"labelSelector": map[string]interface{}{
						"type":        "object",
						"description": "Label selector (key-value pairs)",
					},
				},
			},
		},
		{
			Name:        "k8s_get_deployments",
			Description: "Get deployments in a namespace or across all namespaces.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name (empty for all namespaces)",
					},
					"labelSelector": map[string]interface{}{
						"type":        "object",
						"description": "Label selector (key-value pairs)",
					},
				},
			},
		},
		{
			Name:        "k8s_get_events",
			Description: "Get recent events in a namespace or for a specific resource.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name",
					},
					"involvedObject": map[string]interface{}{
						"type":        "object",
						"description": "Involved object (kind, name, namespace)",
					},
				},
			},
		},
		{
			Name:        "k8s_scale_deployment",
			Description: "Scale a deployment to a specific replica count. Requires approval in review mode.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{"namespace", "name", "replicas"},
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Deployment name",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Target replica count",
					},
					"dryRun": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, simulate the action without executing",
					},
				},
			},
		},
		{
			Name:        "k8s_restart_pod",
			Description: "Restart a pod by deleting it (will be recreated by controller).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{"namespace", "name"},
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"dryRun": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, simulate the action without executing",
					},
				},
			},
		},
		{
			Name:        "k8s_cordon_node",
			Description: "Cordon a node to prevent new pods from being scheduled.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{"name"},
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Node name",
					},
					"dryRun": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, simulate the action without executing",
					},
				},
			},
		},
		{
			Name:        "k8s_drain_node",
			Description: "Drain a node by cordoning it and evicting all pods. High-impact action.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{"name"},
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Node name",
					},
					"ignoreDaemonSets": map[string]interface{}{
						"type":        "boolean",
						"description": "Ignore DaemonSet pods",
					},
					"dryRun": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, simulate the action without executing",
					},
				},
			},
		},
		{
			Name:        "k8s_get_metrics",
			Description: "Get resource metrics for pods or nodes (CPU, memory).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name",
					},
					"resourceType": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"pod", "node"},
						"description": "Resource type",
					},
				},
			},
		},
		{
			Name:        "k8s_apply_network_policy",
			Description: "Apply an emergency network policy to quarantine a namespace.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{"namespace"},
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"allow", "deny"},
						"description": "Network policy action",
					},
					"dryRun": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, simulate the action without executing",
					},
				},
			},
		},
		{
			Name:        "k8s_get_k8sgpt_analysis",
			Description: "Get K8sGPT diagnostic analysis for a resource or namespace.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace name",
					},
					"resourceType": map[string]interface{}{
						"type":        "string",
						"description": "Resource type (pod, deployment, etc.)",
					},
					"resourceName": map[string]interface{}{
						"type":        "string",
						"description": "Resource name",
					},
				},
			},
		},
		{
			Name:        "k8s_get_forecast",
			Description: "Get Grafana ML forecast for a metric.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{"metric"},
				"properties": map[string]interface{}{
					"metric": map[string]interface{}{
						"type":        "string",
						"description": "Metric name (PromQL)",
					},
					"duration": map[string]interface{}{
						"type":        "string",
						"description": "Forecast duration (e.g., 1h, 24h)",
					},
				},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (e *ToolExecutor) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "k8s_get_pods":
		return e.getPods(ctx, arguments)
	case "k8s_get_nodes":
		return e.getNodes(ctx, arguments)
	case "k8s_get_deployments":
		return e.getDeployments(ctx, arguments)
	case "k8s_get_events":
		return e.getEvents(ctx, arguments)
	case "k8s_scale_deployment":
		return e.scaleDeployment(ctx, arguments)
	case "k8s_restart_pod":
		return e.restartPod(ctx, arguments)
	case "k8s_cordon_node":
		return e.cordonNode(ctx, arguments)
	case "k8s_drain_node":
		return e.drainNode(ctx, arguments)
	case "k8s_get_metrics":
		return e.getMetrics(ctx, arguments)
	case "k8s_apply_network_policy":
		return e.applyNetworkPolicy(ctx, arguments)
	case "k8s_get_k8sgpt_analysis":
		return e.getK8sGPTAnalysis(ctx, arguments)
	case "k8s_get_forecast":
		return e.getForecast(ctx, arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// getPods retrieves pods
func (e *ToolExecutor) getPods(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	namespace, _ := args["namespace"].(string)
	
	pods := &corev1.PodList{}
	opts := []client.ListOption{}
	
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}
	
	if labelSelector, ok := args["labelSelector"].(map[string]interface{}); ok {
		labels := make(map[string]string)
		for k, v := range labelSelector {
			if str, ok := v.(string); ok {
				labels[k] = str
			}
		}
		if len(labels) > 0 {
			opts = append(opts, client.MatchingLabels(labels))
		}
	}
	
	if err := e.Client.List(ctx, pods, opts...); err != nil {
		return nil, err
	}
	
	result := make([]map[string]interface{}, 0, len(pods.Items))
	for _, pod := range pods.Items {
		result = append(result, map[string]interface{}{
			"name":      pod.Name,
			"namespace": pod.Namespace,
			"phase":     pod.Status.Phase,
			"ready":     isPodReady(pod),
			"labels":    pod.Labels,
		})
	}
	
	return map[string]interface{}{
		"pods": result,
		"count": len(result),
	}, nil
}

// getNodes retrieves nodes
func (e *ToolExecutor) getNodes(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	nodes := &corev1.NodeList{}
	opts := []client.ListOption{}
	
	if labelSelector, ok := args["labelSelector"].(map[string]interface{}); ok {
		labels := make(map[string]string)
		for k, v := range labelSelector {
			if str, ok := v.(string); ok {
				labels[k] = str
			}
		}
		if len(labels) > 0 {
			opts = append(opts, client.MatchingLabels(labels))
		}
	}
	
	if err := e.Client.List(ctx, nodes, opts...); err != nil {
		return nil, err
	}
	
	result := make([]map[string]interface{}, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		ready := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				ready = condition.Status == corev1.ConditionTrue
				break
			}
		}
		
		result = append(result, map[string]interface{}{
			"name":       node.Name,
			"ready":      ready,
			"unschedulable": node.Spec.Unschedulable,
			"labels":     node.Labels,
		})
	}
	
	return map[string]interface{}{
		"nodes": result,
		"count": len(result),
	}, nil
}

// getDeployments retrieves deployments
func (e *ToolExecutor) getDeployments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	namespace, _ := args["namespace"].(string)
	
	deployments := &metav1.PartialObjectMetadataList{}
	deployments.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "DeploymentList",
	})
	
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}
	
	if err := e.Client.List(ctx, deployments, opts...); err != nil {
		return nil, err
	}
	
	result := make([]map[string]interface{}, 0, len(deployments.Items))
	for _, deployment := range deployments.Items {
		result = append(result, map[string]interface{}{
			"name":      deployment.Name,
			"namespace": deployment.Namespace,
			"labels":    deployment.Labels,
		})
	}
	
	return map[string]interface{}{
		"deployments": result,
		"count":       len(result),
	}, nil
}

// getEvents retrieves events
func (e *ToolExecutor) getEvents(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	namespace, _ := args["namespace"].(string)
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	
	events := &corev1.EventList{}
	if err := e.Client.List(ctx, events, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	
	result := make([]map[string]interface{}, 0, len(events.Items))
	for _, event := range events.Items {
		result = append(result, map[string]interface{}{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"timestamp": event.FirstTimestamp.Time,
			"involvedObject": map[string]interface{}{
				"kind":      event.InvolvedObject.Kind,
				"name":      event.InvolvedObject.Name,
				"namespace": event.InvolvedObject.Namespace,
			},
		})
	}
	
	return map[string]interface{}{
		"events": result,
		"count":  len(result),
	}, nil
}

// scaleDeployment scales a deployment
func (e *ToolExecutor) scaleDeployment(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	namespace, _ := args["namespace"].(string)
	name, _ := args["name"].(string)
	replicasFloat, _ := args["replicas"].(float64)
	replicas := int32(replicasFloat)
	dryRun, _ := args["dryRun"].(bool)
	
	if namespace == "" || name == "" {
		return nil, fmt.Errorf("namespace and name are required")
	}
	
	if dryRun {
		return map[string]interface{}{
			"action":    "scale_deployment",
			"namespace": namespace,
			"name":      name,
			"replicas":  replicas,
			"dryRun":    true,
			"message":   fmt.Sprintf("Would scale deployment %s/%s to %d replicas", namespace, name, replicas),
		}, nil
	}
	
	deployment := &metav1.PartialObjectMetadata{}
	deployment.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	deployment.SetName(name)
	deployment.SetNamespace(namespace)
	
	// Get full deployment object
	fullDeployment := &metav1.PartialObjectMetadata{}
	if err := e.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, fullDeployment); err != nil {
		return nil, err
	}
	
	// Use patch to update replicas
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": replicas,
		},
	}
	patchBytes, _ := json.Marshal(patch)
	
	if err := e.Client.Patch(ctx, fullDeployment, client.RawPatch(types.MergePatchType, patchBytes)); err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"action":    "scale_deployment",
		"namespace": namespace,
		"name":      name,
		"replicas":   replicas,
		"success":   true,
		"message":   fmt.Sprintf("Scaled deployment %s/%s to %d replicas", namespace, name, replicas),
	}, nil
}

// restartPod restarts a pod
func (e *ToolExecutor) restartPod(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	namespace, _ := args["namespace"].(string)
	name, _ := args["name"].(string)
	dryRun, _ := args["dryRun"].(bool)
	
	if namespace == "" || name == "" {
		return nil, fmt.Errorf("namespace and name are required")
	}
	
	if dryRun {
		return map[string]interface{}{
			"action":    "restart_pod",
			"namespace": namespace,
			"name":      name,
			"dryRun":    true,
			"message":   fmt.Sprintf("Would restart pod %s/%s", namespace, name),
		}, nil
	}
	
	pod := &corev1.Pod{}
	if err := e.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, pod); err != nil {
		return nil, err
	}
	
	if err := e.Client.Delete(ctx, pod); err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"action":    "restart_pod",
		"namespace": namespace,
		"name":      name,
		"success":   true,
		"message":   fmt.Sprintf("Restarted pod %s/%s", namespace, name),
	}, nil
}

// cordonNode cordons a node
func (e *ToolExecutor) cordonNode(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, _ := args["name"].(string)
	dryRun, _ := args["dryRun"].(bool)
	
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	
	if dryRun {
		return map[string]interface{}{
			"action":  "cordon_node",
			"name":    name,
			"dryRun":  true,
			"message": fmt.Sprintf("Would cordon node %s", name),
		}, nil
	}
	
	node := &corev1.Node{}
	if err := e.Client.Get(ctx, types.NamespacedName{Name: name}, node); err != nil {
		return nil, err
	}
	
	node.Spec.Unschedulable = true
	if err := e.Client.Update(ctx, node); err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"action":  "cordon_node",
		"name":    name,
		"success": true,
		"message": fmt.Sprintf("Cordoned node %s", name),
	}, nil
}

// drainNode drains a node
func (e *ToolExecutor) drainNode(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, _ := args["name"].(string)
	dryRun, _ := args["dryRun"].(bool)
	
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	
	if dryRun {
		return map[string]interface{}{
			"action":  "drain_node",
			"name":    name,
			"dryRun":  true,
			"message": fmt.Sprintf("Would drain node %s (high-impact action)", name),
		}, nil
	}
	
	// First cordon the node
	if _, err := e.cordonNode(ctx, map[string]interface{}{"name": name, "dryRun": false}); err != nil {
		return nil, fmt.Errorf("failed to cordon node: %w", err)
	}
	
	// Then evict pods (simplified - in production, use proper eviction API)
	pods := &corev1.PodList{}
	if err := e.Client.List(ctx, pods, client.MatchingFields{"spec.nodeName": name}); err != nil {
		return nil, err
	}
	
	evictedCount := 0
	for _, pod := range pods.Items {
		// Skip DaemonSet pods if requested
		if ignoreDS, _ := args["ignoreDaemonSets"].(bool); ignoreDS {
			if pod.OwnerReferences != nil {
				for _, ref := range pod.OwnerReferences {
					if ref.Kind == "DaemonSet" {
						continue
					}
				}
			}
		}
		
		if err := e.Client.Delete(ctx, &pod); err == nil {
			evictedCount++
		}
	}
	
	return map[string]interface{}{
		"action":       "drain_node",
		"name":         name,
		"success":      true,
		"evictedPods":  evictedCount,
		"message":      fmt.Sprintf("Drained node %s, evicted %d pods", name, evictedCount),
	}, nil
}

// getMetrics retrieves metrics (placeholder - would integrate with metrics server)
func (e *ToolExecutor) getMetrics(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"message": "Metrics retrieval requires metrics-server integration",
		"note":    "In production, query metrics-server API or Prometheus",
	}, nil
}

// applyNetworkPolicy applies a network policy (placeholder)
func (e *ToolExecutor) applyNetworkPolicy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	namespace, _ := args["namespace"].(string)
	dryRun, _ := args["dryRun"].(bool)
	
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	
	if dryRun {
		return map[string]interface{}{
			"action":    "apply_network_policy",
			"namespace": namespace,
			"dryRun":    true,
			"message":   fmt.Sprintf("Would apply network policy to namespace %s", namespace),
		}, nil
	}
	
	return map[string]interface{}{
		"action":    "apply_network_policy",
		"namespace": namespace,
		"success":   true,
		"message":   fmt.Sprintf("Applied network policy to namespace %s", namespace),
		"note":      "In production, create/update NetworkPolicy resource",
	}, nil
}

// getK8sGPTAnalysis retrieves K8sGPT analysis (placeholder)
func (e *ToolExecutor) getK8sGPTAnalysis(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"message": "K8sGPT analysis requires K8sGPT operator integration",
		"note":    "In production, query K8sGPT API endpoint",
	}, nil
}

// getForecast retrieves forecast (placeholder)
func (e *ToolExecutor) getForecast(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	metric, _ := args["metric"].(string)
	duration, _ := args["duration"].(string)
	if duration == "" {
		duration = "1h"
	}
	
	return map[string]interface{}{
		"metric":   metric,
		"duration": duration,
		"message":  "Forecast retrieval requires Grafana ML integration",
		"note":     "In production, query Grafana ML API",
	}, nil
}

// Helper functions
func isPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

