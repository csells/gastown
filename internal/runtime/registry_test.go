package runtime

import (
	"context"
	"testing"
)

// mockRuntime is a mock implementation for testing.
type mockRuntime struct {
	name string
}

func (m *mockRuntime) Start(ctx context.Context, opts StartOptions) (*AgentSession, error) {
	return &AgentSession{SessionID: "mock-" + opts.WorkerName}, nil
}

func (m *mockRuntime) Stop(ctx context.Context, sessionID string, force bool) error {
	return nil
}

func (m *mockRuntime) Restart(ctx context.Context, sessionID string, opts StartOptions) (*AgentSession, error) {
	return m.Start(ctx, opts)
}

func (m *mockRuntime) SendPrompt(ctx context.Context, sessionID string, prompt string) error {
	return nil
}

func (m *mockRuntime) StreamResponses(ctx context.Context, sessionID string) (<-chan Response, error) {
	ch := make(chan Response)
	close(ch)
	return ch, nil
}

func (m *mockRuntime) IsRunning(ctx context.Context, sessionID string) (bool, error) {
	return true, nil
}

func (m *mockRuntime) GetStatus(ctx context.Context, sessionID string) (*AgentStatus, error) {
	return &AgentStatus{}, nil
}

func (m *mockRuntime) ListSessions(ctx context.Context, filter SessionFilter) ([]AgentSession, error) {
	return nil, nil
}

func (m *mockRuntime) GetActivity(ctx context.Context, sessionID string) (*ActivityInfo, error) {
	return &ActivityInfo{}, nil
}

func (m *mockRuntime) CaptureOutput(ctx context.Context, sessionID string, lines int) (string, error) {
	return "", nil
}

func (m *mockRuntime) Capabilities() RuntimeCapabilities {
	return RuntimeCapabilities{}
}

func (m *mockRuntime) Close() error {
	return nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()

	mock := &mockRuntime{name: "test"}
	reg.Register("test", mock)

	got, err := reg.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != mock {
		t.Errorf("Get() returned different runtime")
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Error("Get() expected error for nonexistent runtime")
	}
}

func TestRegistrySetActive(t *testing.T) {
	reg := NewRegistry()

	mock := &mockRuntime{name: "test"}
	reg.Register(RuntimeTmux, mock)

	if err := reg.SetActive(RuntimeTmux); err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	if reg.ActiveName() != RuntimeTmux {
		t.Errorf("ActiveName() = %v, want %v", reg.ActiveName(), RuntimeTmux)
	}

	if reg.Active() != mock {
		t.Error("Active() returned different runtime")
	}
}

func TestRegistrySetActiveNotFound(t *testing.T) {
	reg := NewRegistry()

	err := reg.SetActive("nonexistent")
	if err == nil {
		t.Error("SetActive() expected error for nonexistent runtime")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()

	mock1 := &mockRuntime{name: "test1"}
	mock2 := &mockRuntime{name: "test2"}
	reg.Register(RuntimeTmux, mock1)
	reg.Register(RuntimeSDK, mock2)

	names := reg.List()
	if len(names) != 2 {
		t.Errorf("List() returned %d names, want 2", len(names))
	}

	// Check both names are present (order not guaranteed)
	found := make(map[RuntimeName]bool)
	for _, name := range names {
		found[name] = true
	}
	if !found[RuntimeTmux] || !found[RuntimeSDK] {
		t.Errorf("List() missing expected names: got %v", names)
	}
}

func TestRegistryInitialize(t *testing.T) {
	reg := NewRegistry()
	reg.Initialize()

	// Should have tmux registered
	_, err := reg.Get(RuntimeTmux)
	if err != nil {
		t.Errorf("Initialize() should register tmux runtime: %v", err)
	}
}

func TestRegistryCloseAll(t *testing.T) {
	reg := NewRegistry()

	mock1 := &mockRuntime{name: "test1"}
	mock2 := &mockRuntime{name: "test2"}
	reg.Register(RuntimeTmux, mock1)
	reg.Register(RuntimeSDK, mock2)

	err := reg.CloseAll()
	if err != nil {
		t.Errorf("CloseAll() error = %v", err)
	}
}

func TestRuntimeNameConstants(t *testing.T) {
	// Verify runtime name constants have expected string values
	if RuntimeTmux != "tmux" {
		t.Errorf("RuntimeTmux = %q, want %q", RuntimeTmux, "tmux")
	}
	if RuntimeSDK != "sdk" {
		t.Errorf("RuntimeSDK = %q, want %q", RuntimeSDK, "sdk")
	}
}
