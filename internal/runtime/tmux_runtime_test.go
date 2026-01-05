package runtime

import (
	"testing"
)

func TestNewTmuxRuntime(t *testing.T) {
	rt := NewTmuxRuntime()
	if rt == nil {
		t.Fatal("NewTmuxRuntime() returned nil")
	}
	if rt.tmux == nil {
		t.Error("NewTmuxRuntime() tmux field is nil")
	}
}

func TestTmuxRuntimeCapabilities(t *testing.T) {
	rt := NewTmuxRuntime()
	caps := rt.Capabilities()

	// Tmux runtime should support attach and capture
	if !caps.SupportsAttach {
		t.Error("Capabilities().SupportsAttach should be true")
	}
	if !caps.SupportsCapture {
		t.Error("Capabilities().SupportsCapture should be true")
	}

	// Tmux runtime should not support streaming or tool calls
	if caps.SupportsStreaming {
		t.Error("Capabilities().SupportsStreaming should be false")
	}
	if caps.SupportsToolCalls {
		t.Error("Capabilities().SupportsToolCalls should be false")
	}
	if caps.SupportsSystemPrompt {
		t.Error("Capabilities().SupportsSystemPrompt should be false")
	}

	// Concurrency should be unlimited (0)
	if caps.SupportsConcurrency != 0 {
		t.Errorf("Capabilities().SupportsConcurrency = %d, want 0", caps.SupportsConcurrency)
	}
}

func TestTmuxRuntimeTmuxAccessor(t *testing.T) {
	rt := NewTmuxRuntime()
	tmux := rt.Tmux()

	if tmux == nil {
		t.Error("Tmux() returned nil")
	}
	if tmux != rt.tmux {
		t.Error("Tmux() should return the internal tmux instance")
	}
}

func TestTmuxRuntimeClose(t *testing.T) {
	rt := NewTmuxRuntime()
	err := rt.Close()

	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestTmuxRuntimeBuildStartupCommand(t *testing.T) {
	rt := NewTmuxRuntime()

	tests := []struct {
		name     string
		opts     StartOptions
		contains []string
	}{
		{
			name: "default command",
			opts: StartOptions{
				Role: RolePolecat,
			},
			contains: []string{"claude", "GT_ROLE=polecat"},
		},
		{
			name: "custom command",
			opts: StartOptions{
				Role:    RolePolecat,
				Command: "aider",
			},
			contains: []string{"aider"},
		},
		{
			name: "with rig name",
			opts: StartOptions{
				Role:    RolePolecat,
				RigName: "gastown",
			},
			contains: []string{"GT_RIG=gastown"},
		},
		{
			name: "polecat with worker name",
			opts: StartOptions{
				Role:       RolePolecat,
				RigName:    "gastown",
				WorkerName: "toast",
			},
			contains: []string{"GT_POLECAT=toast"},
		},
		{
			name: "crew with worker name",
			opts: StartOptions{
				Role:       RoleCrew,
				RigName:    "gastown",
				WorkerName: "max",
			},
			contains: []string{"GT_CREW=max"},
		},
		{
			name: "with args",
			opts: StartOptions{
				Role:    RolePolecat,
				Command: "claude",
				Args:    []string{"--dangerously-skip-permissions", "--verbose"},
			},
			contains: []string{"--dangerously-skip-permissions", "--verbose"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rt.buildStartupCommand(tt.opts)
			for _, s := range tt.contains {
				if !containsString(cmd, s) {
					t.Errorf("buildStartupCommand() = %q, want to contain %q", cmd, s)
				}
			}
		})
	}
}

func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func TestExtractNewContent(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected string
	}{
		{
			name:     "empty old",
			old:      "",
			new:      "new content",
			expected: "new content",
		},
		{
			name:     "new content appended",
			old:      "old",
			new:      "old new",
			expected: " new",
		},
		{
			name:     "content completely changed",
			old:      "old",
			new:      "completely different",
			expected: "completely different",
		},
		{
			name:     "same content",
			old:      "same",
			new:      "same",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNewContent(tt.old, tt.new)
			if got != tt.expected {
				t.Errorf("extractNewContent() = %q, want %q", got, tt.expected)
			}
		})
	}
}
