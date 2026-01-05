package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/tmux"
)

// TmuxRuntime implements AgentRuntime using tmux sessions and Claude Code CLI.
// This preserves all existing Gas Town behavior.
type TmuxRuntime struct {
	tmux     *tmux.Tmux
	sessions sync.Map // sessionID -> *tmuxSessionState
}

// tmuxSessionState tracks a running tmux session.
type tmuxSessionState struct {
	AgentSession
	workDir string
}

// NewTmuxRuntime creates a new tmux-based runtime.
func NewTmuxRuntime() *TmuxRuntime {
	return &TmuxRuntime{
		tmux: tmux.NewTmux(),
	}
}

// NewTmuxRuntimeWithTmux creates a new tmux-based runtime with an existing Tmux instance.
// This is useful for testing or when you need to share a Tmux instance.
func NewTmuxRuntimeWithTmux(t *tmux.Tmux) *TmuxRuntime {
	return &TmuxRuntime{
		tmux: t,
	}
}

// Tmux returns the underlying Tmux instance.
// This allows access to tmux-specific methods not exposed by the AgentRuntime interface.
func (r *TmuxRuntime) Tmux() *tmux.Tmux {
	return r.tmux
}

// Start implements AgentRuntime.Start
func (r *TmuxRuntime) Start(ctx context.Context, opts StartOptions) (*AgentSession, error) {
	// Generate session ID using existing convention
	sessionID := GenerateSessionID(opts)

	// Check if already running
	running, _ := r.tmux.HasSession(sessionID)
	if running {
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	// Create tmux session
	if err := r.tmux.NewSession(sessionID, opts.WorkDir); err != nil {
		return nil, fmt.Errorf("creating tmux session: %w", err)
	}

	// Set environment variables
	r.setEnvironment(sessionID, opts)

	// Apply theming based on role
	r.applyTheme(sessionID, opts)

	// Build and send startup command
	cmd := r.buildStartupCommand(opts)
	if err := r.tmux.SendKeys(sessionID, cmd); err != nil {
		_ = r.tmux.KillSession(sessionID)
		return nil, fmt.Errorf("sending startup command: %w", err)
	}

	// Wait for Claude to be ready (if requested)
	if opts.WaitForReady {
		timeout := opts.ReadyTimeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		if err := r.waitForReady(ctx, sessionID, timeout); err != nil {
			// Non-fatal: session continues
		}

		// Accept bypass permissions warning if present
		_ = r.tmux.AcceptBypassPermissionsWarning(sessionID)
	}

	// Send initial prompt if provided
	if opts.InitialPrompt != "" {
		// Wait for Claude to be fully ready before sending initial prompt
		time.Sleep(2 * time.Second)
		if err := r.tmux.NudgeSession(sessionID, opts.InitialPrompt); err != nil {
			// Non-fatal
		}
	}

	session := &AgentSession{
		SessionID:   sessionID,
		AgentID:     opts.AgentID,
		Role:        opts.Role,
		RigName:     opts.RigName,
		WorkerName:  opts.WorkerName,
		Running:     true,
		StartedAt:   time.Now(),
		RuntimeType: "tmux",
	}

	r.sessions.Store(sessionID, &tmuxSessionState{
		AgentSession: *session,
		workDir:      opts.WorkDir,
	})

	return session, nil
}

// setEnvironment sets tmux environment variables for the session.
func (r *TmuxRuntime) setEnvironment(sessionID string, opts StartOptions) {
	// Set role
	_ = r.tmux.SetEnvironment(sessionID, "GT_ROLE", string(opts.Role))

	// Set rig name if provided
	if opts.RigName != "" {
		_ = r.tmux.SetEnvironment(sessionID, "GT_RIG", opts.RigName)
	}

	// Set worker name based on role
	switch opts.Role {
	case RolePolecat:
		_ = r.tmux.SetEnvironment(sessionID, "GT_POLECAT", opts.WorkerName)
	case RoleCrew:
		_ = r.tmux.SetEnvironment(sessionID, "GT_CREW", opts.WorkerName)
	}

	// Set CLAUDE_CONFIG_DIR for account selection
	if opts.ClaudeConfigDir != "" {
		_ = r.tmux.SetEnvironment(sessionID, "CLAUDE_CONFIG_DIR", opts.ClaudeConfigDir)
	}

	// Set additional custom environment variables
	for key, value := range opts.Environment {
		_ = r.tmux.SetEnvironment(sessionID, key, value)
	}
}

// applyTheme applies Gas Town theming to the session.
func (r *TmuxRuntime) applyTheme(sessionID string, opts StartOptions) {
	// Assign theme based on rig (or default for town-level)
	theme := tmux.AssignTheme(opts.RigName)

	// Configure the full Gas Town session with theme, status, and bindings
	_ = r.tmux.ConfigureGasTownSession(sessionID, theme, opts.RigName, opts.WorkerName, string(opts.Role))

	// Set pane-died hook for crash detection
	_ = r.tmux.SetPaneDiedHook(sessionID, opts.AgentID)
}

// buildStartupCommand constructs the command to start Claude in the session.
func (r *TmuxRuntime) buildStartupCommand(opts StartOptions) string {
	// Use provided command or default to claude
	command := opts.Command
	if command == "" {
		command = "claude"
	}

	// Build the command with arguments
	var parts []string
	parts = append(parts, command)

	// Add provided arguments
	parts = append(parts, opts.Args...)

	// Export environment variables inline for Claude's role detection
	// This is necessary because tmux SetEnvironment only affects new panes
	var exports []string
	exports = append(exports, fmt.Sprintf("GT_ROLE=%s", opts.Role))
	if opts.RigName != "" {
		exports = append(exports, fmt.Sprintf("GT_RIG=%s", opts.RigName))
	}
	if opts.WorkerName != "" {
		switch opts.Role {
		case RolePolecat:
			exports = append(exports, fmt.Sprintf("GT_POLECAT=%s", opts.WorkerName))
		case RoleCrew:
			exports = append(exports, fmt.Sprintf("GT_CREW=%s", opts.WorkerName))
		}
	}

	// Combine exports with command
	exportStr := strings.Join(exports, " ")
	cmdStr := strings.Join(parts, " ")
	return fmt.Sprintf("export %s && %s", exportStr, cmdStr)
}

// waitForReady waits for Claude to be ready to accept input.
func (r *TmuxRuntime) waitForReady(ctx context.Context, sessionID string, timeout time.Duration) error {
	// First wait for a non-shell process to start
	if err := r.tmux.WaitForCommand(sessionID, constants.SupportedShells, timeout/2); err != nil {
		return err
	}

	// Then wait for Claude's prompt
	return r.tmux.WaitForClaudeReady(sessionID, timeout/2)
}

// Stop implements AgentRuntime.Stop
func (r *TmuxRuntime) Stop(ctx context.Context, sessionID string, force bool) error {
	// Check if session exists
	running, err := r.tmux.HasSession(sessionID)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		r.sessions.Delete(sessionID)
		return nil
	}

	// Graceful shutdown (unless forced)
	if !force {
		_ = r.tmux.SendKeysRaw(sessionID, "C-c")
		time.Sleep(100 * time.Millisecond)
	}

	// Kill session
	if err := r.tmux.KillSession(sessionID); err != nil {
		return fmt.Errorf("killing session: %w", err)
	}

	r.sessions.Delete(sessionID)
	return nil
}

