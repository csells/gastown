# Gas City SDK — Technical Specification

> **Version:** 0.6.0
> **Date:** 2026-02-15
> **Status:** Planning (grounded in Gas Town source exploration)
> **Predecessor:** v0.5.0 (spec-forge pipeline; this revision aligns with Gas Town reality)

---

## 1. Executive Summary

Gas City is an **orchestration-builder SDK** — a Go toolkit for composing multi-agent coding workflows. It is the "Level 8" — the next step beyond Gas Town for users who have outgrown any single orchestrator and want to build their own. It extracts Gas Town's battle-tested subsystems (agents, beads, mail, nudge, formulas, molecules, convoys, plugins, health monitoring) into a configurable SDK where **all role behavior is user-supplied configuration** and the SDK provides only infrastructure. You can build Gas Town in Gas City, or Ralph, or Claude Code Agent Teams, or any other orchestration topology you design — via specific configurations.

**The core principle: ZERO hardcoded roles.** Gas City has no built-in Mayor, Deacon, Polecat, or any other role. The SDK provides the machinery — agent protocol, task system, messaging, formula engine, molecule execution, health monitoring, session management — and the user provides TOML config files and Markdown prompt templates that define what agents exist, what they do, and how they coordinate. Users can create their own roles, teams, coordination rules, and worker instructions — a full configurability surface.

**Progressive capability model:** Users start with a minimal TOML config (~10 lines for a single agent) and add sections as needed. The SDK activates subsystems based on what's configured. Config grows; the SDK is constant. Every level is independently useful.

**Three example configs ship with the SDK:**
- `ralph.toml` — Single agent with task loop
- `ccat.toml` — Claude Code Agent Teams: coordinator + worker pool
- `gastown.toml` — Complete Gas Town replication with all 7 roles, formulas, plugins, multi-project support

These are examples, not defaults. `gc init --file ralph.toml` copies an example to the workspace.

**What the SDK provides (Go code — infrastructure):**
- Agent protocol implementation (start/stop/nudge/prompt agents)
- Template rendering engine (Go templates in Markdown)
- Session management (tmux, subprocess)
- Health monitoring mechanics (patrol loop, stall detection)
- Task dispatch loop with atomic claiming
- Config loading with override resolution
- Formula parsing and molecule execution
- Event bus and websocket streaming
- Mail queue and nudge delivery
- Plugin gate evaluation

**What user config provides (behavior):**
- Which agents exist and what roles they play
- What each role does (prompt templates)
- Health thresholds, session patterns, env vars
- Which subsystems are active (tasks, messaging, formulas, etc.)
- Coordination rules (depends_on, hooks, workflows)
- Formulas, plugins, and their gate conditions

**What this spec does NOT cover:**
- Cloud/remote execution (v1 is local-only)
- Multi-tenant isolation
- Agent marketplace / sharing
- Automatic prompt engineering
- Cost management (responsibility of the underlying runtime)

---

## 2. Agent Protocol

### 2.1 Design Rationale

Every AI coding agent — regardless of implementation — is accessed through a uniform "factory worker" abstraction. This decouples the orchestration logic from any specific coding agent (Claude Code, Codex, Gemini, OpenCode, etc.) or execution substrate (tmux, Docker, Agent SDK, custom code). The rest of the SDK builds exclusively on this abstraction.

v1 ships providers for tmux-based agents and generic subprocess. Docker and Agent SDK substrates are future providers — the interface accommodates them without changes.

The critical operation is `SendPrompt`: delivering a prompt (text + optional images + optional file attachments) to a running agent. This is how Gas Town assigns work — nudging an agent with a prompt that includes context.

### 2.2 The AgentProvider Interface

```go
// AgentProvider is the interface every agent runtime must implement.
type AgentProvider interface {
    // Identity
    Name() string  // e.g., "claude", "codex", "gemini"

    // Lifecycle
    Start(ctx context.Context, config AgentConfig) (AgentHandle, error)
    Stop(handle AgentHandle, graceful bool) error
    Restart(handle AgentHandle) error
    IsRunning(handle AgentHandle) bool

    // Prompt delivery — the core operation
    SendPrompt(handle AgentHandle, prompt Prompt) error

    // Observation
    ReadOutput(handle AgentHandle) (string, error)
    GetState(handle AgentHandle) (AgentState, error)

    // Health
    Ping(handle AgentHandle) (PingResult, error)

    // Session (optional)
    SupportsAttach() bool
    Attach(handle AgentHandle) error
    Detach(handle AgentHandle) error
}

// Prompt represents a message to send to an agent.
// Matches how Gas Town nudges agents: text with optional rich content.
type Prompt struct {
    Text        string            // The prompt text
    Images      [][]byte          // Optional image attachments (screenshots, diagrams)
    Files       []FileAttachment  // Optional file references
    Metadata    map[string]string // Provider-specific metadata
}

type FileAttachment struct {
    Path    string // Absolute path to file
    Content []byte // Or inline content (used when path is empty)
}

// Adopter is an optional interface for crash recovery.
// Providers that can reconnect to a running process after a controller crash
// (e.g., tmux sessions) implement this interface.
type Adopter interface {
    Adopt(ctx context.Context, identity AgentIdentity, metadata map[string]string) (AgentHandle, error)
}
```

### 2.3 Core Data Structures

```go
type AgentIdentity struct {
    Workspace string  // Workspace name
    Project   string  // Project name (empty for workspace-scoped)
    Name      string  // Agent name from config
    Instance  int     // Pool instance index (0 for non-pooled)
}

type AgentHandle struct {
    ID        string            // Unique ID for this runtime instance (UUID)
    Identity  AgentIdentity     // Logical identity (stable across restarts)
    Provider  string            // Provider name (e.g., "claude")
    StartedAt time.Time
    PID       int               // OS process ID (0 if N/A)
    Metadata  map[string]string // Provider-specific metadata (session name, etc.)
}

type AgentConfig struct {
    Name           string
    Role           string
    Provider       string            // "claude", "codex", "gemini", etc.
    Project        string            // Empty for workspace scope
    Command        string            // e.g., "claude"
    Args           []string          // e.g., ["--dangerously-skip-permissions"]
    Env            map[string]string
    WorkDir        string
    PromptTemplate string            // Path to Markdown Go template
    NudgeMessage   string            // Default nudge text
    SessionPattern string            // tmux session name pattern
    DependsOn      []string          // Agent names that must start first
    Ephemeral      bool              // Created/destroyed per-task
    Isolation      string            // "none", "worktree", "directory"
    Pool           *PoolConfig
    Loop           *LoopConfig
    Health         *HealthConfig
    Hooks          *HookConfig
}

type AgentState struct {
    Status        AgentStatus // Running, Idle, Working, Stalled, Stopped
    CurrentTask   string      // Bead ID being worked on
    LastActivity  time.Time
    HookBead      string      // Bead pinned to agent's hook
    Uptime        time.Duration
}

type AgentStatus int
const (
    StatusStopped AgentStatus = iota
    StatusStarting
    StatusIdle
    StatusWorking
    StatusStalled
)
```

### 2.4 Built-in Providers

All providers Gas Town supports today:

