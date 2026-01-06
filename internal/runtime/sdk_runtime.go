package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/steveyegge/gastown/internal/config"
)

// SDKRuntime implements AgentRuntime using either:
// 1. Direct Anthropic API calls (when API key is provided)
// 2. Claude Code CLI subprocess (when no API key - uses user's existing OAuth/auth)
//
// This enables headless operation without terminal dependencies.
type SDKRuntime struct {
	config   *config.SDKRuntimeConfig
	client   *anthropic.Client // nil when using CLI mode
	useCLI   bool              // true when spawning claude CLI subprocess
	sessions sync.Map          // sessionID -> *sdkSession

	// Concurrency control
	semaphore chan struct{}

	// Tool registry
	tools   map[string]ToolConfig
	toolsMu sync.RWMutex
}

// sdkSession tracks a running SDK session.
type sdkSession struct {
	AgentSession

	// SDK state (API mode)
	conversation []anthropic.MessageParam
	systemPrompt string

	// CLI mode state
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	// Control
	ctx    context.Context
	cancel context.CancelFunc

	// Communication
	promptCh   chan string
	responseCh chan Response

	// State
	mu         sync.Mutex
	tokenCount int
	turnCount  int
	lastPrompt time.Time
	lastResp   time.Time

	// Runtime reference for API calls
	runtime *SDKRuntime
}

// NewSDKRuntime creates a new SDK-based runtime.
// By default, it spawns Claude Code CLI subprocesses which use the user's
// existing OAuth/auth configuration. If an API key is explicitly provided
// in the config, it uses direct Anthropic API calls instead.
//
// Note: This does NOT read ANTHROPIC_API_KEY from the environment to avoid
// overriding the user's preferred auth method (e.g., OAuth via Claude Max).
func NewSDKRuntime(cfg *config.SDKRuntimeConfig) (*SDKRuntime, error) {
	if cfg == nil {
		cfg = &config.SDKRuntimeConfig{}
	}

	maxConcurrent := cfg.MaxConcurrentSessions
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	runtime := &SDKRuntime{
		config:    cfg,
		semaphore: make(chan struct{}, maxConcurrent),
		tools:     make(map[string]ToolConfig),
	}

	// Only use direct API mode if API key is EXPLICITLY provided in config
	// Do NOT check environment variables - that would override OAuth auth
	if cfg.APIKey != "" {
		client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
		runtime.client = &client
		runtime.useCLI = false
	} else {
		// CLI mode - spawn claude subprocess (uses user's existing OAuth/auth)
		runtime.useCLI = true
	}

	return runtime, nil
}

