// Package runtime provides the agent runtime abstraction layer.
// This enables Gas Town to support multiple agent backends:
// - TmuxRuntime: Current implementation (Claude Code CLI in tmux)
// - SDKRuntime: Claude Agent SDK (Phase 3)
package runtime

import (
	"context"
	"time"
)

// AgentRuntime defines the interface for managing Claude agent instances.
// Implementations must be safe for concurrent use.
type AgentRuntime interface {
	// Lifecycle
	Start(ctx context.Context, opts StartOptions) (*AgentSession, error)
	Stop(ctx context.Context, sessionID string, force bool) error
	Restart(ctx context.Context, sessionID string, opts StartOptions) (*AgentSession, error)

	// Communication
	SendPrompt(ctx context.Context, sessionID string, prompt string) error
	StreamResponses(ctx context.Context, sessionID string) (<-chan Response, error)

	// Monitoring
	IsRunning(ctx context.Context, sessionID string) (bool, error)
	GetStatus(ctx context.Context, sessionID string) (*AgentStatus, error)
	ListSessions(ctx context.Context, filter SessionFilter) ([]AgentSession, error)

	// Activity
	GetActivity(ctx context.Context, sessionID string) (*ActivityInfo, error)
	CaptureOutput(ctx context.Context, sessionID string, lines int) (string, error)

	// Capabilities
	Capabilities() RuntimeCapabilities

	// Lifecycle
	Close() error
}

// StartOptions configures agent startup.
type StartOptions struct {
	// Identity
	AgentID    string    // e.g., "gastown/polecats/toast"
	Role       AgentRole // polecat, witness, refinery, mayor, deacon, crew
	RigName    string    // e.g., "gastown"
	WorkerName string    // e.g., "toast"

	// Environment
	WorkDir     string            // Working directory
	Environment map[string]string // Additional env vars

	// Configuration
	SystemPrompt  string       // System prompt (SDK only, ignored by tmux)
	Tools         []ToolConfig // Tool configurations (SDK only)
	InitialPrompt string       // First prompt to send after startup

	// Work assignment
	HookBead string // Issue ID to hook on startup

	// Account (tmux only)
	Account         string // Claude account handle
	ClaudeConfigDir string // Resolved CLAUDE_CONFIG_DIR

	// Runtime command (tmux only)
	Command string   // Override command (e.g., "claude")
	Args    []string // Command arguments

	// Behavior
	WaitForReady bool          // Block until agent is responsive
	ReadyTimeout time.Duration // Timeout for ready check
}

// AgentRole defines the type of agent.
type AgentRole string

const (
	RolePolecat  AgentRole = "polecat"
	RoleWitness  AgentRole = "witness"
	RoleRefinery AgentRole = "refinery"
	RoleMayor    AgentRole = "mayor"
	RoleDeacon   AgentRole = "deacon"
	RoleCrew     AgentRole = "crew"
)

// AgentSession represents a running agent instance.
type AgentSession struct {
	// Identity
	SessionID  string    `json:"session_id"`            // Unique session identifier
	AgentID    string    `json:"agent_id"`              // Logical agent ID (e.g., "gastown/polecats/toast")
	Role       AgentRole `json:"role"`                  //nolint:tagliatelle // Role is capitalized intentionally
	RigName    string    `json:"rig_name,omitempty"`    //nolint:tagliatelle
	WorkerName string    `json:"worker_name,omitempty"` //nolint:tagliatelle

	// State
	Running   bool      `json:"running"`
	StartedAt time.Time `json:"started_at"` //nolint:tagliatelle

	// Runtime-specific
	RuntimeType string `json:"runtime_type"`           // "tmux" or "sdk"
	RuntimeMeta any    `json:"runtime_meta,omitempty"` //nolint:tagliatelle
}

// AgentStatus provides detailed status information.
type AgentStatus struct {
	Session     AgentSession `json:"session"`
	Health      HealthState  `json:"health"`
	Activity    ActivityInfo `json:"activity"`
	CurrentWork *WorkInfo    `json:"current_work,omitempty"` //nolint:tagliatelle

	// Runtime-specific status
	TmuxInfo *TmuxStatus `json:"tmux_info,omitempty"` //nolint:tagliatelle
	SDKInfo  *SDKStatus  `json:"sdk_info,omitempty"`  //nolint:tagliatelle
}

// HealthState represents agent health.
type HealthState string

const (
	HealthHealthy   HealthState = "healthy"
	HealthDegraded  HealthState = "degraded"
	HealthUnhealthy HealthState = "unhealthy"
	HealthUnknown   HealthState = "unknown"
)

