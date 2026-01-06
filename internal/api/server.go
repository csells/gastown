// Package api provides a REST/WebSocket server for the Gas Town runtime.
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/steveyegge/gastown/internal/runtime"
)

// Server exposes Gas Town runtime operations via REST and WebSocket.
type Server struct {
	runtime  runtime.AgentRuntime
	upgrader websocket.Upgrader
	addr     string

	// Track WebSocket connections per session
	wsConns   map[string][]*websocket.Conn
	wsConnsMu sync.RWMutex
}

// NewServer creates a new API server.
func NewServer(rt runtime.AgentRuntime, addr string) *Server {
	return &Server{
		runtime: rt,
		addr:    addr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		wsConns: make(map[string][]*websocket.Conn),
	}
}

// Start begins serving HTTP requests.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// REST endpoints
	mux.HandleFunc("POST /sessions", s.handleCreateSession)
	mux.HandleFunc("DELETE /sessions/{id}", s.handleDeleteSession)
	mux.HandleFunc("GET /sessions/{id}", s.handleGetSession)
	mux.HandleFunc("GET /sessions", s.handleListSessions)
	mux.HandleFunc("GET /sessions/{id}/output", s.handleCaptureOutput)

	// WebSocket for streaming
	mux.HandleFunc("GET /sessions/{id}/ws", s.handleWebSocket)

	// Health check
	mux.HandleFunc("GET /health", s.handleHealth)

	log.Printf("Gas Town API server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

// --- Request/Response Types ---

type CreateSessionRequest struct {
	AgentID      string            `json:"agent_id"`
	Role         string            `json:"role"`
	RigName      string            `json:"rig_name,omitempty"`
	WorkerName   string            `json:"worker_name,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
}

type SessionResponse struct {
	SessionID   string    `json:"session_id"`
	AgentID     string    `json:"agent_id"`
	Role        string    `json:"role"`
	RigName     string    `json:"rig_name,omitempty"`
	WorkerName  string    `json:"worker_name,omitempty"`
	Running     bool      `json:"running"`
	StartedAt   time.Time `json:"started_at"`
	RuntimeType string    `json:"runtime_type"`
}

type StatusResponse struct {
	Session  SessionResponse `json:"session"`
	Health   string          `json:"health"`
	Activity ActivityInfo    `json:"activity"`
	SDKInfo  *SDKInfo        `json:"sdk_info,omitempty"`
}

type ActivityInfo struct {
	LastActivity  time.Time `json:"last_activity"`
	IdleDuration  string    `json:"idle_duration"`
	ActivityState string    `json:"activity_state"`
}

type SDKInfo struct {
	TokensUsed int `json:"tokens_used"`
	TurnCount  int `json:"turn_count"`
}

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// WebSocket message types
type WSMessage struct {
	Type      string `json:"type"` // "text", "tool_call", "tool_result", "error", "complete"
	Content   string `json:"content,omitempty"`
	Timestamp string `json:"timestamp"`
	Error     string `json:"error,omitempty"`
}

// --- Handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	opts := runtime.StartOptions{
		AgentID:      req.AgentID,
		Role:         runtime.AgentRole(req.Role),
		RigName:      req.RigName,
		WorkerName:   req.WorkerName,
		SystemPrompt: req.SystemPrompt,
		Environment:  req.Environment,
	}

	session, err := s.runtime.Start(r.Context(), opts)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Start streaming responses to WebSocket clients
	go s.streamToWebSockets(session.SessionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s.sessionToResponse(session))
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "session id required")
		return
	}

	force := r.URL.Query().Get("force") == "true"
	if err := s.runtime.Stop(r.Context(), sessionID, force); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "session id required")
		return
	}

	status, err := s.runtime.GetStatus(r.Context(), sessionID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := StatusResponse{
		Session: s.sessionToResponse(&status.Session),
		Health:  string(status.Health),
		Activity: ActivityInfo{
			LastActivity:  status.Activity.LastActivity,
			IdleDuration:  status.Activity.IdleDuration.String(),
			ActivityState: status.Activity.ActivityState,
		},
	}
	if status.SDKInfo != nil {
		resp.SDKInfo = &SDKInfo{
			TokensUsed: status.SDKInfo.TokensUsed,
			TurnCount:  status.SDKInfo.TurnCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	filter := runtime.SessionFilter{
		RigName: r.URL.Query().Get("rig"),
		Role:    runtime.AgentRole(r.URL.Query().Get("role")),
	}

	sessions, err := s.runtime.ListSessions(r.Context(), filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var resp []SessionResponse
	for _, sess := range sessions {
		resp = append(resp, s.sessionToResponse(&sess))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleCaptureOutput(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "session id required")
		return
	}

	output, err := s.runtime.CaptureOutput(r.Context(), sessionID, 50)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"output": output})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "session id required")
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Register connection
	s.wsConnsMu.Lock()
	s.wsConns[sessionID] = append(s.wsConns[sessionID], conn)
	s.wsConnsMu.Unlock()

	// Handle incoming messages (prompts from client)
	go func() {
		defer func() {
			s.removeWSConn(sessionID, conn)
			conn.Close()
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req PromptRequest
			if err := json.Unmarshal(msg, &req); err != nil {
				continue
			}

			if req.Prompt != "" {
				_ = s.runtime.SendPrompt(context.Background(), sessionID, req.Prompt)
			}
		}
	}()
}

// streamToWebSockets forwards runtime responses to connected WebSocket clients.
func (s *Server) streamToWebSockets(sessionID string) {
	respCh, err := s.runtime.StreamResponses(context.Background(), sessionID)
	if err != nil {
		return
	}

	for resp := range respCh {
		msg := WSMessage{
			Type:      string(resp.Type),
			Timestamp: resp.Timestamp.Format(time.RFC3339),
		}

		switch resp.Type {
		case runtime.ResponseText:
			msg.Content = resp.Content
		case runtime.ResponseError:
			msg.Error = resp.Error.Error()
		case runtime.ResponseComplete:
			msg.Content = "complete"
		}

		s.broadcastToSession(sessionID, msg)
	}
}

func (s *Server) broadcastToSession(sessionID string, msg WSMessage) {
	s.wsConnsMu.RLock()
	conns := s.wsConns[sessionID]
	s.wsConnsMu.RUnlock()

	data, _ := json.Marshal(msg)
	for _, conn := range conns {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (s *Server) removeWSConn(sessionID string, conn *websocket.Conn) {
	s.wsConnsMu.Lock()
	defer s.wsConnsMu.Unlock()

	conns := s.wsConns[sessionID]
	for i, c := range conns {
		if c == conn {
			s.wsConns[sessionID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (s *Server) sessionToResponse(sess *runtime.AgentSession) SessionResponse {
	return SessionResponse{
		SessionID:   sess.SessionID,
		AgentID:     sess.AgentID,
		Role:        string(sess.Role),
		RigName:     sess.RigName,
		WorkerName:  sess.WorkerName,
		Running:     sess.Running,
		StartedAt:   sess.StartedAt,
		RuntimeType: sess.RuntimeType,
	}
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
