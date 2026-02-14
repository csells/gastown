# Gas City SDK Technical Specification

> **Version:** 0.1.0-draft
> **Date:** 2026-02-14
> **Status:** Draft
> **Source:** Spec-forge pipeline (Research → Ideation → Planning)

## 1. Executive Summary

Gas City is an **orchestration-builder SDK** that evolves Gas Town from a hardcoded 7-role multi-agent system into a configurable toolkit. Users start with a single CLI agent and progressively layer in capabilities — work tracking, task loops, worker pools, messaging, workflows, monitoring — until they reach full Gas Town-level orchestration or beyond.

Gas Town, Agent Teams, and Ralph are **named configurations** (shapes) at different points along the complexity progression, not separate products. The SDK uses mainstream terminology while preserving Gas Town's colorful names as an optional theme.

**The foundational design principle:** every agent — regardless of implementation — is accessed through a uniform **Agent Runtime** abstraction. This decouples the orchestration logic from any specific coding agent (Claude Code, Codex, Gemini, OpenCode) or execution substrate (tmux, Docker, Agent SDK, custom code). The rest of the SDK builds exclusively on this abstraction.

---

## 2. The Agent Runtime Abstraction

### 2.1 Design Rationale

Gas Town is currently coupled to Claude Code running in tmux sessions. Every interaction — starting agents, sending work, capturing output, health-checking, nudging — flows through tmux primitives (`send-keys`, `capture-pane`, `has-session`). This coupling means:

- You can't use a different coding agent without rewriting the session layer
- You can't run without tmux (e.g., in containers, CI, or over APIs)
- Bidirectional communication depends on tmux pane scraping
- Presentation (attaching to a tmux pane) is fused with execution

The Agent Runtime abstraction separates **what you do with an agent** from **how a specific agent implementation works**.

### 2.2 The Runtime Interface

```
AgentRuntime
├── Lifecycle
│   ├── Start(config) → AgentHandle
│   ├── Stop(handle, graceful: bool)
│   ├── Restart(handle)
│   └── IsRunning(handle) → bool
│
├── Work
│   ├── Assign(handle, task)         # Give the agent a task
│   ├── Nudge(handle, message)       # Send input/signal to a running agent
│   └── GetState(handle) → State     # Read the agent's current work state
│
├── I/O
│   ├── SendInput(handle, text)      # Feed text to the agent's stdin
│   ├── CaptureOutput(handle) → stream  # Stream the agent's output
│   └── ReadOutput(handle) → string     # Snapshot current output
│
├── Health
│   ├── Ping(handle) → PingResult
│   └── GetMetrics(handle) → Metrics
│
└── Presentation (optional)
    ├── Attach(handle)               # Interactive session (if supported)
    └── Detach(handle)
```

### 2.3 Built-in Runtime Adapters

| Adapter | Implementation | Notes |
|---------|---------------|-------|
| `claude-code` | tmux sessions running `claude --dangerously-skip-permissions` | Current Gas Town behavior, extracted into an adapter |
| `codex` | CLI subprocess for OpenAI Codex | Subprocess-based I/O |
| `gemini-cli` | CLI subprocess for Gemini | Subprocess-based I/O |
| `opencode` | CLI subprocess for OpenCode | Subprocess-based I/O |
| `agent-sdk` | Claude Agent SDK (Python/TypeScript) | API-based, no CLI needed |
| `subprocess` | Any CLI command with stdin/stdout | Generic adapter for arbitrary agents |
| `docker` | Container-based agent execution | Sandboxed execution |

### 2.4 Adapter Configuration

```toml
# In gas-city.toml, each agent specifies its runtime
[[agents]]
name = "coder"
runtime = "claude-code"  # Which adapter to use

# Runtime-specific settings go in [agents.runtime_config]
[agents.runtime_config]
command = "claude --dangerously-skip-permissions"
# claude-code adapter: tmux session management
# agent-sdk adapter: API endpoint, model, tools
# subprocess adapter: command, args, env
```

### 2.5 What the Runtime Replaces in Gas Town

| Current Gas Town code | Becomes |
|----------------------|---------|
| `tmux.NewTmux()` + `t.CreateSession()` | `runtime.Start(config)` |
| `t.KillSession()` / `t.KillSessionWithProcesses()` | `runtime.Stop(handle)` |
| `t.HasSession()` + `t.IsAgentRunning()` | `runtime.IsRunning(handle)` |
| `t.SendKeys(session, message)` | `runtime.Nudge(handle, message)` or `runtime.SendInput(handle, text)` |
| `t.CapturePane()` | `runtime.CaptureOutput(handle)` / `runtime.ReadOutput(handle)` |
| `t.AttachSession()` | `runtime.Attach(handle)` (optional — not all runtimes support it) |
| `NudgeSession()` with retry logic | `runtime.Nudge(handle, message)` with adapter-specific retry |
| Health check via `capture-pane` + prompt detection | `runtime.Ping(handle)` |

### 2.6 How tmux Fits In

tmux is **not eliminated** — it becomes the implementation detail of the `claude-code` adapter. Users who want the current Gas Town experience (attaching to agent sessions, watching agents work in split panes) continue to get it. But tmux is no longer a prerequisite for the orchestration to function.

For adapters that don't support `Attach()` (like `agent-sdk`), the SDK provides alternative observation through `CaptureOutput()` streaming and a web-based dashboard (future work).

