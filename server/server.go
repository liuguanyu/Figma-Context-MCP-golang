package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	router   *mux.Router
	sessions map[string]*Session
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Server    string `json:"server"`
	Version   string `json:"version"`
}

func NewServer() http.Handler {
	s := &Server{
		router:   mux.NewRouter(),
		sessions: make(map[string]*Session),
	}

	s.setupRoutes()

	// 设置CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	return c.Handler(s.router)
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.router.HandleFunc("/mcp", s.mcpHandler).Methods("POST")
	s.router.HandleFunc("/sse", s.sseHandler).Methods("GET")
	s.router.HandleFunc("/messages", s.messagesHandler).Methods("POST")
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Server:    "Figma MCP Server",
		Version:   "0.4.2",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