// ActivityInfo provides activity timing information.
type ActivityInfo struct {
	LastActivity  time.Time     `json:"last_activity"`  //nolint:tagliatelle
	IdleDuration  time.Duration `json:"idle_duration"`  //nolint:tagliatelle
	ActivityState string        `json:"activity_state"` // "active", "stale", "stuck" //nolint:tagliatelle
	LastPrompt    time.Time     `json:"last_prompt,omitempty"`
	LastResponse  time.Time     `json:"last_response,omitempty"`
}

// WorkInfo describes current work assignment.
type WorkInfo struct {
	BeadID     string    `json:"bead_id"`              //nolint:tagliatelle
	Title      string    `json:"title"`                //nolint:tagliatelle
	ConvoyID   string    `json:"convoy_id,omitempty"`  //nolint:tagliatelle
	AssignedAt time.Time `json:"assigned_at,omitempty"` //nolint:tagliatelle
}

// Response represents a response from an agent.
type Response struct {
	Type      ResponseType `json:"type"`
	Content   string       `json:"content"`
	Timestamp time.Time    `json:"timestamp"`

	// For tool calls (SDK)
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`   //nolint:tagliatelle
	ToolResult *ToolResult `json:"tool_result,omitempty"` //nolint:tagliatelle

	// For errors
	Error error `json:"error,omitempty"`
}

// ResponseType categorizes response content.
type ResponseType string

const (
	ResponseText       ResponseType = "text"
	ResponseToolCall   ResponseType = "tool_call"
	ResponseToolResult ResponseType = "tool_result"
	ResponseError      ResponseType = "error"
	ResponseComplete   ResponseType = "complete"
)

// SessionFilter for listing sessions.
type SessionFilter struct {
	RigName string
	Role    AgentRole
	Running *bool
	AgentID string
}

// RuntimeCapabilities describes what a runtime supports.
type RuntimeCapabilities struct {
	SupportsStreaming    bool `json:"supports_streaming"`     //nolint:tagliatelle
	SupportsToolCalls    bool `json:"supports_tool_calls"`    //nolint:tagliatelle
	SupportsSystemPrompt bool `json:"supports_system_prompt"` //nolint:tagliatelle
	SupportsAttach       bool `json:"supports_attach"`        // Terminal attachment //nolint:tagliatelle
	SupportsCapture      bool `json:"supports_capture"`       // Output capture //nolint:tagliatelle
	SupportsConcurrency  int  `json:"supports_concurrency"`   // Max concurrent sessions (0 = unlimited) //nolint:tagliatelle
}

// TmuxStatus contains tmux-specific status info.
type TmuxStatus struct {
	SessionName string `json:"session_name"` //nolint:tagliatelle
	PaneID      string `json:"pane_id"`      //nolint:tagliatelle
	Attached    bool   `json:"attached"`
	Windows     int    `json:"windows"`
	PaneCommand string `json:"pane_command,omitempty"` //nolint:tagliatelle
}

// SDKStatus contains SDK-specific status info.
type SDKStatus struct {
	ConversationID string `json:"conversation_id"` //nolint:tagliatelle
	TokensUsed     int    `json:"tokens_used"`     //nolint:tagliatelle
	TurnCount      int    `json:"turn_count"`      //nolint:tagliatelle
}

// ToolConfig defines a tool available to the agent.
type ToolConfig struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"` //nolint:tagliatelle
	Handler     ToolHandler    `json:"-"`            // Function to execute tool
}

// ToolHandler executes a tool and returns the result.
type ToolHandler func(ctx context.Context, input map[string]any) (any, error)

// ToolCall represents a tool invocation request from the agent.
type ToolCall struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	CallID string `json:"call_id"` //nolint:tagliatelle
	Output any    `json:"output"`
	Error  string `json:"error,omitempty"`
}

// GenerateSessionID creates a session ID following Gas Town conventions.
func GenerateSessionID(opts StartOptions) string {
	switch opts.Role {
	case RoleMayor:
		return "hq-mayor"
	case RoleDeacon:
		return "hq-deacon"
	case RoleWitness:
		return "gt-" + opts.RigName + "-witness"
	case RoleRefinery:
		return "gt-" + opts.RigName + "-refinery"
	case RoleCrew:
		return "gt-" + opts.RigName + "-crew-" + opts.WorkerName
	case RolePolecat:
		return "gt-" + opts.RigName + "-" + opts.WorkerName
	default:
		return "gt-" + opts.RigName + "-" + opts.WorkerName
	}
}
