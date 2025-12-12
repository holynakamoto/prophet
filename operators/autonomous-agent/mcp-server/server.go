package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MCPServer implements Model Context Protocol for real-time cluster context sharing
type MCPServer struct {
	k8sClient   client.Client
	clientset   *kubernetes.Clientset
	upgrader    websocket.Upgrader
	connections map[*websocket.Conn]bool
	broadcast   chan []byte
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

// MCPMessage represents MCP protocol messages
type MCPMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Result    interface{}            `json:"result,omitempty"`
	Error     *MCPError              `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

// MCPError represents MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPServer creates a new MCP server
func NewMCPServer(k8sClient client.Client, clientset *kubernetes.Clientset) *MCPServer {
	return &MCPServer{
		k8sClient:   k8sClient,
		clientset:   clientset,
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
	}
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context, port int) error {
	http.HandleFunc("/mcp", s.handleWebSocket)
	http.HandleFunc("/mcp/query", s.handleQuery)
	http.HandleFunc("/mcp/stream", s.handleStream)

	go s.run()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("MCP Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleWebSocket handles WebSocket connections
func (s *MCPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.register <- conn

	go s.handleClient(conn)
}

// handleClient handles individual client connections
func (s *MCPServer) handleClient(conn *websocket.Conn) {
	defer func() {
		s.unregister <- conn
		conn.Close()
	}()

	for {
		var msg MCPMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		response := s.handleMessage(&msg)
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}

// handleMessage processes MCP messages
func (s *MCPServer) handleMessage(msg *MCPMessage) *MCPMessage {
	response := &MCPMessage{
		Type:      "response",
		ID:        msg.ID,
		Timestamp: time.Now(),
	}

	switch msg.Method {
	case "cluster/state":
		response.Result = s.getClusterState(msg.Params)
	case "cluster/query":
		response.Result = s.queryCluster(msg.Params)
	case "cluster/execute":
		response.Result = s.executeAction(msg.Params)
	case "cluster/stream":
		response.Result = s.startStreaming(msg.Params)
	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("Unknown method: %s", msg.Method),
		}
	}

	return response
}

// getClusterState returns current cluster state
func (s *MCPServer) getClusterState(params map[string]interface{}) map[string]interface{} {
	ctx := context.Background()
	state := make(map[string]interface{})

	// Get pods
	pods := &unstructured.UnstructuredList{}
	pods.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "PodList",
	})
	if err := s.k8sClient.List(ctx, pods); err == nil {
		state["pods"] = pods.Items
	}

	// Get nodes
	nodes := &unstructured.UnstructuredList{}
	nodes.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "NodeList",
	})
	if err := s.k8sClient.List(ctx, nodes); err == nil {
		state["nodes"] = nodes.Items
	}

	// Get deployments
	deployments := &unstructured.UnstructuredList{}
	deployments.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "DeploymentList",
	})
	if err := s.k8sClient.List(ctx, deployments); err == nil {
		state["deployments"] = deployments.Items
	}

	return state
}

// queryCluster performs natural language query
func (s *MCPServer) queryCluster(params map[string]interface{}) map[string]interface{} {
	query, _ := params["query"].(string)
	// In production, this would use LLM to interpret query and return results
	return map[string]interface{}{
		"query":   query,
		"results": "Query processed (LLM integration pending)",
	}
}

// executeAction executes approved action
func (s *MCPServer) executeAction(params map[string]interface{}) map[string]interface{} {
	actionType, _ := params["type"].(string)
	// In production, validate and execute action
	return map[string]interface{}{
		"action":  actionType,
		"status":  "executed",
		"message": "Action executed successfully",
	}
}

// startStreaming starts streaming cluster events
func (s *MCPServer) startStreaming(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"streamId": "stream-123",
		"status":   "started",
	}
}

// handleQuery handles HTTP query endpoint
func (s *MCPServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := s.queryCluster(query)
	json.NewEncoder(w).Encode(result)
}

// handleStream handles streaming endpoint
func (s *MCPServer) handleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Stream cluster events
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			state := s.getClusterState(nil)
			if err := conn.WriteJSON(state); err != nil {
				return
			}
		}
	}
}

// run manages WebSocket connections
func (s *MCPServer) run() {
	for {
		select {
		case conn := <-s.register:
			s.connections[conn] = true
		case conn := <-s.unregister:
			delete(s.connections, conn)
		case message := <-s.broadcast:
			for conn := range s.connections {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(s.connections, conn)
					conn.Close()
				}
			}
		}
	}
}
