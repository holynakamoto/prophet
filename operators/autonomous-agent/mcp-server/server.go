package mcpserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/prophet-aiops/autonomous-agent/mcp-server/tools"
)

// TLSOptions configures TLS for the MCP server.
type TLSOptions struct {
	// CertFile and KeyFile are the server certificate and private key files.
	CertFile string
	KeyFile  string

	// ClientCAFile, if set, enables mTLS by requiring and verifying client certificates.
	ClientCAFile string
}

// MCPServer implements Model Context Protocol for real-time cluster context sharing
type MCPServer struct {
	k8sClient    client.Client
	clientset    *kubernetes.Clientset
	toolExecutor *tools.ToolExecutor
	upgrader     websocket.Upgrader
	connections  map[*websocket.Conn]bool
	broadcast    chan []byte
	register     chan *websocket.Conn
	unregister   chan *websocket.Conn
	mu           sync.RWMutex
	runOnce      sync.Once
}

// MCPMessage represents MCP protocol messages
type MCPMessage struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *MCPError              `json:"error,omitempty"`
}

// MCPError represents MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPTool represents a tool definition
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// NewMCPServer creates a new MCP server
func NewMCPServer(k8sClient client.Client, clientset *kubernetes.Clientset) *MCPServer {
	server := &MCPServer{
		k8sClient:    k8sClient,
		clientset:    clientset,
		toolExecutor: tools.NewToolExecutor(k8sClient),
		upgrader: websocket.Upgrader{
			CheckOrigin:     func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
	}
	return server
}

func (s *MCPServer) newMux() *http.ServeMux {
	mux := http.NewServeMux()

	// MCP protocol endpoints
	mux.HandleFunc("/mcp", s.handleWebSocket)
	mux.HandleFunc("/mcp/initialize", s.handleInitialize)
	mux.HandleFunc("/mcp/tools/list", s.handleToolsList)
	mux.HandleFunc("/mcp/tools/call", s.handleToolCall)
	mux.HandleFunc("/mcp/query", s.handleQuery)
	mux.HandleFunc("/mcp/stream", s.handleStream)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context, port int) error {
	mux := s.newMux()
	s.runOnce.Do(func() { go s.run() })

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("MCP Server (HTTP) starting on %s", addr)

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

func (s *MCPServer) StartTLS(ctx context.Context, port int, opts TLSOptions) error {
	if opts.CertFile == "" || opts.KeyFile == "" {
		return fmt.Errorf("tls enabled but cert/key not provided")
	}

	mux := s.newMux()
	s.runOnce.Do(func() { go s.run() })

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if opts.ClientCAFile != "" {
		pem, err := os.ReadFile(opts.ClientCAFile)
		if err != nil {
			return fmt.Errorf("read client CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(pem); !ok {
			return fmt.Errorf("no certificates found in client CA file")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	log.Printf("MCP Server (HTTPS) starting on %s", addr)

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServeTLS(opts.CertFile, opts.KeyFile)
}

// handleInitialize handles MCP initialization
func (s *MCPServer) handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "prophet-mcp-server",
				"version": "1.0.0",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleToolsList returns list of available tools
func (s *MCPServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toolList := tools.GetTools()
	result := make([]MCPTool, 0, len(toolList))
	for _, tool := range toolList {
		result = append(result, MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}

	response := MCPMessage{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"tools": result,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleToolCall executes an MCP tool
func (s *MCPServer) handleToolCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := req.Params
	if params == nil {
		http.Error(w, "params required", http.StatusBadRequest)
		return
	}

	toolName, ok := params["name"].(string)
	if !ok {
		http.Error(w, "tool name required", http.StatusBadRequest)
		return
	}

	arguments, _ := params["arguments"].(map[string]interface{})
	if arguments == nil {
		arguments = make(map[string]interface{})
	}

	// Execute tool
	result, err := s.toolExecutor.ExecuteTool(r.Context(), toolName, arguments)

	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	if err != nil {
		response.Error = &MCPError{
			Code:    -32603,
			Message: err.Error(),
		}
	} else {
		response.Result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("%v", result),
				},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		JSONRPC: "2.0",
		ID:      msg.ID,
	}

	ctx := context.Background()

	switch msg.Method {
	case "initialize":
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "prophet-mcp-server",
				"version": "1.0.0",
			},
		}
	case "tools/list":
		toolList := tools.GetTools()
		result := make([]MCPTool, 0, len(toolList))
		for _, tool := range toolList {
			result = append(result, MCPTool{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			})
		}
		response.Result = map[string]interface{}{
			"tools": result,
		}
	case "tools/call":
		params := msg.Params
		if params == nil {
			response.Error = &MCPError{
				Code:    -32602,
				Message: "params required",
			}
			break
		}

		toolName, ok := params["name"].(string)
		if !ok {
			response.Error = &MCPError{
				Code:    -32602,
				Message: "tool name required",
			}
			break
		}

		arguments, _ := params["arguments"].(map[string]interface{})
		if arguments == nil {
			arguments = make(map[string]interface{})
		}

		result, err := s.toolExecutor.ExecuteTool(ctx, toolName, arguments)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("%v", result),
					},
				},
			}
		}
	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("Unknown method: %s", msg.Method),
		}
	}

	return response
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

	// Use natural language query to determine which tools to call
	queryText, _ := query["query"].(string)

	// Simple keyword matching (in production, use LLM to interpret query)
	result := map[string]interface{}{
		"query":   queryText,
		"message": "Query processed. Use tools/call endpoint to execute specific actions.",
		"availableTools": []string{
			"k8s_get_pods", "k8s_get_nodes", "k8s_get_deployments",
			"k8s_scale_deployment", "k8s_restart_pod", "k8s_cordon_node",
		},
	}

	w.Header().Set("Content-Type", "application/json")
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

	for range ticker.C {
		// Get cluster state snapshot
		state := map[string]interface{}{
			"timestamp": time.Now(),
			"message":   "Cluster state snapshot",
		}
		if err := conn.WriteJSON(state); err != nil {
			return
		}
	}
}

// ToolExecutor returns the tool executor
func (s *MCPServer) ToolExecutor() *tools.ToolExecutor {
	return s.toolExecutor
}

// run manages WebSocket connections
func (s *MCPServer) run() {
	for {
		select {
		case conn := <-s.register:
			s.mu.Lock()
			s.connections[conn] = true
			s.mu.Unlock()
		case conn := <-s.unregister:
			s.mu.Lock()
			delete(s.connections, conn)
			s.mu.Unlock()
		case message := <-s.broadcast:
			s.mu.RLock()
			for conn := range s.connections {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					s.mu.RUnlock()
					s.mu.Lock()
					delete(s.connections, conn)
					s.mu.Unlock()
					conn.Close()
					s.mu.RLock()
				}
			}
			s.mu.RUnlock()
		}
	}
}
