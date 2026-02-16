# Gas City SDK — Technical Specification

> **Version:** 0.12.0
> **Date:** 2026-02-15
> **Status:** Planning (grounded in Gas Town source exploration)
> **Predecessor:** v0.11.0 (added hello-world.toml as fourth example config, aligned with concepts doc §6 "Four Configs, One SDK")

---

## 1. Executive Summary

Gas City is an **orchestration-builder SDK** — a Go toolkit for composing multi-agent coding workflows. It is the "Level 8" — the next step beyond Gas Town for users who have outgrown any single orchestrator and want to build their own. It extracts Gas Town's battle-tested subsystems (agents, beads, mail, nudge, formulas, molecules, convoys, plugins, health monitoring) into a configurable SDK where **all role behavior is user-supplied configuration** and the SDK provides only infrastructure. You can build Gas Town in Gas City, or Ralph, or Claude Code Agent Teams, or any other orchestration topology you design — via specific configurations.

**The core principle: ZERO hardcoded roles.** Gas City has no built-in Mayor, Deacon, Polecat, or any other role. The SDK provides the machinery — agent protocol, task system, messaging, formula engine, molecule execution, health monitoring, session management — and the user provides TOML config files and Markdown prompt templates that define what agents exist, what they do, and how they coordinate. Users can create their own roles, teams, coordination rules, and worker instructions — a full configurability surface.

**Progressive capability model:** Users start with a minimal TOML config (~10 lines for a single agent) and add sections as needed. The SDK activates subsystems based on what's configured. Config grows; the SDK is constant. Every level is independently useful.

**Four example configs ship with the SDK:**
- `hello-world.toml` — Single agent, single task (simplest possible)
- `ralph.toml` — Single agent with task loop
- `ccat.toml` — Claude Code Agent Teams: coordinator + worker pool
- `gastown.toml` — Complete Gas Town replication with all 8 roles, formulas, plugins, multi-project support

These are examples, not defaults. `gc init --file hello-world.toml` copies an example to the workspace.

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
- Coordination rules (depends_on, lifecycle events, workflows)
- Formulas, plugins, and their gate conditions

**What this spec does NOT cover:**
- Cloud/remote execution (v1 is local-only)
- Multi-tenant isolation
- Agent marketplace / sharing
- Automatic prompt engineering
- Cost management (responsibility of the underlying runtime)

### 1.1 Substrate Layering Principle

**Vision requirement:** "GC builds up higher level concepts on lower level concepts, e.g. the mail system is built on top of the beads system. The core principles are clearly visible and intrinsic whereas the higher level concepts are built up with configuration."

