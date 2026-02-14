# Gas City SDK — Technical Specification

> **Version:** 0.2.0
> **Date:** 2026-02-14
> **Status:** Planning (Spec-Forge Stage 3)
> **Source:** Independent spec-forge pipeline — research-notes.md + ideas-ranked.md

---

## 1. Executive Summary

Gas City is an **orchestration-builder SDK** that extracts Gas Town's hardcoded 7-role multi-agent orchestration into a configurable toolkit. The SDK lets users compose their own agent topologies — from a single task-loop agent up to full multi-project orchestration — using a declarative TOML configuration.

**The core problem:** Gas Town is monolithic. Its 7 roles (Mayor, Deacon, Witness, Refinery, Polecat, Crew, Dog) are hardcoded across 50+ packages, 396+ CLI commands, and deep tmux coupling. You can't use a subset of its capabilities, swap in different agent runtimes, or define custom topologies without forking the codebase.

**The core insight:** Orchestration is composable. Work tracking, messaging, health monitoring, task dispatch, and merge processing are independent capabilities that happen to be welded together in Gas Town. Separating them into composable units, each accessible through a uniform Agent Runtime interface, creates a toolkit that can express Gas Town (and any other topology) as a configuration.

**What this spec covers:**
- The Agent Runtime abstraction (the foundational interface)
- A progressive capability model (Levels 0-7) where each level adds one composable capability
- Three milestone configurations: Ralph, Agent Teams, Gas Town
- Config schema (`gas-city.toml`) with full data structures
- Startup/shutdown sequencing with formal ordering guarantees
- Migration from existing Gas Town workspaces
- CLI design (`gc` commands)
- Testing strategy with contract tests per adapter

**What this spec does NOT cover:**
- Web dashboard UI (future work; the SDK provides the data API)
- Cloud/remote execution (v1 is local-only)
- Multi-tenant isolation
- Agent marketplace / sharing
- Automatic prompt engineering

---

## 2. The Agent Runtime Abstraction

### 2.1 Design Rationale

Every interaction Gas Town performs with an agent — starting, stopping, sending work, reading output, health-checking — currently flows through `internal/tmux/tmux.go`. This creates three problems:

1. **Runtime lock-in**: You can't use Codex, Gemini, or Agent SDK without rewriting the session layer
2. **Environment lock-in**: You can't run without tmux (containers, CI, headless servers)
3. **Observation lock-in**: Monitoring depends on tmux `capture-pane` output scraping

The Agent Runtime abstraction separates **what you do with an agent** from **how a specific agent implementation works**.

### 2.2 The RuntimeAdapter Interface

```go
// RuntimeAdapter is the interface every agent runtime must implement.
// It is the foundational abstraction of the Gas City SDK.
type RuntimeAdapter interface {
    // Identity
    Name() string  // e.g., "claude-code", "codex", "subprocess"

    // Lifecycle
    Start(ctx context.Context, config AgentConfig) (AgentHandle, error)
    Stop(handle AgentHandle, graceful bool) error
    Restart(handle AgentHandle) error
    IsRunning(handle AgentHandle) bool

    // Work assignment
    Assign(handle AgentHandle, task TaskDescriptor) error
    Nudge(handle AgentHandle, message string) error
    GetState(handle AgentHandle) (AgentState, error)

    // I/O
    SendInput(handle AgentHandle, text string) error
    CaptureOutput(handle AgentHandle) (io.ReadCloser, error)
    ReadOutput(handle AgentHandle) (string, error)

    // Health
    Ping(handle AgentHandle) (PingResult, error)
    GetMetrics(handle AgentHandle) (AgentMetrics, error)

    // Presentation (optional)
    SupportsAttach() bool
    Attach(handle AgentHandle) error
    Detach(handle AgentHandle) error
}
```

### 2.3 Core Data Structures

```go
// AgentHandle is an opaque reference to a running agent instance.
type AgentHandle struct {
    ID        string            // Unique ID for this instance (UUID)
    Agent     string            // Agent name from config (e.g., "coder")
    Runtime   string            // Runtime adapter name (e.g., "claude-code")
    Project   string            // Project scope (empty for workspace-scoped)
    StartedAt time.Time
    PID       int               // OS process ID (0 if not applicable)
    Metadata  map[string]string // Runtime-specific metadata
}

// AgentConfig is the resolved config for starting an agent.
type AgentConfig struct {
    Name          string
    Role          string
    Runtime       string
    Project       string            // Empty for workspace scope
    Command       string            // Primary command (e.g., "claude")
    Args          []string
    Env           map[string]string
    WorkDir       string
    SystemPrompt  string            // Path or inline prompt
    NudgePrompt   string            // Initial message to send
    RuntimeConfig map[string]any    // Adapter-specific settings
}

// AgentState represents the current state of an agent.
type AgentState struct {
    Status     AgentStatus // Running, Idle, Working, Stalled, Stopped
    CurrentTask string     // Task ID being worked on (empty if idle)
    LastActivity time.Time
    Uptime      time.Duration
}

type AgentStatus int
const (
    StatusStopped AgentStatus = iota
    StatusStarting
    StatusIdle
    StatusWorking
    StatusStalled
)

// PingResult is the response from a health check.
type PingResult struct {
    OK       bool
    Latency  time.Duration
    Message  string // Optional status message from the agent
}

// AgentMetrics are runtime metrics for an agent.
type AgentMetrics struct {
    TasksCompleted  int
    TasksFailed     int
    TotalUptime     time.Duration
    LastPingLatency time.Duration
    MemoryUsageMB   float64 // 0 if not measurable
}

// TaskDescriptor describes a task to assign to an agent.
type TaskDescriptor struct {
    ID          string
    Title       string
    Description string
    Priority    int
    Labels      []string
    Branch      string // Git branch to work on
    Files       []string // Relevant files (hint for the agent)
}
```

### 2.4 Built-in Runtime Adapters

| Adapter | Transport | Attach? | I/O Model | Key Trade-off |
|---------|-----------|---------|-----------|---------------|
| `claude-code` | tmux session | Yes | `send-keys` / `capture-pane` | Rich observation, requires tmux |
| `codex` | Subprocess | No | stdin/stdout pipes | Sandboxed, one-shot execution model |
| `gemini-cli` | Subprocess | No | stdin/stdout pipes | Similar to codex |
| `agent-sdk` | Subprocess (Python/TS) | No | File-based IPC | API-native, no CLI needed |
| `subprocess` | Subprocess | No | stdin/stdout pipes | Generic, wraps any CLI |
| `docker` | Docker container | No | Docker exec / logs | Strongest isolation |

### 2.5 Adapter Correctness Properties

Every RuntimeAdapter implementation MUST satisfy these properties:

**P1 — Idempotent Stop:** Calling `Stop(handle)` on an already-stopped agent returns `nil`, not an error. This enables retry-safe shutdown sequences.

