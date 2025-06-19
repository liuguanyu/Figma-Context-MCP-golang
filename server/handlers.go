package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"figma-mcp-server/mcp"
	"figma-mcp-server/types"
)

func (s *Server) mcpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Received StreamableHTTP request")

	// 获取会话ID
	sessionID := r.Header.Get("mcp-session-id")

	if sessionID == "" {
		log.Println("[INFO] New initialization request for StreamableHTTP sessionId undefined")
		sessionID = generateSessionID()
	} else {
		log.Printf("[INFO] Reusing existing StreamableHTTP transport for sessionId %s", sessionID)
	}

	// 设置SSE头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("mcp-session-id", sessionID)

	// 读取请求体
	var req types.MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("[INFO] Handling StreamableHTTP request")

	// 处理MCP请求
	response := s.handleMCPRequest(&req, sessionID)

	// 发送SSE响应
	s.sendSSEResponse(w, response)

	log.Println("[INFO] StreamableHTTP request handled")
}

func (s *Server) sseHandler(w http.ResponseWriter, r *http.Request) {
	// SSE连接处理
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 简单的SSE连接保持
	fmt.Fprintf(w, "data: {\"status\":\"connected\"}\n\n")
	w.(http.Flusher).Flush()
}

func (s *Server) messagesHandler(w http.ResponseWriter, r *http.Request) {
	// 消息处理端点
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleMCPRequest(req *types.MCPRequest, sessionID string) *types.MCPResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req, sessionID)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &types.MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *Server) handleInitialize(req *types.MCPRequest, sessionID string) *types.MCPResponse {
	// 创建或获取会话
	_ = s.getOrCreateSession(sessionID)

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]string{
				"name":    "Figma MCP Server",
				"version": "0.4.2",
			},
		},
	}
}

func (s *Server) handleToolsList(req *types.MCPRequest) *types.MCPResponse {
	tools := mcp.GetAvailableTools()

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *Server) handleToolsCall(req *types.MCPRequest) *types.MCPResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &types.MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &types.MCPError{
				Code:    -32602,
				Message: "Missing tool name",
			},
		}
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &types.MCPError{
				Code:    -32602,
				Message: "Missing arguments",
			},
		}
	}

	// 调用工具
	result, err := mcp.CallTool(toolName, arguments)
	if err != nil {
		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &types.MCPError{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) sendSSEResponse(w http.ResponseWriter, response *types.MCPResponse) {
	data, _ := json.Marshal(response)
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(data))
	w.(http.Flusher).Flush()
}

func (s *Server) getOrCreateSession(sessionID string) *Session {
	if session, exists := s.sessions[sessionID]; exists {
		return session
	}

	session := &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
	}
	s.sessions[sessionID] = session
	return session
}

func generateSessionID() string {
	// 生成UUID风格的会话ID
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		time.Now().Unix(),
		time.Now().Nanosecond()%0x10000,
		time.Now().Nanosecond()%0x10000,
		time.Now().Nanosecond()%0x10000,
		time.Now().Nanosecond(),
	)
}