---

## 3. Terminology Mapping

Gas Town uses evocative names tied to its railroad/refinery aesthetic. Gas City uses mainstream SDK terminology. Both vocabularies are valid; the config format uses SDK terms, and Gas Town names are available as an optional theme.

| Gas Town Term | SDK Term | Description |
|---------------|----------|-------------|
| Town | **Workspace** | Top-level orchestrator instance |
| Rig | **Project** | A git repository under management |
| Mayor | **Coordinator** | Agent that dispatches work to others |
| Deacon | **Supervisor** | Health monitor / watchdog |
| Witness | **Observer** | Per-project state monitor |
| Refinery | **Integrator** | Merge queue processor |
| Polecat | **Worker** | Ephemeral task executor |
| Crew | **Agent** (persistent) | Long-lived workspace agent |
| Dog | **Service** | Background infrastructure worker |
| Hook (work state) | **State** | Persistent work state for an agent |
| Bead | **Task** / **Issue** | Individual work item |
| Convoy | **Batch** | Group of related tasks |
| Sling | **Assign** / **Dispatch** | Send work to an agent |
| Formula | **Template** | Reusable workflow definition |
| Molecule | **Workflow** | Running instance of a template |
| Mail | **Messages** | Async inter-agent communication |
| Nudge | **Signal** | Real-time notification to an agent |
| GUPP | **Auto-execute** | Agents auto-run when work appears on their state |
| Overseer | **Operator** | The human controlling the system |

### 3.1 Where Names Come From in Gas Town Today

The 7 hardcoded roles are defined in:
- **`internal/config/roles.go`**: `AllRoles()` returns `["mayor", "deacon", "dog", "witness", "refinery", "polecat", "crew"]`
- **`internal/config/roles/*.toml`**: 7 TOML files defining session patterns, health config, env vars, nudge prompts
- **`internal/constants/constants.go`**: `RoleMayor`, `RoleWitness`, etc. as string constants; directory names (`DirMayor`, `DirPolecats`); session prefixes (`gt-`, `hq-`)
- **`internal/session/identity.go`**: `AgentIdentity` struct with hardcoded `Role` enum and per-role address/session-name generation
- **`internal/session/names.go`**: `MayorSessionName()`, `WitnessSessionName(rig)`, etc.
- **`internal/cmd/role.go`**: `detectRole()` with hardcoded directory-to-role mapping; `getRoleHome()` with per-role path logic
- **`internal/cmd/start.go`**: Hardcoded parallel startup of Mayor, Deacon, then rig agents
- **`internal/cmd/patrol_new.go`**: Per-role `switch` for deacon/witness/refinery patrol molecules

In Gas City, all of these become **data-driven from the config file**.

---

## 4. The Progressive Capability Model

Each level adds a capability. The config file grows accordingly. Every level is independently useful.

### Level 0: Agent Runtime (Foundation)

A single agent wrapped in the standard runtime interface. No orchestration, no work tracking — just a uniform way to start, stop, and interact with any coding agent.

```toml
# gas-city.toml — Level 0: Agent Runtime
[workspace]
name = "my-project"

[[agents]]
name = "coder"
runtime = "claude-code"
```

This is the minimum viable Gas City config. It:
- Wraps a Claude Code instance in the Agent Runtime abstraction
- Provides `gc start` / `gc stop` / `gc status` commands
- Enables `gc attach coder` (if the runtime supports it)
- Establishes the runtime adapter pattern

**With a different runtime:**
```toml
[[agents]]
name = "coder"
runtime = "codex"  # or "gemini-cli", "opencode", "agent-sdk", "subprocess"

[agents.runtime_config]
command = "codex"
args = ["--model", "o3"]
```

**With the Agent SDK (no CLI needed):**
```toml
[[agents]]
name = "coder"
runtime = "agent-sdk"

[agents.runtime_config]
language = "python"            # or "typescript"
entrypoint = "./agents/coder.py"
model = "claude-sonnet-4-5-20250929"
```

### Level 1: + Work Tracking

```toml
# gas-city.toml — Level 1: + Work Tracking
[workspace]
name = "my-project"

[projects.main]
repo = "."  # or a URL

[tasks]
backend = "beads"  # "beads", "github-issues", "linear", "jira"

[[agents]]
name = "coder"
runtime = "claude-code"
```

This adds:
- A pluggable task backend for tracking work items
- `gc task list`, `gc task create`, `gc task assign` commands
- Task state visible to agents via their work state

### Level 2: + Task Loop ("Ralph" shape)

```toml
# gas-city.toml — Level 2: Task Loop (Ralph)
[workspace]
name = "my-project"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "coder"
runtime = "claude-code"

[agents.loop]
enabled = true
auto_execute = true     # GUPP: agent auto-runs when work appears
clear_context = true    # Clear agent context between tasks
# Implicit cycle: read-task → code → test → commit → clear → next
```

This is the **Ralph** milestone — a single agent that autonomously processes a task queue. The loop controller:
1. Checks for available tasks assigned to this agent
2. Sends the task to the agent via `runtime.Assign()`
3. Monitors for completion
4. Clears context (if configured) via `runtime.Stop()` + `runtime.Start()`
5. Picks up the next task

### Level 3: + Worker Agents ("Agent Teams" shape)

