package runtime

import (
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	tests := []struct {
		name     string
		opts     StartOptions
		expected string
	}{
		{
			name: "mayor",
			opts: StartOptions{
				Role: RoleMayor,
			},
			expected: "hq-mayor",
		},
		{
			name: "deacon",
			opts: StartOptions{
				Role: RoleDeacon,
			},
			expected: "hq-deacon",
		},
		{
			name: "witness",
			opts: StartOptions{
				Role:    RoleWitness,
				RigName: "gastown",
			},
			expected: "gt-gastown-witness",
		},
		{
			name: "refinery",
			opts: StartOptions{
				Role:    RoleRefinery,
				RigName: "gastown",
			},
			expected: "gt-gastown-refinery",
		},
		{
			name: "crew",
			opts: StartOptions{
				Role:       RoleCrew,
				RigName:    "gastown",
				WorkerName: "max",
			},
			expected: "gt-gastown-crew-max",
		},
		{
			name: "polecat",
			opts: StartOptions{
				Role:       RolePolecat,
				RigName:    "gastown",
				WorkerName: "toast",
			},
			expected: "gt-gastown-toast",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSessionID(tt.opts)
			if got != tt.expected {
				t.Errorf("GenerateSessionID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAgentRoleConstants(t *testing.T) {
	// Verify role constants have expected string values
	roles := map[AgentRole]string{
		RolePolecat:  "polecat",
		RoleWitness:  "witness",
		RoleRefinery: "refinery",
		RoleMayor:    "mayor",
		RoleDeacon:   "deacon",
		RoleCrew:     "crew",
	}

	for role, expected := range roles {
		if string(role) != expected {
			t.Errorf("Role %v = %q, want %q", role, string(role), expected)
		}
	}
}

func TestHealthStateConstants(t *testing.T) {
	// Verify health state constants have expected string values
	states := map[HealthState]string{
		HealthHealthy:   "healthy",
		HealthDegraded:  "degraded",
		HealthUnhealthy: "unhealthy",
		HealthUnknown:   "unknown",
	}

	for state, expected := range states {
		if string(state) != expected {
			t.Errorf("HealthState %v = %q, want %q", state, string(state), expected)
		}
	}
}

func TestResponseTypeConstants(t *testing.T) {
	// Verify response type constants have expected string values
	types := map[ResponseType]string{
		ResponseText:       "text",
		ResponseToolCall:   "tool_call",
		ResponseToolResult: "tool_result",
		ResponseError:      "error",
		ResponseComplete:   "complete",
	}

	for rt, expected := range types {
		if string(rt) != expected {
			t.Errorf("ResponseType %v = %q, want %q", rt, string(rt), expected)
		}
	}
}