// Start implements AgentRuntime.Start
func (r *SDKRuntime) Start(ctx context.Context, opts StartOptions) (*AgentSession, error) {
	// Acquire semaphore slot
	select {
	case r.semaphore <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("max concurrent sessions reached (%d)", cap(r.semaphore))
	}

	sessionID := GenerateSessionID(opts)

	// Check for existing session
	if _, exists := r.sessions.Load(sessionID); exists {
		<-r.semaphore // Release slot
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	// Build system prompt
	systemPrompt := r.buildSystemPrompt(opts)

	// Create session context
	sessionCtx, cancel := context.WithCancel(context.Background())

	session := &sdkSession{
		AgentSession: AgentSession{
			SessionID:   sessionID,
			AgentID:     opts.AgentID,
			Role:        opts.Role,
			RigName:     opts.RigName,
			WorkerName:  opts.WorkerName,
			Running:     true,
			StartedAt:   time.Now(),
			RuntimeType: "sdk",
		},
		systemPrompt: systemPrompt,
		conversation: make([]anthropic.MessageParam, 0),
		ctx:          sessionCtx,
		cancel:       cancel,
		promptCh:     make(chan string, 10),
		responseCh:   make(chan Response, 100),
		runtime:      r,
	}

	r.sessions.Store(sessionID, session)

	// Start the session loop in background
	if r.useCLI {
		go session.runCLI()
	} else {
		go session.run()
	}

	// Send initial prompt if provided
	if opts.InitialPrompt != "" {
		if err := r.SendPrompt(ctx, sessionID, opts.InitialPrompt); err != nil {
			// Non-fatal: session continues
		}
	}

	return &session.AgentSession, nil
}

// buildSystemPrompt constructs the system prompt for the session.
func (r *SDKRuntime) buildSystemPrompt(opts StartOptions) string {
	if opts.SystemPrompt != "" {
		return opts.SystemPrompt
	}

	// Build a default system prompt based on role
	var prompt string
	switch opts.Role {
	case RoleMayor:
		prompt = "You are the Mayor, the town coordinator for Gas Town. You manage rigs, coordinate work assignments, and oversee the deacon and witnesses."
	case RoleDeacon:
		prompt = "You are the Deacon, the health monitor for Gas Town. You check on agents, detect stuck workers, and ensure the town runs smoothly."
	case RoleWitness:
		prompt = fmt.Sprintf("You are a Witness for rig %s. You monitor polecats, spawn new workers for incoming issues, and report status.", opts.RigName)
	case RoleRefinery:
		prompt = fmt.Sprintf("You are the Refinery for rig %s. You process the merge queue, handle conflicts, and ensure code gets merged cleanly.", opts.RigName)
	case RoleCrew:
		prompt = fmt.Sprintf("You are crew member %s working on rig %s. You are a human-supervised worker with full access to the codebase.", opts.WorkerName, opts.RigName)
	case RolePolecat:
		prompt = fmt.Sprintf("You are polecat %s working on rig %s. You are an autonomous worker that handles issues and creates pull requests.", opts.WorkerName, opts.RigName)
	default:
		prompt = "You are a Gas Town agent."
	}

	return prompt
}

// run is the main loop for an SDK session (API mode).
func (s *sdkSession) run() {
	defer func() {
		close(s.responseCh)
		s.mu.Lock()
		s.Running = false
		s.mu.Unlock()
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case prompt, ok := <-s.promptCh:
			if !ok {
				return
			}
			s.handlePrompt(prompt)
		}
	}
}

// runCLI is the main loop for a CLI-mode session.
// It spawns `claude` as a subprocess and communicates via stdin/stdout.
func (s *sdkSession) runCLI() {
	defer func() {
		close(s.responseCh)
		s.mu.Lock()
		s.Running = false
		s.mu.Unlock()
	}()

	// Start claude CLI with print mode for non-interactive output
	// Using --output-format stream-json for streaming JSON responses
	args := []string{"--output-format", "stream-json", "--verbose"}

	// Add system prompt if provided
	if s.systemPrompt != "" {
		args = append(args, "--system-prompt", s.systemPrompt)
	}

	s.cmd = exec.CommandContext(s.ctx, "claude", args...)

	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		s.responseCh <- Response{
			Type:      ResponseError,
			Error:     fmt.Errorf("failed to get stdin pipe: %w", err),
			Timestamp: time.Now(),
		}
		return
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		s.responseCh <- Response{
			Type:      ResponseError,
			Error:     fmt.Errorf("failed to get stdout pipe: %w", err),
			Timestamp: time.Now(),
		}
		return
	}

	if err := s.cmd.Start(); err != nil {
		s.responseCh <- Response{
			Type:      ResponseError,
			Error:     fmt.Errorf("failed to start claude: %w", err),
			Timestamp: time.Now(),
		}
		return
	}

	// Read stdout in background
	go s.readCLIOutput()

	// Process prompts
	for {
		select {
		case <-s.ctx.Done():
			s.stdin.Close()
			s.cmd.Wait()
			return
		case prompt, ok := <-s.promptCh:
			if !ok {
				s.stdin.Close()
				s.cmd.Wait()
				return
			}
			s.handleCLIPrompt(prompt)
		}
	}
}