**P2 — Liveness After Start:** If `Start(ctx, config)` returns `(handle, nil)`, then `IsRunning(handle)` returns `true` within the agent's configured `ping_timeout`. Formally: `Start(cfg) = (h, nil) ⟹ ∃t ≤ ping_timeout: IsRunning(h) @ t = true`.

**P3 — Graceful Degradation:** If `SupportsAttach()` returns `false`, then `Attach(handle)` returns `ErrNotSupported` (not a panic, not a hang).

**P4 — Bounded Resource Cleanup:** `Stop(handle, graceful=false)` guarantees all OS resources (processes, file descriptors, temp directories) are released within `kill_cooldown`. Formally: let `R(h, t)` be the set of OS resources held by handle `h` at time `t`. Then `Stop(h, false) @ t₀ ⟹ R(h, t₀ + kill_cooldown) = ∅`.

**P5 — Thread Safety:** Concurrent calls to any combination of `Nudge()`, `CaptureOutput()`, `ReadOutput()`, `GetState()`, and `Ping()` on the same handle are safe. Adapters MUST use internal synchronization (mutexes or channels).

### 2.6 The `claude-code` Adapter Implementation Map

This adapter wraps Gas Town's existing tmux integration:

| RuntimeAdapter Method | Gas Town Code | Implementation |
|----------------------|--------------|----------------|
| `Start()` | `tmux.NewSessionWithCommandAndEnv()` | Create tmux session, inject env vars from `AgentConfig.Env`, send start command |
| `Stop(graceful=true)` | `tmux.KillSession()` | Send `ESC` + wait 5s + `tmux kill-session` |
| `Stop(graceful=false)` | `tmux.KillSessionWithProcesses()` | `kill -9` all pane PIDs + `tmux kill-session` |
| `IsRunning()` | `tmux.HasSession() && tmux.IsAgentRunning()` | Check tmux session exists AND agent process is alive in pane |
| `Assign()` | Write task to agent's bead hook | Create/update the agent's hooked bead, then `Nudge()` |
| `Nudge()` | `NudgeSession()` | `tmux.SendKeys()` with literal mode, debounce, separate Enter |
| `GetState()` | Read agent's hooked bead | Query beads DB for agent's current assignment |
| `SendInput()` | `tmux.SendKeys()` | Raw text input to tmux pane |
| `CaptureOutput()` | `tmux.CapturePane()` | Return piped output from `tmux capture-pane -p` |
| `ReadOutput()` | `tmux.CapturePane()` | Snapshot (non-streaming) version |
| `Ping()` | Capture-pane + prompt detection | Read pane output, detect prompt indicator (e.g., `$`, `❯`) |
| `Attach()` | `tmux.AttachSession()` | `tmux attach-session -t <name>` |

### 2.7 The `subprocess` Adapter Implementation

Generic adapter for wrapping any CLI tool:

```go
type SubprocessAdapter struct{}

func (s *SubprocessAdapter) Start(ctx context.Context, config AgentConfig) (AgentHandle, error) {
    cmd := exec.CommandContext(ctx, config.Command, config.Args...)
    cmd.Dir = config.WorkDir
    cmd.Env = mergeEnv(os.Environ(), config.Env)

    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()

    if err := cmd.Start(); err != nil {
        return AgentHandle{}, fmt.Errorf("subprocess start: %w", err)
    }

    handle := AgentHandle{
        ID:        uuid.New().String(),
        Agent:     config.Name,
        Runtime:   "subprocess",
        PID:       cmd.Process.Pid,
        StartedAt: time.Now(),
    }

    // Store pipes in internal registry keyed by handle.ID
    s.register(handle.ID, cmd, stdin, stdout, stderr)
    return handle, nil
}

func (s *SubprocessAdapter) Nudge(handle AgentHandle, message string) error {
    stdin := s.getStdin(handle.ID)
    _, err := fmt.Fprintln(stdin, message)
    return err
}

func (s *SubprocessAdapter) CaptureOutput(handle AgentHandle) (io.ReadCloser, error) {
    return s.getStdout(handle.ID), nil
}

func (s *SubprocessAdapter) IsRunning(handle AgentHandle) bool {
    return s.getCmd(handle.ID).ProcessState == nil // not yet exited
}

func (s *SubprocessAdapter) SupportsAttach() bool { return false }
```

### 2.8 Adapter Registry

```go
var adapterRegistry = map[string]RuntimeAdapter{
    "claude-code": &ClaudeCodeAdapter{},
    "codex":       &SubprocessAdapter{DefaultCommand: "codex"},
    "gemini-cli":  &SubprocessAdapter{DefaultCommand: "gemini"},
    "agent-sdk":   &AgentSDKAdapter{},
    "subprocess":  &SubprocessAdapter{},
    "docker":      &DockerAdapter{},
}

func GetAdapter(name string) (RuntimeAdapter, error) {
    adapter, ok := adapterRegistry[name]
    if !ok {
        return nil, fmt.Errorf("unknown runtime adapter: %q", name)
    }
    return adapter, nil
}

func RegisterAdapter(name string, adapter RuntimeAdapter) {
    adapterRegistry[name] = adapter
}
```

---

## 3. Configuration Schema

### 3.1 File Format and Location