// Restart implements AgentRuntime.Restart
func (r *TmuxRuntime) Restart(ctx context.Context, sessionID string, opts StartOptions) (*AgentSession, error) {
	// Stop the existing session
	if err := r.Stop(ctx, sessionID, false); err != nil {
		return nil, fmt.Errorf("stopping session: %w", err)
	}

	// Wait a moment for cleanup
	time.Sleep(500 * time.Millisecond)

	// Start a new session
	return r.Start(ctx, opts)
}

// SendPrompt implements AgentRuntime.SendPrompt
func (r *TmuxRuntime) SendPrompt(ctx context.Context, sessionID string, prompt string) error {
	return r.tmux.NudgeSession(sessionID, prompt)
}

// StreamResponses implements AgentRuntime.StreamResponses
// Note: Tmux doesn't support true streaming, so we poll the pane content.
func (r *TmuxRuntime) StreamResponses(ctx context.Context, sessionID string) (<-chan Response, error) {
	ch := make(chan Response, 100)

	go func() {
		defer close(ch)

		lastContent := ""
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				content, err := r.tmux.CapturePane(sessionID, 50)
				if err != nil {
					ch <- Response{Type: ResponseError, Error: err, Timestamp: time.Now()}
					return
				}

				if content != lastContent {
					// Extract new content (simple diff)
					newContent := extractNewContent(lastContent, content)
					if newContent != "" {
						ch <- Response{
							Type:      ResponseText,
							Content:   newContent,
							Timestamp: time.Now(),
						}
					}
					lastContent = content
				}
			}
		}
	}()

	return ch, nil
}