```toml
# gas-city.toml — Level 3: Worker Pool (Agent Teams)
[workspace]
name = "my-project"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "lead"
role = "coordinator"
runtime = "claude-code"

[[agents]]
name = "workers"
role = "worker"
runtime = "claude-code"
ephemeral = true
pool = { min = 0, max = 5 }

[agents.loop]
enabled = true
auto_execute = true
clear_context = true
```

This is the **Agent Teams** milestone. It adds:
- A coordinator agent that breaks work into subtasks and dispatches
- A pool of ephemeral worker agents that spin up/down on demand
- The coordinator uses the task system to assign work to the pool
- Workers are created/destroyed by the pool manager via the Agent Runtime

### Level 4: + Inter-Agent Messaging

```toml
# gas-city.toml — Level 4: + Messaging
# ...previous config...

[messaging]
backend = "beads"  # "beads", "redis", "filesystem"
channels = ["updates", "reviews"]
```

This adds:
- Async message queues between agents
- Named channels for topic-based communication
- `gc mail send`, `gc mail inbox` commands
- Agents can send status updates, request reviews, escalate issues

### Level 5: + Workflow Templates

```toml
# gas-city.toml — Level 5: + Workflows
# ...previous config...

[workflows]
templates_dir = "workflows/"  # Directory containing template TOML files
```