// readCLIOutput reads streaming JSON output from claude CLI.
func (s *sdkSession) readCLIOutput() {
	scanner := bufio.NewScanner(s.stdout)
	// Increase buffer size for potentially large responses
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse streaming JSON response
		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
			Error   string `json:"error,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not JSON, treat as raw text
			s.responseCh <- Response{
				Type:      ResponseText,
				Content:   line,
				Timestamp: time.Now(),
			}
			continue
		}

		switch msg.Type {
		case "text", "content":
			s.responseCh <- Response{
				Type:      ResponseText,
				Content:   msg.Content,
				Timestamp: time.Now(),
			}
		case "error":
			s.responseCh <- Response{
				Type:      ResponseError,
				Error:     fmt.Errorf("%s", msg.Error),
				Timestamp: time.Now(),
			}
		case "done", "complete", "result":
			s.mu.Lock()
			s.lastResp = time.Now()
			s.mu.Unlock()
			s.responseCh <- Response{
				Type:      ResponseComplete,
				Timestamp: time.Now(),
			}
		}
	}
}

// handleCLIPrompt sends a prompt to the claude CLI subprocess.
func (s *sdkSession) handleCLIPrompt(prompt string) {
	s.mu.Lock()
	s.lastPrompt = time.Now()
	s.turnCount++
	s.mu.Unlock()

	// Send prompt as a line to stdin
	_, err := fmt.Fprintf(s.stdin, "%s\n", prompt)
	if err != nil {
		s.responseCh <- Response{
			Type:      ResponseError,
			Error:     fmt.Errorf("failed to send prompt: %w", err),
			Timestamp: time.Now(),
		}
	}
}

// handlePrompt processes a prompt and generates a response (API mode).
func (s *sdkSession) handlePrompt(prompt string) {
	s.mu.Lock()
	s.lastPrompt = time.Now()
	s.turnCount++
	s.mu.Unlock()

	// Add user message to conversation
	s.conversation = append(s.conversation, anthropic.NewUserMessage(
		anthropic.NewTextBlock(prompt),
	))

	// Get model from config
	model := s.runtime.config.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	// Get max tokens from config
	maxTokens := int64(s.runtime.config.MaxTokens)
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	// Build tools for the request
	tools := s.runtime.buildToolParams()

	// Create message request
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: maxTokens,
		Messages:  s.conversation,
	}

	// Add system prompt if set
	if s.systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{
				Text: s.systemPrompt,
				Type: "text",
			},
		}
	}

	// Add tools if available
	if len(tools) > 0 {
		params.Tools = tools
	}

	// Call the API
	response, err := (*s.runtime.client).Messages.New(s.ctx, params)
	if err != nil {
		s.responseCh <- Response{
			Type:      ResponseError,
			Error:     err,
			Timestamp: time.Now(),
		}
		return
	}

	s.mu.Lock()
	s.lastResp = time.Now()
	if response.Usage.InputTokens > 0 || response.Usage.OutputTokens > 0 {
		s.tokenCount += int(response.Usage.InputTokens + response.Usage.OutputTokens)
	}
	s.mu.Unlock()

	// Process response content
	var assistantContent []anthropic.ContentBlockParamUnion
	for _, block := range response.Content {
		switch block.Type {
		case "text":
			s.responseCh <- Response{
				Type:      ResponseText,
				Content:   block.Text,
				Timestamp: time.Now(),
			}
			assistantContent = append(assistantContent, anthropic.NewTextBlock(block.Text))

		case "tool_use":
			// Convert input to map
			inputMap := make(map[string]any)
			if err := json.Unmarshal(block.Input, &inputMap); err != nil {
				inputMap = map[string]any{"raw": string(block.Input)}
			}

			toolCall := &ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: inputMap,
			}
			s.responseCh <- Response{
				Type:      ResponseToolCall,
				ToolCall:  toolCall,
				Timestamp: time.Now(),
			}
			assistantContent = append(assistantContent, anthropic.NewToolUseBlock(block.ID, inputMap, block.Name))

			// Execute tool and send result
			result := s.runtime.executeTool(s.ctx, toolCall)
			s.responseCh <- Response{
				Type:       ResponseToolResult,
				ToolResult: result,
				Timestamp:  time.Now(),
			}
		}
	}

	// Add assistant message to conversation
	if len(assistantContent) > 0 {
		s.conversation = append(s.conversation, anthropic.NewAssistantMessage(assistantContent...))
	}

	// Check if we need to continue (tool use requires follow-up)
	if response.StopReason == "tool_use" {
		// Add tool results and continue conversation
		s.handleToolResults()
	} else {
		s.responseCh <- Response{
			Type:      ResponseComplete,
			Timestamp: time.Now(),
		}
	}
}

// handleToolResults processes tool results and continues the conversation.
func (s *sdkSession) handleToolResults() {
	// Collect pending tool results from the last assistant message
	var toolResults []anthropic.ContentBlockParamUnion

	// Find tool use blocks in the last assistant message and execute them
	if len(s.conversation) > 0 {
		lastMsg := s.conversation[len(s.conversation)-1]
		for _, block := range lastMsg.Content {
			// The block is ContentBlockParamUnion - check its underlying type
			// For tool use blocks added via NewToolUseBlock, we need to extract the ID
			blockJSON, _ := json.Marshal(block)
			var blockData struct {
				Type  string         `json:"type"`
				ID    string         `json:"id"`
				Name  string         `json:"name"`
				Input map[string]any `json:"input"`
			}
			if err := json.Unmarshal(blockJSON, &blockData); err != nil {
				continue
			}

			if blockData.Type == "tool_use" && blockData.ID != "" {
				toolCall := &ToolCall{
					ID:    blockData.ID,
					Name:  blockData.Name,
					Input: blockData.Input,
				}
				result := s.runtime.executeTool(s.ctx, toolCall)

				// Create tool result block
				resultContent := fmt.Sprintf("%v", result.Output)
				if result.Error != "" {
					resultContent = fmt.Sprintf("Error: %s", result.Error)
				}
				toolResults = append(toolResults, anthropic.NewToolResultBlock(
					blockData.ID,
					resultContent,
					result.Error != "",
				))
			}
		}
	}

	if len(toolResults) == 0 {
		return
	}

	// Add tool results as user message
	s.conversation = append(s.conversation, anthropic.NewUserMessage(toolResults...))

	// Continue the conversation
	model := s.runtime.config.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	maxTokens := int64(s.runtime.config.MaxTokens)
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: maxTokens,
		Messages:  s.conversation,
	}
	if s.systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{
				Text: s.systemPrompt,
				Type: "text",
			},
		}
	}
	tools := s.runtime.buildToolParams()
	if len(tools) > 0 {
		params.Tools = tools
	}

	response, err := (*s.runtime.client).Messages.New(s.ctx, params)
	if err != nil {
		s.responseCh <- Response{
			Type:      ResponseError,
			Error:     err,
			Timestamp: time.Now(),
		}
		return
	}

	// Process response (recursive tool handling)
	var assistantContent []anthropic.ContentBlockParamUnion
	hasToolUse := false

	for _, block := range response.Content {
		switch block.Type {
		case "text":
			s.responseCh <- Response{
				Type:      ResponseText,
				Content:   block.Text,
				Timestamp: time.Now(),
			}
			assistantContent = append(assistantContent, anthropic.NewTextBlock(block.Text))

		case "tool_use":
			hasToolUse = true
			inputMap := make(map[string]any)
			if err := json.Unmarshal(block.Input, &inputMap); err != nil {
				inputMap = map[string]any{"raw": string(block.Input)}
			}
			toolCall := &ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: inputMap,
			}
			s.responseCh <- Response{
				Type:      ResponseToolCall,
				ToolCall:  toolCall,
				Timestamp: time.Now(),
			}
			assistantContent = append(assistantContent, anthropic.NewToolUseBlock(block.ID, inputMap, block.Name))
		}
	}

	if len(assistantContent) > 0 {
		s.conversation = append(s.conversation, anthropic.NewAssistantMessage(assistantContent...))
	}

	if hasToolUse && response.StopReason == "tool_use" {
		s.handleToolResults() // Recursive tool handling
	} else {
		s.responseCh <- Response{
			Type:      ResponseComplete,
			Timestamp: time.Now(),
		}
	}
}

// buildToolParams converts registered tools to API parameters.
func (r *SDKRuntime) buildToolParams() []anthropic.ToolUnionParam {
	r.toolsMu.RLock()
	defer r.toolsMu.RUnlock()

	if len(r.tools) == 0 {
		return nil
	}

	params := make([]anthropic.ToolUnionParam, 0, len(r.tools))
	for _, tool := range r.tools {
		// Convert input schema to the expected format
		inputSchema := anthropic.ToolInputSchemaParam{
			Properties: tool.InputSchema,
		}

		params = append(params, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: inputSchema,
			},
		})
	}
	return params
}

// executeTool runs a tool and returns the result.
func (r *SDKRuntime) executeTool(ctx context.Context, call *ToolCall) *ToolResult {
	r.toolsMu.RLock()
	tool, ok := r.tools[call.Name]
	r.toolsMu.RUnlock()

	if !ok {
		return &ToolResult{
			CallID: call.ID,
			Error:  fmt.Sprintf("unknown tool: %s", call.Name),
		}
	}

	if tool.Handler == nil {
		return &ToolResult{
			CallID: call.ID,
			Error:  fmt.Sprintf("tool %s has no handler", call.Name),
		}
	}

	output, err := tool.Handler(ctx, call.Input)
	if err != nil {
		return &ToolResult{
			CallID: call.ID,
			Error:  err.Error(),
		}
	}

	return &ToolResult{
		CallID: call.ID,
		Output: output,
	}
}

// Stop implements AgentRuntime.Stop
func (r *SDKRuntime) Stop(ctx context.Context, sessionID string, force bool) error {
	stored, ok := r.sessions.Load(sessionID)
	if !ok {
		return nil
	}

	session := stored.(*sdkSession)

	// For CLI mode, close stdin to signal the subprocess to exit
	if session.stdin != nil {
		session.stdin.Close()
	}

	session.cancel() // Cancel the context to stop the run loop
	close(session.promptCh)

	// For CLI mode, wait for the process to exit
	if session.cmd != nil {
		if force {
			session.cmd.Process.Kill()
		}
		session.cmd.Wait()
	}

	r.sessions.Delete(sessionID)
	<-r.semaphore // Release semaphore slot

	return nil
}

// Restart implements AgentRuntime.Restart
func (r *SDKRuntime) Restart(ctx context.Context, sessionID string, opts StartOptions) (*AgentSession, error) {
	if err := r.Stop(ctx, sessionID, false); err != nil {
		return nil, fmt.Errorf("stopping session: %w", err)
	}
	return r.Start(ctx, opts)
}

// SendPrompt implements AgentRuntime.SendPrompt
func (r *SDKRuntime) SendPrompt(ctx context.Context, sessionID string, prompt string) error {
	stored, ok := r.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session := stored.(*sdkSession)

	select {
	case session.promptCh <- prompt:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-session.ctx.Done():
		return fmt.Errorf("session closed")
	}
}

// StreamResponses implements AgentRuntime.StreamResponses
func (r *SDKRuntime) StreamResponses(ctx context.Context, sessionID string) (<-chan Response, error) {
	stored, ok := r.sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session := stored.(*sdkSession)

	// Create a new channel that forwards responses
	ch := make(chan Response, 100)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-session.responseCh:
				if !ok {
					return
				}
				select {
				case ch <- resp:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// IsRunning implements AgentRuntime.IsRunning
func (r *SDKRuntime) IsRunning(ctx context.Context, sessionID string) (bool, error) {
	stored, ok := r.sessions.Load(sessionID)
	if !ok {
		return false, nil
	}

	session := stored.(*sdkSession)
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.Running, nil
}

// GetStatus implements AgentRuntime.GetStatus
func (r *SDKRuntime) GetStatus(ctx context.Context, sessionID string) (*AgentStatus, error) {
	stored, ok := r.sessions.Load(sessionID)
	if !ok {
		return &AgentStatus{
			Session: AgentSession{SessionID: sessionID, Running: false, RuntimeType: "sdk"},
			Health:  HealthUnknown,
		}, nil
	}

	session := stored.(*sdkSession)
	session.mu.Lock()
	defer session.mu.Unlock()

	health := HealthHealthy
	if !session.Running {
		health = HealthUnhealthy
	}

	activityState := "active"
	idleDuration := time.Since(session.lastResp)
	if session.lastResp.IsZero() {
		idleDuration = time.Since(session.StartedAt)
	}
	if idleDuration > 5*time.Minute {
		activityState = "stuck"
	} else if idleDuration > 1*time.Minute {
		activityState = "stale"
	}

	return &AgentStatus{
		Session: session.AgentSession,
		Health:  health,
		Activity: ActivityInfo{
			LastActivity:  session.lastResp,
			IdleDuration:  idleDuration,
			ActivityState: activityState,
			LastPrompt:    session.lastPrompt,
			LastResponse:  session.lastResp,
		},
		SDKInfo: &SDKStatus{
			ConversationID: sessionID,
			TokensUsed:     session.tokenCount,
			TurnCount:      session.turnCount,
		},
	}, nil
}

// ListSessions implements AgentRuntime.ListSessions
func (r *SDKRuntime) ListSessions(ctx context.Context, filter SessionFilter) ([]AgentSession, error) {
	var result []AgentSession

	r.sessions.Range(func(key, value any) bool {
		session := value.(*sdkSession)
		session.mu.Lock()
		defer session.mu.Unlock()

		// Apply filters
		if filter.RigName != "" && session.RigName != filter.RigName {
			return true
		}
		if filter.Role != "" && session.Role != filter.Role {
			return true
		}
		if filter.AgentID != "" && session.AgentID != filter.AgentID {
			return true
		}
		if filter.Running != nil && session.Running != *filter.Running {
			return true
		}

		result = append(result, session.AgentSession)
		return true
	})

	return result, nil
}

// GetActivity implements AgentRuntime.GetActivity
func (r *SDKRuntime) GetActivity(ctx context.Context, sessionID string) (*ActivityInfo, error) {
	status, err := r.GetStatus(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &status.Activity, nil
}

// CaptureOutput implements AgentRuntime.CaptureOutput
// SDK runtime doesn't have terminal output, so this returns conversation history.
func (r *SDKRuntime) CaptureOutput(ctx context.Context, sessionID string, lines int) (string, error) {
	stored, ok := r.sessions.Load(sessionID)
	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	session := stored.(*sdkSession)
	session.mu.Lock()
	defer session.mu.Unlock()

	// Return last N conversation turns as text
	var output string
	start := 0
	if lines > 0 && len(session.conversation) > lines {
		start = len(session.conversation) - lines
	}

	for i := start; i < len(session.conversation); i++ {
		msg := session.conversation[i]
		output += fmt.Sprintf("[%s]\n", msg.Role)
		for _, block := range msg.Content {
			// Marshal block to check its type
			blockJSON, _ := json.Marshal(block)
			var blockData struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if json.Unmarshal(blockJSON, &blockData) == nil && blockData.Type == "text" {
				output += blockData.Text + "\n"
			}
		}
		output += "\n"
	}

	return output, nil
}

// Capabilities implements AgentRuntime.Capabilities
func (r *SDKRuntime) Capabilities() RuntimeCapabilities {
	return RuntimeCapabilities{
		SupportsStreaming:    true,  // Real streaming support
		SupportsToolCalls:    true,  // Native tool support
		SupportsSystemPrompt: true,  // Direct system prompt
		SupportsAttach:       false, // No terminal
		SupportsCapture:      true,  // Conversation history
		SupportsConcurrency:  cap(r.semaphore),
	}
}

// Close implements AgentRuntime.Close
func (r *SDKRuntime) Close() error {
	// Stop all sessions
	r.sessions.Range(func(key, value any) bool {
		sessionID := key.(string)
		_ = r.Stop(context.Background(), sessionID, true)
		return true
	})
	return nil
}

// RegisterTool adds a tool to the SDK runtime.
// Tools are available to all sessions managed by this runtime.
func (r *SDKRuntime) RegisterTool(tool ToolConfig) {
	r.toolsMu.Lock()
	defer r.toolsMu.Unlock()
	r.tools[tool.Name] = tool
}

// UnregisterTool removes a tool from the SDK runtime.
func (r *SDKRuntime) UnregisterTool(name string) {
	r.toolsMu.Lock()
	defer r.toolsMu.Unlock()
	delete(r.tools, name)
}

// ListTools returns all registered tools.
func (r *SDKRuntime) ListTools() []ToolConfig {
	r.toolsMu.RLock()
	defer r.toolsMu.RUnlock()

	tools := make([]ToolConfig, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