Gas City's subsystems form a strict dependency hierarchy. Each layer builds exclusively on the layers below it. No layer bypasses a lower layer to access its substrate directly. This makes each layer independently testable, replaceable, and comprehensible.

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 5: CONTROLLER (§7)                                    │
│   Workspace daemon — orchestrates all layers below           │
│   Depends on: everything                                     │
├─────────────────────────────────────────────────────────────┤
│ Layer 4: DISPATCH & COORDINATION                             │
│   Sling (§14) ─── Convoy (§13) ─── Health/Patrol (§9)      │
│   Depends on: formulas, molecules, beads, mail, pool,       │
│   event bus, agent registry, nudge                           │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: WORKFLOW ENGINE                                     │
│   Formulas (§11) ─── Molecules (§12) ─── Plugins (§16)     │
│   Depends on: beads (molecules are beads), event bus         │
│   (plugin triggers), agent pool (execution targets)          │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: MESSAGING & INTERACTION                             │
│   Mail (§15.1) ─── Nudge (§15.1) ─── Channels (§15.6)     │
│   Lifecycle Events (§19)                                     │
│   Mail depends on: beads (messages stored as beads)          │
│   Nudge depends on: tmux, agent registry                     │
│   Lifecycle depends on: event bus                            │
├─────────────────────────────────────────────────────────────┤
│ Layer 1: TASK SYSTEM & AGENTS                                │
│   Beads/TaskBackend (§10) ─── Agent Registry (§18.3)        │
│   Pool Manager (§2.7) ─── Name Pool                         │
│   Depends on: agent protocol, session management             │
├─────────────────────────────────────────────────────────────┤
│ Layer 0: INFRASTRUCTURE (no Gas City dependencies)           │
│   Agent Protocol (§2) ─── Event Bus (§8) ─── Config (§3)   │
│   Session/tmux ─── Sandbox/Isolation                        │
│   These are self-contained: pure interfaces, pure I/O,       │
│   or pure parsing with no upward dependencies.               │
└─────────────────────────────────────────────────────────────┘
```

**Detailed substrate relationships:**

| Subsystem | Layer | Builds On | Provides To |
|-----------|-------|-----------|-------------|
| **Agent Protocol** (`AgentProvider`) | 0 | Nothing — pure interface | Everything that touches agents |
| **Event Bus** | 0 | Nothing — JSONL append + tiered subscribers (§8.2/§8.3) | Lifecycle events, health, plugins, dashboard |
| **Config Parser** | 0 | Nothing — pure TOML parsing | Everything (config is universal input) |
| **Session/tmux** | 0 | Nothing — OS-level tmux management | Agent protocol (tmux provider), nudge |
| **Sandbox/Isolation** | 0 | Nothing — worktree/directory creation | Agent protocol (workspace setup) |
| **Beads** (`TaskBackend`) | 1 | Config (data_dir), event bus (change notifications) | Mail, formulas, molecules, convoys, plugins (wisps), sling, agent registry (identity beads) |
| **Agent Registry** | 1 | Agent protocol (handle persistence) | Pool manager, health patrol, sling, nudge |
| **Pool Manager** | 1 | Agent protocol, name pool | Sling (auto-spawn), health (scale-down) |
| **Mail** | 2 | Beads (messages stored as mail-type beads) | Sling (notifications), health (alerts), agents (inbox) |
| **Nudge** | 2 | Tmux (`send-keys`), agent registry (session lookup) | Sling (wake agents), health (restart notification) |
| **Channels** | 2 | Nudge (delivery), config (member lists) | Broadcast, group nudge |
| **Lifecycle Events** | 2 | Event bus (subscribe to events), shell (`sh -c`) | Health alerting, custom integrations |
| **Formula Parser** | 3 | Nothing (pure TOML parsing) | Molecules (instantiation template) |
| **Molecules** | 3 | Beads (root bead + step children), formula parser | Sling (work units), agents (DAG execution) |
| **Plugins** | 3 | Event bus (gate triggers), beads (wisps), agent pool (execution) | Automated maintenance, custom actions |
| **Sling** | 4 | Formulas, molecules, beads, mail, nudge, pool manager, convoy | Agents (unified work dispatch) |
| **Convoy** | 4 | Beads (tracking via `Tracks` dependencies), molecules (legs) | Progress tracking, stranded detection |
| **Health/Patrol** | 4 | Agent registry, event bus, beads (stale hooks), nudge (restart) | Agent reliability, stall recovery |
| **Controller/Daemon** | 5 | All of the above | Workspace lifecycle, websocket transparency |

**Layering invariants:**

1. **No upward dependencies.** Layer N never imports or calls Layer N+1. The event bus (Layer 0) has no knowledge of health monitoring (Layer 4). Beads (Layer 1) has no knowledge of formulas (Layer 3).

2. **Beads is the universal persistence substrate for domain state.** Mail messages, molecule steps, convoy tracking, plugin wisps, agent identities — all are beads. This means a single storage engine (Dolt or filesystem) serves domain state for the entire system. Swapping backends (e.g., `beads` → `filesystem`) swaps domain persistence for everything at once. **Exceptions (operational/runtime state, not beads):** agent registry handles (`.gc/agents/*.json`), durable event log (`.gc/events.jsonl`), config pointer (`.gc/config.toml`), sequence counter (`.gc/events.seq`). These are infrastructure artifacts managed by their respective subsystems.

3. **Event bus is the universal observation substrate.** Every layer publishes events. No layer consumes events from a higher layer. The event bus is how transparency (§7.4) works — the websocket streams events from all layers without any layer needing to know about the websocket.

4. **Config is the universal activation mechanism.** Each layer activates based on config sections present (§4). If `[messaging]` is absent, Layer 2 messaging subsystems don't initialize. If `[formulas]` is absent, Layer 3 formula engine doesn't initialize. The controller (Layer 5) inspects config to determine which layers to start.

5. **The progressive capability model (§4) generally follows the layering.** Capabilities activate bottom-up: Level 0-1 activates Layers 0-1, Level 2-4 activates Layers 1-2, Level 5 activates Layer 3, Level 6 activates Layer 4 (health/patrol requires the daemon), Level 7 activates Layer 3 plugins, Level 8 spans Layers 4-5. The strict bottom-up rule has one exception: health monitoring (Level 6) reaches into Layer 4 because the patrol loop requires the daemon (Layer 5). Users still build up from the bottom — you can't configure formulas (Layer 3) without having beads (Layer 1) active.

**Why this matters for extensibility:** To add a new subsystem (e.g., a cost tracker), you identify its layer, declare its dependencies, and wire it into the event bus. The subsystem reads config to activate, uses beads for persistence, and publishes events for transparency. No existing subsystem needs modification.

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

// SandboxProvider is the interface for agent workspace isolation.
// Decouples orchestration from the isolation mechanism (worktree, directory, Docker, etc.).
// Vision requirement: "wiring in sandboxes, plugins, and hooks (explicit extensibility)"
type SandboxProvider interface {
    // Name returns the sandbox type (e.g., "worktree", "directory", "docker")
    Name() string

    // Create sets up an isolated workspace for an agent.
    // Returns a Sandbox handle (not just a path) — the handle carries the metadata
    // needed by Destroy/Exists/Sync and by callers who need container IDs, branch
    // names, or "where to run commands" for non-filesystem sandboxes (Docker, remote).
    Create(ctx context.Context, config SandboxConfig) (Sandbox, error)

    // Destroy tears down the isolated workspace.
    // For worktrees: git worktree remove. For directories: rm -rf. For Docker: container rm.
    // Idempotent: returns nil if sandbox already destroyed.
    // DestroyOpts.Force bypasses safety checks (e.g., uncommitted changes in worktrees).
    Destroy(ctx context.Context, sandbox Sandbox, opts DestroyOpts) error

    // Exists checks if the sandbox workspace still exists.
    Exists(sandbox Sandbox) bool

    // Sync pulls latest changes into the sandbox (e.g., git pull in worktree).
    Sync(ctx context.Context, sandbox Sandbox) error
}

// Sandbox is the handle returned by Create — it carries everything needed to
// interact with the isolated workspace. This is a struct (not just a path)
// so that Docker/remote providers can carry container IDs, volume mounts, etc.
type Sandbox struct {
    WorkDir    string            // Absolute filesystem path (primary for worktree/directory)
    Kind       string            // Provider type that created this ("worktree", "directory", "docker")
    Branch     string            // Git branch name (worktree only; empty otherwise)
    Metadata   map[string]string // Provider-specific: container ID, volume, remote host, etc.
    BeadsDir   string            // Resolved beads storage path (see "Beads in sandboxes" below)
}

type SandboxConfig struct {
    Project    string // Project name
    AgentName  string // Agent name (for branch naming)
    BaseBranch string // Base branch to branch from (default: "main")
    BaseDir    string // Parent directory for the sandbox
}

type DestroyOpts struct {
    Force bool // Bypass safety checks (e.g., uncommitted changes in worktrees)
}
```

**Built-in sandbox providers:**

| Provider | Isolation | Lifecycle | Use Case |
|----------|-----------|-----------|----------|
| `worktree` (default) | Git worktree on unique branch | `git worktree add` / `git worktree remove` | Ephemeral agents (polecats) — full git isolation |
| `directory` | Copy of working tree | `cp -r` / `rm -rf` | Non-git projects or quick isolation |
| `none` | Shared workspace | No-op create/destroy | Persistent agents (crew, mayor) working on main |

Docker and remote container providers are future implementations — the `Sandbox` handle struct accommodates them without interface changes (Docker carries container ID and volume mounts in `Metadata`). The `AgentConfig.Isolation` field (§3.2) selects which `SandboxProvider` to use.

**Beads storage in sandboxes:** Sandboxes must locate the shared beads storage without duplicating the task database. The resolution order:
1. If `GC_BEADS_DIR` env var is set, use that path (providers inject this into agent sessions).
2. If `.beads/redirect` exists in the sandbox `WorkDir`, follow the redirect to the shared store (Gas Town pattern: worktrees contain a redirect file pointing to the main repo's `.beads/`).
3. Otherwise, use `{WorkDir}/.beads/` (in-sandbox storage, appropriate for `directory` and `none` providers).

The `Sandbox.BeadsDir` field is resolved at `Create` time and passed to agents via `GC_BEADS_DIR`. This ensures worktree sandboxes share the main repo's beads store while directory-copy sandboxes get their own snapshot.

### 2.3 Core Data Structures

> **Note:** `AgentConfig` represents the **resolved** configuration after merging workspace TOML settings with role definitions (`roles/*.toml`). Fields like `PromptTemplate` and `NudgeMessage` are resolved from the role file, not directly specified in the workspace config.

```go
type AgentIdentity struct {
    Workspace string  // Workspace name
    Project   string  // Project name (empty for workspace-scoped)
    Name      string  // Agent name from config
    Instance  string  // Pool instance name ("Toast", "Furiosa"; empty for non-pooled)
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
    Scope          string            // "workspace" or "project"
    Project        string            // Resolved project name (empty for workspace scope)
    Session        *SessionConfig
    Env            map[string]string
    PromptTemplate string            // Path to Markdown Go template (from role file)
    NudgeMessage   string            // Default nudge text (from role file)
    DependsOn      []string          // Agent names that must start first
    Ephemeral      bool              // Created/destroyed per-task
    Isolation      string            // "none", "worktree", "directory"
    Pool           *PoolConfig
    Loop           *LoopConfig
    Health         *HealthConfig
    Lifecycle      *LifecycleConfig  // Shell commands on lifecycle events (§19)
    Resume         *ResumeConfig     // Provider resume metadata (§2.8)
}

type SessionConfig struct {
    Pattern       string // tmux session name pattern, e.g., "gc-{project}-{name}"
    WorkDir       string // Working directory pattern, e.g., "{workspace}/{project}"
    StartCommand  string // Override command (default from provider)
    NeedsPreSync  bool   // Git pull before session start
}

type PoolConfig struct {
    Min               int
    Max               int
    IdleTimeout       time.Duration
    Theme             string   // Name pool theme (§2.7)
    Names             []string // Custom name list (overrides theme)
    MaxBeforeOverflow int      // Switch to numbered overflow (default 50)
}

type LoopConfig struct {
    Enabled     bool
    AutoExecute bool          // GUPP: auto-start when work on hook
    PollInterval time.Duration // Default: 10s
}

type HealthConfig struct {
    PingTimeout         time.Duration // Default: 30s
    StuckThreshold      time.Duration // Default: 1h (varies by role)
    ConsecutiveFailures int           // Default: 3
    KillCooldown        time.Duration // Default: 5m
}

type LifecycleConfig struct {
    OnStart        string // Shell command run when agent starts
    OnStop         string // Shell command run when agent stops
    OnTaskAssign   string // Shell command run when bead hooked to agent
    OnTaskComplete string // Shell command run when bead closed
    OnStall        string // Shell command run when agent detected stalled
}

type PingResult struct {
    Alive   bool
    Latency time.Duration
    Output  string // Last line of agent output (provider-specific)
}

type AgentState struct {
    Status        AgentStatus // Stopped, Starting, Idle, Working, Stalled
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

**Target provider matrix** — all providers Gas Town supports today. Phase column indicates when each provider ships (§22):

| Provider | Command | Flags | Provider Hooks | Resume | Session | Phase |
|----------|---------|-------|----------------|--------|---------|-------|
| `claude` (default) | `claude` | `--dangerously-skip-permissions` | Native | `--resume` (flag) | tmux | 1 |
| `subprocess` | (any) | (any) | No | No | stdin/stdout | 1 |
| `codex` | `codex` | `--yolo` | No | `resume <id>` (subcommand) | tmux | 5 |
| `gemini` | `gemini` | `--approval-mode yolo` | Yes | `--resume` (flag) | tmux | 5 |
| `opencode` | `opencode` | env: `OPENCODE_PERMISSION` | Yes (plugin) | No | tmux | 5 |
| `cursor` | `cursor-agent` | `-f` | No | `--resume` (flag) | tmux | 5 |
| `auggie` | `auggie` | `--allow-indexing` | No | `--resume` (flag) | tmux | 5 |
| `amp` | `amp` | `--dangerously-allow-all --no-ide` | No | `threads continue <id>` (subcommand) | tmux | 5 |

All tmux-based providers use the same session management: create session, inject env vars, send prompt via `send-keys`. The `subprocess` provider is the generic fallback for any CLI tool via stdin/stdout pipes.

### 2.5 Provider Auto-Detection

When `provider` is omitted from config, the SDK checks for available binaries in order: `claude` → `codex` → `gemini` → `opencode` → `cursor-agent` → `auggie` → `amp` → fallback to `subprocess`.

### 2.6 Provider Correctness Properties

**P1 — Idempotent Stop:** `Stop(handle)` on an already-stopped agent returns nil.
**P2 — Liveness After Start:** `Start()` returning `(handle, nil)` implies `IsRunning(handle)` is true within `ping_timeout`.
**P3 — Graceful Attach:** If `SupportsAttach()` returns false, `Attach()` returns `ErrNotSupported`.
**P4 — Bounded Cleanup:** `Stop(handle, false)` releases all OS resources within `kill_cooldown`.
**P5 — Thread Safety:** Concurrent calls to `SendPrompt()`, `ReadOutput()`, `GetState()`, `Ping()` on the same handle are safe.

**Standard sentinel errors:**

```go
var (
    ErrNotSupported          = errors.New("operation not supported by provider")
    ErrNotFound              = errors.New("agent or bead not found")
    ErrConflict              = errors.New("concurrent modification conflict")
    ErrInvalidState          = errors.New("invalid state transition")
    ErrNotAssignee           = errors.New("agent is not the current assignee")
    ErrTemporarilyUnavailable = errors.New("resource temporarily unavailable, retry later")
)
```

**Idempotency rules for mutating operations:**
- `Hook()`: Returns `ErrConflict` if bead already hooked by another agent. Returns nil if already hooked by the same agent (idempotent).
- `Unhook()`: Returns `ErrNotAssignee` if agent is not the current assignee. Returns nil if bead already open (idempotent).
- `Close()`: Returns `ErrNotAssignee` if agent is not the current assignee. Returns nil if bead already closed (idempotent).
- `Pin()`: Returns nil if bead already pinned (idempotent).
- `Stop()`: Returns nil if agent already stopped (P1).

All sentinel errors support `errors.Is()` wrapping — providers and backends may wrap these with context while preserving identity.

### 2.7 Pool Naming

Ephemeral pool agents receive human-readable names from themed name pools rather than numeric indices. This matches Gas Town's naming system where polecats get names like "Toast", "Furiosa", "Obsidian".

Name pool fields are part of `PoolConfig` (§2.3): `Theme`, `Names`, `MaxBeforeOverflow`.

**Built-in Themes:**

| Theme | Examples | Count |
|-------|----------|-------|
| `mad-max` (default) | furiosa, nux, slit, rictus, toast, dag, cheedo, valkyrie | 50 |
| `minerals` | obsidian, quartz, jasper, onyx, opal, topaz, garnet, ruby | 50 |
| `wasteland` | rust, chrome, nitro, guzzle, shiny, fury, witness | 50 |

**Allocation rules:**
1. Theme selection is deterministic: hash of project name selects theme (variety across projects).
2. Names allocated in order from the theme list.
3. Reserved infrastructure names (`witness`, `mayor`, `deacon`, `refinery`, `crew`, `polecats`) are filtered out.
4. When a pool agent is destroyed, its name is released for reuse.
5. When theme names are exhausted (>50 agents), overflow naming: `{project}-51`, `{project}-52`, etc.
6. Custom names (via config) override the theme entirely.

**Session naming:** Agent name "Toast" in project "gastown" → tmux session `gc-gastown-toast` (lowercase), agent ID `gastown/polecats/Toast` (original case preserved).

### 2.8 Resume Capabilities

Some providers support session resume (continuing a previous conversation). Resume is provider-specific and not part of the `AgentProvider` interface — instead, resume metadata is stored in the provider's config:

```go
type ResumeConfig struct {
    Flag    string // e.g., "--resume" (Claude), "resume" (Codex)
    Style   string // "flag" (appended to command) or "subcommand" (prefixed)
    IDEnv   string // Env var storing session ID (e.g., "CLAUDE_SESSION_ID")
}
```

| Provider | Resume Flag | Style | Session ID Env |
|----------|-------------|-------|----------------|
| `claude` | `--resume` | flag | `CLAUDE_SESSION_ID` |
| `codex` | `resume` | subcommand | (captured from JSONL output) |
| `gemini` | `--resume` | flag | `GEMINI_SESSION_ID` |
| `cursor` | `--resume` | flag | (uses chatId directly) |
| `auggie` | `--resume` | flag | — |
| `amp` | `threads continue` | subcommand | — |
| `opencode` | — | — | (manages sessions internally) |

Resume is distinct from crash recovery (`Adopter` interface). Resume provides conversation continuity; `Adopter` reconnects to a surviving process after controller crash.

---

## 3. Configuration Schema

### 3.1 File Format and Conventions

- **Main workspace config:** TOML — named by the user (e.g., `ralph.toml`, `ccat.toml`, `gastown.toml`)
- **Role definitions:** TOML files in `roles/*.toml` — operational settings per role
- **Prompt templates:** Markdown with Go templates in `roles/*.md.tmpl` — behavioral instructions
- **Formulas:** TOML files in `.beads/formulas/*.formula.toml`
- **Plugins:** Markdown with TOML frontmatter in `plugins/*.md`
- **Beads config:** YAML in `.beads/config.yaml`

**Config resolution order:**
1. `--config <path>` flag (explicit) → use that file
2. `GC_CONFIG` environment variable → use that path
3. `.gc/config.toml` pointer file (contains path to workspace TOML) → follow pointer
4. Single `*.toml` with `[workspace]` section in current directory → use that file
5. Multiple `*.toml` with `[workspace]` → error: "multiple Gas City configs found: X, Y. Use --config to select."

**Mode detection:**
1. Config resolved per above → Gas City mode
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
theme = ""                      # Name pool theme: "mad-max", "minerals", "wasteland"
                                # Default: auto-selected by hash of project name
names = []                      # Custom name list (overrides theme)
max_before_overflow = 50        # Switch to numbered names after this count

[agents.loop]                   # Task loop configuration
enabled = false
auto_execute = false            # GUPP: auto-start when work on hook
poll_interval = "10s"

[agents.health]                 # Health check thresholds
ping_timeout = "30s"
stuck_threshold = "1h"          # 1h (coordinator), 2h (workers), 4h (persistent)
consecutive_failures = 3
kill_cooldown = "5m"
max_restarts_per_window = 5     # Quarantine after this many restarts (§9.1)
restart_window = "1h"           # Window for counting restarts
max_restart_backoff = "30m"     # Cap on exponential backoff

[agents.lifecycle]              # Lifecycle event handlers (shell commands)
on_start = ""
on_stop = ""
on_task_assign = ""
on_task_complete = ""
on_stall = ""

[agents.resume]                 # Provider resume configuration (§2.8)
flag = ""                       # e.g., "--resume" (Claude), "resume" (Codex)
style = ""                      # "flag" or "subcommand"
id_env = ""                     # Env var storing session ID

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

# === CHANNELS (nudge groups) ===
[channels.<name>]
members = []                    # Agent names; pool names expand to all instances

# === DAEMON ===
[daemon]
websocket_port = 8765           # Websocket port for transparency streaming
patrol_interval = "3m"          # Deacon-style heartbeat interval
```

### 3.3 Defaults

Every setting has a reasonable default so users only configure what they need:

| Setting | Default | Notes |
|---------|---------|-------|
| **Workspace** | | |
| `workspace.version` | `1` | Schema version (for future migrations) |
| **Agent identity** | | |
| `provider` | Auto-detected | Scans PATH for known binaries |
| `scope` | `"project"` | Most agents are project-scoped |
| `ephemeral` | `false` | Persistent by default |
| `isolation` | `"worktree"` | Git worktree (Gas Town's default) |
| `depends_on` | `[]` | No dependencies |
| **Session** | | |
| `session.pattern` | `"gc-{project}-{name}"` | Tmux session name |
| `session.work_dir` | `"{workspace}/{project}"` | Working directory |
| `session.start_command` | `""` | Uses provider default |
| `session.needs_pre_sync` | `false` | No git pull before start |
| **Pool** | | |
| `pool.min` | `0` | Scale to zero when idle |
| `pool.max` | `5` | Reasonable concurrency limit |
| `pool.idle_timeout` | `"5m"` | Release idle instances |
| `pool.theme` | Auto-selected | Hash of project name picks theme |
| `pool.names` | `[]` | Use theme names |
| `pool.max_before_overflow` | `50` | Switch to numbered names after this |
| **Loop** | | |
| `loop.enabled` | `false` | Must opt-in to task loop |
| `loop.auto_execute` | `false` | Must opt-in to GUPP |
| `loop.poll_interval` | `"10s"` | Balance responsiveness vs overhead |
| **Health** | | |
| `health.ping_timeout` | `"30s"` | |
| `health.stuck_threshold` | `"1h"` | Varies by role (1h–4h) |
| `health.consecutive_failures` | `3` | |
| `health.kill_cooldown` | `"5m"` | |
| `health.max_restarts_per_window` | `5` | Quarantine threshold |
| `health.restart_window` | `"1h"` | Window for counting restarts |
| `health.max_restart_backoff` | `"30m"` | Cap on exponential backoff |
| **Lifecycle** | | |
| `lifecycle.on_start` | `""` | No handler |
| `lifecycle.on_stop` | `""` | No handler |
| `lifecycle.on_task_assign` | `""` | No handler |
| `lifecycle.on_task_complete` | `""` | No handler |
| `lifecycle.on_stall` | `""` | No handler |
| **Resume** | | |
| `resume.flag` | `""` | Provider default or none |
| `resume.style` | `""` | Provider default or none |
| `resume.id_env` | `""` | Provider default or none |
| **Subsystems** | | |
| `tasks.backend` | `"beads"` | Beads is the primary backend |
| `tasks.beads.data_dir` | `".beads"` | Beads storage location |
| `messaging.backend` | `"beads"` | |
| `formulas.dir` | `".beads/formulas"` | Formula discovery path |
| `plugins.dir` | `"plugins"` | Plugin discovery path |
| `projects.*.branch` | `"main"` | Default git branch |
| **Daemon** | | |
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
| Pool min ≤ max | "pool min (X) exceeds max (Y)" |
| Loop requires tasks | "agents.loop requires [tasks] section" |
| Worktree requires git | "worktree isolation requires git repository" |
| Formula step needs refs exist | "step X references unknown dependency Y" |
| Convoy synthesis depends_on refs exist | "synthesis depends_on references unknown leg Y" |
| Scope is valid enum | "invalid scope: X (must be 'workspace' or 'project')" |
| Isolation is valid enum | "invalid isolation: X (must be 'none', 'worktree', or 'directory')" |
| Backend is valid enum | "invalid backend: X (must be 'beads' or 'filesystem')" |
| Pool theme is known or custom names set | "unknown pool theme: X" |
| Duration fields parse | "invalid duration: X (must be Go duration like '5m', '1h')" |
| Channel members reference agents | "channel X member Y not found in agents" |
| Health stuck_threshold ≥ ping_timeout | "stuck_threshold (X) must be ≥ ping_timeout (Y)" |

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
| 6 | + Health monitoring | Add `[agents.health]` on any agent + `[daemon]` for patrol loop |
| 7 | + Plugins | Add `[plugins]` — automated actions with gate conditions |
| 8 | Full orchestration | Multiple projects, all roles, multi-project formulas |

> **Note on Level 6:** Health monitoring has two aspects: (a) per-agent `[agents.health]` thresholds (available at any level) and (b) the daemon patrol loop (`[daemon]`) that runs the actual monitoring cycle. Level 6 activates the patrol loop. Plugins (Level 7) are orthogonal to health monitoring — they can be added at any level but are listed here for progressive ordering.

**Level detection** is automatic: the config parser examines which sections are present and determines the capability level. Each level is independently useful — you don't need Level 8 to benefit from Level 2.

---

## 5. Four Example Configs

### 5.1 hello-world.toml — Single Agent, Single Task (Level 1)

```toml
[workspace]
name = "my-project"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "worker"
```

~8 lines. Provider auto-detected. Beads for task tracking. User creates beads, agent claims and executes them. No loop — agent works a single task per session. This is the simplest possible Gas City config: one agent, tracked work, context survival across sessions.

### 5.2 ralph.toml — Single Agent with Task Loop (Level 2)

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

### 5.3 ccat.toml — Claude Code Agent Teams (Level 4)

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

### 5.4 gastown.toml — Full Gas Town Replication (Level 8)

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
4. gc done (pushes, creates MR, closes bead, cleans up)

**Work is not done until `gc done` completes.**
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
| boot | workspace | Special dog that monitors the Deacon itself every 5 minutes — the watchdog's watchdog. |
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
4. Starts control loops: pool managers, patrol cycle, lifecycle executor, plugin gate evaluator
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

Clients connect and receive a catch-up replay of recent history, then real-time streaming.

### 7.5 Controller/CLI Contract

**The controller is the single writer for all mutable state.** When the controller is running, CLI commands (`gc bead`, `gc hook`, `gc sling`, etc.) act as thin clients that communicate with the controller via a local Unix domain socket (`.gc/controller.sock`). The controller serializes all mutations and emits events.

**When the controller is NOT running**, CLI commands fall back to direct file/database access with advisory locking. This enables simple Level 0-2 workflows without a daemon.

**Why this matters:** Agents running `gc` commands inside tmux sessions are concurrent with the controller's patrol, pool management, and lifecycle loops. Without a single-writer contract, Hook/Unhook/Close operations from agents race with stale hook scanning, pool scaling, and lifecycle handlers — causing missed events, double-hooks, and phantom state.

**RPC contract:**
- CLI detects controller presence via `.gc/controller.sock` existence + liveness ping
- All mutating operations (`Hook`, `Unhook`, `Close`, `Pin`, `Create`, `Update`, mail send, nudge) go through the controller when running
- Controller emits events for every mutation, ensuring the event bus captures all state changes
- Read operations (`List`, `Get`, `Ready`, `gc status`) may bypass the controller for lower latency

**Atomic write pattern:** All state persistence (`.gc/agents/*.json`, registry files) uses atomic writes: write to temp file, then `os.Rename()`. This prevents corruption from controller crashes during writes.

Three built-in consumers:

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
    EventBeadUnhooked   EventType = "bead.unhooked"
    EventBeadClosed     EventType = "bead.closed"

    // Health
    EventHealthPingOK       EventType = "health.ping_ok"
    EventHealthPingFail     EventType = "health.ping_fail"
    EventHealthRestart      EventType = "health.restart"
    EventAgentQuarantined   EventType = "agent.quarantined"

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

    // Plugin
    EventPluginFired    EventType = "plugin.fired"
    EventPluginFailed   EventType = "plugin.failed"
)
```

### 8.2 Tiered Subscriber Model

The event bus has two tiers:
- **Critical subscribers** (supervisor, structured logger, lifecycle executor): Delivered via bounded queues (capacity: 1000 events per subscriber). If a critical subscriber falls behind, excess events are written to a per-subscriber overflow log (`.gc/events-overflow-{subscriber}.jsonl`) recording delivery failures (event seq + timestamp + error reason). The overflow log is distinct from the durable event log: the durable log records *all published events*; overflow logs record *delivery failures for a specific subscriber*. The subscriber receives a gap notification on next read.
- **Optional subscribers** (websocket streamer, CLI feed, metrics): Fire-and-forget via unbounded channel. Slow consumers may miss events but never stall the bus.

**Backpressure protection:** Critical subscribers have a per-event timeout (default: 5s). If a subscriber (e.g., lifecycle executor running a shell command) exceeds the timeout, the event is queued and processing continues. This prevents lifecycle handlers from stalling the entire event bus. Lifecycle handlers that need long execution should spawn background processes.

A ring buffer (10k events) provides **in-memory** catch-up replay when new subscribers connect (e.g., websocket clients joining late).

### 8.3 Durable Event Log

In addition to the in-memory ring buffer, all events are appended to `.gc/events.jsonl` (one JSON object per line). This satisfies the vision requirement for "historical data" — events survive controller restarts and can be queried for post-mortem analysis.

**Event schema:** Every event is a JSON object with the following fields:

```go
type Event struct {
    Seq       uint64                 // Monotonically increasing sequence number
    Type      string                 // Event type (e.g., "bead.hooked", "agent.started")
    Timestamp time.Time              // Wall-clock time of publication
    Actor     string                 // Agent or subsystem that caused the event
    Payload   map[string]interface{} // Event-type-specific data
    Visibility string               // "internal" (system events) or "external" (user-visible)
}
```

The same `Event` struct is used for the JSONL durable log, the websocket stream, and the in-memory ring buffer. Fields are required except `Visibility` (defaults to `"external"`).

**Event log management:**
- **Location:** `.gc/events.jsonl`
- **Rotation:** When the log exceeds 50MB, it is rotated to `.gc/events.jsonl.1` (max 3 rotated files, ~200MB total cap)
- **Format:** JSON-lines, one event per line: `{"seq": 1, "type": "bead.hooked", "timestamp": "...", "actor": "gastown/main/polecat-Toast", "payload": {...}}`
- **Replay:** Websocket clients send `{"replay_from": <seq>}` on connect. The server reads from the durable log for events older than the ring buffer, then switches to live streaming. If `replay_from` requests a seq older than the oldest retained rotated file, the server returns `{"error": "seq_expired", "oldest_available": <seq>}` and begins streaming from the oldest available event.

**Sequence monotonicity:** Events have monotonically increasing sequence numbers, persisted in `.gc/events.seq` (single integer, atomically written). Clients detect gaps by comparing received sequence numbers and request replay.

**Cross-process writer safety:** When the controller is running, it is the sole writer to `.gc/events.jsonl` and `.gc/events.seq` (events from CLI fallback commands are forwarded via the Unix socket per §7.5). When the controller is *not* running, CLI commands may write events directly. In this fallback mode:
- `.gc/events.jsonl` is protected by `flock` (advisory file lock) — the same mechanism Gas Town uses (`internal/events/events.go`).
- `.gc/events.seq` is atomically updated: read current seq → increment → write to `.gc/events.seq.tmp` → `rename()` over `.gc/events.seq`. The `flock` on the JSONL file serializes this read-modify-write cycle across processes.

**Security and retention:**
- **Permissions:** `.gc/events.jsonl` is created with mode `0600` (owner-only read/write). Rotated files inherit the same permissions.
- **Sensitive content:** Events may include agent prompt/output summaries (per §7.4 and §20A). The durable log follows the same redaction stance as websocket streaming (§20A.2): output is logged but secrets are never included in event payloads. Operators who need full prompt/output history should use provider-specific session logs, not the event log.
- **Retention:** The 3-file rotation cap (~200MB) provides natural retention. For stricter requirements, operators can configure external log rotation (e.g., `logrotate`) on `.gc/events.jsonl`. A future `[events.retention]` config section may add TTL-based pruning per event type.

---

## 9. Health Monitoring / Supervisor

### 9.1 Patrol Cycle

Health monitoring follows Gas Town's Deacon patrol pattern — configured via role config, not hardcoded.

The patrol cycle runs at `daemon.patrol_interval` (default 3m):

```
FOR each running agent (skip self, skip quarantined):
  result = Ping(agent)
  IF ping fails:
    agent.consecutive_failures++
  ELSE:
    agent.consecutive_failures = 0
    agent.restart_count = 0      // Reset on success

  IF consecutive_failures >= agent.health.consecutive_failures:
    backoff = kill_cooldown * (2 ^ agent.restart_count)  // Exponential backoff
    backoff = min(backoff, max_restart_backoff)           // Cap at 30m default
    IF time since last restart < backoff:
      publish EventAgentStalled  // In backoff window, don't restart yet
    ELSE IF agent.restart_count >= max_restarts_per_window:
      quarantine(agent)          // Stop restarting, notify operator
      publish EventAgentQuarantined
    ELSE:
      agent.restart_count++
      Restart(agent)
      publish EventHealthRestart
```

**Restart throttling:** Exponential backoff prevents restart thrashing. After `max_restarts_per_window` (default: 5) restarts within `restart_window` (default: 1h), the agent enters quarantine. Quarantined agents are stopped and require manual intervention (`gc agent restart <name>` or `gc agent unquarantine <name>`).

### 9.2 Stall Detection

An agent is stalled when:
- Status = Working for longer than `stuck_threshold`
- AND Ping() returns OK (alive but not progressing)

Stall thresholds from Gas Town defaults:
- Mayor/Deacon: 1h
- Polecat/Refinery: 2h
- Crew: 4h

### 9.3 Stale Hook Scanning

Hooked beads **persist across agent crashes** — the system does not automatically unhook on failure. A periodic stale hook scan detects orphaned work:

```
FOR each bead with status = hooked:
  session = resolve assignee → tmux session name
  IF session is confirmed dead:
    check worktree for uncommitted changes / unpushed commits
    IF partial work detected:
      publish warning (operator notification)
    unhook bead (status → open, available for re-claim)
  ELSE IF session unknown AND bead age > max_age (default 1h):
    unhook bead (age-based fallback)
```

**Worktree safety check:** Before unhooking a bead from a dead agent, the system checks the agent's worktree for uncommitted changes or unpushed commits. If partial work is detected, a warning is published so the operator can recover the work before it is re-assigned.

**Two unhook criteria:**
1. **Session confirmed dead** — tmux session no longer exists (or `IsRunning()` returns false for non-tmux providers) → unhook immediately
2. **Unknown agent + age-based fallback** — assignee can't be resolved AND bead older than `max_age` → unhook as safety net

**Multi-signal confirmation:** For non-tmux providers where session liveness isn't directly observable, the scan requires TWO consecutive patrol cycles confirming "not running" before unhooking. This prevents false positives from transient failures. The Unhook operation uses compare-and-swap semantics — it fails if the bead's assignee changed between check and unhook, preventing races with concurrent re-assignment.

The stale hook scan runs as part of the patrol cycle (§9.1). In Gas Town, this is the Deacon's `stale-hooks` command.

### 9.4 Who Runs the Patrol?

In a full Gas Town config, the deacon agent runs the patrol. In simpler configs, the daemon itself runs the patrol loop. The user controls this via config — if a supervisor-like agent exists with health config, it takes over monitoring.

---

## 10. Task System / Beads

### 10.1 Beads-First Design

Gas Town uses beads (Dolt-backed structured data) as its primary task system. Gas City follows suit — beads is the default and recommended backend. A filesystem backend exists as a zero-dependency fallback for development and testing.

### 10.2 Bead Types

Bead types fall into two categories:

**Reserved system types** have SDK-level semantics that subsystems depend on. These type strings are reserved and the SDK implements specific behaviors for them:
- `agent` — has hook slot, identity binding, used by registry and hook system
- `molecule` — root bead with child step beads, used by molecule executor
- `convoy` — batch tracking with `Tracks` dependencies, used by convoy manager
- `mail` — priority-level queue entry, used by mail system
- `wisp` — ephemeral, destroyed after run completes, used by plugin executor and patrol

**User-defined types** are opaque strings with no SDK-level behavior. The SDK stores and filters them but attaches no special semantics. Users can define whatever types make sense for their orchestration (e.g., `task`, `bug`, `feature`, `epic`).

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
| `wisp` | Ephemeral bead — destroyed after the run completes (patrol cycles, plugin executions) |
| `gate` | Async coordination point — park agents until conditions are met |
| `slot` | Exclusive access control (e.g., merge slots) |
| `queue` | Message queue routing |
| `event` | Session/cost tracking events |
| `role` | Agent role definition bead |
| `rig` | Rig (project) identity bead |

Users can define additional types as needed. The SDK's subsystems (formulas, convoys, mail) create beads of the appropriate type but don't restrict what types can exist.

### 10.3 Bead Status Lifecycle

Gas Town's actual status model:

```
open → hooked → closed
         ↑
       pinned (infrastructure only)
```

- **open**: Available for claiming
- **hooked**: Pinned to an agent's hook (being worked on)
- **closed**: Completed
- **pinned**: Permanent infrastructure record (agent identity beads, role definitions). Pinned beads are never claimed or closed — they represent persistent system entities.

The `hooked` state is Gas Town's equivalent of "in-progress" — a bead is physically attached to an agent's hook bead via a dependency relationship. Only one agent can hook a bead at a time (atomic claiming via Dolt SQL transactions).

**Failure recovery:** Hooked beads **survive agent crashes**. When an agent crashes or is restarted, its hooked bead remains in `hooked` status. Two recovery paths:

1. **Agent restarts with work on hook** → GUPP applies: agent detects hooked work on startup, resumes execution.
2. **Agent is dead (session gone)** → Stale hook scan (§9.3) detects the dead session, checks the worktree for partial work, and unhooks the bead (status → `open`), making it available for re-claim.

Beads have **no built-in TTL or timeout**. Cleanup depends entirely on the health monitoring patrol cycle (§9).

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
    Pin(beadID string) error                       // Mark as permanent infrastructure (§10.3)
    Ready(filter TaskFilter) ([]Bead, error)       // Beads with no blockers, ordered by priority, filtered by scope
}

type Bead struct {
    ID          string
    Type        string            // "task", "bug", "feature", "agent", "molecule", "convoy", "mail", "merge-request", or user-defined
    Status      string            // "open", "hooked", "closed", "pinned"
    Pinned      bool              // True for permanent infrastructure beads (agent identity, role defs)
    Title       string
    Description string
    Project     string            // Project scope (empty for workspace-level)
    Assignee    string            // Agent ID currently hooked (empty if open)
    Priority    int               // 0 = highest
    Labels      []string
    Needs       []string          // Bead IDs that must close before this is ready
    Tracks      []string          // Non-blocking tracking refs (convoys)
    CreatedAt   time.Time
    ClosedAt    *time.Time
    Metadata    map[string]string // Extensible key-value pairs
}

type BeadUpdates struct {
    Title       *string
    Description *string
    Priority    *int
    Labels      []string          // Replaces entire label set
    Needs       []string          // Replaces entire needs set
    Metadata    map[string]string // Merged with existing metadata
}

type TaskFilter struct {
    Status  string   // Filter by status ("open", "hooked", "closed", "" for all)
    Type    string   // Filter by bead type ("task", "bug", etc., "" for all)
    Project string   // Filter by project ("" for all)
    Labels  []string // Filter by labels (AND logic)
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
  ready = backend.Ready(TaskFilter{Project: self.project})  // Scope to agent's project
  IF len(ready) > 0:
    bead = ready[0]  // highest priority
    err = backend.Hook(bead.id, self.id)  // Atomic claim
    IF err != nil: GOTO LOOP  // Someone else got it (or conflict)
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

[inputs]
[inputs.feature]
description = "The feature being implemented"
type = "string"
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
required = true

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

# Presets — preconfigured leg selections users can choose
[presets.gate]
description = "Light review for automatic flow"
legs = ["security", "correctness"]

[presets.full]
description = "Comprehensive review — all legs"
legs = ["correctness", "security", "performance"]
```

**Presets** are preconfigured selections of legs (convoy) or aspects that ship with a formula. Users select a preset at runtime via `--preset <name>`. If no preset is specified, all legs/aspects run. Presets are optional — formulas work without them.

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

### 11.3 Template Variable Mechanisms

Formulas use two template mechanisms:

**Simple substitution** (`{{key}}`): Used in step titles, descriptions, and workflow/expansion formulas. Variables are resolved from `[inputs]` at instantiation time via `--var key=value`. Validated by regex `{{[a-zA-Z_][a-zA-Z0-9_]*}}`.

> **Naming:** `[inputs]` is the unified section name for all formula types in Gas City. Gas Town's codebase uses separate `[inputs]` (convoy-specific) and `[vars]` (workflow-specific) as distinct struct fields. Gas City unifies these into a single `[inputs]` section that works across all formula types. The SDK parser accepts `[vars]` as an alias for migration compatibility, mapping it to `[inputs]` internally. New formulas should always use `[inputs]`.

```toml
# In step title:
title = "Design {{feature}}"

# Defined in [inputs]:
[inputs.feature]
description = "The feature being implemented"
type = "string"            # "string" (default), "number", "boolean"
required = true            # Must be provided via --var
# required_unless = ["other_input"]  # Required unless another input is provided
```

**Input field reference:**

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable description |
| `type` | string | `"string"` (default), `"number"`, `"boolean"` |
| `required` | bool | Must be provided at instantiation |
| `required_unless` | []string | Required unless one of the named inputs is provided |

**Go template rendering** (`{{.field}}`): Used in convoy/aspect prompt templates and output patterns. The SDK injects a structured context map with both user inputs and runtime fields, accessed via Go `text/template` dot notation.

| Formula Type | Context Fields |
|-------------|---------------|
| All types | `{{.formula_name}}`, `{{.target_description}}`, all `[inputs]` values |
| Convoy | `{{.leg.id}}`, `{{.leg.focus}}`, `{{.leg.title}}` |
| Aspect | `{{.aspect.id}}`, `{{.aspect.title}}` |

Example: `"Review PR #{{.pr}} focusing on {{.leg.focus}}"` — `.pr` comes from `[inputs]`, `.leg.focus` is a runtime context field injected per-leg.

### 11.4 Formula Validation Pipeline

```
Parse TOML → Infer type (if ambiguous) → Validate schema per type → Cycle detection (workflow/expansion) → Input validation
```

- Workflow/expansion: Steps form a DAG. Cycle detection via topological sort.
- Convoy: Legs are parallel by definition; synthesis depends on all legs.
- Inputs: Required inputs checked at instantiation time (`--var key=value`). Type validation applied.

### 11.5 Formula Discovery

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
| `backoff` | For wait-type steps: backoff configuration (see below) |

**BackoffConfig** (for wait-type steps):

```go
type BackoffConfig struct {
    Base       time.Duration // Initial wait duration (e.g., 10s)
    Multiplier float64       // Backoff multiplier (e.g., 2.0)
    Max        time.Duration // Maximum wait duration (e.g., 5m)
}
```

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

# Project-scoped target (auto-spawn polecat from name pool)
gc sling gt-abc myproject

# Specific named worker (name from pool, e.g., "Toast" from mad-max theme)
gc sling gt-abc myproject/Toast
gc sling gt-abc myproject/polecats/Toast   # Explicit long form

# Dog pool
gc sling gt-abc deacon/dogs
```

**Resolution pipeline:**
1. Contains `/` → parse as `project/role/name` or `project/name`
2. Single token → resolve against `[[agents]]` entries by name. If the workspace defines an agent named "mayor", that agent is the target. There are **no hardcoded role shortcuts** — all resolution is config-driven. (Gas Town compatibility mode auto-generates aliases from `gastown.toml` agent names.)
3. If target session not found and target looks like a pool agent → auto-spawn fresh instance from name pool (default: auto-spawn; suppress with `--no-create`)
4. **Cross-project guard:** bead's project must match target agent's project scope

### 14.4 Sling Flags

| Flag | Description |
|------|-------------|
| `--create` | Create pool agent if target doesn't exist (default: true for pool targets) |
| `--no-create` | Fail instead of auto-spawning if target doesn't exist |
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

### 15.4 Nudge Target Resolution

Nudge targets resolve against `[[agents]]` entries by name — there are **no hardcoded shortcuts**. Resolution order:

1. **Agent name** → match against `[[agents]]` entries (e.g., `gc nudge mayor` works if an agent named "mayor" exists in config)
2. **`channel:<name>`** → all members of named channel (§15.6)
3. **`<project>/<name>`** → project-scoped agent lookup

The target is resolved to a tmux session name via the agent's `session.pattern` config. For example, an agent named "mayor" with `session.pattern = "hq-mayor"` resolves to tmux session `hq-mayor`. Gas Town's `gastown.toml` example defines these patterns, making `gc nudge mayor` resolve to `hq-mayor` — but this is config-driven, not hardcoded.

### 15.5 DND (Do Not Disturb)

Agents can enter DND mode to suppress incoming nudges. When DND is active, nudges are silently dropped unless `--force` is used.

**DND is runtime state**, not configuration — it is set via CLI and stored in the agent registry (`.gc/agents/*.json`):

| Command | Description |
|---------|-------------|
| `gc agent dnd <name> on` | Enable DND for agent |
| `gc agent dnd <name> off` | Disable DND for agent |
| `gc agent dnd <name>` | Toggle DND state |

DND state is cleared on agent restart. Nudges suppressed by DND are not queued — they are lost. Use mail for messages that must survive DND.

### 15.6 Channels

Channels are named groups of agents that can be targeted as a single nudge destination. A nudge to `channel:<name>` is delivered to all member agents.

**Channel configuration** in the workspace TOML:

```toml
[channels.workers]
members = ["polecats", "crew"]    # Agent names (pools expand to all instances)

[channels.monitors]
members = ["witness", "deacon"]
```

**Rules:**
1. Channel names must be unique within the workspace.
2. Members reference agent names from `[[agents]]` entries.
3. Pool agents (e.g., "polecats") expand to all running instances in the pool.
4. Nudge to a channel delivers to all running members; stopped members are skipped.
5. `gc broadcast` is equivalent to nudging all running agents — it does not use channels.

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

### 16.3 Plugin Execution Model

The **plugin gate evaluator** (§7.1 control loop) periodically checks all registered plugins:

```
FOR each plugin in workspace + project plugin dirs:
  IF gate condition met (cooldown elapsed, cron fired, event matched, condition true):
    create wisp bead (type "wisp", ephemeral)
    resolve target agent: first available pool agent, or spawn new if pool has capacity
    hook wisp to target agent
    nudge agent with plugin's Markdown body as prompt
    ON completion: destroy wisp bead (ephemeral cleanup)
    ON failure: publish EventPluginFailed, retry up to 2 times with exponential backoff
```

**Manual gate trigger:** `gc plugin run <name>` — bypasses gate condition and fires immediately.

**Plugin CLI commands:**

| Command | Description |
|---------|-------------|
| `gc plugin list` | List all discovered plugins with gate status |
| `gc plugin run <name>` | Manually trigger a plugin (bypasses gate) |
| `gc plugin disable <name>` | Disable a plugin (skip during evaluation) |
| `gc plugin enable <name>` | Re-enable a disabled plugin |
| `gc plugin status <name>` | Show last execution, next scheduled, gate state |

### 16.4 Plugin Locations

- **Workspace-level:** `plugins/` — applies to all projects
- **Project-level:** `<project>/plugins/` — applies to specific project

---

## 17. CLI Design

### 17.1 Command Structure

The Gas City CLI is `gc`.

```
# Workspace lifecycle
gc init [--file hello-world.toml|ralph.toml|ccat.toml|gastown.toml]  # Init from example config
gc init                                               # Interactive wizard
gc start [--daemon] [--config <path>]                 # Config flag (§3.1)
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
gc agent dnd <name> [on|off]                          # Toggle/set DND mode (§15.5)
gc agent unquarantine <name>                          # Release from quarantine (§9.1)

# Work dispatch
gc sling <bead-or-formula> [target] [--var key=value]
gc hook                                               # Show hook status (bead + molecule)
gc hook <bead-id>                                     # Attach bead to own hook
gc hook --verbose                                     # Detailed hook view (DAG, dependencies)
gc hook clear                                         # Remove bead from hook (unhook)
gc handoff <bead-id>                                  # Hook + restart with fresh context
gc done                                               # Signal work complete (push, MR, cleanup)

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
gc nudge <target> -m "message"                        # Explicit message flag
gc nudge <target> --stdin                             # Read nudge from stdin
gc nudge --if-fresh <target> "msg"                    # Only nudge if session < 60s old
gc broadcast <message>

# Beads (task management)
gc bead list [--status=X] [--type=X]                  # List beads with filters
gc bead create <title> [--type=task]                   # Create a bead
gc bead show <id> [--json]                            # Detailed bead view
gc bead close <id> [--reason "..."]                   # Close/complete a bead
gc bead ready                                         # List ready (unblocked) beads
gc bead move <id> <target-prefix>                     # Move bead to different project

# Monitoring / transparency
gc dashboard                                          # Web UI
gc dashboard --tui                                    # Terminal UI
gc activity [--follow]                                # Event stream
gc stats                                              # Agent metrics

# Session inspection
gc seance <session-id>                                # Resume/inspect a predecessor's session
gc seance --talk <session-id>                         # Interactive conversation with previous agent

# Utility
gc validate                                           # Config validation
gc prime                                              # Render role prompt for current agent
gc migrate [--dry-run]                                # Generate config from Gas Town workspace
gc doctor                                             # Health checks (config, providers, tmux, beads)
gc config show                                        # Display resolved config
gc test-provider <name>                               # Run conformance suite against a provider
gc version
```

### 17.2 `gc init --file`

```
$ gc init --file hello-world.toml
Copied hello-world.toml to ./hello-world.toml
Run `gc start` to begin.

$ gc init
Welcome to Gas City SDK!

Example configs available:
  > hello-world.toml -- Single agent, single task (simplest)
    ralph.toml       -- Single agent with task loop
    ccat.toml        -- Coordinator + worker pool (Agent Teams)
    gastown.toml     -- Full multi-project orchestration (Gas Town)
    (custom)         -- Start from scratch

Select [hello-world.toml]:

Which coding agent do you use?
  > Claude Code
    Codex (OpenAI)
    Gemini CLI
    OpenCode
    Other

Select [Claude Code]:

Created hello-world.toml (Level 1 - Single Agent)
Run `gc start` to begin.
```

### 17.3 `gc done`

The task completion command for all agent types. Behavior adapts based on agent lifecycle:

**For ephemeral agents (polecats):**
1. **Push** — commit and push all changes to the working branch
2. **Merge request** — create MR/PR for the merge queue (unless `--no-merge`)
3. **Close bead** — mark the hooked bead as closed
4. **Cleanup** — destroy worktree, release name back to pool, terminate session

**For persistent agents (crew, mayor, witness, refinery):**
1. **Push** — commit and push all changes
2. **Merge request** — create MR/PR if on a feature branch (skip if on main)
3. **Close bead** — mark the hooked bead as closed
4. **No cleanup** — session and workspace are preserved

Exit codes communicate outcome to the orchestration layer:

| Exit | Code | Meaning |
|------|------|---------|
| `COMPLETED` | 0 | Work finished successfully |
| `ESCALATED` | 1 | Work needs human or senior agent review |
| `DEFERRED` | 2 | Work paused, bead stays open |
| `PHASE_COMPLETE` | 3 | Molecule step done, more steps remain |

Exit codes are emitted by `gc done` only. Other `gc` commands use standard Unix exit codes (0 = success, 1 = error). The symbolic names are set as `GC_EXIT_REASON` environment variable for lifecycle handlers.

### 17.4 `gc seance`

Session inspection/resume command. Uses provider-specific resume capabilities (§2.8) to interact with a previous agent's conversation:

```
gc seance <session-id>                   # Read-only inspection of session history
gc seance --talk <session-id>            # Interactive conversation with predecessor
```

For Claude Code, this uses `--fork-session --resume <id>` to create a read-only fork of the previous conversation. Useful for debugging stalled agents or understanding what a predecessor did.

---

## 18. Identity and Addressing

### 18.1 Address Format

Gas City uses a hierarchical addressing scheme matching Gas Town's patterns.

**Canonical (fully-qualified) format:**
```
<workspace>/<agent-name>                    # Workspace-scoped
<workspace>/<project>/<agent-name>          # Project-scoped
<workspace>/<project>/<pool>/<instance>     # Pool instance (named, e.g., "Toast")
```

**CLI shorthand rules** (workspace prefix is implicit within a workspace):
```
mayor                     → <workspace>/mayor          (agent name lookup)
gastown/witness           → <workspace>/gastown/witness (project/agent)
gastown/Toast             → <workspace>/gastown/polecats/Toast (project/instance, pool inferred)
gastown/polecats/Toast    → <workspace>/gastown/polecats/Toast (explicit)
channel:workers           → expand to all channel members
```

The CLI always resolves shorthand against the current workspace's `[[agents]]` entries. Ambiguity (e.g., an agent named "gastown" and a project named "gastown") is resolved by checking agents first, then projects.

### 18.2 Gas Town Address Compatibility

| Gas Town Format | Gas City Format |
|----------------|-----------------|
| `mayor/` | `<workspace>/mayor` |
| `deacon/` | `<workspace>/deacon` |
| `<rig>/crew/<name>` | `<workspace>/<project>/<name>` |
| `<rig>/polecats/<name>` | `<workspace>/<project>/polecats/<name>` |
| `<rig>/witness` | `<workspace>/<project>/witness` |
| `<rig>/refinery` | `<workspace>/<project>/refinery` |

### 18.3 Agent Registry

The agent registry maps logical identities to runtime handles, persisted to `.gc/agents/*.json` for crash recovery. On controller restart, it reconnects to surviving tmux sessions via the `Adopter` interface.

---

## 19. Lifecycle Events

> **Terminology note:** This section covers lifecycle event handlers — shell commands triggered by agent lifecycle events. These are configured via `[agents.lifecycle]` in config. The term "hook" in the rest of this spec refers exclusively to bead hooks (attaching work to an agent's hook slot, §10.3). Provider-specific hooks (e.g., Claude Code's native hook system) are called "provider hooks" when disambiguation is needed.

### 19.1 Lifecycle Event Handlers

Lifecycle handlers are shell commands triggered by agent events:

| Event | Trigger | Template Variables |
|------|---------|-------------------|
| `on_start` | Agent started | `{{.Agent}}`, `{{.Project}}` |
| `on_stop` | Agent stopped | `{{.Agent}}`, `{{.Project}}` |
| `on_task_assign` | Bead hooked to agent | `{{.Agent}}`, `{{.Task}}`, `{{.TaskTitle}}` |
| `on_task_complete` | Bead closed | `{{.Agent}}`, `{{.Task}}` |
| `on_stall` | Agent detected stalled | `{{.Agent}}`, `{{.Project}}` |

### 19.2 Execution Model

Lifecycle handlers run as shell commands via `sh -c` with a default 30s timeout. They are executed by the lifecycle executor, which is a critical event bus subscriber — handler failures are not silently dropped.

**Security note:** Template variables in lifecycle commands are shell-escaped before expansion to prevent injection. A task title containing `; rm -rf /` will be escaped to a safe string. See §20A for details.

---

## 20A. Security

### 20A.1 Shell Injection Mitigation

Lifecycle handlers (§19) execute shell commands with template variables expanded from bead metadata. Since bead titles and descriptions are user-supplied (and may originate from issue trackers), all template variables are shell-escaped before expansion using Go's `shellescape` package.

**Escaping rules:**
- **All** template variables (`{{.Agent}}`, `{{.Task}}`, `{{.TaskTitle}}`, and any custom variables from `[inputs]` or `[env]`) are passed through `shellescape.Quote()` before shell expansion.
- This wraps values in single quotes and escapes embedded single quotes.
- Example: a task title `fix bug; rm -rf /` becomes `'fix bug; rm -rf /'` — the semicolon is not interpreted as a command separator.

**What is NOT escaped:**
- The command template itself (the `on_start`, `on_stop`, etc. strings from config). These are trusted — they come from the workspace TOML, which is under the operator's control.
- Environment variables set via `[agents.env]` — also operator-controlled config.

### 20A.2 Secret Management

Gas City does not provide a built-in secret store. Secrets are managed via:

1. **Environment variables** in `[agents.env]` — suitable for API keys passed to providers.
2. **Shell profile** — providers inherit the user's shell environment (e.g., `~/global_env.sh`).
3. **Config file permissions** — workspace TOML files should be `chmod 600` in multi-user environments.

**Secrets MUST NOT appear in:**
- Bead metadata (stored in Dolt, visible to all agents)
- Mail messages (stored in Dolt)
- Websocket event streams (observable via dashboard)
- Formula `[inputs]` defaults (committed to repo)

**Streaming output redaction:** The websocket transparency stream (§7.4) may relay agent I/O that incidentally contains secrets (e.g., an agent printing an API key during debugging). The event bus does NOT perform automatic redaction — operators should be aware that raw agent output on the websocket is equivalent to watching the tmux session directly. A future enhancement could add configurable redaction patterns (`[daemon.redact_patterns]`), but v1 treats the websocket as a same-trust-level channel.

### 20A.3 Websocket Authentication

The daemon websocket endpoint (`ws://localhost:{port}`) is local-only by default:

- **v1:** Listens on `127.0.0.1` only. No authentication required (same-user access model, like tmux).
- **Future:** Token-based auth for remote dashboard access.

The websocket streams all event bus traffic. Any process on the local machine can connect and observe agent activity, bead state changes, mail delivery, etc. This is by design — transparency is a core feature.

### 20A.4 Tmux Session Access

Tmux sessions are owned by the user running `gc start`. Standard Unix permissions apply:
- Only the session owner can attach (`gc agent attach`).
- Worktree directories inherit the parent repo's permissions.
- Agent processes run as the invoking user — there is no privilege escalation.

### 20A.5 Provider Permission Models

Each provider has its own auto-approve mechanism (§2.4). These flags bypass the provider's built-in confirmation prompts, allowing autonomous operation. Operators should understand the implications:

| Provider | Auto-approve Flag | Scope |
|----------|------------------|-------|
| `claude` | `--dangerously-skip-permissions` | All file/command operations |
| `codex` | `--yolo` | All file/command operations |
| `gemini` | `--approval-mode yolo` | All file/command operations |
| `opencode` | env: `OPENCODE_PERMISSION={"*":"allow"}` | All tool calls |
| `cursor` | `-f` | Force mode |
| `auggie` | `--allow-indexing` | Indexing + file operations |
| `amp` | `--dangerously-allow-all --no-ide` | All operations, headless |

Using these flags is required for autonomous agent operation but means the agent can execute arbitrary shell commands and modify any file the user can access. This is the intended operating mode for Gas City agents.

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
| Name pool | Theme selection, allocation, release, overflow, reserved names | None (pure logic) |
| Stale hook scanner | Session liveness check, age-based fallback, worktree safety | AgentProvider + TaskBackend mocks |
| Task loop | Hook/claim cycle, mail check, poll interval | AgentProvider + TaskBackend mocks |
| Startup sequencer | DAG ordering, parallel start, rollback | AgentProvider mock |
| Event bus | Pub/sub, critical vs optional tiers, replay | None (pure concurrency) |
| Formula parser | All 4 types, input validation, cycle detection | None (pure parsing) |
| Molecule executor | DAG step ordering, ready detection, crash recovery | TaskBackend mock |
| Lifecycle executor | Template expansion, timeout, error propagation | EventBus (inject events) |
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
| Agent Teams smoke test | Coordinator dispatches to worker pool (named from pool) |
| Formula → molecule → execute | Sling formula → molecule created → steps execute |
| Convoy lifecycle | Create → track issues → auto-close |
| Stale hook recovery | Agent crash → bead stays hooked → scan → unhook → re-claim |
| Done exit flow | `gc done` → push → MR → close bead → cleanup worktree |
| Migration test | Gas Town workspace → gc migrate → verify config |
| Websocket streaming | Connect → receive events → verify catch-up replay |

---

## 22. Implementation Phases

> **Note on event bus progression:** A minimal in-process event bus (pub/sub, no websocket, no replay) ships in Phase 1 as part of the controller. Lifecycle handlers (Phase 3) subscribe to this minimal bus. Phase 5 upgrades it with websocket streaming, ring buffer replay, and the dashboard consumer. §7/§8 describe the full system; earlier phases use a subset.

### Phase 1: Agent Protocol + Config (3 weeks)

- AgentProvider interface with `SendPrompt`
- Provider implementations: `claude` (tmux), `subprocess` (generic)
- TOML config parser with level detection and validation
- Identity and addressing (agent registry, address resolution)
- `gc init --file`, `gc start`, `gc stop`, `gc status`, `gc level`, `gc validate`
- `gc config show`, `gc doctor`, `gc version`
- Contract test suite (`gc test-provider`)
- hello-world.toml working end-to-end (single agent, single task)

### Phase 2: Task System + Ralph (2 weeks)

- TaskBackend interface (beads + filesystem)
- Bead status lifecycle (open → hooked → closed)
- Task loop with atomic Hook/Claim
- Role system: TOML config loading, template rendering, override resolution
- `gc bead` (including `gc bead close`), `gc hook` (with --verbose), `gc prime`
- Ralph config working end-to-end

### Phase 3: Pool + Messaging + Agent Teams (2 weeks)

- Pool manager with min/max bounds and name pool allocation (themed names)
- Startup/shutdown sequencer with DAG ordering
- Mail system (beads-backed, priority levels)
- Nudge system (tmux send-keys, DND support)
- Channels (named agent groups for nudge targeting)
- Lifecycle event handlers (shell commands on events)
- `gc mail`, `gc nudge`, `gc broadcast`, `gc done`, `gc agent dnd`
- ccat.toml (Agent Teams) working end-to-end

### Phase 4: Formulas + Molecules + Sling + Convoys (2 weeks)

- Formula parser (all 4 types: workflow, convoy, expansion, aspect)
- Formula presets (named leg/aspect selections)
- Molecule instantiation (root bead + child step beads)
- Molecule execution (DAG step ordering, ready detection)
- Basic convoy creation and tracking (auto-convoy from sling)
- Sling dispatch command
- `gc formula`, `gc mol`, `gc sling`, `gc convoy create`, `gc convoy status`
- `gc handoff`

### Phase 5: Full Gas Town + Migration (3 weeks)

- Full convoy management (`gc convoy list`, `gc convoy check`, `gc convoy stranded`, `gc convoy close`)
- Plugin system (Markdown + TOML frontmatter, gate types)
- Health monitoring / patrol cycle with stale hook scanning
- Event bus with websocket streaming
- Dashboard (web + TUI)
- Migration tooling
- Session inspection (`gc seance`)
- `gc bead move`, `gc dashboard`, `gc activity`, `gc stats`, `gc migrate`, `gc seance`
- gastown.toml working end-to-end
- All remaining providers: codex, gemini, opencode, cursor, auggie, amp

---

## 23. Exclusions

| Exclusion | Rationale |
|-----------|-----------|
| Web dashboard (standalone, production-grade) | `gc dashboard` provides a dev/debug UI (§17.1, Phase 5). A production-grade standalone dashboard is a separate deliverable. |
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

### 24.5 Name Pool Uniqueness

For any pool within a project, no two running instances share the same name at time `t`:
```
∀i,j ∈ running_instances(pool, t): i ≠ j → name(i) ≠ name(j)
```

Enforced by the name pool allocator tracking in-use names. Names are released on instance destruction. All pool operations (scale-up, scale-down, name allocation, name release, `gc done` cleanup) are serialized through the controller (§7.5) to prevent races between concurrent sling commands, pool scaling, and agent cleanup.

### 24.6 Molecule DAG Correctness

A molecule step `s` with dependencies `D = {d1, d2, ...dk}` becomes ready if and only if all dependencies are complete:
```
ready(s) ⟺ ∀di ∈ D: status(di) = closed
```

---

## 25. Open Questions

1. **`gc` vs `gt` CLI namespace.** Options: (a) `gc` as separate binary, (b) `gt gc` subcommand, (c) `gt` detects Gas City config and adjusts. Recommendation: (a) separate binary.

2. **Plugin system scope.** The plan includes plugins with gate types. Gas Town's plugin system may still be evolving. Confirm scope and gate type implementations before Phase 5.

3. ~~**Websocket protocol.**~~ Resolved: JSON-lines format with `{"replay_from": <seq>}` handshake. Durable event log at `.gc/events.jsonl` with rotation. See §8.2/§8.3.

4. ~~**Formula preset system.**~~ Resolved: presets are v1. Added to convoy schema (§11.2) as preconfigured leg selections.

5. **Non-interactive mode for providers.** Several providers support non-interactive/batch modes (e.g., `codex exec`, `gemini -p`). Define when to use interactive vs batch mode. Gas Town implementation shows `NonInteractiveConfig` with `PromptFlag`, `OutputFlag`, and `Subcommand` fields per provider.

6. **Config hot-reload.** Can you add agents to a running workspace? v1: no (restart required). v2: file watch + diff.

7. **Log management.** The spec defines `gc agent logs <name>` but doesn't specify log storage location, rotation policy, or retention. Recommend `.gc/logs/` with configurable rotation.

8. **Provider image support.** The `Prompt` struct supports `Images [][]byte` but most CLI providers are text-only. Define fallback behavior when a provider doesn't support image attachments (error vs silent drop vs text description).

9. **Secret management in formulas.** Formula `[inputs]` values end up in bead metadata (Dolt). §20A.2 says secrets must not appear in bead metadata. Define a mechanism for environment-based secret injection that avoids persistence (e.g., `type = "secret"` inputs resolved from env vars at execution time, never stored).

10. ~~**Event durability.**~~ Resolved: durable event log at `.gc/events.jsonl` with 50MB rotation, 3-file cap. Event struct defined. See §8.3.

11. **Provider capability negotiation.** `Prompt.Images` and `Prompt.Files` exist (§2.2) but most CLI providers are text-only. Define capability flags or optional interfaces (`SupportsImages() bool`, `SupportsFiles() bool`) and required fallback behavior (error vs transform-to-text vs drop-with-warning).

12. **Worktree filesystem layout.** §2.3 and §3.2 enumerate `isolation` modes but don't specify the exact directory layout under `.gc/`, naming conventions for worktree branches, or cleanup procedures for ephemeral agents. Define before Phase 2.

13. **`gc seance` provider abstraction.** §17.4 describes Claude-specific `--fork-session --resume` but this isn't captured in the `AgentProvider` interface or `ResumeConfig`. Formalize "session inspection" as an optional provider capability.

---

## Appendix A: Vision Alignment Checklist

| Vision Requirement | Spec Section | Status |
|-------------------|-------------|--------|
| Orchestration-builder toolkit | §1 Executive Summary | Covered |
| Multiple town shapes via config | §5 Four Example Configs | hello-world, ralph, ccat, gastown |
| Progressive capability model | §4 Levels 0-8 | Covered (plugins now Level 7) |
| Reasonable defaults | §3.3 Defaults table | Every setting has default |
| Full configurability surface | §6 Role System | TOML + templates + override resolution |
| Roles external, not hardcoded | §6.1 Three-Part Role Stack, §14.3, §15.4 | ZERO hardcoded roles; sling/nudge resolve against config |
| Sandboxes, plugins, lifecycle events | §2.2 SandboxProvider, §16 Plugins, §19 Lifecycle Events | Pluggable sandbox interface + implementations |
| Uniform agent abstraction | §2 Agent Protocol | AgentProvider interface |
| Same subsystems as GT, configurable | §10-16 | Beads, mail, nudge, formulas, molecules, convoys, plugins |
| Higher-level on lower-level layering | §1.1 Substrate Layering Principle | 6-layer hierarchy, invariants, substrate table |
| Daemon with websocket transparency | §7.4 Websocket, §8.2/§8.3 Durable Event Log | Historical (JSONL) + real-time streaming, formal Event struct |
| Worker naming (name pools) | §2.7 Pool Naming | Themed names matching Gas Town |
| Failure recovery (stale hooks) | §9.3, §10.3 | Session liveness + worktree safety |
| Complete CLI parity with GT | §17 CLI Design | done, seance, hook subcommands, task close |

## Appendix B: Gas Town Subsystem Coverage

| Gas Town Subsystem | Spec Section | Notes |
|-------------------|-------------|-------|
| Agents (tmux sessions) | §2 Agent Protocol | All 8 providers |
| Beads (Dolt task system) | §10 Task System | Beads-first, failure recovery |
| Mail (async messaging) | §15.1 Mail | Priority levels, bead-backed |
| Nudge (tmux send-keys) | §15.1 Nudge | Synchronous delivery |
| Roles (TOML + templates) | §6 Role System | Three-part stack |
| Formulas (4 types) | §11 Formulas | Full schemas, unified [inputs] |
| Molecules (instantiated formulas) | §12 Molecules | DAG execution, step lifecycle |
| Convoys (batch tracking) | §13 Convoys | Auto-close, stranded detection |
| Sling (work dispatch) | §14 Sling | Unified dispatch, target resolution pipeline |
| Plugins (Markdown + gates) | §16 Plugin System | 5 gate types |
| Deacon patrol (health) | §9 Health Monitoring | Patrol cycle, stale hook scanning |
| Lifecycle events | §19 Lifecycle Events | Shell commands on events |
| Session management (tmux) | §7 Controller | Session patterns, crash recovery |
| Controller/CLI contract | §7.5 | Single-writer via Unix socket, atomic writes |
| Websocket streaming | §7.4, §8.2/§8.3 | Historical + real-time, Event struct |
| Name pool (worker naming) | §2.7 Pool Naming | Themed names, allocation/release |
| Resume (session continuity) | §2.8 Resume Capabilities | Provider-specific, separate from crash recovery |
| Wisps (ephemeral beads) | §10.2 Bead Types | Ephemeral beads for patrol cycles, transient ops |
| Gates (async coordination) | §10.2 Bead Types | Park/resume agents on conditions |
| Boot (deacon watchdog) | §6.5 Role Reference | Watchdog's watchdog |
| Done (task exit flow) | §17.3 `gc done` | Push, MR, close, cleanup (all agent types) |
| Seance (session inspection) | §17.4 `gc seance` | Fork/resume predecessor sessions |