Workflow templates (Gas Town's "formulas") define reusable multi-step processes:

```toml
# workflows/code-review.toml
formula = "code-review"
description = "Multi-aspect code review"
type = "aspect"
version = 1

[[aspects]]
id = "correctness"
title = "Correctness Review"
focus = "Logic errors, edge cases, off-by-one"

[[aspects]]
id = "security"
title = "Security Review"
focus = "Injection, auth, data exposure"

[[aspects]]
id = "performance"
title = "Performance Review"
focus = "Algorithmic complexity, resource leaks"
```

### Level 6: + Monitoring & Health

```toml
# gas-city.toml — Level 6: + Monitoring
# ...previous config...

[[agents]]
name = "supervisor"
role = "supervisor"
runtime = "claude-code"
scope = "workspace"

[agents.health]
ping_timeout = "30s"
stuck_threshold = "1h"
consecutive_failures = 3
kill_cooldown = "5m"
```

This adds:
- A supervisor agent that monitors all other agents
- Health checking via `runtime.Ping()` on each agent
- Automatic restart of stuck or crashed agents
- Configurable thresholds per agent

### Level 7: + Per-Project Agents & Full Orchestration ("Gas Town" shape)

```toml
# gas-city.toml — Level 7: Full Orchestration (Gas Town)
[workspace]
name = "my-town"

[projects.gastown]
repo = "https://github.com/steveyegge/gastown"

[projects.beads]
repo = "https://github.com/steveyegge/beads"

[tasks]
backend = "beads"

[messaging]
backend = "beads"

[workflows]
templates_dir = "workflows/"

[batches]
enabled = true
auto_track = true

# --- Town-level agents ---

[[agents]]
name = "coordinator"
role = "coordinator"
runtime = "claude-code"
scope = "workspace"

[[agents]]
name = "supervisor"
role = "supervisor"
runtime = "claude-code"
scope = "workspace"

[[agents]]
name = "services"
role = "service"
runtime = "claude-code"
scope = "workspace"
ephemeral = true
pool = { min = 0, max = 3 }

# --- Per-project agents ---

[[agents]]
name = "observer"
role = "observer"
scope = "project"      # One instance per project

[[agents]]
name = "integrator"
role = "integrator"
scope = "project"

[[agents]]
name = "workers"
role = "worker"
runtime = "claude-code"
scope = "project"
ephemeral = true
pool = { min = 0, max = 5 }

[agents.loop]
enabled = true
auto_execute = true
clear_context = true

[[agents]]
name = "persistent-agents"
role = "agent"
runtime = "claude-code"
scope = "project"
# Persistent agents are created imperatively: gc agent add gastown/joe
```

This is the full **Gas Town** milestone — multiple projects, full agent topology, messaging, workflows, batches, and health monitoring.

---

## 5. Configuration Schema

### 5.1 Top-Level Structure

```toml
[workspace]
name = "string"             # Required. Workspace identifier.
theme = "gas-town"          # Optional. Enables Gas Town naming in CLI output.

[projects.<name>]
repo = "string"             # Git repository URL or local path
branch = "main"             # Default branch (optional)

[tasks]
backend = "beads"           # Task backend: "beads", "github-issues", "linear", "jira"
# Backend-specific config follows

[messaging]
backend = "beads"           # Messaging backend: "beads", "redis", "filesystem"
channels = ["string"]       # Named channels

[workflows]
templates_dir = "string"    # Directory containing workflow templates

[batches]
enabled = false             # Enable batch (convoy) tracking
auto_track = false          # Auto-create batches for multi-task dispatches

[[agents]]
name = "string"             # Required. Agent identifier.
role = "string"             # Role: "coordinator", "supervisor", "observer", "integrator", "worker", "agent", "service"
runtime = "string"          # Runtime adapter: "claude-code", "codex", "agent-sdk", "subprocess", etc.
scope = "workspace"         # "workspace" (one total) or "project" (one per project)
ephemeral = false           # Whether agent is created/destroyed on demand
pool = { min = 0, max = 5 } # Pool sizing for ephemeral agents

[agents.loop]               # Task loop configuration
enabled = false
auto_execute = false        # GUPP: auto-run when work appears on state
clear_context = false       # Clear agent context between tasks

[agents.health]             # Health check configuration
ping_timeout = "30s"
stuck_threshold = "1h"
consecutive_failures = 3
kill_cooldown = "5m"

[agents.runtime_config]     # Runtime-adapter-specific settings
# claude-code: command, flags, env
# agent-sdk: entrypoint, model, tools
# subprocess: command, args, env, working_dir

[agents.env]                # Environment variables injected into the agent
KEY = "VALUE"

[agents.prompts]
system = "string"           # Path to system prompt template
nudge = "string"            # Initial nudge message
```

### 5.2 Built-in Roles

The SDK provides built-in behavior for these roles. Custom roles use the same config surface but with user-provided prompt templates and no built-in behavior.

| Role | Scope | Built-in Behavior |
|------|-------|-------------------|
| `coordinator` | workspace | Dispatches tasks to workers; breaks down large tasks |
| `supervisor` | workspace | Health-checks all agents; restarts stuck/dead agents |
| `observer` | project | Monitors project state; tracks worker progress |
| `integrator` | project | Processes merge queue; runs verification gates |
| `worker` | project | Executes individual tasks; typically ephemeral with pool |
| `agent` | project | Persistent workspace agent; user-managed lifecycle |
| `service` | workspace | Background infrastructure tasks; dispatched by supervisor |

### 5.3 Scope Semantics

- **`scope = "workspace"`**: One instance of this agent exists per workspace. Equivalent to Gas Town's "town-scoped" roles (Mayor, Deacon).
- **`scope = "project"`**: One instance per project. Equivalent to Gas Town's "rig-scoped" roles (Witness, Refinery). The SDK automatically instantiates one per registered project.

---

## 6. The Three Milestone Shapes

### 6.1 Ralph

**What it is:** A single agent with a task loop that autonomously processes queued work.

**Config:**
```toml
[workspace]
name = "ralph"

[projects.main]
repo = "."

[tasks]
backend = "github-issues"

[[agents]]
name = "ralph"
runtime = "claude-code"

[agents.loop]
enabled = true
auto_execute = true
clear_context = true
```

**Gas Town mapping:** No direct equivalent — Ralph predates Gas Town. It's what you get when one agent runs the full read→code→test→commit→clear cycle on a task queue without any other agents.

### 6.2 Agent Teams

**What it is:** A coordinator that dispatches tasks to a pool of ephemeral workers.

**Config:**
```toml
[workspace]
name = "agent-teams"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "lead"
role = "coordinator"
runtime = "claude-code"

[[agents]]
name = "dev-pool"
role = "worker"
runtime = "claude-code"
ephemeral = true
pool = { min = 0, max = 5 }

[agents.loop]
enabled = true
auto_execute = true
clear_context = true
```

**Gas Town mapping:** Mayor (coordinator) + Polecats (worker pool). No supervisor, observer, or integrator.

### 6.3 Gas Town

**What it is:** Full multi-project orchestration with all agent roles.

See the Level 7 config in Section 4 above. Maps to the current Gas Town architecture:

| SDK Config | Gas Town Role |
|-----------|---------------|
| `coordinator` (workspace scope) | Mayor |
| `supervisor` (workspace scope) | Deacon |
| `service` (workspace scope, pooled) | Dog |
| `observer` (project scope) | Witness |
| `integrator` (project scope) | Integrator/Refinery |
| `worker` (project scope, pooled, ephemeral) | Polecat |
| `agent` (project scope, persistent) | Crew |

---

## 7. Custom Roles and Topologies

### 7.1 Defining a Custom Role

Any agent definition with a `role` value not in the built-in set becomes a custom role. Custom roles have no built-in behavior — their behavior comes entirely from their prompt templates and the capabilities they're given access to.

```toml
[[agents]]
name = "docs-writer"
role = "docs-writer"        # Custom role — no built-in behavior
runtime = "claude-code"
scope = "project"

[agents.prompts]
system = "prompts/docs-writer.md"
nudge = "Check for undocumented public APIs and write docs."

[agents.loop]
enabled = true
auto_execute = true

[agents.health]
ping_timeout = "60s"
stuck_threshold = "2h"
```

### 7.2 Custom Coordination Patterns

The built-in `coordinator` role uses a simple dispatch pattern. Custom coordination can be achieved by:

1. **Custom prompt templates** that define the coordination logic
2. **Messaging channels** for custom communication patterns
3. **Workflow templates** for structured multi-agent processes

Example: A "review-gate" pattern where every PR must pass through a reviewer agent before the integrator merges it:

```toml
[[agents]]
name = "reviewer"
role = "reviewer"
scope = "project"

[agents.prompts]
system = "prompts/reviewer.md"
nudge = "Check for pending review requests."

[messaging]
backend = "beads"
channels = ["review-requests", "review-results"]
```

### 7.3 Custom Topologies

The combination of agents, scopes, roles, and messaging channels defines the topology. Some example shapes beyond the three milestones:

**Hub-and-Spoke** (one coordinator, many specialized workers):
```toml
[[agents]]
name = "dispatcher"
role = "coordinator"
scope = "workspace"

[[agents]]
name = "frontend"
role = "worker"
scope = "project"

[agents.prompts]
system = "prompts/frontend-specialist.md"

[[agents]]
name = "backend"
role = "worker"
scope = "project"

[agents.prompts]
system = "prompts/backend-specialist.md"

[[agents]]
name = "tester"
role = "worker"
scope = "project"

[agents.prompts]
system = "prompts/test-specialist.md"
```

**Peer Network** (no coordinator, agents communicate directly):
```toml
[[agents]]
name = "alice"
role = "agent"
runtime = "claude-code"

[[agents]]
name = "bob"
role = "agent"
runtime = "codex"

[messaging]
backend = "beads"
channels = ["collaboration"]
```

**Pipeline** (sequential handoff):
```toml
# Defined as a workflow template
# workflows/pipeline.toml
formula = "feature-pipeline"
type = "workflow"

[[steps]]
id = "design"
title = "Design"
description = "Create technical design doc"

[[steps]]
id = "implement"
title = "Implement"
needs = ["design"]

[[steps]]
id = "test"
title = "Test"
needs = ["implement"]

[[steps]]
id = "review"
title = "Review"
needs = ["test"]
```

---

## 8. Migration from Gas Town to Gas City

### 8.1 Migration Strategy

Gas City is built **on top of** Gas Town's existing infrastructure. The migration is incremental:

1. **Phase 1: Extract the Agent Runtime** — Wrap the existing tmux session management into the `claude-code` adapter. All current code continues to work; the adapter is a thin wrapper.

2. **Phase 2: Config-driven role registry** — Replace `AllRoles()`, `TownRoles()`, `RigRoles()` with a dynamic registry loaded from config. The 7 built-in roles become default entries.

3. **Phase 3: Config-driven startup** — Replace the hardcoded boot sequence in `start.go` with a config-driven startup that reads the agents list and starts them according to scope and dependencies.

4. **Phase 4: Config-driven identity** — Replace the hardcoded `ParseAddress()`, `ParseSessionName()`, and per-role name functions with pattern-based resolution from config.

5. **Phase 5: Additional runtime adapters** — Implement adapters for Codex, Agent SDK, subprocess, etc.

### 8.2 File-by-File Migration Map

| File | Current State | Migration |
|------|--------------|-----------|
| `internal/config/roles.go` | `AllRoles()` returns hardcoded 7 roles; `LoadRoleDefinition()` loads from embedded TOML | Add `LoadFromConfig()` that reads `gas-city.toml` and registers roles dynamically. Keep embedded TOML as fallback defaults for the 7 built-in roles. |
| `internal/config/roles/*.toml` | 7 embedded role definitions | Become the default "Gas Town shape" definition. Custom shapes load their own role definitions from the config. |
| `internal/session/identity.go` | `AgentIdentity` with hardcoded `Role` enum; `ParseAddress()` with per-role switch | Replace `Role` enum with string type. `ParseAddress()` becomes pattern-matching against registered roles from config. |
| `internal/session/names.go` | `MayorSessionName()`, `WitnessSessionName(rig)`, etc. | Replace with `AgentSessionName(agent, project)` that uses the session pattern from the agent's config. Existing functions become convenience wrappers for the Gas Town shape. |
| `internal/cmd/start.go` | Hardcoded parallel startup: Mayor+Deacon first, then rig agents | Read agent list from config, group by startup priority (supervisors first, then coordinators, then project-scoped agents), start groups in parallel. |
| `internal/cmd/role.go` | `detectRole()` with hardcoded directory paths; `getRoleHome()` with per-role switch | Directory detection uses registered agent scope + work_dir patterns. `getRoleHome()` reads from agent config. |
| `internal/cmd/patrol_new.go` | Per-role switch: deacon/witness/refinery | Each role's patrol config comes from its agent definition in config. Custom roles that define a patrol template get patrol support automatically. |
| `internal/constants/constants.go` | Hardcoded role constants, directory names, session prefixes | Role constants become dynamic lookups. Directory names stay as defaults but are overridable via agent config. Session prefixes are configurable. |
| `internal/plugin/` | Plugin system with town/rig scoping, gate types, TOML frontmatter | Becomes the Gas City extension point. Plugins gain access to the Agent Runtime interface and can define custom agent behaviors. |

### 8.3 Backward Compatibility

A Gas Town workspace without a `gas-city.toml` file continues to work exactly as it does today. The presence of the config file opts into Gas City mode. The SDK detects which mode to use:

1. If `gas-city.toml` exists → Gas City mode (config-driven)
2. If `mayor/town.json` exists without `gas-city.toml` → Gas Town compatibility mode (hardcoded roles)
3. If neither exists → Fresh workspace, default to Gas City

---

## 9. Agent Runtime Deep Dive

### 9.1 Adapter Registration

Runtime adapters register themselves with the SDK at init time:

```go
// Pseudocode — actual implementation will be in Go
type RuntimeAdapter interface {
    Name() string
    Start(config AgentConfig) (AgentHandle, error)
    Stop(handle AgentHandle, graceful bool) error
    IsRunning(handle AgentHandle) bool
    Assign(handle AgentHandle, task Task) error
    Nudge(handle AgentHandle, message string) error
    GetState(handle AgentHandle) (State, error)
    SendInput(handle AgentHandle, text string) error
    CaptureOutput(handle AgentHandle) (io.Reader, error)
    ReadOutput(handle AgentHandle) (string, error)
    Ping(handle AgentHandle) (PingResult, error)
    GetMetrics(handle AgentHandle) (Metrics, error)
    SupportsAttach() bool
    Attach(handle AgentHandle) error
    Detach(handle AgentHandle) error
}

var registry = map[string]RuntimeAdapter{}

func RegisterAdapter(adapter RuntimeAdapter) {
    registry[adapter.Name()] = adapter
}
```

### 9.2 The `claude-code` Adapter

This adapter encapsulates Gas Town's existing tmux-based session management:

```
claude-code adapter
├── Start   → tmux.CreateSession() + send start_command
├── Stop    → tmux.KillSessionWithProcesses() (or graceful ESC + wait)
├── IsRunning → tmux.HasSession() + tmux.IsAgentRunning()
├── Assign  → Write task to agent's state (bead hook) + Nudge
├── Nudge   → tmux.SendKeys() with retry logic (existing NudgeSession)
├── GetState → Read agent's hook bead
├── SendInput → tmux.SendKeys() (raw text)
├── CaptureOutput → tmux.CapturePane()
├── ReadOutput → tmux.CapturePane() snapshot
├── Ping    → tmux.CapturePane() + detect prompt indicator
├── Attach  → tmux.AttachSession()
├── Detach  → tmux.DetachClient()
```

### 9.3 The `agent-sdk` Adapter

For agents implemented via the Claude Agent SDK (Python or TypeScript), no CLI or tmux is needed:

```
agent-sdk adapter
├── Start   → Launch subprocess: python agents/coder.py (or node agents/coder.js)
│             The subprocess connects to the Claude API and runs autonomously
├── Stop    → Send SIGTERM to subprocess; wait for graceful shutdown
├── IsRunning → Check subprocess PID
├── Assign  → Write task to a shared file/queue that the SDK agent reads
├── Nudge   → Write to subprocess stdin (if piped) or shared signal file
├── GetState → Read from shared state file maintained by the SDK agent
├── SendInput → Write to subprocess stdin
├── CaptureOutput → Read subprocess stdout/stderr
├── Ping    → Heartbeat check (SDK agent writes heartbeat to a file)
├── Attach  → Not supported (SupportsAttach() returns false)
```

### 9.4 The `subprocess` Adapter

A generic adapter for wrapping any CLI tool:

```
subprocess adapter
├── Start   → os/exec: spawn the configured command with args
├── Stop    → Send SIGTERM, wait grace period, SIGKILL
├── IsRunning → Check PID
├── Assign  → Write task description to stdin
├── Nudge   → Write message to stdin
├── SendInput → Write to stdin
├── CaptureOutput → Read stdout pipe
├── Ping    → Write "ping\n" to stdin, expect "pong\n" on stdout within timeout
├── Attach  → Not supported
```

---

## 10. Startup and Shutdown

### 10.1 Startup Order

Gas Town currently starts agents in this order:
1. Mayor + Deacon (parallel)
2. Rig agents: Witnesses + Refineries (parallel, per rig)
3. Configured crew (parallel)
4. Polecats (on demand)

Gas City generalizes this into **startup groups** derived from agent config:

| Priority | Criteria | Gas Town Equivalent |
|----------|----------|-------------------|
| 0 | `role = "supervisor"` | Deacon |
| 1 | `role = "coordinator"` | Mayor |
| 2 | `scope = "project"` and not ephemeral | Witness, Refinery, Crew |
| 3 | Everything else (started on demand) | Polecat, Dog |

Within each priority group, agents start in parallel.

### 10.2 Shutdown Order

Reverse of startup, matching Gas Town's current behavior:
1. Workers and persistent agents (stop before monitors can restart them)
2. Integrators (work processors)
3. Observers (monitors)
4. Coordinator, then Supervisor

### 10.3 The `gc` CLI

Gas City provides a `gc` command (or extends `gt` with a `gc` subcommand):

```
gc start                    # Start workspace per config
gc stop                     # Graceful shutdown
gc status                   # Show all agents and their state

gc agent list               # List all configured agents
gc agent start <name>       # Start a specific agent
gc agent stop <name>        # Stop a specific agent
gc agent attach <name>      # Attach to agent (if runtime supports it)
gc agent logs <name>        # Stream agent output

gc task list                # List tasks
gc task create "..."        # Create a task
gc task assign <task> <agent>  # Assign task to agent

gc mail send <to> "..."     # Send message
gc mail inbox               # Check messages

gc workflow run <template>  # Run a workflow template
gc workflow status          # Show running workflows
```

---

## 11. Extension Points

### 11.1 Plugin System

Gas Town's existing plugin system (80% complete) becomes Gas City's primary extension mechanism. Plugins are periodic automation tasks that run during supervisor patrol cycles.

Plugins gain access to the Agent Runtime interface, allowing them to:
- Query agent state across all runtime adapters
- Trigger custom actions based on agent metrics
- Implement custom health checks
- Define custom coordination patterns

### 11.2 Custom Runtime Adapters

Users can implement custom runtime adapters for agents not covered by the built-in set:

```toml
[[agents]]
name = "my-agent"
runtime = "custom"

[agents.runtime_config]
adapter_path = "./adapters/my-adapter"
# Adapter-specific config
```

The adapter is a Go plugin (or subprocess with a JSON-RPC protocol) that implements the `RuntimeAdapter` interface.

### 11.3 Task Backend Plugins

The task system is pluggable. The built-in backends are:
- `beads` — Gas Town's native task tracking (Dolt-based)
- `github-issues` — GitHub Issues as task source
- `linear` — Linear issues
- `jira` — Jira issues

Custom backends implement a `TaskBackend` interface.

### 11.4 Messaging Backend Plugins

Similarly, the messaging system is pluggable:
- `beads` — Beads-based message queues
- `redis` — Redis pub/sub
- `filesystem` — File-based message queues (simple, no external deps)

---

## 12. Addressing the Gas City Vision

The Gas City vision document (`specs/gas-city-vision.md`) identifies 6 key requirements. Here's how this spec addresses each:

| Vision Requirement | How Addressed |
|-------------------|---------------|
| "Orchestration-builder toolkit, not just an orchestrator" | The progressive capability model (Levels 0-7) lets users build their own orchestration shape from composable pieces |
| "Kit for building different town shapes / topologies" | Section 7 (Custom Roles and Topologies) shows how Agent Teams, Gas Town, hub-and-spoke, peer network, and pipeline shapes are all expressible in the same config format |
| "Create your own roles, teams, coordination rules, worker instructions" | Any `role` value in agent config creates a custom role; coordination is defined by prompt templates + messaging channels + workflow templates |
| "Roles expressed in dynamic format external to code" | All role definitions live in `gas-city.toml` and associated prompt template files. The hardcoded role registry becomes a config-driven registry |
| "Sandboxes, plugins, and hooks" | Plugin system (Section 11.1), custom runtime adapters including Docker-based sandboxing (Section 11.2), Agent Runtime hooks for lifecycle events |
| "Full configurability surface" | Single `gas-city.toml` file covers all aspects: agents, roles, runtimes, tasks, messaging, workflows, health, and extensions |

---

## 13. Architecture Design Principles

### 13.1 Layered Architecture

The SDK follows a strict layered architecture where each layer depends only on layers below it:

```
┌─────────────────────────────────────────┐
│  CLI (gc commands)                       │  Layer 4: User Interface
├─────────────────────────────────────────┤
│  Orchestration (startup, shutdown,       │  Layer 3: Coordination
│  pool management, task dispatch)         │
├─────────────────────────────────────────┤
│  Capabilities (tasks, messaging,         │  Layer 2: Services
│  workflows, health, plugins)             │
├─────────────────────────────────────────┤
│  Agent Runtime (RuntimeAdapter interface)│  Layer 1: Abstraction
├─────────────────────────────────────────┤
│  Adapters (claude-code, codex,           │  Layer 0: Implementation
│  agent-sdk, subprocess, docker)          │
└─────────────────────────────────────────┘
```

**Invariant:** No code in Layer 2+ may import or reference adapter-specific types. All agent interaction flows through the `RuntimeAdapter` interface. This guarantee is enforced by Go's package visibility — adapter packages are not importable from orchestration packages.

### 13.2 Monotonic Capability Growth

The progressive capability model has a formal monotonicity property: adding a capability level never removes or breaks capabilities from lower levels.

Let `C(n)` represent the set of capabilities available at level `n`. The guarantee is:

`∀ n: C(n) ⊂ C(n+1)`

This is enforced by the config parser: a `gas-city.toml` valid at level `n` is always valid at level `n+1` without modification. New sections add capabilities; they never invalidate existing sections.

### 13.3 Adapter Correctness

Each runtime adapter must satisfy these safety properties:

1. **Idempotent Stop:** `Stop(handle)` on an already-stopped agent is a no-op (returns nil, not an error)
2. **Liveness after Start:** After `Start(config)` returns a handle without error, `IsRunning(handle)` returns true within `ping_timeout`
3. **Graceful degradation:** If `Attach()` is not supported, `SupportsAttach()` returns false and `Attach()` returns `ErrNotSupported` (never panics)
4. **Bound resource cleanup:** `Stop(handle, graceful=false)` guarantees all OS resources (processes, file descriptors, network connections) are released within `kill_cooldown`

### 13.4 Pool Scaling Bounds

For ephemeral agent pools configured with `pool = { min = M, max = N }`:

- **Invariant:** At any time `t`, the number of running instances `|R(t)|` satisfies: `M ≤ |R(t)| ≤ N`
- **Convergence:** When pending tasks exceed pool capacity, the pool scales to `N` within `O(N × start_latency)` amortized time
- **Drain guarantee:** When no tasks are pending and `min = 0`, the pool scales to 0 within `idle_timeout` (configurable, default 5 minutes)

---

## 14. Testing Strategy

### 14.1 Unit Tests

Each subsystem has isolated unit tests:

| Subsystem | Test Focus | Mock Boundary |
|-----------|-----------|---------------|
| Config parser | TOML parsing, validation, progressive levels | Filesystem (embed test fixtures) |
| Role registry | Registration, lookup, built-in vs custom | None (pure logic) |
| Agent Runtime interface | Contract tests for each adapter | OS/tmux (interface mock) |
| Pool manager | Scaling up/down, min/max bounds | RuntimeAdapter (mock) |
| Task loop | Cycle execution, context clearing | RuntimeAdapter + TaskBackend (mocks) |
| Startup sequencer | Priority grouping, parallel start | RuntimeAdapter (mock) |
| Shutdown sequencer | Reverse ordering, graceful/immediate | RuntimeAdapter (mock) |

### 14.2 Integration Tests

| Test | What It Validates |
|------|-------------------|
| `claude-code` adapter roundtrip | Start → Nudge → CaptureOutput → Stop on a real tmux session |
| `subprocess` adapter roundtrip | Start → SendInput → ReadOutput → Stop on a real subprocess |
| Config → startup | Parse `gas-city.toml` → start all agents → verify all running → shutdown |
| Shape smoke tests | Ralph, Agent Teams, Gas Town shapes start and run a simple task end-to-end |
| Migration test | Existing Gas Town workspace → `gc migrate` → verify equivalent config → start in Gas City mode |

### 14.3 Contract Tests for Runtime Adapters

Every `RuntimeAdapter` implementation must pass a standard contract test suite:

```go
func TestAdapterContract(t *testing.T, adapter RuntimeAdapter) {
    // Test: Start returns valid handle
    // Test: IsRunning returns true after Start
    // Test: Stop makes IsRunning return false
    // Test: Double-Stop is idempotent (no error)
    // Test: Nudge delivers message (verify via CaptureOutput)
    // Test: Ping returns within timeout
    // Test: SupportsAttach matches Attach behavior
}
```

### 14.4 Acceptance Criteria per Phase

| Phase | Acceptance |
|-------|-----------|
| Phase 1: Runtime extraction | All existing Gas Town tests pass with the adapter wrapper. No behavioral change. |
| Phase 2: Dynamic registry | `gc role list` shows custom roles. `gc validate` accepts unknown role names. |
| Phase 3: Config startup | `gc start` with a custom `gas-city.toml` starts the defined agents. |
| Phase 4: Config identity | `gc status` shows agent names from config, not hardcoded role names. |
| Phase 5: New adapters | `subprocess` adapter passes contract tests. At least one non-Claude agent runs. |

---

## 15. Exclusions and Non-Goals

The following are explicitly **out of scope** for Gas City v1:

| Exclusion | Rationale |
|-----------|-----------|
| **Web dashboard / GUI** | Important but separate concern. The Agent Runtime's `CaptureOutput()` provides the data; the UI is a future project. |
| **Cloud/remote execution** | v1 targets local execution only. Docker adapter provides sandboxing but not remote deployment. Cloud adapters (EC2, Cloud Run) are future work. |
| **Built-in CI/CD integration** | Gas City orchestrates agents, not deployment pipelines. Users can trigger CI from agents but the SDK doesn't manage CI systems. |
| **Agent marketplace / registry** | Discovering and sharing agent configurations is a community feature, not a core SDK feature. |
| **Multi-tenant isolation** | One workspace per machine. Multi-tenant (multiple users sharing a workspace) requires security boundaries that are out of scope. |
| **Automatic prompt engineering** | The SDK provides the plumbing (prompt template paths); it doesn't generate or optimize prompts. |
| **Cost management / billing** | The SDK doesn't track API costs. That's the responsibility of the underlying agent runtime (Claude Code already has cost tracking). |
| **Real-time collaboration** | Agents don't share a real-time editing session. They coordinate through tasks and messages, not cursor positions. |

---

## 16. Open Questions

1. **Runtime adapter protocol for out-of-process adapters.** Go plugins are fragile. JSON-RPC over stdin/stdout (like LSP) may be more robust for custom adapters. Decision needed before Phase 5.

2. **Multi-runtime coordination.** When a coordinator is Claude Code and workers are Codex, how does task result serialization work? Likely needs a standard task result format that all runtimes emit.

3. **State persistence across restarts.** The `claude-code` adapter can use tmux session survival + `/resume`. Other runtimes need a different persistence mechanism. The Agent Runtime should define a standard state checkpoint interface.

4. **Web dashboard.** For runtimes that don't support `Attach()`, users need another way to observe agents. A web dashboard streaming `CaptureOutput()` is the likely answer but is out of scope for this spec.

5. **Agent-to-agent runtime heterogeneity.** Can a coordinator running on Claude Code effectively dispatch to workers running on Codex? The answer is yes (they communicate through the task system, not directly), but needs validation.

6. **Config hot-reload.** Can you add agents to a running workspace without restarting? Gas Town doesn't support this today. Gas City should, but the implementation is non-trivial.

---

## 17. Verification Checklist

- [x] The spec expresses all 3 milestone shapes (Ralph, Agent Teams, Gas Town) in the proposed config format — Section 6
- [x] Each capability level is independently useful and builds on the previous — Section 4
- [x] The config format is a single file at each level (no unnecessary config sprawl) — All examples use one `gas-city.toml`
- [x] Every hardcoded location has a clear migration path — Section 8.2
- [x] At least one custom role is defined end-to-end as an example — Section 7.1 (`docs-writer`) and Section 7.2 (`reviewer`)
- [x] The spec covers all 6 points from the Gas City vision document — Section 12
- [x] The Agent Runtime abstraction decouples orchestration from agent implementation — Sections 2 and 9
- [x] Multiple coding agent runtimes are supported (Claude Code, Codex, Gemini, Agent SDK, subprocess) — Section 2.3
- [x] tmux is an implementation detail, not an architectural requirement — Section 2.6