// extractNewContent finds the difference between old and new content.
func extractNewContent(old, new string) string {
	if old == "" {
		return new
	}

	// Simple approach: find where old content ends in new content
	// This is a basic implementation; could be improved with proper diff
	if strings.HasPrefix(new, old) {
		return strings.TrimPrefix(new, old)
	}

	// Content completely changed, return all new content
	return new
}

// IsRunning implements AgentRuntime.IsRunning
func (r *TmuxRuntime) IsRunning(ctx context.Context, sessionID string) (bool, error) {
	return r.tmux.HasSession(sessionID)
}

// GetStatus implements AgentRuntime.GetStatus
func (r *TmuxRuntime) GetStatus(ctx context.Context, sessionID string) (*AgentStatus, error) {
	running, err := r.tmux.HasSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Get stored session info
	stored, ok := r.sessions.Load(sessionID)
	var session AgentSession
	if ok {
		session = stored.(*tmuxSessionState).AgentSession
	} else {
		session = AgentSession{SessionID: sessionID, Running: running, RuntimeType: "tmux"}
	}
	session.Running = running

	status := &AgentStatus{
		Session: session,
		Health:  HealthUnknown,
	}

	if !running {
		status.Health = HealthUnhealthy
		return status, nil
	}

	// Get tmux-specific info
	tmuxInfo, err := r.tmux.GetSessionInfo(sessionID)
	if err == nil {
		status.TmuxInfo = &TmuxStatus{
			SessionName: tmuxInfo.Name,
			Attached:    tmuxInfo.Attached,
			Windows:     tmuxInfo.Windows,
		}

		// Get pane command if available
		if cmd, err := r.tmux.GetPaneCommand(sessionID); err == nil {
			status.TmuxInfo.PaneCommand = cmd
		}
	}

	// Check if Claude is running
	if r.tmux.IsClaudeRunning(sessionID) {
		status.Health = HealthHealthy
	} else {
		status.Health = HealthDegraded
	}

	// Get activity info
	if tmuxInfo != nil && tmuxInfo.Activity != "" {
		var activityUnix int64
		if _, err := fmt.Sscanf(tmuxInfo.Activity, "%d", &activityUnix); err == nil {
			lastActivity := time.Unix(activityUnix, 0)
			idle := time.Since(lastActivity)

			state := "active"
			if idle > 5*time.Minute {
				state = "stuck"
			} else if idle > 1*time.Minute {
				state = "stale"
			}

			status.Activity = ActivityInfo{
				LastActivity:  lastActivity,
				IdleDuration:  idle,
				ActivityState: state,
			}
		}
	}

	return status, nil
}