- **Format:** TOML (consistent with Gas Town's formula system)
- **File:** `gas-city.toml` at the workspace root
- **Detection logic:**
  1. `gas-city.toml` exists → Gas City mode
  2. `mayor/town.json` exists without `gas-city.toml` → Gas Town compatibility mode
  3. Neither exists → fresh workspace, `gc init` required

### 3.2 Full Schema

```toml
# === REQUIRED ===
[workspace]
name = "string"                 # Workspace identifier
schema_version = 1              # Config schema version (for forward compat)
theme = "default"               # "default" or "gas-town" (enables GT naming in output)

# === LEVEL 0: Agent Runtime ===
[[agents]]
name = "string"                 # Required. Agent identifier (unique within workspace).
runtime = "string"              # Runtime adapter: "claude-code", "codex", "agent-sdk", "subprocess", "docker"
                                # Omit for auto-detection.
role = "worker"                 # "coordinator", "supervisor", "observer", "integrator",
                                # "worker", "agent", "service", or any custom string
scope = "project"               # "workspace" (one total) or "project" (one per project)
ephemeral = false               # Whether this agent is created/destroyed per-task
isolation = "worktree"          # "none", "worktree", "directory", "container"

[agents.runtime_config]         # Adapter-specific settings
command = "claude"              # Command to run
args = ["--dangerously-skip-permissions"]
env = { KEY = "VALUE" }         # Extra env vars for the runtime
resume_flag = "--resume"        # How this runtime resumes sessions
resume_style = "flag"           # "flag" or "subcommand"

[agents.pool]                   # Pool sizing for ephemeral agents
min = 0                         # Minimum instances
max = 5                         # Maximum instances
idle_timeout = "5m"             # Time before idle instances are killed

[agents.loop]                   # Task loop configuration
enabled = false
auto_execute = false            # GUPP: auto-start when work appears
clear_context = false           # Clear agent context between tasks
poll_interval = "10s"           # How often to check for new tasks

[agents.health]                 # Health check configuration
ping_timeout = "30s"
stuck_threshold = "1h"
consecutive_failures = 3        # Failures before restart
kill_cooldown = "5m"            # Time between forced restarts

[agents.prompts]                # Prompt templates
system = "prompts/agent.md"     # Path to system prompt (Go template)
nudge = "Check for work."       # Default nudge message

[agents.hooks]                  # Lifecycle hooks (shell commands)
on_start = ""                   # Runs after agent starts
on_stop = ""                    # Runs after agent stops
on_task_assign = ""             # Runs when task is assigned
on_task_complete = ""           # Runs when task completes
on_task_fail = ""               # Runs when task fails
on_stall = ""                   # Runs when agent is detected stalled

# === LEVEL 1: Work Tracking ===
[projects.<name>]               # One or more projects
repo = "string"                 # Git repository URL or "." for local
branch = "main"                 # Default branch

[tasks]
backend = "beads"               # "beads", "github-issues", "linear", "jira", "filesystem"

[tasks.beads]                   # Backend-specific config (beads)
dolt_port = 3307
data_dir = ".dolt-data"

[tasks.github]                  # Backend-specific config (github-issues)
owner = "org"
repo = "repo"
labels = ["agent-work"]

# === LEVEL 4: Messaging ===
[messaging]
backend = "beads"               # "beads", "redis", "filesystem"
channels = ["updates", "reviews"]

# === LEVEL 5: Workflows ===
[workflows]
templates_dir = "workflows/"    # Directory containing TOML workflow templates

# === LEVEL 6: Monitoring (implicit when supervisor role exists) ===
# No extra config section needed — monitoring is enabled when a
# supervisor agent is configured.

# === LEVEL 7: Multi-project (implicit when 2+ projects defined) ===
[batches]
enabled = false                 # Enable batch (convoy) tracking
auto_track = false              # Auto-create batches for multi-task dispatches
```

### 3.3 Progressive Level Detection

The config parser determines the workspace's capability level:

```go
func DetectLevel(cfg *WorkspaceConfig) int {
    level := 0
    if len(cfg.Agents) > 0 {
        level = 0 // Agent Runtime
    }
    if cfg.Tasks != nil && len(cfg.Projects) > 0 {
        level = 1 // Work Tracking
    }
    if hasLoopEnabled(cfg) {
        level = 2 // Task Loop (Ralph shape)
    }
    if hasPooledWorkers(cfg) || hasCoordinator(cfg) {
        level = 3 // Worker Pool (Agent Teams shape)
    }
    if cfg.Messaging != nil {
        level = 4 // Inter-Agent Messaging
    }
    if cfg.Workflows != nil {
        level = 5 // Workflow Templates
    }
    if hasSupervisor(cfg) {
        level = 6 // Monitoring
    }
    if len(cfg.Projects) >= 2 && hasProjectScopedAgents(cfg) {
        level = 7 // Full Orchestration (Gas Town shape)
    }
    return level
}
```

**Monotonicity invariant:** Adding a config section never breaks lower-level functionality. Formally: if config `C` is valid at level `n`, then `C ∪ {new_section}` is valid at level `n` or higher. The parser validates this: it never rejects a config that was valid before adding a section.

### 3.4 Config Validation Rules

| Rule | Check | Error Message |
|------|-------|---------------|
| Agent names unique | No duplicate `agents.name` | "duplicate agent name: X" |
| Pool only on ephemeral | `agents.pool` requires `ephemeral = true` | "pool config requires ephemeral = true" |
| Scope consistency | `scope = "project"` requires `[projects]` | "project-scoped agent X requires at least one project" |
| Runtime exists | `agents.runtime` in adapter registry | "unknown runtime: X" |
| Backend exists | `tasks.backend` in backend registry | "unknown task backend: X" |
| Coordinator limit | At most 1 coordinator per workspace | "multiple coordinators not supported" |
| Supervisor limit | At most 1 supervisor per workspace | "multiple supervisors not supported" |

---

## 4. The Three Milestone Shapes

### 4.1 Ralph (Level 2)

A single agent with a task loop that autonomously processes queued work.

```toml
[workspace]
name = "ralph-demo"
schema_version = 1

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
poll_interval = "30s"
```

**Task loop algorithm:**

```
LOOP:
  tasks = taskBackend.ListReady(filter: {assignee: self.name})
  if len(tasks) == 0:
    sleep(poll_interval)
    goto LOOP
  task = tasks[0]  // highest priority
  taskBackend.SetStatus(task.id, IN_PROGRESS)
  runtime.Assign(self.handle, task)
  WAIT_COMPLETION:
    state = runtime.GetState(self.handle)
    if state.Status == Working:
      sleep(5s)
      goto WAIT_COMPLETION
    if state.Status == Idle:
      taskBackend.SetStatus(task.id, COMPLETED)
      if clear_context:
        runtime.Stop(self.handle, graceful=true)
        self.handle = runtime.Start(ctx, self.config)
      goto LOOP
    if state.Status == Stalled:
      taskBackend.SetStatus(task.id, FAILED, reason: "agent stalled")
      runtime.Restart(self.handle)
      goto LOOP
```

**Edge cases:**
- Agent crashes mid-task: Task stays IN_PROGRESS. On restart, the loop picks it back up.
- No tasks available: Loop sleeps for `poll_interval`, doesn't spin.
- `clear_context = false`: Agent keeps context between tasks (useful for related tasks).
- Task backend unavailable: Log error, retry with exponential backoff (max 5m).

### 4.2 Agent Teams (Level 3)

A coordinator dispatches tasks to a pool of ephemeral workers.

```toml
[workspace]
name = "agent-teams-demo"
schema_version = 1

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "lead"
role = "coordinator"
runtime = "claude-code"

[[agents]]
name = "devs"
role = "worker"
runtime = "claude-code"
ephemeral = true
isolation = "worktree"

[agents.pool]
min = 0
max = 5
idle_timeout = "5m"

[agents.loop]
enabled = true
auto_execute = true
clear_context = true
```

**Pool scaling algorithm:**

```
// PoolManager runs continuously while the workspace is active.
func (pm *PoolManager) Run(ctx context.Context) {
    for {
        pending = pm.taskBackend.CountPending(assignee: pm.agentName)
        running = pm.countRunning()

        // Scale up: spawn workers for pending tasks up to max
        needed = min(pending, pm.config.Pool.Max) - running
        if needed > 0:
            for i := 0; i < needed; i++:
                workerConfig = pm.makeWorkerConfig(i)
                handle, err = pm.runtime.Start(ctx, workerConfig)
                if err != nil:
                    log.Error("pool scale-up failed", err)
                    break
                pm.workers = append(pm.workers, handle)

        // Scale down: kill idle workers beyond min
        idle = pm.findIdleWorkers(older_than: pm.config.Pool.IdleTimeout)
        excess = max(0, len(idle) - pm.config.Pool.Min)
        for i := 0; i < excess; i++:
            pm.runtime.Stop(idle[i], graceful=true)
            pm.removeWorker(idle[i])

        sleep(10s)
    }
}
```

**Pool invariant:** At any time `t`, `Min ≤ |running_workers(t)| ≤ Max`. Enforced by the PoolManager which is the sole creator/destroyer of pool instances.

**Edge cases:**
- Worker crashes: PoolManager detects via `IsRunning()` check, removes from pool. Pending tasks are reassigned.
- All workers busy, new task arrives: Task stays in queue until a worker completes and picks it up, or pool scales up.
- Coordinator crashes: Workers continue running. On coordinator restart, it rediscovers workers via the handle registry.
- `min = 0` and no tasks: Pool scales to zero. No resource usage when idle.

### 4.3 Gas Town (Level 7)

Full multi-project orchestration with all agent roles.

```toml
[workspace]
name = "my-town"
schema_version = 1
theme = "gas-town"

[projects.gastown]
repo = "https://github.com/steveyegge/gastown"

[projects.beads]
repo = "https://github.com/steveyegge/beads"

[tasks]
backend = "beads"
[tasks.beads]
dolt_port = 3307

[messaging]
backend = "beads"
channels = ["updates", "reviews", "escalations"]

[workflows]
templates_dir = "workflows/"

[batches]
enabled = true
auto_track = true

# Town-level agents
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

[agents.health]
ping_timeout = "30s"
stuck_threshold = "1h"
consecutive_failures = 3

[[agents]]
name = "services"
role = "service"
runtime = "claude-code"
scope = "workspace"
ephemeral = true
[agents.pool]
min = 0
max = 3

# Per-project agents
[[agents]]
name = "observer"
role = "observer"
scope = "project"
runtime = "claude-code"

[[agents]]
name = "integrator"
role = "integrator"
scope = "project"
runtime = "claude-code"

[[agents]]
name = "workers"
role = "worker"
runtime = "claude-code"
scope = "project"
ephemeral = true
isolation = "worktree"
[agents.pool]
min = 0
max = 5
[agents.loop]
enabled = true
auto_execute = true
clear_context = true
```

**Gas Town role mapping:**

| Gas Town Role | SDK Role | Scope | Behavior |
|--------------|----------|-------|----------|
| Mayor | coordinator | workspace | Dispatches tasks, breaks down epics |
| Deacon | supervisor | workspace | Health monitoring, agent restarts |
| Dog | service | workspace (pooled) | Background infrastructure tasks |
| Witness | observer | project | Monitors project state, tracks worker progress |
| Refinery | integrator | project | Merge queue processing, verification gates |
| Polecat | worker | project (pooled, ephemeral) | Executes individual tasks |
| Crew | agent | project (persistent) | Long-lived workspace agents |

---

## 5. Startup and Shutdown

### 5.1 Startup Sequencer

Agents start in priority groups. Within each group, agents start in parallel. Groups start sequentially.

```go
// Priority groups derived from agent config
var startupPriority = map[string]int{
    "supervisor":  0,  // Health monitor must be first to watch everything
    "coordinator": 1,  // Dispatcher starts after monitor is watching
    "observer":    2,  // Per-project monitors
    "integrator":  2,  // Per-project merge processors
    "agent":       2,  // Persistent agents
    "service":     3,  // Background services (on demand)
    "worker":      3,  // Workers (on demand)
}
// Custom roles default to priority 2 (same as per-project agents)
```

**Startup algorithm:**

```
func StartWorkspace(cfg *WorkspaceConfig) error {
    groups = groupByPriority(cfg.Agents)
    sort(groups by priority ascending)

    for _, group in groups:
        if group.priority >= 3:
            continue  // On-demand agents don't start at workspace startup

        var wg sync.WaitGroup
        for _, agentCfg in group.agents:
            if agentCfg.Scope == "project":
                // Start one instance per project
                for _, project in cfg.Projects:
                    wg.Add(1)
                    go startAgent(agentCfg, project, &wg)
            else:
                wg.Add(1)
                go startAgent(agentCfg, "", &wg)
        wg.Wait()

        // Verify all agents in this group are running
        for _, handle in group.handles:
            if !runtime.IsRunning(handle):
                return fmt.Errorf("agent %s failed to start", handle.Agent)
    return nil
}
```

**Startup ordering invariant:** For any two agents `a` and `b`, if `priority(a) < priority(b)`, then `a` is confirmed running before `b` starts. Formally: `IsRunning(a) = true` before `Start(b)` is called.

### 5.2 Shutdown Sequencer

Reverse of startup. Workers stop first (before monitors try to restart them), then observers, then coordinator, then supervisor.

```
func StopWorkspace(cfg *WorkspaceConfig) error {
    groups = groupByPriority(cfg.Agents)
    sort(groups by priority DESCENDING)  // Reverse order

    for _, group in groups:
        var wg sync.WaitGroup
        for _, handle in group.handles:
            wg.Add(1)
            go func(h AgentHandle) {
                runtime.Stop(h, graceful=true)
                wg.Done()
            }(handle)
        wg.Wait()

    return nil
}
```

**Edge cases:**
- Agent doesn't stop within timeout: Escalate to `Stop(handle, graceful=false)`.
- Supervisor restarts an agent during shutdown: Shutdown sets a `shuttingDown` flag that the supervisor checks before restarting.
- Orphaned processes: `gc stop --force` kills all processes matching the workspace's process group.

---

## 6. Event Bus

### 6.1 Event Schema

```go
type Event struct {
    Timestamp time.Time
    Type      EventType
    Agent     string  // Agent name (empty for system events)
    Project   string  // Project name (empty for workspace events)
    Payload   any     // Type-specific payload
}

type EventType string
const (
    EventAgentStarted   EventType = "agent.started"
    EventAgentStopped   EventType = "agent.stopped"
    EventAgentStalled   EventType = "agent.stalled"
    EventAgentCrashed   EventType = "agent.crashed"
    EventTaskCreated    EventType = "task.created"
    EventTaskAssigned   EventType = "task.assigned"
    EventTaskCompleted  EventType = "task.completed"
    EventTaskFailed     EventType = "task.failed"
    EventHealthPingOK   EventType = "health.ping_ok"
    EventHealthPingFail EventType = "health.ping_fail"
    EventHealthRestart  EventType = "health.restart"
    EventMessageSent    EventType = "message.sent"
    EventMessageRead    EventType = "message.read"
    EventPoolScaleUp    EventType = "pool.scale_up"
    EventPoolScaleDown  EventType = "pool.scale_down"
    EventWorkflowStep   EventType = "workflow.step_completed"
)
```

### 6.2 Bus Implementation

```go
type EventBus struct {
    mu          sync.RWMutex
    subscribers []chan<- Event
    buffer      *ring.Buffer[Event]  // Last 10,000 events for replay
}

func (b *EventBus) Publish(event Event) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    b.buffer.Push(event)
    for _, ch := range b.subscribers {
        select {
        case ch <- event:
        default:
            // Drop event if subscriber is slow (log warning)
        }
    }
}

func (b *EventBus) Subscribe() <-chan Event {
    ch := make(chan Event, 100)
    b.mu.Lock()
    b.subscribers = append(b.subscribers, ch)
    b.mu.Unlock()
    return ch
}
```

### 6.3 Built-in Consumers

| Consumer | Subscribes To | Output |
|----------|--------------|--------|
| CLI activity feed | All events | `gc activity --follow` real-time display |
| Structured logger | All events | JSON lines to `workspace/.gc/events.jsonl` |
| Hook executor | Configurable per event type | Runs user-defined shell commands |
| Metrics aggregator | Health + task events | Powers `gc stats` command |
| Supervisor | Agent lifecycle + health events | Triggers restart logic |

---

## 7. Health Monitoring

### 7.1 Supervisor Patrol Cycle

The supervisor agent runs a patrol cycle at regular intervals:

```
PATROL:
  for agent in workspace.allRunningAgents():
    if agent.role == "supervisor":
      continue  // Don't self-monitor

    result = runtime.Ping(agent.handle)
    bus.Publish(Event{Type: result.OK ? PingOK : PingFail, Agent: agent.name})

    if !result.OK:
      agent.consecutiveFailures++
      if agent.consecutiveFailures >= agent.health.consecutive_failures:
        if time.Since(agent.lastRestart) < agent.health.kill_cooldown:
          bus.Publish(Event{Type: AgentStalled, Agent: agent.name})
          continue  // Cooldown period, don't restart yet
        runtime.Restart(agent.handle)
        agent.consecutiveFailures = 0
        agent.lastRestart = time.Now()
        bus.Publish(Event{Type: HealthRestart, Agent: agent.name})
    else:
      agent.consecutiveFailures = 0

  sleep(30s)
  goto PATROL
```

### 7.2 Stall Detection

An agent is stalled when:
- It reports `Status = Working` for longer than `stuck_threshold`
- AND `Ping()` returns OK (agent is alive but not making progress)

Stall response: publish `EventAgentStalled`. The supervisor's default behavior is to wait for one more patrol cycle, then force-restart. Custom behavior via hooks:
```toml
[agents.hooks]
on_stall = "./scripts/page-oncall.sh {{.Agent}} {{.Project}}"
```

---

## 8. Task System

### 8.1 Task Backend Interface

```go
type TaskBackend interface {
    List(filter TaskFilter) ([]Task, error)
    Get(id string) (Task, error)
    Create(task Task) (string, error)  // Returns ID
    Update(id string, updates TaskUpdates) error
    CountPending(assignee string) (int, error)
}

type Task struct {
    ID          string
    Title       string
    Description string
    Status      TaskStatus  // Open, InProgress, Completed, Failed
    Assignee    string      // Agent name
    Priority    int
    Labels      []string
    Project     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Dependencies []string   // Task IDs that must complete first
}

type TaskFilter struct {
    Status   TaskStatus
    Assignee string
    Project  string
    Labels   []string
    NoDeps   bool  // Only tasks with no unresolved dependencies
}
```

### 8.2 Built-in Backends

**Beads backend:** Wraps Gas Town's existing Dolt-based beads system. Each task is a bead. Dependencies are bead dependencies. Status maps: Open → new, InProgress → in_progress, Completed → done, Failed → blocked.

**GitHub Issues backend:** Maps tasks to GitHub Issues via the `gh` CLI. Labels filter issues. Status maps: Open → open issue, InProgress → open + "in-progress" label, Completed → closed issue.

**Filesystem backend:** JSON files in `.gc/tasks/`. Zero dependencies. Good for simple setups and testing.

```go
// FilesystemBackend stores tasks as JSON files.
type FilesystemBackend struct {
    dir string  // e.g., ".gc/tasks/"
}

func (f *FilesystemBackend) Create(task Task) (string, error) {
    task.ID = uuid.New().String()[:8]
    task.CreatedAt = time.Now()
    data, _ := json.MarshalIndent(task, "", "  ")
    return task.ID, os.WriteFile(filepath.Join(f.dir, task.ID+".json"), data, 0644)
}
```

---

## 9. Messaging System

### 9.1 Messaging Interface

```go
type MessageBackend interface {
    Send(msg Message) error
    Inbox(agent string, unreadOnly bool) ([]Message, error)
    MarkRead(id string) error
    Delete(id string) error
    Subscribe(agent string) (<-chan Message, error)
}

type Message struct {
    ID        string
    From      string    // Agent name
    To        string    // Agent name or channel name
    Subject   string
    Body      string
    Timestamp time.Time
    Read      bool
    Channel   string    // Empty for direct messages
}
```

### 9.2 Channel Semantics

- Direct messages: `To = "agent-name"`, `Channel = ""`
- Channel messages: `To = ""`, `Channel = "channel-name"`
- All agents subscribed to a channel receive the message
- Channels are configured in `[messaging] channels = [...]`

---

## 10. Workflow Templates

### 10.1 Template Format

Workflow templates use TOML (consistent with Gas Town's formula system):

```toml
# workflows/code-review.toml
formula = "code-review"
description = "Multi-aspect parallel code review"
type = "aspect"   # "workflow" (sequential), "aspect" (parallel), "expansion" (macro)
version = 1

[[aspects]]
id = "correctness"
title = "Correctness Review"
description = "Logic errors, edge cases, off-by-one"

[[aspects]]
id = "security"
title = "Security Review"
description = "Injection, auth, data exposure"

[[aspects]]
id = "performance"
title = "Performance Review"
description = "Algorithmic complexity, resource leaks"
```

Sequential workflow:
```toml
# workflows/feature.toml
formula = "feature-pipeline"
type = "workflow"

[[steps]]
id = "design"
title = "Create design doc"

[[steps]]
id = "implement"
title = "Implement feature"
needs = ["design"]

[[steps]]
id = "test"
title = "Write and run tests"
needs = ["implement"]

[[steps]]
id = "review"
title = "Code review"
needs = ["test"]
```

### 10.2 Workflow Execution

```
func ExecuteWorkflow(template WorkflowTemplate, agents []AgentHandle) error {
    switch template.Type {
    case "workflow":
        return executeSequential(template.Steps, agents)
    case "aspect":
        return executeParallel(template.Aspects, agents)
    case "expansion":
        return expandAndExecute(template, agents)
    }
}

func executeSequential(steps []Step, agents []AgentHandle) error {
    completed = set{}
    for _, step in topologicalSort(steps):
        // Wait for dependencies
        for _, dep in step.Needs:
            waitUntil(completed.contains(dep))
        // Assign to next available agent
        agent = findAvailable(agents)
        runtime.Assign(agent, stepToTask(step))
        waitForCompletion(agent)
        completed.add(step.ID)
}

func executeParallel(aspects []Aspect, agents []AgentHandle) error {
    var wg sync.WaitGroup
    for i, aspect in aspects:
        agent = agents[i % len(agents)]
        wg.Add(1)
        go func(a Aspect, h AgentHandle) {
            runtime.Assign(h, aspectToTask(a))
            waitForCompletion(h)
            wg.Done()
        }(aspect, agent)
    wg.Wait()
}
```

---

## 11. CLI Design

### 11.1 Command Structure

The Gas City CLI is `gc`. It extends or replaces `gt` depending on configuration mode.

```
gc init [--shape ralph|agent-teams|gas-town|custom]
gc start
gc stop [--force]
gc status
gc level                           # Show capability level and progression

gc agent list
gc agent start <name>
gc agent stop <name>
gc agent attach <name>             # Interactive session (if runtime supports)
gc agent logs <name> [--follow]

gc task list [--status=X]
gc task create <title>
gc task assign <task-id> <agent>
gc task show <task-id>

gc mail send <to> -s "Subject" -m "Body"
gc mail inbox [--all]
gc mail read <id>

gc workflow run <template>
gc workflow status
gc workflow list

gc activity [--follow]             # Real-time event feed
gc stats                           # Agent metrics

gc validate                        # Config validation with diagnostics
gc migrate [--dry-run]             # Generate gas-city.toml from Gas Town workspace
gc doctor                          # Health checks (extended from gt doctor)

gc test-adapter <name>             # Run contract tests against an adapter

gc config show                     # Display resolved config
gc version
```

### 11.2 `gc init` Wizard

```
$ gc init

Welcome to Gas City SDK!

What kind of orchestration do you need?

  ▸ Ralph         — Single agent with task loop (simplest)
    Agent Teams   — Coordinator + worker pool
    Gas Town      — Full multi-project orchestration
    Custom        — Build your own topology

Select [Ralph]:

Which coding agent do you use?

  ▸ Claude Code   — claude --dangerously-skip-permissions
    Codex         — codex (OpenAI)
    Gemini CLI    — gemini
    Other         — Custom command

Select [Claude Code]:

Created gas-city.toml (Level 2 - Task Loop)
Run `gc start` to begin.
```

### 11.3 `gc level` Output

```
$ gc level

Workspace "my-project" is at Level 3 (Worker Pool)

  ✓ Level 0: Agent Runtime
  ✓ Level 1: Work Tracking
  ✓ Level 2: Task Loop
  ✓ Level 3: Worker Pool       ← you are here
  ○ Level 4: Messaging          — add [messaging] section
  ○ Level 5: Workflows          — add [workflows] section
  ○ Level 6: Health Monitoring  — add a supervisor agent
  ○ Level 7: Multi-Project      — add 2+ projects with project-scoped agents
```

---

## 12. Migration from Gas Town

### 12.1 Migration Algorithm

```
func Migrate(townRoot string) (*WorkspaceConfig, error) {
    cfg := &WorkspaceConfig{}

    // 1. Read town identity
    town := readJSON(townRoot + "/mayor/town.json")
    cfg.Workspace.Name = town.Name

    // 2. Read rig registry → projects
    rigs := readJSON(townRoot + "/mayor/rigs.json")
    for _, rig in rigs:
        cfg.Projects[rig.Name] = ProjectConfig{Repo: rig.URL}

    // 3. Read agent registries → agents
    townAgents := readJSON(townRoot + "/settings/agents.json")
    for _, agent in townAgents:
        cfg.Agents = append(cfg.Agents, convertAgent(agent))

    // 4. Map hardcoded roles → agent entries
    cfg.Agents = append(cfg.Agents, AgentEntry{Name: "coordinator", Role: "coordinator", Scope: "workspace"})
    cfg.Agents = append(cfg.Agents, AgentEntry{Name: "supervisor", Role: "supervisor", Scope: "workspace"})
    for _, rig in rigs:
        cfg.Agents = append(cfg.Agents, AgentEntry{Name: "observer-" + rig.Name, Role: "observer", Scope: "project"})
        cfg.Agents = append(cfg.Agents, AgentEntry{Name: "integrator-" + rig.Name, Role: "integrator", Scope: "project"})

    // 5. Read formulas → workflows
    formulas := readTOML(townRoot + "/.beads/formulas/*.toml")
    cfg.Workflows.TemplatesDir = "workflows/"

    // 6. Set task and messaging backends
    cfg.Tasks = &TaskConfig{Backend: "beads"}
    cfg.Messaging = &MessagingConfig{Backend: "beads"}

    return cfg, nil
}
```

### 12.2 Compatibility Guarantees

1. A Gas Town workspace without `gas-city.toml` continues to work with `gt` commands unchanged.
2. `gc migrate --dry-run` generates the equivalent config without modifying anything.
3. After migration, both `gt` and `gc` commands work (Gas Town compatibility mode persists).
4. Migration is additive: it creates `gas-city.toml` but doesn't modify existing Gas Town configs.

### 12.3 File-by-File Migration Map

| Gas Town File | Migration Target | Notes |
|--------------|-----------------|-------|
| `mayor/town.json` | `[workspace] name` | Direct mapping |
| `mayor/rigs.json` | `[projects.*]` | One entry per rig |
| `settings/agents.json` | `[[agents]]` entries | Agent configs with runtime_config |
| `settings/config.json` | Various sections | Town settings spread across config sections |
| `internal/config/roles/*.toml` | Built-in role behaviors | Not migrated — hardcoded in SDK |
| `.beads/formulas/*.toml` | `workflows/*.toml` | Copy with variable renaming ({{rig}} → {{project}}) |
| `~/.gt/hooks-base.json` | Not migrated | Claude Code hooks remain managed by `gt hooks` |
| `~/.gt/hooks-overrides/*.json` | `[agents.hooks]` | Flagged for manual migration |

---

## 13. Identity and Addressing

### 13.1 Agent Address Format

Gas City uses a simplified addressing scheme:

```
<workspace>/<agent-name>
<workspace>/<project>/<agent-name>
<workspace>/<project>/<agent-name>/<instance-id>
```

Examples:
```
my-town/coordinator          # Workspace-scoped agent
my-town/gastown/observer     # Project-scoped agent
my-town/gastown/workers/a3f  # Specific worker instance
```

### 13.2 Identity Resolution

```go
func ResolveAgent(address string) (*AgentHandle, error) {
    parts := strings.Split(address, "/")
    switch len(parts) {
    case 2:
        // workspace/agent — workspace-scoped agent
        return registry.FindByName(parts[1])
    case 3:
        // workspace/project/agent — project-scoped agent
        return registry.FindByNameAndProject(parts[2], parts[1])
    case 4:
        // workspace/project/agent/instance — specific pool instance
        return registry.FindByInstance(parts[3])
    default:
        return nil, fmt.Errorf("invalid address: %s", address)
    }
}
```

### 13.3 Backward Compatibility with Gas Town Addresses

Gas Town uses `gastown/polecats/Toast` format. The migration layer maps:
- `<rig>/polecats/<name>` → `<workspace>/<project>/workers/<instance>`
- `<rig>/crew/<name>` → `<workspace>/<project>/<name>`
- `mayor` → `<workspace>/coordinator`
- `deacon` → `<workspace>/supervisor`

---

## 14. Lifecycle Hooks

### 14.1 Hook Execution Model

```go
type HookExecutor struct {
    bus      *EventBus
    hooks    map[EventType][]HookConfig
    timeout  time.Duration  // Default 30s
}

func (h *HookExecutor) Run() {
    events := h.bus.Subscribe()
    for event := range events {
        configs, ok := h.hooks[event.Type]
        if !ok {
            continue
        }
        for _, cfg := range configs {
            go h.executeHook(cfg, event)
        }
    }
}

func (h *HookExecutor) executeHook(cfg HookConfig, event Event) {
    cmd := expandTemplate(cfg.Command, event)
    ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
    defer cancel()

    result := exec.CommandContext(ctx, "sh", "-c", cmd).Run()
    if result != nil {
        log.Warn("hook failed", "event", event.Type, "error", result)
    }
}
```

### 14.2 Template Variables

| Variable | Available On | Description |
|----------|-------------|-------------|
| `{{.Agent}}` | All agent events | Agent name |
| `{{.Project}}` | Project-scoped events | Project name |
| `{{.Task}}` | Task events | Task ID |
| `{{.TaskTitle}}` | Task events | Task title |
| `{{.Error}}` | Failure events | Error message |
| `{{.Timestamp}}` | All events | ISO-8601 timestamp |
| `{{.Status}}` | State change events | New status |

---

## 15. Testing Strategy

### 15.1 Unit Tests

| Subsystem | What It Tests | Mock Boundary |
|-----------|-------------|---------------|
| Config parser | TOML parsing, level detection, validation rules | Filesystem (embed test TOML) |
| Adapter registry | Registration, lookup, unknown adapter error | None (pure logic) |
| Pool manager | Scale up/down, min/max bounds, idle timeout | RuntimeAdapter mock |
| Task loop | Cycle execution, clear context, stall handling | RuntimeAdapter + TaskBackend mocks |
| Startup sequencer | Priority grouping, parallel start, failure handling | RuntimeAdapter mock |
| Shutdown sequencer | Reverse ordering, graceful timeout escalation | RuntimeAdapter mock |
| Event bus | Publish/subscribe, slow subscriber drop, replay | None (pure concurrency) |
| Hook executor | Template expansion, timeout, error handling | EventBus (inject events) |
| Identity resolver | Address parsing, backward compatibility | None (pure logic) |
| Migration | Town→City config conversion, edge cases | Filesystem (test fixtures) |

### 15.2 Contract Tests

Every RuntimeAdapter passes the same contract test suite:

```go
func RunAdapterContractTests(t *testing.T, adapter RuntimeAdapter, cfg AgentConfig) {
    t.Run("Start returns valid handle", func(t *testing.T) {
        handle, err := adapter.Start(ctx, cfg)
        require.NoError(t, err)
        require.NotEmpty(t, handle.ID)
        defer adapter.Stop(handle, true)
    })

    t.Run("IsRunning true after Start", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        require.Eventually(t, func() bool {
            return adapter.IsRunning(handle)
        }, cfg.Health.PingTimeout, 1*time.Second)
    })

    t.Run("Stop makes IsRunning false", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        adapter.Stop(handle, true)
        require.Eventually(t, func() bool {
            return !adapter.IsRunning(handle)
        }, 30*time.Second, 1*time.Second)
    })

    t.Run("Double Stop is idempotent", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        require.NoError(t, adapter.Stop(handle, true))
        require.NoError(t, adapter.Stop(handle, true))  // No error
    })

    t.Run("Nudge delivers message", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        waitForRunning(t, adapter, handle)
        require.NoError(t, adapter.Nudge(handle, "test-message-"+uuid()))
        output, _ := adapter.ReadOutput(handle)
        require.Contains(t, output, "test-message")
    })

    t.Run("Ping within timeout", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        waitForRunning(t, adapter, handle)
        result, err := adapter.Ping(handle)
        require.NoError(t, err)
        require.True(t, result.OK)
        require.Less(t, result.Latency, cfg.Health.PingTimeout)
    })

    t.Run("Attach matches SupportsAttach", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        if adapter.SupportsAttach() {
            require.NoError(t, adapter.Attach(handle))
            adapter.Detach(handle)
        } else {
            require.ErrorIs(t, adapter.Attach(handle), ErrNotSupported)
        }
    })

    t.Run("Concurrent access is safe", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        waitForRunning(t, adapter, handle)
        var wg sync.WaitGroup
        for i := 0; i < 10; i++ {
            wg.Add(3)
            go func() { adapter.Nudge(handle, "msg"); wg.Done() }()
            go func() { adapter.ReadOutput(handle); wg.Done() }()
            go func() { adapter.Ping(handle); wg.Done() }()
        }
        wg.Wait()
    })
}
```

### 15.3 Integration Tests

| Test | What It Validates | Prerequisites |
|------|-------------------|---------------|
| Claude-code adapter roundtrip | Start → Nudge → CaptureOutput → Stop | tmux, claude CLI |
| Subprocess adapter roundtrip | Start → SendInput → ReadOutput → Stop | None |
| Config → startup → shutdown | Parse TOML → start all → verify → stop all | claude CLI |
| Ralph smoke test | Single agent processes one task end-to-end | claude CLI, task backend |
| Agent Teams smoke test | Coordinator dispatches task to worker pool | claude CLI, task backend |
| Migration test | Existing Gas Town workspace → gc migrate → verify config | Gas Town workspace fixture |
| Level progression | Level 0 → 7 configs all start correctly | claude CLI |

### 15.4 Acceptance Criteria per Implementation Phase

| Phase | Acceptance Criteria |
|-------|-------------------|
| 1: Runtime Abstraction | All existing Gas Town tests pass with adapter wrapper. `gc agent start/stop/status` work. Contract tests pass for `claude-code` adapter. |
| 2: Config Parser | `gc validate` accepts all 8 levels. `gc level` reports correct level. `gc init` generates valid configs for all 3 shapes. |
| 3: Task System | `gc task create/list/assign` work with filesystem backend. Task loop runs for Ralph shape. |
| 4: Pool Manager | Agent Teams shape starts. Pool scales between min and max. Workers are created/destroyed. |
| 5: Messaging | `gc mail send/inbox` work. Direct messages and channels deliver. |
| 6: Workflows | `gc workflow run` executes sequential and parallel templates. |
| 7: Health Monitor | Supervisor detects stalled agents. Automatic restart with cooldown. |
| 8: Full Gas Town | Level 7 config starts all agents. Equivalent behavior to current Gas Town. |

---

## 16. Implementation Phases

### Phase 1: Agent Runtime Extraction (2-3 weeks)

**Goal:** Extract the `claude-code` adapter from Gas Town's existing tmux code without changing behavior.

**Deliverables:**
- `internal/runtime/adapter.go` — RuntimeAdapter interface
- `internal/runtime/registry.go` — Adapter registry
- `internal/runtime/claude_code/` — Claude Code adapter wrapping `internal/tmux/`
- `internal/runtime/subprocess/` — Generic subprocess adapter
- Contract test suite in `internal/runtime/contract_test.go`
- `gc agent start/stop/status` commands

**Acceptance:** All existing Gas Town tests pass. Contract tests pass for both adapters.

### Phase 2: Config Parser + CLI Foundation (1-2 weeks)

**Goal:** Parse `gas-city.toml`, detect levels, validate, and provide `gc init` + `gc level` + `gc validate`.

**Deliverables:**
- `internal/config/gas_city.go` — TOML parser with level detection
- `internal/config/validate.go` — Validation rules
- `gc init`, `gc level`, `gc validate`, `gc config show` commands
- Test fixtures for all 8 levels

**Acceptance:** All validation rules enforced. Level detection correct. Init wizard generates valid configs.

### Phase 3: Task System + Ralph Shape (2 weeks)

**Goal:** Task backend interface + filesystem backend + task loop = working Ralph.

**Deliverables:**
- `internal/tasks/backend.go` — TaskBackend interface
- `internal/tasks/filesystem/` — Filesystem backend
- `internal/tasks/beads/` — Beads backend adapter
- `internal/loop/` — Task loop controller
- `gc task create/list/assign/show` commands

**Acceptance:** Ralph shape processes tasks end-to-end. Task loop handles stalls and restarts.

### Phase 4: Pool Manager + Agent Teams Shape (2 weeks)

**Goal:** Pool management + coordinator dispatch = working Agent Teams.

**Deliverables:**
- `internal/pool/` — Pool manager with min/max/idle-timeout
- Startup/shutdown sequencer in `internal/orchestration/`
- Coordinator role behavior

**Acceptance:** Pool scales correctly. Agent Teams shape dispatches tasks to workers.

### Phase 5: Messaging + Workflows (1-2 weeks)

**Goal:** Inter-agent messaging and workflow template execution.

**Deliverables:**
- `internal/messaging/backend.go` — MessageBackend interface
- `internal/messaging/filesystem/` — Filesystem backend
- `internal/workflows/` — Workflow template parser and executor
- `gc mail` and `gc workflow` commands

**Acceptance:** Direct messages and channels work. Sequential and parallel workflows execute.

### Phase 6: Health Monitoring + Event Bus (1-2 weeks)

**Goal:** Supervisor patrol cycle + event bus + lifecycle hooks.

**Deliverables:**
- `internal/events/bus.go` — Event bus
- `internal/health/supervisor.go` — Supervisor patrol cycle
- `internal/hooks/executor.go` — Lifecycle hook executor
- `gc activity` and `gc stats` commands

**Acceptance:** Stalled agents detected and restarted. Events published. Hooks fire.

### Phase 7: Migration + Gas Town Shape (2 weeks)

**Goal:** Full Gas Town shape + migration tooling.

**Deliverables:**
- `internal/migration/` — Gas Town → Gas City migration
- `gc migrate` command with `--dry-run`
- Level 7 config with all role behaviors
- Observer, integrator, service role behaviors

**Acceptance:** `gc migrate --dry-run` generates correct config from Gas Town workspace. Gas Town shape starts all agents.

### Phase 8: Additional Adapters (ongoing)

**Goal:** Codex, Gemini, Agent SDK, Docker adapters.

**Deliverables:**
- Per-adapter package in `internal/runtime/`
- Contract tests passing for each

---

## 17. Exclusions with Rationale

| Exclusion | Rationale |
|-----------|-----------|
| Web dashboard | Important but separate concern. Event bus provides the data API. Dashboard is a consumer, not core SDK. |
| Cloud/remote execution | v1 is local-only. Docker adapter provides sandboxing. Cloud adapters (EC2, Cloud Run) are Phase 9+. |
| Multi-tenant isolation | One workspace per user. Multi-tenant requires security boundaries beyond v1 scope. |
| Agent marketplace | Sharing agent configs/templates is a community feature. The SDK provides the config format; sharing is orthogonal. |
| Automatic prompt engineering | SDK provides prompt template paths. Generating or optimizing prompts is AI-specific, not orchestration. |
| Cost management | Responsibility of the underlying runtime (Claude Code has built-in cost tracking). SDK tracks metrics, not billing. |
| Real-time collaboration | Agents coordinate through tasks and messages, not shared editing. |
| Hot config reload | Adding agents to a running workspace is desirable but complex. v1 requires restart. |

---

## 18. Formal Properties

### 18.1 Monotonic Capability Growth

Let `C(n)` be the set of config sections valid at level `n`. The progressive model guarantees:

```
∀ n ∈ [0,7]: C(n) ⊂ C(n+1)
```

**Proof sketch:** Each level adds a new optional config section ([tasks], [messaging], etc.) without modifying the schema of lower-level sections. The config parser accepts any superset of a valid config. Adding a section maps to a new level; it never invalidates existing sections.

### 18.2 Startup Ordering

For agents `a` and `b` with `priority(a) < priority(b)`:

```
IsRunning(a) = true → Start(b) is called
```

This is enforced by the sequencer's `sync.WaitGroup` barrier between priority groups.

### 18.3 Pool Bounds

For a pool with config `{min: M, max: N}`, at any time `t`:

```
M ≤ |running_instances(t)| ≤ N
```

**Invariant enforcement:** The PoolManager is the sole creator/destroyer of pool instances. It checks bounds before every scale operation:
- `scaleUp`: `if running < N then spawn(min(needed, N - running))`
- `scaleDown`: `if idle > M then kill(idle - M)`

### 18.4 Event Bus Liveness

Every event published to the bus is delivered to every subscriber with a non-full channel within one bus cycle (< 1ms under normal load). Events are dropped for slow subscribers (channel full) to prevent backpressure from stalling the bus. The ring buffer provides replay capability for up to 10,000 events.

---

## 19. Open Questions

1. **Out-of-process adapter protocol.** Go plugins are fragile. JSON-RPC over stdin/stdout (LSP-style) is more robust for custom adapters. Decision: implement JSON-RPC in Phase 8.

2. **Config hot-reload.** Can you add agents to a running workspace? v1: no (requires restart). v2: yes (via file watch + diff).

3. **Multi-runtime task serialization.** When results cross runtimes (Claude Code coordinator → Codex worker), the standard TaskResult schema (Section 8) provides the bridge. Validate in Phase 4.

4. **State persistence across restarts.** Claude-code uses tmux session survival + `/resume`. Other runtimes need checkpoint files. Define `StateCheckpoint` interface in Phase 3.

5. **`gc` vs `gt` CLI namespace.** Options: (a) `gc` is a separate binary, (b) `gt gc` subcommand, (c) `gt` detects Gas City mode and adjusts behavior. Recommendation: (a) separate binary for clarity.