| Provider | Command | Flags | Hooks | Resume | Session |
|----------|---------|-------|-------|--------|---------|
| `claude` (default) | `claude` | `--dangerously-skip-permissions` | Native | `--resume` (flag) | tmux |
| `codex` | `codex` | `--yolo` | No | `resume <id>` (subcommand) | tmux |
| `gemini` | `gemini` | `--approval-mode yolo` | Yes | No | tmux |
| `opencode` | `opencode` | — | Yes (plugin) | No | tmux |
| `cursor` | `cursor-agent` | `-f` | No | No | tmux |
| `auggie` | `auggie` | `--allow-indexing` | No | No | tmux |
| `amp` | `amp` | `--dangerously-allow-all --no-ide` | No | `threads continue <id>` (subcommand) | tmux |
| `subprocess` | (any) | (any) | No | No | stdin/stdout |

All tmux-based providers use the same session management: create session, inject env vars, send prompt via `send-keys`. The `subprocess` provider is the generic fallback for any CLI tool via stdin/stdout pipes.

### 2.5 Provider Auto-Detection

When `provider` is omitted from config, the SDK checks for available binaries in order: `claude` → `codex` → `gemini` → `opencode` → `cursor-agent` → `auggie` → `amp` → fallback to `subprocess`.

### 2.6 Provider Correctness Properties

**P1 — Idempotent Stop:** `Stop(handle)` on an already-stopped agent returns nil.
**P2 — Liveness After Start:** `Start()` returning `(handle, nil)` implies `IsRunning(handle)` is true within `ping_timeout`.
**P3 — Graceful Attach:** If `SupportsAttach()` returns false, `Attach()` returns `ErrNotSupported`.
**P4 — Bounded Cleanup:** `Stop(handle, false)` releases all OS resources within `kill_cooldown`.
**P5 — Thread Safety:** Concurrent calls to `SendPrompt()`, `ReadOutput()`, `GetState()`, `Ping()` on the same handle are safe.

---

## 3. Configuration Schema

### 3.1 File Format and Conventions

- **Main workspace config:** TOML — named by the user (e.g., `ralph.toml`, `ccat.toml`, `gastown.toml`)
- **Role definitions:** TOML files in `roles/*.toml` — operational settings per role
- **Prompt templates:** Markdown with Go templates in `roles/*.md.tmpl` — behavioral instructions
- **Formulas:** TOML files in `.beads/formulas/*.formula.toml`
- **Plugins:** Markdown with TOML frontmatter in `plugins/*.md`
- **Beads config:** YAML in `.beads/config.yaml`

**Detection logic:**
1. `*.toml` with `[workspace]` section exists → Gas City mode
2. `mayor/town.json` exists without Gas City config → Gas Town compatibility mode
3. Neither → fresh workspace, `gc init` required

**Environment variable expansion:** Values in `env` maps support `${VAR}` expansion from the host environment.

### 3.2 Full Schema

```toml
# === WORKSPACE IDENTITY ===
[workspace]
name = "string"                 # Required. Workspace identifier.
version = 1                     # Config schema version

# === AGENTS ===
[[agents]]
name = "string"                 # Required. Agent identifier (unique within workspace).
role = "string"                 # Role name — any string. Maps to roles/*.toml + roles/*.md.tmpl.
provider = "string"             # "claude", "codex", "gemini", "opencode", "cursor",
                                # "auggie", "amp", "subprocess"
                                # Default: auto-detected from available binaries.
scope = "project"               # "workspace" (one total) or "project" (one per project)
ephemeral = false               # Created/destroyed per-task (polecats)
isolation = "worktree"          # "none", "worktree", "directory"
depends_on = []                 # Agent names that must start before this one

[agents.session]                # Session management
pattern = "gc-{project}-{name}" # tmux session name pattern
work_dir = "{workspace}/{project}"
start_command = ""              # Override command (default from provider)
needs_pre_sync = false          # Git pull before session start

[agents.env]                    # Extra environment variables
GC_ROLE = "{role}"
GC_SCOPE = "{scope}"

[agents.pool]                   # Pool sizing for ephemeral agents
min = 0
max = 5
idle_timeout = "5m"

[agents.loop]                   # Task loop configuration
enabled = false
auto_execute = false            # GUPP: auto-start when work on hook
poll_interval = "10s"

[agents.health]                 # Health check thresholds
ping_timeout = "30s"
stuck_threshold = "1h"          # 1h (coordinator), 2h (workers), 4h (persistent)
consecutive_failures = 3
kill_cooldown = "5m"

[agents.hooks]                  # Lifecycle hooks (shell commands)
on_start = ""
on_stop = ""
on_task_assign = ""
on_task_complete = ""
on_stall = ""

# === PROJECTS (RIGS) ===
[projects.<name>]
repo = "string"                 # Git repository URL or "." for local
branch = "main"                 # Default branch

# === TASK SYSTEM ===
[tasks]
backend = "beads"               # "beads" or "filesystem"

[tasks.beads]
data_dir = ".beads"

# === MESSAGING ===
[messaging]
backend = "beads"               # "beads" or "filesystem"

# === FORMULAS ===
[formulas]
dir = ".beads/formulas"         # Directory containing *.formula.toml files

# === PLUGINS ===
[plugins]
dir = "plugins"                 # Directory containing *.md plugin files

# === DAEMON ===
[daemon]
websocket_port = 8765           # Websocket port for transparency streaming
patrol_interval = "3m"          # Deacon-style heartbeat interval
```

### 3.3 Defaults

Every setting has a reasonable default so users only configure what they need:

