package runtime

import (
	"context"
	"os"
	"testing"

	"github.com/steveyegge/gastown/internal/config"
)

func TestNewSDKRuntime_CLIModeWithoutAPIKey(t *testing.T) {
	// Ensure ANTHROPIC_API_KEY is not set for this test
	original := os.Getenv("ANTHROPIC_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	defer func() {
		if original != "" {
			os.Setenv("ANTHROPIC_API_KEY", original)
		}
	}()

	// Without API key, SDK runtime should use CLI mode (spawn claude subprocess)
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{})
	if err != nil {
		t.Errorf("NewSDKRuntime() error = %v, expected nil (CLI mode)", err)
	}
	if rt == nil {
		t.Fatal("NewSDKRuntime() returned nil")
	}
	if !rt.useCLI {
		t.Error("Expected useCLI = true when no API key provided")
	}
	if rt.client != nil {
		t.Error("Expected client = nil in CLI mode")
	}
}

func TestNewSDKRuntime_APIModeWithAPIKey(t *testing.T) {
	// With API key, SDK runtime should use direct API mode
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}
	if rt == nil {
		t.Fatal("NewSDKRuntime() returned nil")
	}
	if rt.useCLI {
		t.Error("Expected useCLI = false when API key provided")
	}
	if rt.client == nil {
		t.Error("Expected client != nil in API mode")
	}
}

func TestNewSDKRuntime_ConfigDefaults(t *testing.T) {
	// Test with a mock API key
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	// Check default concurrency
	caps := rt.Capabilities()
	if caps.SupportsConcurrency != 10 {
		t.Errorf("Default concurrency = %d, want 10", caps.SupportsConcurrency)
	}
}

func TestNewSDKRuntime_CustomConcurrency(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey:                "test-key-for-unit-test",
		MaxConcurrentSessions: 5,
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	caps := rt.Capabilities()
	if caps.SupportsConcurrency != 5 {
		t.Errorf("Custom concurrency = %d, want 5", caps.SupportsConcurrency)
	}
}

func TestSDKRuntimeCapabilities(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	caps := rt.Capabilities()

	// SDK runtime should support streaming and tool calls
	if !caps.SupportsStreaming {
		t.Error("Capabilities().SupportsStreaming should be true")
	}
	if !caps.SupportsToolCalls {
		t.Error("Capabilities().SupportsToolCalls should be true")
	}
	if !caps.SupportsSystemPrompt {
		t.Error("Capabilities().SupportsSystemPrompt should be true")
	}
	if !caps.SupportsCapture {
		t.Error("Capabilities().SupportsCapture should be true")
	}

	// SDK runtime should NOT support attach
	if caps.SupportsAttach {
		t.Error("Capabilities().SupportsAttach should be false")
	}
}

func TestSDKRuntime_BuildSystemPrompt(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	tests := []struct {
		name     string
		opts     StartOptions
		contains string
	}{
		{
			name: "custom prompt",
			opts: StartOptions{
				SystemPrompt: "Custom system prompt",
			},
			contains: "Custom system prompt",
		},
		{
			name: "mayor role",
			opts: StartOptions{
				Role: RoleMayor,
			},
			contains: "Mayor",
		},
		{
			name: "deacon role",
			opts: StartOptions{
				Role: RoleDeacon,
			},
			contains: "Deacon",
		},
		{
			name: "witness role",
			opts: StartOptions{
				Role:    RoleWitness,
				RigName: "gastown",
			},
			contains: "Witness",
		},
		{
			name: "polecat role",
			opts: StartOptions{
				Role:       RolePolecat,
				RigName:    "gastown",
				WorkerName: "toast",
			},
			contains: "polecat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := rt.buildSystemPrompt(tt.opts)
			if !containsString(prompt, tt.contains) {
				t.Errorf("buildSystemPrompt() = %q, want to contain %q", prompt, tt.contains)
			}
		})
	}
}

func TestSDKRuntime_RegisterTool(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	tool := ToolConfig{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{
					"type": "string",
				},
			},
		},
		Handler: func(ctx context.Context, input map[string]any) (any, error) {
			return "result", nil
		},
	}

	rt.RegisterTool(tool)

	tools := rt.ListTools()
	if len(tools) != 1 {
		t.Errorf("ListTools() returned %d tools, want 1", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Errorf("Tool name = %q, want %q", tools[0].Name, "test_tool")
	}
}

func TestSDKRuntime_UnregisterTool(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	rt.RegisterTool(ToolConfig{Name: "tool1"})
	rt.RegisterTool(ToolConfig{Name: "tool2"})

	if len(rt.ListTools()) != 2 {
		t.Error("Expected 2 tools after registration")
	}

	rt.UnregisterTool("tool1")

	tools := rt.ListTools()
	if len(tools) != 1 {
		t.Errorf("ListTools() returned %d tools after unregister, want 1", len(tools))
	}
	if tools[0].Name != "tool2" {
		t.Errorf("Remaining tool name = %q, want %q", tools[0].Name, "tool2")
	}
}

func TestSDKRuntime_ExecuteTool(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	// Test unknown tool
	result := rt.executeTool(context.Background(), &ToolCall{
		ID:   "call1",
		Name: "unknown_tool",
	})
	if result.Error == "" {
		t.Error("Expected error for unknown tool")
	}

	// Register a tool
	rt.RegisterTool(ToolConfig{
		Name: "echo",
		Handler: func(ctx context.Context, input map[string]any) (any, error) {
			return input["message"], nil
		},
	})

	// Test registered tool
	result = rt.executeTool(context.Background(), &ToolCall{
		ID:    "call2",
		Name:  "echo",
		Input: map[string]any{"message": "hello"},
	})
	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
	if result.Output != "hello" {
		t.Errorf("Tool output = %v, want %q", result.Output, "hello")
	}
}

func TestSDKRuntime_Close(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	err = rt.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestSDKRuntime_GetStatus_NotFound(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	status, err := rt.GetStatus(context.Background(), "nonexistent")
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}
	if status.Session.Running {
		t.Error("Expected Running = false for nonexistent session")
	}
	if status.Health != HealthUnknown {
		t.Errorf("Health = %v, want %v", status.Health, HealthUnknown)
	}
}

func TestSDKRuntime_IsRunning_NotFound(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	running, err := rt.IsRunning(context.Background(), "nonexistent")
	if err != nil {
		t.Errorf("IsRunning() error = %v", err)
	}
	if running {
		t.Error("Expected running = false for nonexistent session")
	}
}

func TestSDKRuntime_SendPrompt_NotFound(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	err = rt.SendPrompt(context.Background(), "nonexistent", "hello")
	if err == nil {
		t.Error("Expected error for nonexistent session")
	}
}

func TestSDKRuntime_StreamResponses_NotFound(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	_, err = rt.StreamResponses(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent session")
	}
}

func TestSDKRuntime_CaptureOutput_NotFound(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	_, err = rt.CaptureOutput(context.Background(), "nonexistent", 10)
	if err == nil {
		t.Error("Expected error for nonexistent session")
	}
}

func TestSDKRuntime_ListSessions_Empty(t *testing.T) {
	rt, err := NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey: "test-key-for-unit-test",
	})
	if err != nil {
		t.Fatalf("NewSDKRuntime() error = %v", err)
	}

	sessions, err := rt.ListSessions(context.Background(), SessionFilter{})
	if err != nil {
		t.Errorf("ListSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("ListSessions() returned %d sessions, want 0", len(sessions))
	}
}