// ListSessions implements AgentRuntime.ListSessions
func (r *TmuxRuntime) ListSessions(ctx context.Context, filter SessionFilter) ([]AgentSession, error) {
	sessions, err := r.tmux.ListSessions()
	if err != nil {
		return nil, err
	}

	var result []AgentSession
	for _, sessionName := range sessions {
		// Check stored session info
		if stored, ok := r.sessions.Load(sessionName); ok {
			state := stored.(*tmuxSessionState)
			session := state.AgentSession
			session.Running = true

			// Apply filters
			if filter.RigName != "" && session.RigName != filter.RigName {
				continue
			}
			if filter.Role != "" && session.Role != filter.Role {
				continue
			}
			if filter.AgentID != "" && session.AgentID != filter.AgentID {
				continue
			}
			if filter.Running != nil && session.Running != *filter.Running {
				continue
			}

			result = append(result, session)
		} else {
			// Session exists in tmux but not in our map
			// Try to parse the session name for basic info
			session := AgentSession{
				SessionID:   sessionName,
				Running:     true,
				RuntimeType: "tmux",
			}

			// Parse Gas Town session names: gt-<rig>-<worker> or hq-<role>
			if strings.HasPrefix(sessionName, "hq-") {
				role := strings.TrimPrefix(sessionName, "hq-")
				session.Role = AgentRole(role)
			} else if strings.HasPrefix(sessionName, "gt-") {
				parts := strings.SplitN(strings.TrimPrefix(sessionName, "gt-"), "-", 2)
				if len(parts) >= 1 {
					session.RigName = parts[0]
				}
				if len(parts) >= 2 {
					session.WorkerName = parts[1]
				}
			}

			// Apply filters
			if filter.RigName != "" && session.RigName != filter.RigName {
				continue
			}
			if filter.Role != "" && session.Role != filter.Role {
				continue
			}
			if filter.AgentID != "" && session.AgentID != filter.AgentID {
				continue
			}
			if filter.Running != nil && session.Running != *filter.Running {
				continue
			}

			result = append(result, session)
		}
	}

	return result, nil
}

// GetActivity implements AgentRuntime.GetActivity
func (r *TmuxRuntime) GetActivity(ctx context.Context, sessionID string) (*ActivityInfo, error) {
	status, err := r.GetStatus(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &status.Activity, nil
}

// CaptureOutput implements AgentRuntime.CaptureOutput
func (r *TmuxRuntime) CaptureOutput(ctx context.Context, sessionID string, lines int) (string, error) {
	return r.tmux.CapturePane(sessionID, lines)
}

// Capabilities implements AgentRuntime.Capabilities
func (r *TmuxRuntime) Capabilities() RuntimeCapabilities {
	return RuntimeCapabilities{
		SupportsStreaming:    false, // Polling only
		SupportsToolCalls:    false, // Tools handled by Claude Code
		SupportsSystemPrompt: false, // Uses CLAUDE.md files
		SupportsAttach:       true,  // Can attach to terminal
		SupportsCapture:      true,  // Can capture pane output
		SupportsConcurrency:  0,     // Unlimited (tmux handles)
	}
}

// Close implements AgentRuntime.Close
func (r *TmuxRuntime) Close() error {
	// TmuxRuntime doesn't own the tmux server, so nothing to close
	return nil
}

// Attach attaches the current terminal to a tmux session.
// This is tmux-specific and not part of the AgentRuntime interface.
func (r *TmuxRuntime) Attach(sessionID string) error {
	return r.tmux.AttachSession(sessionID)
}

// EnsureSessionFresh ensures a session is available and healthy.
// If the session exists but is a zombie, it kills the session first.
// This is tmux-specific and not part of the AgentRuntime interface.
func (r *TmuxRuntime) EnsureSessionFresh(name, workDir string) error {
	return r.tmux.EnsureSessionFresh(name, workDir)
}