| Setting | Default | Notes |
|---------|---------|-------|
| `provider` | Auto-detected | Scans PATH for known binaries |
| `scope` | `"project"` | Most agents are project-scoped |
| `ephemeral` | `false` | Persistent by default |
| `isolation` | `"worktree"` | Git worktree (Gas Town's default) |
| `pool.min` | `0` | Scale to zero when idle |
| `pool.max` | `5` | Reasonable concurrency limit |
| `loop.poll_interval` | `"10s"` | Balance responsiveness vs overhead |
| `health.ping_timeout` | `"30s"` | |
| `health.stuck_threshold` | `"1h"` | |
| `health.consecutive_failures` | `3` | |
| `health.kill_cooldown` | `"5m"` | |
| `tasks.backend` | `"beads"` | Beads is the primary backend |
| `messaging.backend` | `"beads"` | |
| `daemon.websocket_port` | `8765` | |
| `daemon.patrol_interval` | `"3m"` | Gas Town's deacon heartbeat interval |

### 3.4 Config Validation Rules

| Rule | Error |
|------|-------|
| Agent names unique | "duplicate agent name: X" |
| Pool requires ephemeral | "pool config requires ephemeral = true" |
| Project scope requires projects | "project-scoped agent X needs at least one project" |
| Provider exists or auto-detectable | "unknown provider: X" |
| depends_on references exist | "agent X depends on unknown agent Y" |
| depends_on is acyclic | "dependency cycle: X → Y → X" |
| Role files exist (if referenced) | "role file not found: roles/X.toml" |
| Template files exist (if referenced) | "template not found: roles/X.md.tmpl" |

---

## 4. Progressive Capability Model

Each level adds one capability. Config grows; the SDK is constant.

| Level | What the User Configures | What Activates |
|-------|-------------------------|----------------|
| 0 | Single agent | `[[agents]]` — one agent, auto-detect provider |
| 1 | + Work tracking | Add `[tasks]` section — beads backend |
| 2 | + Task loop | Add `[agents.loop]` — agent polls for work |
| 3 | + Coordinator + workers | Multiple agents, worker pool with `[agents.pool]` |
| 4 | + Messaging | Add `[messaging]` — mail + nudge |
| 5 | + Formulas & molecules | Add `[formulas]` — workflow templates |
| 6 | + Health monitoring | Add supervisor agent with `[agents.health]` patrol config |
| 7 | Full orchestration | Multiple projects, all roles, multi-project formulas |

**Level detection** is automatic: the config parser examines which sections are present and determines the capability level. Each level is independently useful — you don't need Level 7 to benefit from Level 2.

---

## 5. Three Example Configs

### 5.1 ralph.toml — Single Agent with Task Loop (Level 2)

```toml
[workspace]
name = "my-project"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "ralph"

[agents.loop]
enabled = true
auto_execute = true
poll_interval = "30s"
```

~10 lines. Provider auto-detected. Beads for task tracking. Agent polls for work and executes it.

### 5.2 ccat.toml — Claude Code Agent Teams (Level 4)

```toml
[workspace]
name = "agent-teams"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[messaging]
backend = "beads"

# Coordinator — dispatches work
[[agents]]
name = "lead"
role = "coordinator"
scope = "workspace"

# Worker pool — executes tasks
[[agents]]
name = "devs"
role = "worker"
ephemeral = true
isolation = "worktree"
depends_on = ["lead"]

[agents.pool]
min = 0
max = 5

[agents.loop]
enabled = true
auto_execute = true
```

### 5.3 gastown.toml — Full Gas Town Replication (Level 7)

```toml
[workspace]
name = "my-town"
version = 1

# === Projects (Rigs) ===
[projects.gastown]
repo = "https://github.com/steveyegge/gastown"

[projects.beads]
repo = "https://github.com/steveyegge/beads"

# === Subsystems ===
[tasks]
backend = "beads"

[messaging]
backend = "beads"

[formulas]
dir = ".beads/formulas"

[plugins]
dir = "plugins"

[daemon]
patrol_interval = "3m"

# === Town-level agents ===

# Mayor — global coordinator
[[agents]]
name = "mayor"
role = "mayor"
scope = "workspace"

[agents.session]
pattern = "hq-mayor"

[agents.health]
stuck_threshold = "1h"

# Deacon — daemon beacon, health monitoring
[[agents]]
name = "deacon"
role = "deacon"
scope = "workspace"

[agents.session]
pattern = "hq-deacon"

[agents.health]
stuck_threshold = "1h"

# Dogs — town-level infrastructure workers
[[agents]]
name = "dogs"
role = "dog"
scope = "workspace"
ephemeral = true
depends_on = ["deacon"]

[agents.pool]
min = 0
max = 3

# === Per-project agents ===

# Witness — per-rig worker monitor
[[agents]]
name = "witness"
role = "witness"
scope = "project"
depends_on = ["mayor"]

[agents.session]
pattern = "gt-{project}-witness"

# Refinery — merge queue processor
[[agents]]
name = "refinery"
role = "refinery"
scope = "project"
depends_on = ["mayor"]

[agents.session]
pattern = "gt-{project}-refinery"

[agents.health]
stuck_threshold = "2h"

# Polecats — ephemeral batch workers
[[agents]]
name = "polecats"
role = "polecat"
scope = "project"
ephemeral = true
isolation = "worktree"
depends_on = ["mayor", "witness"]

[agents.pool]
min = 0
max = 5

[agents.loop]
enabled = true
auto_execute = true

[agents.health]
stuck_threshold = "2h"

# Crew — persistent per-rig workspace agents
[[agents]]
name = "crew"
role = "crew"
scope = "project"

[agents.session]
pattern = "gt-{project}-crew-{name}"

[agents.health]
stuck_threshold = "4h"
```

---

## 6. Role System

### 6.1 The Three-Part Role Stack

Each role's behavior is defined by a three-part stack, all user-supplied:

1. **TOML config** (`roles/*.toml`): Operational settings — session patterns, health thresholds, env vars, nudge message, prompt template reference
2. **Markdown Go template** (`roles/*.md.tmpl`): Full behavioral instructions — rendered at session start. Contains philosophy, capabilities, startup protocol, work guidance, command reference
3. **Workspace TOML** (e.g., `gastown.toml`): Agent entries reference roles and override settings

The SDK provides ONLY template rendering and config loading. All role behavior lives in user-supplied files.

### 6.2 Role Definition Format

```toml
# roles/polecat.toml
role = "polecat"
scope = "project"
nudge = "You have work on your hook. Check gc mol status."
prompt_template = "polecat.md.tmpl"

[session]
pattern = "gt-{project}-{name}"
work_dir = "{workspace}/{project}"
needs_pre_sync = true
start_command = "exec claude --dangerously-skip-permissions"

[env]
GC_ROLE = "polecat"
GC_SCOPE = "project"

[health]
ping_timeout = "30s"
consecutive_failures = 3
kill_cooldown = "5m"
stuck_threshold = "2h"
```

### 6.3 Prompt Template Format

```markdown
{{/* roles/polecat.md.tmpl */}}
# {{.Role}} Worker Context — {{.Project}}

> **Context Recovery**: Run `gc prime` for full context after compaction.

## The Propulsion Principle (GUPP)

**If you find work on your hook, YOU RUN IT.**
No confirmation. No waiting. The hook having work IS the assignment.

## Startup Protocol

1. Check your hook: `gc mol status`
2. If work is hooked → EXECUTE
3. If hook empty → Check mail: `gc mail inbox`
4. Still nothing? Wait for instructions.

## Session End Checklist

1. git status (check what changed)
2. git add <files>
3. git commit -m "..."
4. git push
5. gc mol step done (if working on molecule)

**Work is not done until pushed.**
```

Templates use Go `text/template` with variables: `.Role`, `.Project`, `.Workspace`, `.AgentName`, `.Instance`, and any custom variables from the role's `[env]` section.

### 6.4 Override Resolution

Role settings cascade with specificity:

```
SDK defaults (reasonable defaults for every setting)
  → Workspace-level role definitions (roles/*.toml)
    → Project-level overrides (<project>/roles/*.toml)
      → Inline settings in workspace TOML ([agents.health], etc.)
```

More specific settings win. This lets you define a base `polecat.toml` role and override `stuck_threshold` for a specific project.

### 6.5 Gas Town Role Reference

These roles ship as example files with the SDK, not as hardcoded behavior:

| Role | Scope | Behavior (defined in prompt template) |
|------|-------|---------------------------------------|
| mayor | workspace | Global coordinator. Dispatches tasks, breaks down epics, manages projects. |
| deacon | workspace | Daemon beacon. Receives heartbeats, watches witnesses, manages dogs. |
| dog | workspace | Workspace-level infrastructure worker. Cross-project tasks, cleanup, maintenance. |
| witness | project | Per-project monitor. Tracks worker progress, detects stalls, reports to deacon. |
| refinery | project | Merge queue processor. Verification gates, conflict resolution. |
| polecat | project | Ephemeral batch worker. Executes individual tasks, self-cleans on completion. |
| crew | project | Persistent workspace agent. Long-lived, direct push to main. |

---

## 7. Workspace Controller / Daemon

### 7.1 Controller Lifecycle

The workspace controller is the long-lived daemon that hosts all control loops.

1. **`gc start`** launches the controller (foreground by default, `--daemon` to background)
2. Acquires `.gc/controller.lock` via `flock` — at most one controller per workspace
3. Starts event bus, loads providers, runs startup sequencer
4. Starts control loops: pool managers, patrol cycle, hook executor
5. Opens websocket for transparency streaming
6. **`gc stop`** sends shutdown signal, runs shutdown sequencer

**Crash recovery:** Agents are independent tmux sessions that survive controller crashes. On restart, the controller rediscovers running agents via the persisted registry (`.gc/agents/*.json`) and reconnects via the `Adopter` interface.

### 7.2 Startup Sequencer

Agents start according to a dependency DAG built from `depends_on`. Within a dependency tier, agents start in parallel. If any agent fails, all previously started agents are stopped in reverse order (rollback).

**Ordering invariant:** For agents `a` and `b` where `a` is in `b.depends_on`: `IsRunning(a) = true` before `Start(b)` is called.

### 7.3 Shutdown Sequencer

Reverse of startup. Workers stop first (before monitors try to restart them).

- `gc stop` (graceful): Waits for in-progress tasks up to `--drain-timeout` (default 5m), then re-queues.
- `gc stop --force`: Stops agents immediately, re-queues in-progress tasks.

A `shuttingDown` flag prevents the health monitor from restarting agents during shutdown.

### 7.4 Websocket Transparency

**Vision requirement:** "GC is completely transparent and can be traced via the daemon that provides a websocket that shows the historical data and streaming changes to data."

The daemon provides a websocket endpoint that streams all system activity:
- Agent request/response pairs (prompts sent, output captured)
- Mail messages sent and received
- Bead state changes (created, hooked, closed)
- Agent active status (started, stopped, stalled)
- Formula/molecule progress
- Health patrol results

Clients connect and receive a catch-up replay of recent history, then real-time streaming. Three built-in consumers:

- **`gc dashboard`** — Web UI consuming websocket
- **`gc dashboard --tui`** — Terminal UI consuming websocket
- **`gc activity --follow`** — Raw event stream

---

## 8. Event Bus

### 8.1 Event Types

```go
type EventType string
const (
    // Agent lifecycle
    EventAgentStarted   EventType = "agent.started"
    EventAgentStopped   EventType = "agent.stopped"
    EventAgentStalled   EventType = "agent.stalled"
    EventAgentCrashed   EventType = "agent.crashed"

    // Tasks / beads
    EventBeadCreated    EventType = "bead.created"
    EventBeadHooked     EventType = "bead.hooked"
    EventBeadClosed     EventType = "bead.closed"

    // Health
    EventHealthPingOK   EventType = "health.ping_ok"
    EventHealthPingFail EventType = "health.ping_fail"
    EventHealthRestart  EventType = "health.restart"

    // Messaging
    EventMailSent       EventType = "mail.sent"
    EventMailRead       EventType = "mail.read"
    EventNudgeSent      EventType = "nudge.sent"

    // Formulas / molecules
    EventMolCreated     EventType = "molecule.created"
    EventMolStepDone    EventType = "molecule.step_done"
    EventMolCompleted   EventType = "molecule.completed"

    // Pool
    EventPoolScaleUp    EventType = "pool.scale_up"
    EventPoolScaleDown  EventType = "pool.scale_down"

    // Convoy
    EventConvoyCreated  EventType = "convoy.created"
    EventConvoyClosed   EventType = "convoy.closed"
)
```

### 8.2 Tiered Subscriber Model

The event bus has two tiers:
- **Critical subscribers** (supervisor, structured logger, hook executor): Block `Publish()` — if they fail, the event is not silently lost.
- **Optional subscribers** (websocket streamer, CLI feed, metrics): Fire-and-forget. Slow consumers may miss events but never stall the bus.

A ring buffer (10k events) provides catch-up replay when new subscribers connect (e.g., websocket clients joining late).

**Sequence monotonicity:** Events have monotonically increasing sequence numbers, enabling clients to detect gaps and request replay.

---

## 9. Health Monitoring / Supervisor

### 9.1 Patrol Cycle

Health monitoring follows Gas Town's Deacon patrol pattern — configured via role config, not hardcoded.

The patrol cycle runs at `daemon.patrol_interval` (default 3m):

```
FOR each running agent (skip self):
  result = Ping(agent)
  IF ping fails:
    agent.consecutive_failures++
  ELSE:
    agent.consecutive_failures = 0

  IF consecutive_failures >= agent.health.consecutive_failures:
    IF time since last restart < kill_cooldown:
      publish EventAgentStalled  // In cooldown, don't restart yet
    ELSE:
      Restart(agent)
      publish EventHealthRestart
```

### 9.2 Stall Detection

An agent is stalled when:
- Status = Working for longer than `stuck_threshold`
- AND Ping() returns OK (alive but not progressing)

Stall thresholds from Gas Town defaults:
- Mayor/Deacon: 1h
- Polecat/Refinery: 2h
- Crew: 4h

### 9.3 Who Runs the Patrol?

In a full Gas Town config, the deacon agent runs the patrol. In simpler configs, the daemon itself runs the patrol loop. The user controls this via config — if a supervisor-like agent exists with health config, it takes over monitoring.

---

## 10. Task System / Beads

### 10.1 Beads-First Design

Gas Town uses beads (Dolt-backed structured data) as its primary task system. Gas City follows suit — beads is the default and recommended backend. A filesystem backend exists as a zero-dependency fallback for development and testing.

### 10.2 Bead Types

Bead types are **not hardcoded into the SDK**. They are part of the workspace's configuration — users can define whatever types make sense for their orchestration. The SDK treats the type field as an opaque string; specific semantics (like the hook slot on agent beads, or step children on molecule beads) are implemented as behaviors keyed on type, not as a fixed enum.

Gas Town's bead taxonomy (shipped as examples):

| Type | Purpose |
|------|---------|
| `task` | Regular work item |
| `bug` | Bug report |
| `feature` | Feature request |
| `agent` | Agent identity bead (has hook slot) |
| `molecule` | Instantiated formula (root bead with step children) |
| `convoy` | Batch tracking bead |
| `mail` | Mail message |
| `merge-request` | Merge queue entry |

Users can define additional types as needed. The SDK's subsystems (formulas, convoys, mail) create beads of the appropriate type but don't restrict what types can exist.

### 10.3 Bead Status Lifecycle

Gas Town's actual status model:

```
open → hooked → closed
```

- **open**: Available for claiming
- **hooked**: Pinned to an agent's hook (being worked on)
- **closed**: Completed

The `hooked` state is Gas Town's equivalent of "in-progress" — a bead is physically attached to an agent's hook bead via a dependency relationship. Only one agent can hook a bead at a time (atomic claiming via Dolt SQL transactions).

### 10.4 Task Backend Interface

```go
type TaskBackend interface {
    List(filter TaskFilter) ([]Bead, error)
    Get(id string) (Bead, error)
    Create(bead Bead) (string, error)
    Update(id string, updates BeadUpdates) error
    Hook(beadID string, agentID string) error     // Atomic: attach to agent's hook
    Unhook(beadID string, agentID string) error
    Close(beadID string, agentID string) error
    Ready() ([]Bead, error)                        // Beads with no blockers
}
```

### 10.5 Task Loop

The task loop (enabled via `[agents.loop]`) polls for available work:

```
LOOP:
  IF check hook → work exists:
    execute hooked work
    GOTO LOOP
  IF check mail → messages exist:
    process mail (may contain assignments)
    GOTO LOOP
  ready = backend.Ready()
  IF len(ready) > 0:
    bead = ready[0]  // highest priority
    ok = backend.Hook(bead.id, self.id)  // Atomic claim
    IF !ok: GOTO LOOP  // Someone else got it
    execute bead
    GOTO LOOP
  sleep(poll_interval)
  GOTO LOOP
```

**Claim atomicity:** `Hook()` is atomic. For beads, this uses Dolt SQL transactions. For filesystem, this uses flock. At most one agent can hook a bead.

---

## 11. Formulas

### 11.1 Overview

A formula is a static template that defines a structured workflow. Gas Town has four formula types, all defined as TOML files in `.beads/formulas/`.

### 11.2 Formula Types

**Workflow** — Sequential steps with DAG dependencies:

```toml
formula = "shiny"
description = "Feature development pipeline"
type = "workflow"
version = 1

[[steps]]
id = "design"
title = "Design {{feature}}"
description = "Create design document for the feature"

[[steps]]
id = "implement"
title = "Implement {{feature}}"
description = "Write the implementation code"
needs = ["design"]

[[steps]]
id = "test"
title = "Test {{feature}}"
description = "Write and run tests"
needs = ["implement"]

[vars]
[vars.feature]
description = "The feature being implemented"
required = true
```

**Convoy** — Parallel legs with synthesis:

```toml
formula = "code-review"
description = "Multi-aspect parallel code review"
type = "convoy"
version = 1

[inputs]
[inputs.pr]
description = "Pull request number"
type = "number"

[prompts]
base = "Review PR #{{.pr}} focusing on {{.leg.focus}}"

[[legs]]
id = "correctness"
title = "Correctness Review"
focus = "Logic errors, bugs, edge cases"

[[legs]]
id = "security"
title = "Security Review"
focus = "Injection, auth, data exposure"

[[legs]]
id = "performance"
title = "Performance Review"
focus = "Algorithmic complexity, resource leaks"

[synthesis]
title = "Combine findings from all review legs"
depends_on = ["correctness", "security", "performance"]

[output]
directory = ".reviews/{{.review_id}}"
leg_pattern = "{{.leg.id}}-findings.md"
synthesis = "review-summary.md"
```

**Expansion** — Pre-computed step sequences:

```toml
formula = "tower-of-hanoi"
description = "Durability proof via pre-computed steps"
type = "expansion"
version = 1

[[steps]]
id = "move-1"
title = "Move disk 1: A → C"

[[steps]]
id = "move-2"
title = "Move disk 2: A → B"
needs = ["move-1"]
```

**Aspect** — Multi-aspect parallel analysis (similar to convoy but without synthesis):

```toml
formula = "security-audit"
description = "Parallel security analysis"
type = "aspect"
version = 1

[[aspects]]
id = "sast"
title = "Static Analysis"
description = "Run static analysis tools"

[[aspects]]
id = "secrets"
title = "Secrets Scanning"
description = "Check for leaked credentials"
```

### 11.3 Formula Validation Pipeline

```
Parse TOML → Infer type (if ambiguous) → Validate schema per type → Cycle detection (workflow/expansion) → Variable resolution
```

- Workflow/expansion: Steps form a DAG. Cycle detection via topological sort.
- Convoy: Legs are parallel by definition; synthesis depends on all legs.
- Variables: `{{feature}}` resolved at instantiation time via `--var key=value`.

### 11.4 Formula Discovery

Formulas are discovered from:
1. `[formulas].dir` path (default: `.beads/formulas/`)
2. Files matching `*.formula.toml`

Gas Town ships ~32 formulas including `code-review`, `mol-polecat-code-review`, `mol-deacon-patrol`, `mol-convoy-cleanup`, `towers-of-hanoi-10` (1023-step durability proof).

---

## 12. Molecules

### 12.1 Overview

A molecule is an **instantiated formula** — a root bead with child step beads forming a DAG. Molecules are the runtime representation of formulas.

### 12.2 Lifecycle

```
Formula (static template)
  → [sling] → Molecule (root bead + child step beads)
    → [attach to hook] → Agent's hook bead gets molecule reference
      → [execute steps] → Steps completed via `gc mol step done`
        → [squash/burn] → Molecule finalized or discarded
```

### 12.3 Step Structure

Each step in a molecule is a child bead with:

| Field | Description |
|-------|-------------|
| `ref` | Step reference ID (from formula) |
| `title` | Step title (with variables resolved) |
| `instructions` | Prose instructions for the agent |
| `needs` | Dependencies (step refs that must complete first) |
| `tier` | Model tier hint: haiku, sonnet, opus |
| `type` | "task" (default) or "wait" (polling step with backoff) |
| `waits_for` | Dynamic wait conditions (e.g., "all-children") |
| `backoff` | For wait-type steps: base, multiplier, max |

### 12.4 DAG Execution

- **Ready steps** = all dependencies (`needs`) complete
- **Execution tier** = topological depth in the DAG
- **Progress** = count of completed steps / total steps
- **Crash recovery**: Molecule state persists in beads. On agent restart, `gc mol status` shows current state and next ready step.

### 12.5 Molecule Commands

| Command | Description |
|---------|-------------|
| `gc mol status` | What's on my hook? Show molecule + current step. |
| `gc mol current` | What step should I work on next? |
| `gc mol step done [step-id]` | Complete step, auto-find next ready step. |
| `gc mol progress <molecule-id>` | DAG progress view (completed/total). |
| `gc mol dag <molecule-id>` | Visualize dependency graph. |
| `gc mol attach <molecule-id>` | Link molecule to agent's hook. |
| `gc mol detach` | Unlink molecule from hook. |
| `gc mol squash` | Finalize molecule as digest (permanent record). |
| `gc mol burn` | Discard molecule (no record). |

---

## 13. Convoys

### 13.1 Overview

A convoy is a persistent tracking bead that groups related issues across projects. It enables batch progress tracking and coordinated completion.

### 13.2 Key Concepts

- A convoy is a bead of type `convoy` at the town level
- Tracks multiple issues across projects via non-blocking "tracks" dependency
- Convoy status is derived from its tracked issues
- Auto-closes when all tracked issues close
- Can detect "stranded" convoys (ready work, no active workers)

### 13.3 Auto-Convoy

When slinging work via `gc sling`, a convoy is automatically created unless `--no-convoy` is passed. This groups related work for tracking.

### 13.4 Convoy Commands

| Command | Description |
|---------|-------------|
| `gc convoy create "name" issue1 issue2...` | Create convoy tracking specified issues. |
| `gc convoy list` | Dashboard of active convoys with progress. |
| `gc convoy status <id>` | Detailed progress for a convoy. |
| `gc convoy add <convoy-id> <issue-id>` | Add issue to existing convoy. |
| `gc convoy close <id>` | Manually close a convoy. |
| `gc convoy check` | Auto-close convoys where all issues are done. |
| `gc convoy stranded` | Find convoys with ready work but no active workers. |

---

## 14. Sling

### 14.1 Overview

Sling is THE unified work dispatch command. It creates molecules from formulas and assigns them to agents.

### 14.2 Flow

1. **Resolve source**: Is it a bead ID or a formula name?
2. **Resolve target**: Which agent? Auto-spawn polecat if needed.
3. **Cook** (if formula): Instantiate formula → create molecule with step children, resolve variables.
4. **Hook**: Attach bead/molecule to agent's hook bead.
5. **Nudge**: Send prompt to agent's tmux session to start working.
6. **Auto-convoy**: Create convoy unless `--no-convoy`.

### 14.3 Target Resolution

```
gc sling <bead-or-formula> [target]

# Self-assignment (hook to current agent)
gc sling gt-abc

# Specific role target
gc sling gt-abc crew
gc sling gt-abc mayor

# Project-scoped target (auto-spawn polecat)
gc sling gt-abc myproject

# Specific named worker
gc sling gt-abc myproject/Toast

# Dog pool
gc sling gt-abc deacon/dogs
```

### 14.4 Sling Flags

| Flag | Description |
|------|-------------|
| `--create` | Create polecat if target doesn't exist |
| `--force` | Ignore unread mail on target |
| `--agent <preset>` | Override provider (e.g., `--agent gemini`) |
| `--no-convoy` | Skip auto-convoy creation |
| `--no-merge` | Skip merge queue on completion |
| `--var key=value` | Formula variable substitution |
| `--on <bead>` | Apply formula to an existing bead |
| `--max-concurrent N` | Limit concurrent spawns in batch mode |
| `--base-branch <branch>` | Override base branch for worktree |

---

## 15. Messaging

### 15.1 Two Mechanisms

Gas Town has two distinct messaging mechanisms:

**Mail** — Asynchronous, persistent, Dolt-backed queue:
- Priority levels: `low`, `normal` (default), `high`
- Address format: `<project>/crew/<name>`, `mayor/`, `deacon/`, `<project>/witness`
- Stored as beads with type "mail"
- Persistent across restarts

**Nudge** — Synchronous, immediate delivery via tmux `send-keys`:
- Literal text injected directly into agent's tmux session
- 500ms paste delay + separate Enter keystroke
- DND (Do Not Disturb) support (skip unless `--force`)
- Used for waking agents, delivering prompts, sending work notifications

### 15.2 Mail Commands

| Command | Description |
|---------|-------------|
| `gc mail inbox` | Check messages |
| `gc mail read <id>` | Read specific message |
| `gc mail send <addr> -s "Subject" -m "Body"` | Send mail |
| `gc mail send --human` | Send to human overseer |
| `gc mail mark-read <id>` | Mark as read |
| `gc mail hook <mail-id>` | Hook mail as assignment |
| `gc mail archive` | Archive old mail |

### 15.3 Nudge Commands

| Command | Description |
|---------|-------------|
| `gc nudge <target> [message]` | Send nudge to agent |
| `gc nudge <target> -m "message"` | Nudge with explicit message |
| `gc nudge <target> --stdin` | Read nudge from stdin |
| `gc nudge --if-fresh <target> "msg"` | Only nudge if session < 60s old |
| `gc broadcast <message>` | Town-wide broadcast to all agents |

### 15.4 Nudge Target Shortcuts

| Shortcut | Resolves To |
|----------|-------------|
| `mayor` | `hq-mayor` session |
| `deacon` | `hq-deacon` session |
| `witness` | `gt-{project}-witness` session |
| `refinery` | `gt-{project}-refinery` session |
| `channel:<name>` | All members of named channel |

---

## 16. Plugin System

### 16.1 Plugin Format

Plugins are Markdown files with TOML frontmatter. They define automated actions with gate conditions that control when they can fire.

```markdown
+++
name = "rebuild-gt"
description = "Rebuild stale binary from source"
version = 1

[gate]
type = "cooldown"
duration = "1h"

[tracking]
labels = ["plugin:rebuild-gt", "category:maintenance"]
digest = true

[execution]
timeout = "5m"
notify_on_failure = true
severity = "medium"
+++

# Rebuild Gas Town Binary

Check if the `gt` binary is stale (older than source files).
If so, rebuild from source:

1. `cd /path/to/gastown`
2. `go build -o gt ./cmd/gt`
3. Verify: `gt version`
```

### 16.2 Gate Types

| Gate | Description | Config |
|------|-------------|--------|
| `cooldown` | Minimum time between firings | `duration = "1h"` |
| `cron` | Fire on schedule | `schedule = "0 */6 * * *"` |
| `condition` | Fire when condition is true | `check = "command to evaluate"` |
| `event` | Fire on specific event | `event = "bead.closed"` |
| `manual` | Only fire when explicitly triggered | (no extra config) |

### 16.3 Plugin Locations

- **Workspace-level:** `plugins/` — applies to all projects
- **Project-level:** `<project>/plugins/` — applies to specific project

---

## 17. CLI Design

### 17.1 Command Structure

The Gas City CLI is `gc`.

```
# Workspace lifecycle
gc init [--file ralph.toml|ccat.toml|gastown.toml]  # Init from example config
gc init                                               # Interactive wizard
gc start [--daemon]
gc stop [--force] [--drain-timeout 5m]
gc status
gc level                                              # Show capability level

# Agent management
gc agent list
gc agent start <name>
gc agent stop <name>
gc agent restart <name>
gc agent attach <name>                                # tmux attach
gc agent logs <name> [--follow]

# Work dispatch
gc sling <bead-or-formula> [target] [--var key=value]
gc hook <bead-id>                                     # Attach to own hook
gc handoff <bead-id>                                  # Hook + restart with fresh context

# Molecule management
gc mol status
gc mol current
gc mol step done [step-id]
gc mol progress <molecule-id>
gc mol dag <molecule-id>
gc mol attach <molecule-id>
gc mol detach
gc mol squash
gc mol burn

# Formula management
gc formula list
gc formula show <name>
gc formula create <name>

# Convoy tracking
gc convoy list
gc convoy create "name" <issues...>
gc convoy status <id>
gc convoy add <convoy-id> <issue-id>
gc convoy close <id>
gc convoy check
gc convoy stranded

# Messaging
gc mail inbox
gc mail read <id>
gc mail send <addr> -s "Subject" -m "Body"
gc mail send --human                                  # Send to human overseer
gc mail mark-read <id>
gc mail hook <mail-id>                                # Hook mail as assignment
gc mail archive                                       # Archive old mail
gc nudge <target> [message]
gc broadcast <message>

# Beads (task management)
gc task list [--status=X]
gc task create <title>
gc task show <id>
gc task ready                                         # List ready (unblocked) beads

# Monitoring / transparency
gc dashboard                                          # Web UI
gc dashboard --tui                                    # Terminal UI
gc activity [--follow]                                # Event stream
gc stats                                              # Agent metrics

# Utility
gc validate                                           # Config validation
gc prime                                              # Render role prompt for current agent
gc migrate [--dry-run]                                # Generate config from Gas Town workspace
gc doctor                                             # Health checks
gc config show                                        # Display resolved config
gc test-provider <name>                               # Run conformance suite against a provider
gc version
```

### 17.2 `gc init --file`

```
$ gc init --file ralph.toml
Copied ralph.toml to ./ralph.toml
Run `gc start` to begin.

$ gc init
Welcome to Gas City SDK!

Example configs available:
  > ralph.toml      -- Single agent with task loop (simplest)
    ccat.toml       -- Coordinator + worker pool (Agent Teams)
    gastown.toml    -- Full multi-project orchestration (Gas Town)
    (custom)        -- Start from scratch

Select [ralph.toml]:

Which coding agent do you use?
  > Claude Code
    Codex (OpenAI)
    Gemini CLI
    OpenCode
    Other

Select [Claude Code]:

Created ralph.toml (Level 2 - Task Loop)
Run `gc start` to begin.
```

---

## 18. Identity and Addressing

### 18.1 Address Format

Gas City uses a hierarchical addressing scheme matching Gas Town's patterns:

```
<workspace>/<agent-name>                    # Workspace-scoped
<workspace>/<project>/<agent-name>          # Project-scoped
<workspace>/<project>/<agent-name>[<idx>]   # Pool instance
```

### 18.2 Gas Town Address Compatibility

| Gas Town Format | Gas City Format |
|----------------|-----------------|
| `mayor/` | `<workspace>/mayor` |
| `deacon/` | `<workspace>/deacon` |
| `<rig>/crew/<name>` | `<workspace>/<project>/<name>` |
| `<rig>/polecats/<name>` | `<workspace>/<project>/polecats[<idx>]` |
| `<rig>/witness` | `<workspace>/<project>/witness` |
| `<rig>/refinery` | `<workspace>/<project>/refinery` |

### 18.3 Agent Registry

The agent registry maps logical identities to runtime handles, persisted to `.gc/agents/*.json` for crash recovery. On controller restart, it reconnects to surviving tmux sessions via the `Adopter` interface.

---

## 19. Lifecycle Hooks

### 19.1 Hook Events

Hooks are shell commands triggered by lifecycle events:

| Hook | Trigger | Template Variables |
|------|---------|-------------------|
| `on_start` | Agent started | `{{.Agent}}`, `{{.Project}}` |
| `on_stop` | Agent stopped | `{{.Agent}}`, `{{.Project}}` |
| `on_task_assign` | Bead hooked to agent | `{{.Agent}}`, `{{.Task}}`, `{{.TaskTitle}}` |
| `on_task_complete` | Bead closed | `{{.Agent}}`, `{{.Task}}` |
| `on_stall` | Agent detected stalled | `{{.Agent}}`, `{{.Project}}` |

### 19.2 Execution Model

Hooks run as shell commands via `sh -c` with a default 30s timeout. They are executed by the hook executor, which is a critical event bus subscriber — hook failures are not silently dropped.

---

## 20. Migration from Gas Town

### 20.1 Migration Algorithm

```
1. Check for active sessions → warn if found
2. Read mayor/town.json → [workspace] name
3. Read mayor/rigs.json → [projects.*] entries
4. Read settings/config.json → provider settings, role agent mappings
5. Read internal/config/roles/*.toml → copy to roles/ directory
6. Read internal/templates/roles/*.md.tmpl → copy to roles/ directory
7. Read .beads/formulas/*.formula.toml → keep in place (same format)
8. Read plugins/*.md → keep in place (same format)
9. Map Gas Town roles to [[agents]] entries
10. Generate workspace TOML file
11. Flag Claude hooks for manual review
```

### 20.2 File-by-File Migration Map

| Gas Town File | Gas City Target | Action |
|--------------|-----------------|--------|
| `mayor/town.json` | `[workspace] name` | Extract name |
| `mayor/rigs.json` | `[projects.*]` | One entry per rig |
| `settings/config.json` | Various sections | Map settings |
| `internal/config/roles/*.toml` | `roles/*.toml` | Copy (same format) |
| `internal/templates/roles/*.md.tmpl` | `roles/*.md.tmpl` | Copy (same format) |
| `.beads/formulas/*.formula.toml` | Same location | No change needed |
| `.beads/config.yaml` | Same location | No change needed |
| `plugins/*.md` | Same location | No change needed |
| `.claude/settings.json` | Flag for review | Claude hooks stay as-is |

### 20.3 Compatibility

1. Gas Town workspaces without Gas City config continue to work with `gt` unchanged.
2. `gc migrate --dry-run` generates equivalent config without modifying anything.
3. After migration, both `gt` and `gc` commands work.
4. Migration is additive — creates Gas City config alongside existing Gas Town files.

---

## 21. Testing Strategy

### 21.1 Unit Tests

| Subsystem | What It Tests | Mock Boundary |
|-----------|---------------|---------------|
| Config parser | TOML parsing, level detection, validation, defaults | Filesystem (embed test TOML) |
| Provider registry | Registration, lookup, auto-detection | None (pure logic) |
| Pool manager | Scale up/down, min/max bounds, idle timeout | AgentProvider mock |
| Task loop | Hook/claim cycle, mail check, poll interval | AgentProvider + TaskBackend mocks |
| Startup sequencer | DAG ordering, parallel start, rollback | AgentProvider mock |
| Event bus | Pub/sub, critical vs optional tiers, replay | None (pure concurrency) |
| Formula parser | All 4 types, variable resolution, cycle detection | None (pure parsing) |
| Molecule executor | DAG step ordering, ready detection, crash recovery | TaskBackend mock |
| Hook executor | Template expansion, timeout, error propagation | EventBus (inject events) |
| Identity resolver | Address parsing, Gas Town compat | None (pure logic) |
| Migration | Town→City conversion, edge cases, report | Filesystem (test fixtures) |

### 21.2 Contract Tests

`gc test-provider` runs a conformance suite against any registered provider, testing:
- Start returns valid handle
- IsRunning true after Start
- Stop makes IsRunning false
- Double Stop is idempotent (P1)
- SendPrompt delivers content
- Ping within timeout (P2)
- Attach matches SupportsAttach (P3)
- Concurrent access is safe (P5)
- Start creates unique handles

### 21.3 Integration Tests

| Test | What It Validates |
|------|-------------------|
| Provider roundtrip | Start → SendPrompt → ReadOutput → Stop |
| Config → startup → shutdown | Parse TOML → start all → verify → stop all |
| Ralph smoke test | Single agent processes one task end-to-end |
| Agent Teams smoke test | Coordinator dispatches to worker pool |
| Formula → molecule → execute | Sling formula → molecule created → steps execute |
| Convoy lifecycle | Create → track issues → auto-close |
| Migration test | Gas Town workspace → gc migrate → verify config |
| Websocket streaming | Connect → receive events → verify catch-up replay |

---

## 22. Implementation Phases

### Phase 1: Agent Protocol + Config (3 weeks)

- AgentProvider interface with `SendPrompt`
- Provider implementations: `claude` (tmux), `subprocess` (generic)
- TOML config parser with level detection and validation
- Identity and addressing (agent registry, address resolution)
- `gc init --file`, `gc start`, `gc stop`, `gc status`, `gc level`, `gc validate`
- Contract test suite (`gc test-provider`)

### Phase 2: Task System + Ralph (2 weeks)

- TaskBackend interface (beads + filesystem)
- Bead status lifecycle (open → hooked → closed)
- Task loop with atomic Hook/Claim
- Role system: TOML config loading, template rendering, override resolution
- `gc task`, `gc prime`
- Ralph config working end-to-end

### Phase 3: Pool + Messaging + Agent Teams (2 weeks)

- Pool manager with min/max bounds
- Startup/shutdown sequencer with DAG ordering
- Mail system (beads-backed, priority levels)
- Nudge system (tmux send-keys)
- Lifecycle hooks (event-driven shell commands)
- `gc mail`, `gc nudge`, `gc broadcast`
- ccat.toml (Agent Teams) working end-to-end

### Phase 4: Formulas + Molecules + Sling (2 weeks)

- Formula parser (all 4 types: workflow, convoy, expansion, aspect)
- Molecule instantiation (root bead + child step beads)
- Molecule execution (DAG step ordering, ready detection)
- Sling dispatch command
- `gc formula`, `gc mol`, `gc sling`

### Phase 5: Full Gas Town + Migration (3 weeks)

- Convoy tracking system
- Plugin system (Markdown + TOML frontmatter, gate types)
- Health monitoring / patrol cycle
- Event bus with websocket streaming
- Dashboard (web + TUI)
- Migration tooling
- `gc convoy`, `gc dashboard`, `gc activity`, `gc migrate`
- gastown.toml working end-to-end
- All remaining providers: codex, gemini, opencode, cursor, auggie, amp

---

## 23. Exclusions

| Exclusion | Rationale |
|-----------|-----------|
| Web dashboard (standalone) | Event bus + websocket provide the data API. Dashboard consumes it. Separate deliverable. |
| Cloud/remote execution | v1 is local-only. All providers run locally. |
| Multi-tenant isolation | One workspace per user. |
| Agent marketplace | SDK provides config format; sharing is orthogonal. |
| Hot config reload | v1 requires restart. Watch-and-diff is v2. |
| GitHub Issues / Linear / Jira backends | Gas Town uses beads. Filesystem is the dev fallback. External issue trackers are a future integration. |
| Redis messaging | Gas Town uses beads-backed mail. No Redis needed. |
| Docker/container isolation | Gas Town uses tmux + worktrees. Container isolation is a future provider. |
| Formal coordination grammar/DSL | Coordination is expressed via config + prompt templates + formulas. No need for a custom DSL in v1. |

---

## 24. Formal Properties

### 24.1 Task Claim Atomicity

For any bead `b` with status `open`, at most one agent can successfully call `Hook(b.id, agent)` and transition it to `hooked`. All other concurrent hooks receive an error. Enforced by Dolt SQL transactions (beads) or flock (filesystem).

### 24.2 Startup Ordering (DAG)

For agents `a` and `b` where `a` is in `b.depends_on`:
```
IsRunning(a) = true  BEFORE  Start(b) is called
```

Enforced by topological group execution. Cycles detected at config validation time.

### 24.3 Pool Bounds

For a pool with `{min: M, max: N}`, at any time `t`:
```
M <= |running_instances(t)| <= N
```

Enforced by serialized scale operations under mutex.

### 24.4 Startup Rollback Safety

If `StartWorkspace()` fails at agent `k` of `n`, agents `1..k-1` are stopped in reverse order. No orphaned agents after failed startup.

### 24.5 Molecule DAG Correctness

A molecule step `s` with dependencies `D = {d1, d2, ...dk}` becomes ready if and only if all dependencies are complete:
```
ready(s) ⟺ ∀di ∈ D: status(di) = closed
```

---

## 25. Open Questions

1. **`gc` vs `gt` CLI namespace.** Options: (a) `gc` as separate binary, (b) `gt gc` subcommand, (c) `gt` detects Gas City config and adjusts. Recommendation: (a) separate binary.

2. **Plugin system scope.** The plan includes plugins with gate types. Gas Town's plugin system may still be evolving. Confirm scope and gate type implementations before Phase 5.

3. **Websocket protocol.** Define the exact message format for websocket streaming (JSON-lines? Protobuf?). JSON-lines is simplest and matches Gas Town's existing JSONL patterns.

4. **Formula preset system.** Gas Town's convoy formulas support presets (e.g., `[presets.gate]` for light code review). Determine if presets are v1 or v2.

5. **Non-interactive mode for providers.** Several providers support non-interactive/batch modes (e.g., `codex exec`, `gemini -p`). Define when to use interactive vs batch mode.

6. **Config hot-reload.** Can you add agents to a running workspace? v1: no (restart required). v2: file watch + diff.

---

## Appendix A: Vision Alignment Checklist

| Vision Requirement | Spec Section | Status |
|-------------------|-------------|--------|
| Orchestration-builder toolkit | §1 Executive Summary | Covered |
| Multiple town shapes via config | §5 Three Example Configs | ralph, ccat, gastown |
| Progressive capability model | §4 Levels 0-7 | Covered |
| Reasonable defaults | §3.3 Defaults table | Every setting has default |
| Full configurability surface | §6 Role System | TOML + templates + override resolution |
| Roles external, not hardcoded | §6.1 Three-Part Role Stack | ZERO hardcoded roles |
| Sandboxes, plugins, hooks | §16 Plugins, §19 Hooks, §3.2 isolation | Covered |
| Uniform agent abstraction | §2 Agent Protocol | AgentProvider interface |
| Same subsystems as GT, configurable | §10-16 | Beads, mail, nudge, formulas, molecules, convoys, plugins |
| Daemon with websocket transparency | §7.4 Websocket Transparency | Streams all data types |

## Appendix B: Gas Town Subsystem Coverage

| Gas Town Subsystem | Spec Section | Notes |
|-------------------|-------------|-------|
| Agents (tmux sessions) | §2 Agent Protocol | All 7 providers |
| Beads (Dolt task system) | §10 Task System | Beads-first |
| Mail (async messaging) | §15.1 Mail | Priority levels, bead-backed |
| Nudge (tmux send-keys) | §15.1 Nudge | Synchronous delivery |
| Roles (TOML + templates) | §6 Role System | Three-part stack |
| Formulas (4 types) | §11 Formulas | Full schemas |
| Molecules (instantiated formulas) | §12 Molecules | DAG execution, step lifecycle |
| Convoys (batch tracking) | §13 Convoys | Auto-close, stranded detection |
| Sling (work dispatch) | §14 Sling | Unified dispatch command |
| Plugins (Markdown + gates) | §16 Plugin System | 5 gate types |
| Deacon patrol (health) | §9 Health Monitoring | Configurable patrol cycle |
| Hooks (lifecycle events) | §19 Lifecycle Hooks | Shell commands on events |
| Session management (tmux) | §7 Controller | Session patterns, crash recovery |
| Websocket streaming | §7.4, §8.2 | Historical + real-time |
