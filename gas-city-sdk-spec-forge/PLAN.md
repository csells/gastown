# Gas City SDK — Technical Specification

> **Version:** 0.5.0
> **Date:** 2026-02-14
> **Status:** Planning (Spec-Forge Stage 3, Round 3 integrated)
> **Source:** Independent spec-forge pipeline — research-notes.md + ideas-ranked.md
> **Review Round 1:** Codex (gpt-5.3), Gemini, Claude — feedback integrated
> **Review Round 2:** Codex (gpt-5.3), Gemini, Claude — feedback integrated
> **Review Round 3:** Codex (gpt-5.3), Gemini, Claude — feedback integrated (converged)

---

## 1. Executive Summary

Gas City is an **orchestration-builder SDK** that extracts Gas Town's hardcoded 7-role multi-agent orchestration into a configurable toolkit. The SDK lets users compose their own agent topologies — from a single task-loop agent up to full multi-project orchestration — using a declarative TOML configuration.

**The core problem:** Gas Town is monolithic. Its 7 roles (Mayor, Deacon, Witness, Refinery, Polecat, Crew, Dog) are hardcoded across 50+ packages, 396+ CLI commands, and deep tmux coupling. You can't use a subset of its capabilities, swap in different agent runtimes, or define custom topologies without forking the codebase.

**The core insight:** Orchestration is composable. Work tracking, messaging, health monitoring, task dispatch, and merge processing are independent capabilities that happen to be welded together in Gas Town. Separating them into composable units, each accessible through a uniform Agent Runtime interface, creates a toolkit that can express Gas Town (and any other topology) as a configuration.

**The positioning:** Gas City is the "Level 8" — for users who have outgrown any single orchestrator and want to build their own. Gas Town becomes one possible configuration of Gas City, alongside Ralph, Agent Teams, and any custom topology the user designs.

**What this spec covers:**
- The Agent Runtime abstraction (the foundational interface — the uniform "factory worker" abstraction)
- A progressive capability model (Levels 0-7) where each level adds one composable capability
- Three milestone configurations: Ralph, Agent Teams, Gas Town
- Custom roles, coordination rules, and worker instructions — all externalized in config and prompts
- Config schema (`gas-city.toml`) with full data structures
- Startup/shutdown sequencing with DAG-based dependency ordering
- Extensibility: sandboxes (isolation modes), plugins (adapter registry), hooks (lifecycle events)
- Migration from existing Gas Town workspaces with detailed reporting
- CLI design (`gc` commands)
- Testing strategy with contract tests per adapter (exposed via `gc test-adapter`)

**What this spec does NOT cover:**
- Web dashboard UI (future work; the SDK provides the data API)
- Cloud/remote execution (v1 is local-only)
- Multi-tenant isolation
- Agent marketplace / sharing
- Automatic prompt engineering
- Out-of-process plugin protocol (deferred to Phase 8; see Section 20.1)
- Pre-event hooks (post-event only in v1; see Section 20.2)
- Beads backend implementation details (existing Gas Town artifact; see Section 20.6)

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
    WaitForResult(ctx context.Context, handle AgentHandle) (TaskResult, error)  // Blocks until task completes or ctx expires
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

// ReadinessProbe is an optional interface adapters can implement
// to provide explicit readiness checking beyond simple process liveness.
type ReadinessProbe interface {
    IsReady(handle AgentHandle) (bool, error)
    WaitReady(ctx context.Context, handle AgentHandle) error
}

// Adopter is an optional interface for crash recovery.
// Adapters that can reconnect to a running process after a controller crash
// (e.g., tmux sessions, Docker containers) implement this interface.
// Adapters that lose their connection on crash (e.g., subprocess pipes) do NOT
// implement this — their agents will be restarted instead.
type Adopter interface {
    Adopt(ctx context.Context, identity AgentIdentity, metadata map[string]string) (AgentHandle, error)
}
```

**Interface design rationale:** The interface is intentionally broad (lifecycle + work + I/O + health) because all these concerns are coupled at the transport layer (tmux, subprocess, docker). Splitting into sub-interfaces would create combinatorial explosion for adapter authors. Instead, optional capabilities (Attach, ReadinessProbe, Adopter) use Go's interface assertion pattern: `if probe, ok := adapter.(ReadinessProbe); ok { ... }`.

### 2.2.1 Agent Task Protocol (Completion Detection)

`WaitForResult()` blocks until a task completes and returns a `TaskResult`. But how does an adapter *know* the task is done? The SDK defines three completion modes, configured per-agent:

```toml
[agents.runtime_config]
completion_mode = "prompt"    # "file", "prompt", or "exit"
```

**Mode 1: `file` (cooperative agents)**
The agent writes a result file to `<workdir>/.gc/result-<task-id>.json` when done. The adapter watches for this file (via `fsnotify` or polling). The file conforms to the `TaskResult` JSON schema. This is the most reliable mode and the recommended default for custom adapters.

**Mode 2: `prompt` (interactive agents — default for `claude-code`)**
The adapter detects the agent returning to an idle prompt after receiving a task. For `claude-code`, this means the tmux pane shows a prompt indicator. For other interactive CLIs, the adapter watches stdout for a configurable prompt pattern. This is a heuristic and may produce false positives if the agent prints a prompt-like string mid-task.

**Mode 3: `exit` (one-shot agents — default for `codex`, `subprocess`)**
The agent process exits after completing the task. Exit code 0 = Completed, non-zero = Failed. Stdout is captured as `TaskResult.Output`. Artifacts are detected by diffing the working directory before and after execution.

```go
func (a *SubprocessAdapter) WaitForResult(ctx context.Context, handle AgentHandle) (TaskResult, error) {
    cmd := a.getCmd(handle.ID)
    err := cmd.Wait()  // Blocks until process exits

    output, _ := a.readAllOutput(handle.ID)
    artifacts := a.detectChangedFiles(handle)

    result := TaskResult{
        TaskID:    handle.Metadata["current_task"],
        Output:    output,
        Artifacts: artifacts,
    }

    if err != nil {
        result.Status = StatusFailed
        result.Error = err.Error()
    } else {
        result.Status = StatusCompleted
    }
    return result, nil
}
```

The completion mode is part of the adapter contract: each built-in adapter has a default mode, and the contract tests verify it works correctly.

### 2.3 Core Data Structures

```go
// AgentIdentity is the stable, logical identity of an agent.
// It persists across restarts and is used for addressing/messaging.
type AgentIdentity struct {
    Workspace string  // Workspace name
    Project   string  // Project name (empty for workspace-scoped)
    Name      string  // Agent name from config (e.g., "coordinator")
    Instance  int     // Pool instance index (0 for non-pooled agents)
}

func (id AgentIdentity) String() string {
    if id.Project == "" {
        return fmt.Sprintf("%s/%s", id.Workspace, id.Name)
    }
    if id.Instance > 0 {
        return fmt.Sprintf("%s/%s/%s[%d]", id.Workspace, id.Project, id.Name, id.Instance)
    }
    return fmt.Sprintf("%s/%s/%s", id.Workspace, id.Project, id.Name)
}

// AgentHandle is a transient reference to a running agent instance.
// A new handle is created on each Start(). The Identity persists.
type AgentHandle struct {
    ID        string            // Unique ID for this runtime instance (UUID)
    Identity  AgentIdentity     // Logical identity (stable across restarts)
    Runtime   string            // Runtime adapter name (e.g., "claude-code")
    StartedAt time.Time
    PID       int               // OS process ID (0 if not applicable)
    Metadata  map[string]string // Runtime-specific metadata
}

// AgentConfig is the resolved config for starting an agent.
type AgentConfig struct {
    Name           string
    Role           string
    Runtime        string
    Project        string            // Empty for workspace scope
    Command        string            // Primary command (e.g., "claude")
    Args           []string
    Env            map[string]string
    WorkDir        string
    SystemPrompt   string            // Path or inline prompt
    NudgePrompt    string            // Initial message to send
    CompletionMode string            // "file", "prompt", or "exit"
    RuntimeConfig  map[string]any    // Adapter-specific settings
    DependsOn      []string          // Agent names that must start first
    Ephemeral      bool              // Created/destroyed per-task
    Isolation      string            // "none", "worktree", "directory", "container"
    Pool           *PoolConfig       // Pool sizing (nil = no pool)
    Loop           *LoopConfig       // Task loop config (nil = no loop)
    Health         *HealthConfig     // Health check config
    Hooks          *HookConfig       // Lifecycle hooks
}

type PoolConfig struct {
    Min         int
    Max         int
    IdleTimeout time.Duration
}

type LoopConfig struct {
    Enabled        bool
    AutoExecute    bool              // GUPP: auto-start on work
    ContextPolicy  string            // "clear", "keep", "summarize"
    PollInterval   time.Duration
}

type HealthConfig struct {
    PingTimeout         time.Duration
    StuckThreshold      time.Duration
    ConsecutiveFailures int
    KillCooldown        time.Duration
}

type HookConfig struct {
    OnStart      string
    OnStop       string
    OnTaskStart  string
    OnTaskEnd    string
    Critical     bool              // Critical hooks block; best-effort hooks are async
}

// AgentState represents the current state of an agent.
type AgentState struct {
    Status             AgentStatus // Running, Idle, Working, Stalled, Stopped
    CurrentTask        string      // Task ID being worked on (empty if idle)
    LastActivity       time.Time
    LastProgressUpdate time.Time   // When agent last reported progress
    ProgressMessage    string      // e.g., "compiling foo.rs", "running tests"
    WorkStartedAt      time.Time   // When current task was assigned
    Uptime             time.Duration
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
    TokensIn        int     // Input tokens consumed (0 if runtime doesn't report)
    TokensOut       int     // Output tokens generated
    EstimatedCostUSD float64 // Estimated cost in USD (0 if not available)
}

// TaskDescriptor describes a task to assign to an agent.
type TaskDescriptor struct {
    ID          string
    Title       string
    Description string
    Priority    int
    Labels      []string
    Branch      string   // Git branch to work on
    Files       []string // Relevant files (hint for the agent)
}

// TaskResult normalizes completion data across runtimes.
// Every adapter's WaitForResult() returns this structure.
type TaskResult struct {
    TaskID    string       `json:"task_id"`
    Status    TaskStatus   `json:"status"`    // Completed, Failed
    Output    string       `json:"output"`    // Summary text
    Artifacts []string     `json:"artifacts"` // File paths created/modified
    Metrics   AgentMetrics `json:"metrics"`   // Cost, tokens, time
    Error     string       `json:"error,omitempty"`
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

**P1 — Idempotent Stop:** Calling `Stop(handle)` on an already-stopped agent returns `nil`, not an error. The adapter distinguishes "already stopped" from "never existed" by tracking handle IDs in an internal registry. This enables retry-safe shutdown sequences.

**P2 — Liveness After Start:** If `Start(ctx, config)` returns `(handle, nil)`, then `IsRunning(handle)` returns `true` within the agent's configured `ping_timeout`. Formally: `Start(cfg) = (h, nil) => IsRunning(h) @ t = true` for some `t <= ping_timeout`.

**P3 — Graceful Degradation:** If `SupportsAttach()` returns `false`, then `Attach(handle)` returns `ErrNotSupported` (not a panic, not a hang).

**P4 — Bounded Resource Cleanup:** `Stop(handle, graceful=false)` guarantees all OS resources (processes, file descriptors, temp directories) are released within `kill_cooldown`. Formally: let `R(h, t)` be the set of OS resources held by handle `h` at time `t`. Then `Stop(h, false) @ t0 => R(h, t0 + kill_cooldown) = {}`.

**P5 — Thread Safety:** Concurrent calls to any combination of `Nudge()`, `CaptureOutput()`, `ReadOutput()`, `GetState()`, and `Ping()` on the same handle are safe. Adapters MUST use internal synchronization (mutexes or channels).

**P6 — Start Uniqueness:** Calling `Start()` with the same config creates a new, independent agent instance with a unique handle ID. Two concurrent `Start()` calls never return the same handle.

### 2.6 The `claude-code` Adapter Implementation Map

This adapter wraps Gas Town's existing tmux integration:

| RuntimeAdapter Method | Gas Town Code | Implementation |
|----------------------|--------------|----------------|
| `Start()` | `tmux.NewSessionWithCommandAndEnv()` | Create tmux session, inject env vars from `AgentConfig.Env`, send start command |
| `Stop(graceful=true)` | `tmux.KillSession()` | Send `ESC` + wait 5s + `tmux kill-session` |
| `Stop(graceful=false)` | `tmux.KillSessionWithProcesses()` | `kill -9` all pane PIDs + `tmux kill-session` |
| `IsRunning()` | `tmux.HasSession() && tmux.IsAgentRunning()` | Check tmux session exists AND agent process is alive in pane |
| `Assign()` | Write task to agent's bead hook | Create/update the agent's hooked bead, then `Nudge()` |
| `WaitForResult()` | Poll `GetState()` until idle or stalled | Loop until task completes, then read output and construct `TaskResult` |
| `Nudge()` | `NudgeSession()` | `tmux.SendKeys()` with literal mode, debounce, separate Enter |
| `GetState()` | Read agent's hooked bead | Query beads DB for agent's current assignment |
| `SendInput()` | `tmux.SendKeys()` | Raw text input to tmux pane |
| `CaptureOutput()` | `tmux.CapturePane()` | Return piped output from `tmux capture-pane -p` |
| `ReadOutput()` | `tmux.CapturePane()` | Snapshot (non-streaming) version |
| `Ping()` | Capture-pane + prompt detection | Read pane output, detect prompt indicator (e.g., `$`, `>`) |
| `Attach()` | `tmux.AttachSession()` | `tmux attach-session -t <name>` |

The `claude-code` adapter also implements `ReadinessProbe`:

```go
func (a *ClaudeCodeAdapter) IsReady(handle AgentHandle) (bool, error) {
    output, err := a.ReadOutput(handle)
    if err != nil {
        return false, err
    }
    return containsPrompt(output), nil
}

func (a *ClaudeCodeAdapter) WaitReady(ctx context.Context, handle AgentHandle) error {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            ready, err := a.IsReady(handle)
            if err != nil {
                return err
            }
            if ready {
                return nil
            }
        }
    }
}
```

### 2.7 The `subprocess` Adapter Implementation

Generic adapter for wrapping any CLI tool:

```go
type SubprocessAdapter struct{
    DefaultCommand string  // Fallback when AgentConfig.Command is empty
}

func (s *SubprocessAdapter) Start(ctx context.Context, config AgentConfig) (AgentHandle, error) {
    command := config.Command
    if command == "" {
        command = s.DefaultCommand  // Use adapter default (e.g., "codex", "gemini")
    }
    cmd := exec.CommandContext(ctx, command, config.Args...)
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
        Identity:  AgentIdentity{Name: config.Name},
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

// mergeEnv: system env vars form the base, config.Env overrides.
func mergeEnv(systemEnv []string, configEnv map[string]string) []string {
    envMap := make(map[string]string)
    for _, pair := range systemEnv {
        parts := strings.SplitN(pair, "=", 2)
        if len(parts) == 2 {
            envMap[parts[0]] = parts[1]
        }
    }
    for k, v := range configEnv {
        envMap[k] = v
    }
    result := make([]string, 0, len(envMap))
    for k, v := range envMap {
        result = append(result, k+"="+v)
    }
    return result
}
```

### 2.8 Adapter Registry

```go
var (
    adapterMu       sync.RWMutex
    adapterRegistry = map[string]RuntimeAdapter{
        "claude-code": &ClaudeCodeAdapter{},
        "codex":       &SubprocessAdapter{DefaultCommand: "codex"},
        "gemini-cli":  &SubprocessAdapter{DefaultCommand: "gemini"},
        "agent-sdk":   &AgentSDKAdapter{},
        "subprocess":  &SubprocessAdapter{},
        "docker":      &DockerAdapter{},
    }
)

func GetAdapter(name string) (RuntimeAdapter, error) {
    adapterMu.RLock()
    defer adapterMu.RUnlock()
    adapter, ok := adapterRegistry[name]
    if !ok {
        return nil, fmt.Errorf("unknown runtime adapter: %q", name)
    }
    return adapter, nil
}

func RegisterAdapter(name string, adapter RuntimeAdapter) {
    adapterMu.Lock()
    defer adapterMu.Unlock()
    adapterRegistry[name] = adapter
}
```

### 2.9 Runtime Auto-Detection

When `runtime` is omitted from an agent config, the SDK detects the available runtime:

```go
func DetectRuntime() string {
    // 1. Check for claude binary
    if _, err := exec.LookPath("claude"); err == nil {
        return "claude-code"
    }
    // 2. Check for codex binary
    if _, err := exec.LookPath("codex"); err == nil {
        return "codex"
    }
    // 3. Check for gemini binary
    if _, err := exec.LookPath("gemini"); err == nil {
        return "gemini-cli"
    }
    // 4. Check for docker
    if _, err := exec.LookPath("docker"); err == nil {
        return "docker"
    }
    // 5. Fallback
    return "subprocess"
}
```

---

## 3. Configuration Schema

### 3.1 File Format and Location

- **Format:** TOML (consistent with Gas Town's formula system)
- **File:** `gas-city.toml` at the workspace root
- **Detection logic:**
  1. `gas-city.toml` exists -> Gas City mode
  2. `mayor/town.json` exists without `gas-city.toml` -> Gas Town compatibility mode
  3. Neither exists -> fresh workspace, `gc init` required

- **Environment variable expansion:** Values in `env` maps support `${VAR}` expansion from the host environment and optional `.env` files in the workspace root. API keys and secrets should use this mechanism rather than being hardcoded in `gas-city.toml`.

- **TOML ordering note:** Sub-tables like `[agents.pool]` and `[agents.health]` attach to the most recently declared `[[agents]]` entry. Each agent's sub-tables must appear immediately after its `[[agents]]` header and before the next `[[agents]]` header. The config validator detects and reports orphaned sub-tables.

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
runtime = "string"              # Runtime adapter: "claude-code", "codex", "agent-sdk",
                                # "subprocess", "docker"
                                # If omitted, auto-detected:
                                #   1. Check for `claude` binary -> "claude-code"
                                #   2. Check for `codex` binary -> "codex"
                                #   3. Check for `gemini` binary -> "gemini-cli"
                                #   4. Check for `docker` binary -> "docker"
                                #   5. Fallback -> "subprocess"
role = "worker"                 # "coordinator", "supervisor", "observer", "integrator",
                                # "worker", "agent", "service", or any custom string
scope = "project"               # "workspace" (one total) or "project" (one per project)
ephemeral = false               # Whether this agent is created/destroyed per-task
isolation = "worktree"          # Isolation strategy:
                                # "none"      - Runs in workspace root (simple, shared state)
                                # "worktree"  - Git worktree (Gas Town default, branch isolation)
                                # "directory" - Copy of files to temp dir (no git overhead)
                                # "container" - Docker/OCI container (strongest isolation)
depends_on = []                 # Agent names that must start before this one

[agents.runtime_config]         # Adapter-specific settings
command = "claude"              # Command to run
args = ["--dangerously-skip-permissions"]
env = { KEY = "${MY_SECRET}" }  # Extra env vars (supports ${VAR} expansion)
completion_mode = "prompt"      # How task completion is detected:
                                # "file"   - Agent writes .gc/result-<task-id>.json (recommended)
                                # "prompt" - Adapter scans output for completion pattern
                                # "exit"   - Process exit = task done (subprocess only)
resume_flag = "--resume"        # How this runtime resumes sessions
resume_style = "flag"           # "flag" or "subcommand"

[agents.pool]                   # Pool sizing for ephemeral agents
min = 0                         # Minimum instances
max = 5                         # Maximum instances
idle_timeout = "5m"             # Time before idle instances are killed

[agents.loop]                   # Task loop configuration
enabled = false
auto_execute = false            # GUPP: auto-start when work appears
context_policy = "keep"         # Context management between tasks:
                                # "clear"     - Full reset, start fresh (safest)
                                # "keep"      - Retain full context (risks overflow)
                                # "summarize" - Summarize before next task (balanced)
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

The config parser determines the workspace's capability level. Each level requires the previous level as a prerequisite (monotonic progression):

```go
func DetectLevel(cfg *WorkspaceConfig) int {
    if cfg.Agents == nil || len(cfg.Agents) == 0 {
        return -1 // Invalid: no agents
    }

    level := 0 // Level 0: Agent Runtime (agents exist)

    if cfg.Tasks != nil && len(cfg.Projects) > 0 {
        level = 1 // Level 1: Work Tracking
    }

    if level >= 1 && hasAgentWithLoop(cfg) {
        level = 2 // Level 2: Task Loop (Ralph shape)
    }

    if level >= 2 && (hasPooledWorkers(cfg) || hasCoordinator(cfg)) {
        level = 3 // Level 3: Worker Pool (Agent Teams shape)
    }

    if level >= 3 && cfg.Messaging != nil {
        level = 4 // Level 4: Inter-Agent Messaging
    }

    if level >= 4 && cfg.Workflows != nil {
        level = 5 // Level 5: Workflow Templates
    }

    if level >= 5 && hasSupervisor(cfg) {
        level = 6 // Level 6: Monitoring
    }

    if level >= 6 && len(cfg.Projects) >= 2 && hasProjectScopedAgents(cfg) {
        level = 7 // Level 7: Full Orchestration (Gas Town shape)
    }

    return level
}
```

**Monotonicity invariant:** Each level check is guarded by `level >= n-1`, ensuring levels can only progress forward. Adding a config section never breaks lower-level functionality. Formally: if config `C` is valid at level `n`, then `C U {new_section}` is valid at level `n` or higher.

### 3.4 Config Validation Rules

| Rule | Check | Error Message |
|------|-------|---------------|
| Agent names unique | No duplicate `agents.name` | "duplicate agent name: X" |
| Pool only on ephemeral | `agents.pool` requires `ephemeral = true` | "pool config requires ephemeral = true" |
| Scope consistency | `scope = "project"` requires `[projects]` | "project-scoped agent X requires at least one project" |
| Runtime exists | `agents.runtime` in adapter registry (or omit for auto-detect) | "unknown runtime: X" |
| Backend exists | `tasks.backend` in backend registry | "unknown task backend: X" |
| Coordinator limit | At most 1 coordinator per workspace | "multiple coordinators not supported" |
| Supervisor limit | At most 1 supervisor per workspace | "multiple supervisors not supported" |
| Dependency validity | `depends_on` references existing agent names | "agent X depends on unknown agent Y" |
| Dependency acyclic | `depends_on` graph has no cycles | "dependency cycle: X -> Y -> X" |

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
context_policy = "clear"
poll_interval = "30s"
```

**Task loop algorithm:**

```
LOOP:
  tasks = taskBackend.List(filter: {
    status: Ready,
    claimableBy: self.name,  // Matches tasks assigned to self, self's pool, or unassigned
    sortBy: "priority",
    sortOrder: "desc"
  })
  if len(tasks) == 0:
    sleep(poll_interval)
    goto LOOP
  task = tasks[0]  // highest priority
  ok = taskBackend.Claim(task.id, self.identity)  // Atomic claim (uses identity string)
  if !ok:
    goto LOOP  // Another agent claimed it first
  runtime.Assign(self.handle, task)
  WAIT_COMPLETION:
    result = runtime.WaitForResult(ctx, self.handle)  // Blocks until done or ctx expires
    if result.Status == Completed:
      taskBackend.MarkCompleted(task.id, self.identity, result)
      if context_policy == "clear":
        runtime.Stop(self.handle, graceful=true)
        self.handle = runtime.Start(ctx, self.config)
      goto LOOP
    if result.Status == Failed:
      taskBackend.MarkFailed(task.id, self.identity, result.Error)
      runtime.Restart(self.handle)
      goto LOOP
```

**Edge cases:**
- Agent crashes mid-task: Task stays InProgress. On restart, the loop checks for orphaned InProgress tasks assigned to self and resumes or re-queues them.
- No tasks available: Loop sleeps for `poll_interval`, doesn't spin.
- `context_policy = "keep"`: Agent retains context between tasks (useful for related tasks). `"clear"` resets fully, `"summarize"` injects a task-history summary before the next task.
- Task backend unavailable: Log error, retry with exponential backoff (max 5m).
- Concurrent claim race: `Claim()` is atomic — only one agent wins. Losers retry.

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
context_policy = "clear"
```

**Pool scaling algorithm:**

```go
// PoolManager runs continuously while the workspace is active.
type PoolManager struct {
    agentName    string
    config       *PoolConfig
    runtime      RuntimeAdapter
    taskBackend  TaskBackend
    workers      []*AgentHandle
    mu           sync.Mutex  // Serializes scale operations with task dispatch
}

func (pm *PoolManager) Run(ctx context.Context) {
    // Initial scale-up to min (enforce pool invariant from the start)
    pm.mu.Lock()
    for i := pm.countRunning(); i < pm.config.Pool.Min; i++ {
        handle, err := pm.runtime.Start(ctx, pm.makeWorkerConfig(i))
        if err != nil {
            log.Error("initial pool scale-up failed", err)
            break
        }
        pm.workers = append(pm.workers, handle)
    }
    pm.mu.Unlock()

    // Main scaling loop
    for {
        // Phase 1: Compute scaling decisions under lock (fast, in-memory)
        pm.mu.Lock()
        pending, _ := pm.taskBackend.CountReady(pm.agentName)
        running := pm.countRunning()
        needed := min(pending, pm.config.Pool.Max) - running
        idle := pm.findIdleWorkers(pm.config.Pool.IdleTimeout)
        excess := max(0, len(idle) - pm.config.Pool.Min)
        pm.mu.Unlock()

        // Phase 2: Execute scaling operations outside lock (slow I/O)
        // Scale up
        var newHandles []*AgentHandle
        for i := 0; i < needed; i++ {
            workerConfig := pm.makeWorkerConfig(running + i)
            handle, err := pm.runtime.Start(ctx, workerConfig)
            if err != nil {
                log.Error("pool scale-up failed", err)
                break
            }
            if probe, ok := pm.runtime.(ReadinessProbe); ok {
                readyCtx, cancel := context.WithTimeout(ctx, workerConfig.Health.PingTimeout)
                if err := probe.WaitReady(readyCtx, handle); err != nil {
                    cancel()
                    pm.runtime.Stop(handle, true)
                    log.Error("worker failed readiness check", err)
                    continue
                }
                cancel()
            }
            newHandles = append(newHandles, handle)
        }

        // Scale down
        var removed []*AgentHandle
        for i := 0; i < excess; i++ {
            state, err := pm.runtime.GetState(idle[i])
            if err != nil || state.Status != StatusIdle {
                continue
            }
            stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
            err = pm.runtime.Stop(idle[i], true)
            cancel()
            if err != nil {
                log.Warn("graceful stop failed, escalating", "worker", idle[i], "error", err)
                pm.runtime.Stop(idle[i], false)
            }
            removed = append(removed, idle[i])
        }

        // Phase 3: Update worker list under lock
        pm.mu.Lock()
        pm.workers = append(pm.workers, newHandles...)
        for _, r := range removed {
            pm.removeWorker(r)
        }
        pm.mu.Unlock()
        sleep(10s)
    }
}
```

**Pool invariant:** At any time `t`, `Min <= |running_workers(t)| <= Max`. Enforced by the PoolManager holding `mu` during all scale operations, ensuring no concurrent scale-up/down races with task dispatch.

**Edge cases:**
- Worker crashes: PoolManager detects via `IsRunning()` check, removes from pool. Task stays InProgress and is reassigned on next cycle.
- All workers busy, new task arrives: Task stays in queue until a worker completes and picks it up, or pool scales up.
- Coordinator crashes: Workers continue running. On coordinator restart, it rediscovers workers via the handle registry.
- `min = 0` and no tasks: Pool scales to zero. No resource usage when idle.
- Scale-down race: The double-check on `GetState()` after acquiring `mu` prevents killing a worker that just received work.

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
depends_on = []  # Starts first (no dependencies)

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
depends_on = ["supervisor"]
[agents.pool]
min = 0
max = 3

# Per-project agents
[[agents]]
name = "observer"
role = "observer"
scope = "project"
runtime = "claude-code"
depends_on = ["coordinator"]

[[agents]]
name = "integrator"
role = "integrator"
scope = "project"
runtime = "claude-code"
depends_on = ["coordinator"]

[[agents]]
name = "workers"
role = "worker"
runtime = "claude-code"
scope = "project"
ephemeral = true
isolation = "worktree"
depends_on = ["coordinator", "observer"]
[agents.pool]
min = 0
max = 5
[agents.loop]
enabled = true
auto_execute = true
context_policy = "clear"
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

### 4.4 Custom Roles and Externalized Role Definitions

A core vision requirement: **roles are expressed externally, not hardcoded into the codebase.** Gas Town hardcodes role behaviors in Go packages (`internal/config/roles/*.toml` plus Go logic). Gas City externalizes them entirely.

**Role behavior is defined by three things:**
1. **Role name** in config: `role = "reviewer"` (any string, not a fixed enum)
2. **System prompt** in `[agents.prompts]`: defines the agent's behavior, instructions, and personality
3. **Coordination rules** in config: `depends_on`, `scope`, `ephemeral`, `isolation`, `pool`, `loop`, `hooks`

There is no hardcoded role logic in the SDK. The SDK provides the infrastructure (task dispatch, health monitoring, messaging, etc.) and the role definition provides the policy (what the agent does, when, and how).

**Example: Custom "reviewer" role**

```toml
[[agents]]
name = "reviewer"
role = "reviewer"                     # Custom role — no hardcoded behavior
scope = "project"
runtime = "claude-code"
depends_on = ["coordinator"]

[agents.prompts]
system = "roles/reviewer.md"          # External markdown file defines behavior
nudge = "Review the next pending PR."

[agents.hooks]
on_task_assign = "./scripts/checkout-pr.sh {{.Task}}"
on_task_complete = "./scripts/post-review.sh {{.Task}}"

[agents.loop]
enabled = true
auto_execute = true
poll_interval = "1m"
```

**`roles/reviewer.md`** (externalized role definition):
```markdown
You are a code reviewer for the {{.Project}} project.

Your responsibilities:
- Review pull requests assigned to you
- Check for correctness, security, and performance issues
- Post your review as a GitHub comment
- Mark the task as complete when done

Coordination rules:
- Wait for the coordinator to assign PRs
- If you find critical issues, send a message to the "escalations" channel
- If unsure, ask the coordinator for guidance via direct message
```

**Custom coordination rules** are expressed through the combination of:
- `depends_on`: startup ordering
- `[agents.hooks]`: event-driven actions (scripts run on lifecycle events)
- `[agents.prompts]`: agent instructions that reference messaging/task APIs
- `[agents.loop]`: autonomous polling behavior
- Workflow templates: multi-step processes with dependency chains

This means users can create entirely new orchestration topologies — a QA pipeline, a documentation generator, a multi-stage deployment system — without modifying SDK source code. Everything is configuration and prompts.

**Contrast with Gas Town:**

| Aspect | Gas Town | Gas City |
|--------|----------|----------|
| Role set | 7 hardcoded roles | Any string (user-defined) |
| Role behavior | Go code in `internal/` | External prompts + config |
| Coordination | Hardcoded in Mayor/Deacon | `depends_on` + hooks + workflows |
| Worker instructions | Embedded in role package | `[agents.prompts]` files |
| Adding a new role | Fork the codebase | Add `[[agents]]` to TOML + write a prompt |

---

## 5. Workspace Controller and Lifecycle

### 5.0 Workspace Controller

The workspace controller is the long-lived process that hosts all control loops: the event bus, agent registry, pool managers, supervisor patrol, hook executor, and workflow executor. It is the central coordination point that all other subsystems depend on.

```go
type WorkspaceController struct {
    config       *WorkspaceConfig
    registry     *AgentRegistry
    eventBus     *EventBus
    taskBackend  TaskBackend
    poolManagers map[string]*PoolManager
    supervisor   *Supervisor
    hookExec     *HookExecutor
    workflows    *WorkflowExecutor
    lockFile     string  // .gc/controller.lock
    socketPath   string  // .gc/controller.sock
}
```

**Lifecycle:**

1. **`gc start`** launches the controller. By default it runs in the foreground. Use `gc start --daemon` to background it.
2. The controller acquires `.gc/controller.lock` (via `flock`) to prevent multiple controllers per workspace.
3. It starts the event bus, loads adapters, then runs the startup sequencer (Section 5.1).
4. After all agents are running, it starts the control loops: pool managers, supervisor patrol, hook executor.
5. CLI commands (`gc status`, `gc task list`, etc.) connect to the controller via Unix socket at `.gc/controller.sock`.
6. **`gc stop`** sends a shutdown signal to the controller, which runs the shutdown sequencer (Section 5.2).

**Crash recovery:** If the controller crashes, agents continue running (they are independent OS processes). On next `gc start`, the controller rediscovers running agents via the persisted agent registry (`.gc/agents/*.json`). For each persisted agent, it checks whether the runtime adapter implements `Adopter`:
- **Adopter runtimes** (tmux, Docker): Controller calls `Adopt()` to reconnect to the running process and resume control without restart.
- **Non-Adopter runtimes** (subprocess): The original pipes are gone. Controller starts a fresh instance and re-queues any in-progress tasks.

**Lock semantics:** The lock file uses `flock()` (not exclusive create), so it is automatically released if the controller process dies. A subsequent `gc start` can acquire the lock cleanly.

**Formal guarantee — Single controller:** At most one controller process is active per workspace at any time. *Proof:* `flock(LOCK_EX|LOCK_NB)` on `.gc/controller.lock` returns `EWOULDBLOCK` if another process holds the lock. On failure, `gc start` exits with a "controller already running" error. On success, the lock persists until process exit (automatic flock release). There is no window where two controllers can both hold the lock.

### 5.1 Startup Sequencer

Agents start according to a dependency DAG. If no `depends_on` is specified, agents use default priority groups based on role. Within a dependency level, agents start in parallel.

```go
// Default priorities (used when depends_on is not specified)
var defaultPriority = map[string]int{
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

**Startup algorithm with DAG resolution and rollback:**

```go
func StartWorkspace(cfg *WorkspaceConfig) error {
    // Build dependency graph from depends_on + default priorities
    graph := buildStartupGraph(cfg.Agents)
    if cycle := graph.DetectCycle(); cycle != nil {
        return fmt.Errorf("dependency cycle: %s", strings.Join(cycle, " -> "))
    }

    groups := graph.TopologicalGroups()  // Groups of agents with no inter-dependencies
    started := []*AgentHandle{}

    rollback := func() {
        // Reverse order: stop most recently started first
        for i := len(started) - 1; i >= 0; i-- {
            runtime.Stop(started[i], graceful: true)
        }
    }

    for _, group := range groups {
        if group.onDemand {
            continue  // On-demand agents don't start at workspace startup
        }

        var wg sync.WaitGroup
        var startErrs []error
        var errMu sync.Mutex
        startedBefore := len(started)

        for _, agentCfg := range group.agents {
            instances := resolveInstances(agentCfg, cfg.Projects)
            for _, inst := range instances {
                wg.Add(1)
                go func(ac AgentConfig) {
                    defer wg.Done()
                    handle, err := startAgent(ac)
                    if err != nil {
                        errMu.Lock()
                        startErrs = append(startErrs, fmt.Errorf("agent %s failed: %w", ac.Name, err))
                        errMu.Unlock()
                        return
                    }
                    errMu.Lock()
                    started = append(started, handle)
                    errMu.Unlock()
                }(inst)
            }
        }
        wg.Wait()

        if len(startErrs) > 0 {
            rollback()
            return errors.Join(startErrs...)  // Report ALL failures, not just the last
        }

        // Verify all agents started in this group (use startedBefore, not group.agents count,
        // because project-scoped agents expand into multiple instances)
        for _, handle := range started[startedBefore:] {
            if !runtime.IsRunning(handle) {
                rollback()
                return fmt.Errorf("agent %s failed to start", handle.Identity.Name)
            }
        }
    }
    return nil
}
```

**Startup ordering invariant:** For any two agents `a` and `b`, if `a` is in `b.depends_on`, then `a` is confirmed running before `b.Start()` is called. Formally: `IsRunning(a) = true` before `Start(b)` is called.

**Rollback guarantee:** If any agent fails to start, all previously started agents are stopped in reverse order, leaving the workspace in a clean state.

### 5.2 Shutdown Sequencer

Reverse of startup. Workers stop first (before monitors try to restart them), then observers, then coordinator, then supervisor.

**In-progress task handling:**
- `gc stop` (graceful): Waits for in-progress tasks to complete up to `--drain-timeout` (default 5m). Tasks not completed within the timeout are re-queued as `Ready` with a `shutdown_interrupted` flag.
- `gc stop --force`: Stops agents immediately. Tasks left `InProgress` are marked with `shutdown_interrupted = true` for recovery on next `gc start`.

```go
func StopWorkspace(cfg *WorkspaceConfig, drainTimeout time.Duration, force bool) error {
    groups := buildStartupGraph(cfg.Agents).TopologicalGroups()
    slices.Reverse(groups)  // Reverse order for shutdown

    // Set shuttingDown flag so supervisor doesn't restart agents
    setShuttingDown(true)

    if !force && drainTimeout > 0 {
        // Wait for in-progress tasks to complete
        drainCtx, cancel := context.WithTimeout(ctx, drainTimeout)
        defer cancel()
        waitForInProgressTasks(drainCtx, taskBackend)
        // Any tasks still InProgress after drain timeout are re-queued
        requeueInterruptedTasks(taskBackend)
    } else if force {
        requeueInterruptedTasks(taskBackend)
    }

    for _, group := range groups {
        var wg sync.WaitGroup
        for _, handle := range group.handles {
            wg.Add(1)
            go func(h AgentHandle) {
                defer wg.Done()
                // Graceful stop with timeout escalation
                stopCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
                err := runtime.Stop(h, graceful: true)
                cancel()
                if err != nil {
                    log.Warn("graceful stop failed, forcing", "agent", h.Identity.Name)
                    runtime.Stop(h, graceful: false)
                }
            }(handle)
        }
        wg.Wait()
    }

    return nil
}
```

**Edge cases:**
- Agent doesn't stop within timeout: Escalate to `Stop(handle, graceful=false)`.
- Supervisor restarts an agent during shutdown: Shutdown sets a `shuttingDown` flag (field on WorkspaceController) that the supervisor checks before restarting.

**Formal guarantee — No work loss on shutdown:** For every task `T` with `Status == InProgress` at shutdown time, exactly one of: (a) `T` completes within the drain timeout and is marked `Completed`, or (b) `T` is re-queued as `Ready` with `ClaimedBy = ""`. No task can remain `InProgress` with a dead agent after shutdown completes.
- Orphaned processes: `gc stop --force` kills all processes matching the workspace's process group.

---

## 6. Event Bus

### 6.1 Event Schema

```go
type Event struct {
    Sequence  int64     // Monotonically increasing sequence number
    Timestamp time.Time
    Type      EventType
    Agent     string    // Agent name (empty for system events)
    Project   string    // Project name (empty for workspace events)
    Payload   any       // Type-specific payload
}

type EventType string
const (
    EventAgentStarted   EventType = "agent.started"
    EventAgentStopped   EventType = "agent.stopped"
    EventAgentStalled   EventType = "agent.stalled"
    EventAgentCrashed   EventType = "agent.crashed"
    EventAgentReady     EventType = "agent.ready"
    EventTaskCreated    EventType = "task.created"
    EventTaskClaimed    EventType = "task.claimed"
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
    EventWorkflowFailed EventType = "workflow.step_failed"
)
```

### 6.2 Bus Implementation

The event bus uses tiered subscribers to prevent dropping critical events while tolerating slow optional consumers:

```go
type EventBus struct {
    mu          sync.RWMutex
    sequence    int64
    critical    []EventSubscriber  // Block on these (supervisor, structured logger)
    optional    []EventSubscriber  // Fire-and-forget (CLI feed, metrics)
    buffer      *ring.Buffer[Event]  // Ring buffer (10k events) for catch-up on subscribe
}

type EventSubscriber interface {
    OnEvent(event Event) error
    IsCritical() bool
}

func (b *EventBus) Publish(event Event) error {
    b.mu.Lock()
    event.Sequence = atomic.AddInt64(&b.sequence, 1)
    b.buffer.Push(event)
    // Snapshot subscriber lists under lock to avoid races with Subscribe/Unsubscribe
    criticals := append([]EventSubscriber(nil), b.critical...)
    optionals := append([]EventSubscriber(nil), b.optional...)
    b.mu.Unlock()

    // Critical subscribers: block and propagate errors
    for _, sub := range criticals {
        if err := sub.OnEvent(event); err != nil {
            return fmt.Errorf("critical subscriber failed: %w", err)
        }
    }

    // Optional subscribers: fire-and-forget
    for _, sub := range optionals {
        go sub.OnEvent(event)
    }

    return nil
}

func (b *EventBus) Subscribe(sub EventSubscriber) {
    b.mu.Lock()
    // Snapshot buffer under lock
    snapshot := b.buffer.Snapshot()
    if sub.IsCritical() {
        b.critical = append(b.critical, sub)
    } else {
        b.optional = append(b.optional, sub)
    }
    b.mu.Unlock()

    // Replay outside lock so publishes are not blocked during replay
    for _, event := range snapshot {
        sub.OnEvent(event)
    }
}

func (b *EventBus) Unsubscribe(sub EventSubscriber) {
    b.mu.Lock()
    defer b.mu.Unlock()
    // Remove from appropriate list
    b.critical = removeSubscriber(b.critical, sub)
    b.optional = removeSubscriber(b.optional, sub)
}
```

**Subscriber tiers:**
- **Critical:** Supervisor, structured logger, hook executor. These block `Publish()` — if they fail, the event is not silently lost.
- **Optional:** CLI activity feed, metrics aggregator. These receive events asynchronously. Slow consumers may miss events but never stall the bus.

**Formal guarantee — Critical delivery:** For every event `e` published to the bus, if `Publish(e)` returns `nil`, then every critical subscriber `s` has processed `e` (i.e., `s.OnEvent(e)` returned `nil`). If any critical subscriber returns an error, `Publish` propagates that error and no further subscribers (critical or optional) are invoked. *Corollary:* Critical subscribers form a total order — subscriber `s_i` processes event `e` before `s_{i+1}` sees it.

**Formal guarantee — Sequence monotonicity:** Event sequences are monotonically increasing: `forall e1, e2: e1 published before e2 => e1.Sequence < e2.Sequence`. This is enforced by incrementing the sequence counter under the write lock before snapshot.

### 6.3 Built-in Consumers

| Consumer | Tier | Subscribes To | Output |
|----------|------|--------------|--------|
| Structured logger | Critical | All events + agent logs | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
| Hook executor | Critical | Configurable per event type | Runs user-defined shell commands |
| Supervisor | Critical | Agent lifecycle + health events | Triggers restart logic |
| CLI activity feed | Optional | All events | `gc activity --follow` real-time display |
| Metrics aggregator | Optional | Health + task events | Powers `gc stats` command |

---

## 7. Health Monitoring

### 7.1 Supervisor Patrol Cycle

The supervisor agent runs a patrol cycle at regular intervals:

```go
func (s *Supervisor) PatrolLoop(ctx context.Context) {
    for {
        if isShuttingDown() {
            return  // Don't interfere with shutdown
        }

        for _, agent := range s.workspace.AllRunningAgents() {
            if agent.Identity.Name == s.identity.Name {
                continue  // Don't self-monitor
            }

            result, err := s.runtime.Ping(agent.handle)
            if err != nil {
                s.bus.Publish(Event{Type: EventHealthPingFail, Agent: agent.Identity.Name})
                agent.consecutiveFailures++
            } else if result.OK {
                s.bus.Publish(Event{Type: EventHealthPingOK, Agent: agent.Identity.Name})
                agent.consecutiveFailures = 0
            } else {
                agent.consecutiveFailures++
            }

            if agent.consecutiveFailures >= agent.health.ConsecutiveFailures {
                if time.Since(agent.lastRestart) < agent.health.KillCooldown {
                    s.bus.Publish(Event{Type: EventAgentStalled, Agent: agent.Identity.Name})
                    continue  // Cooldown period, don't restart yet
                }
                s.runtime.Restart(agent.handle)
                agent.consecutiveFailures = 0
                agent.lastRestart = time.Now()
                s.bus.Publish(Event{Type: EventHealthRestart, Agent: agent.Identity.Name})
            }
        }

        select {
        case <-ctx.Done():
            return
        case <-time.After(30 * time.Second):
        }
    }
}
```

### 7.2 Stall Detection

An agent is stalled when:
- It reports `Status = Working` for longer than `stuck_threshold`
- AND it has not reported progress within `stuck_threshold` (via `LastProgressUpdate`)
- AND `Ping()` returns OK (agent is alive but not making progress)

```go
func (s *Supervisor) isStalled(handle *AgentHandle, state AgentState) bool {
    if state.Status != StatusWorking {
        return false
    }

    elapsed := time.Since(state.WorkStartedAt)
    noProgress := time.Since(state.LastProgressUpdate) > s.config.StuckThreshold

    // Not stalled if:
    // 1. Agent has reported progress recently, OR
    // 2. Elapsed time < threshold
    if !noProgress || elapsed < s.config.StuckThreshold {
        return false
    }

    // Final check: is the agent actually responsive?
    result, _ := s.runtime.Ping(handle)
    return result.OK  // Alive but not progressing = stalled
}
```

**Progress reporting:** Agents can update `LastProgressUpdate` and `ProgressMessage` by writing to a well-known progress file (`<workdir>/.gc/progress.json`) or through a heartbeat mechanism. The runtime adapter's `GetState()` reads this file to populate the fields. This distinguishes "slow legitimate work" (frequent progress updates) from "actual stall" (no updates for `stuck_threshold`).

Stall response: publish `EventAgentStalled`. The supervisor's default behavior is to wait for one more patrol cycle, then force-restart. Custom behavior via hooks:
```toml
[agents.hooks]
on_stall = "./scripts/page-oncall.sh {{.Agent}} {{.Project}}"
```

---

## 8. Task System

### 8.1 Task Backend Interface

```go
type TaskStatus int
const (
    StatusOpen       TaskStatus = iota
    StatusReady                        // All deps resolved, available for claim
    StatusInProgress                   // Claimed and being worked on
    StatusCompleted                    // Successfully finished
    StatusFailed                       // Terminal failure
    StatusBlocked                      // Blocked on unresolved dependencies
)

type TaskBackend interface {
    // CRUD
    List(filter TaskFilter) ([]Task, error)
    Get(id string) (Task, error)
    Create(task Task) (string, error)  // Returns ID
    Update(id string, updates TaskUpdates) error
    CountReady(claimableBy string) (int, error)  // Count tasks claimable by agent

    // Atomic operations for concurrency safety
    Claim(id string, agent string) (bool, error)                     // Atomic: claim or return false
    MarkInProgress(id string, agent string) error
    MarkCompleted(id string, claimedBy string, result TaskResult) error
    MarkFailed(id string, agent string, reason string) error
}

type Task struct {
    ID           string
    Title        string
    Description  string
    Status       TaskStatus
    Assignee     string      // Target agent/pool name (routing: WHO should do this)
    ClaimedBy    string      // Instance that claimed it (tracking: WHO is doing this)
    Priority     int
    Labels       []string
    Project      string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    Dependencies []string    // Task IDs that must complete first
    Version      int64       // Optimistic lock version (incremented on every update)
    Result       *TaskResult // Non-nil when completed
}

type TaskFilter struct {
    Status      TaskStatus
    Assignee    string   // Filter by target agent/pool name
    ClaimableBy string   // Filter for tasks claimable by this agent (Assignee matches or empty)
    Project     string
    Labels      []string
    NoDeps      bool     // Only tasks with no unresolved dependencies
    SortBy      string   // "priority", "created", "updated"
    SortOrder   string   // "asc", "desc"
}

type TaskUpdates struct {
    Status   *TaskStatus
    Assignee *string
    Result   *TaskResult
    Error    *string
    Version  int64  // Must match current version (optimistic lock)
}
```

**Status transition rules:**

```
Open -> Ready       (when all dependencies complete)
Ready -> InProgress (via Claim())
InProgress -> Completed (via MarkCompleted())
InProgress -> Failed    (via MarkFailed())
Failed -> Ready     (manual retry via gc task retry)
Blocked -> Ready    (when blocking dependency completes)
```

Invalid transitions (return error): `Completed -> *`, `Ready -> Open`, any backward transition not listed above.

**Dependency resolution:** When a task is created with `Dependencies` (list of task IDs), it starts in `StatusBlocked`. On each `MarkCompleted(taskID)`, the backend queries all tasks that list `taskID` in their `Dependencies`. For each, it checks whether *all* dependencies are now completed. If so, it transitions the task from `Blocked` to `Ready`, making it available for claiming. This resolution is triggered inside `MarkCompleted()` to ensure atomicity.

**Formal guarantee — No double-execution:** For any task `T`, exactly one of these holds at any time:
1. `T.ClaimedBy == ""` (unclaimed — no agent is executing it), OR
2. `T.ClaimedBy == agent_i` for exactly one agent `agent_i` (single executor)

*Proof:* The `Claim()` operation is atomic (flock/SQL transaction). It checks `ClaimedBy == ""` and sets `ClaimedBy = agent` in a single critical section. If two agents race, the backend's atomic primitive ensures at most one sees the empty state and succeeds. The `ClaimedBy` field is only cleared on explicit `MarkCompleted`/`MarkFailed` + retry, which transitions Status back to Ready before re-enabling claims.

**Formal guarantee — Dependency correctness:** A task `T` with dependencies `D = {d1, d2, ..., dk}` transitions to `Ready` if and only if `forall di in D: di.Status == Completed`. *Proof:* `MarkCompleted(di)` scans dependents and only transitions those where all deps are completed. Since `MarkCompleted` is called under a per-task lock (or transaction), the check-then-transition is atomic.

**Claim semantics:** `Claim(id, agent)` atomically checks `Status == Ready && ClaimedBy == ""`, sets `Status = InProgress, ClaimedBy = agent`, and increments `Version`. If the task is already claimed, it returns `(false, nil)`. This prevents two agents from claiming the same task. The `Assignee` field is for routing (which agent/pool should receive this task) and is set at creation time; `ClaimedBy` tracks which specific instance is executing it.

**Task assignment models:**
- **Push model (coordinator dispatches):** Coordinator creates tasks with `Assignee = "worker-pool"`. Workers in the pool list tasks where `ClaimableBy` matches their pool name and claim them.
- **Pull model (agents self-select):** Tasks created with empty `Assignee`. Any agent with matching labels/project can claim.
- **Direct assignment:** Task created with `Assignee = "specific-agent"`. Only that agent can claim it.

### 8.2 Built-in Backends

**Beads backend:** Wraps Gas Town's existing Dolt-based beads system. Each task is a bead. Dependencies are bead dependencies. Status maps: Open -> new, Ready -> ready, InProgress -> in_progress, Completed -> done, Failed -> blocked. `Claim()` uses Dolt's SQL transactions for atomicity.

**GitHub Issues backend:** Maps tasks to GitHub Issues via the `gh` CLI. Labels filter issues. Status maps: Open -> open issue, InProgress -> open + "in-progress" label, Completed -> closed issue. `Claim()` uses issue assignment + label as a two-step atomic operation (check-then-assign with retry).

**Filesystem backend:** JSON files in `.gc/tasks/`. Zero dependencies. Good for simple setups and testing. `Claim()` uses filesystem locking (`flock`).

```go
// FilesystemBackend stores tasks as JSON files.
type FilesystemBackend struct {
    dir string  // e.g., ".gc/tasks/"
}

func (f *FilesystemBackend) Create(task Task) (string, error) {
    task.ID = uuid.New().String()[:8]
    task.CreatedAt = time.Now()
    task.Version = 1
    data, _ := json.MarshalIndent(task, "", "  ")
    return task.ID, os.WriteFile(filepath.Join(f.dir, task.ID+".json"), data, 0644)
}

func (f *FilesystemBackend) Claim(id string, agent string) (bool, error) {
    path := filepath.Join(f.dir, id+".json")
    lockFile := path + ".lock"

    // Advisory lock via flock (automatically released if process crashes)
    lock, err := os.OpenFile(lockFile, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return false, err
    }
    defer lock.Close()

    if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
        return false, nil  // Lock held by another process
    }
    defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

    task, err := f.Get(id)
    if err != nil {
        return false, err
    }
    if task.Status != StatusReady || task.ClaimedBy != "" {
        return false, nil  // Already claimed
    }

    task.Status = StatusInProgress
    task.ClaimedBy = agent
    task.Version++
    task.UpdatedAt = time.Now()
    data, _ := json.MarshalIndent(task, "", "  ")
    return true, os.WriteFile(path, data, 0644)
}
```

### 8.3 Batch (Convoy) Tracking

A **batch** groups related tasks that were dispatched together, enabling aggregate progress tracking and coordinated completion. This maps to Gas Town's "convoy" concept where multiple beads move through the pipeline as a unit.

```go
type Batch struct {
    ID        string
    Name      string       // Human-readable label (e.g., "feature-auth")
    TaskIDs   []string     // Tasks in this batch
    CreatedAt time.Time
    Status    BatchStatus  // Pending, InProgress, Completed, PartialFailure
}

type BatchStatus int
const (
    BatchPending        BatchStatus = iota
    BatchInProgress                         // At least one task started
    BatchCompleted                          // All tasks completed
    BatchPartialFailure                     // Some tasks failed
)
```

**Behavior:**
- When `auto_track = true`, creating multiple tasks in a single coordinator dispatch automatically groups them into a batch.
- Batch status is derived: `InProgress` when any task is claimed, `Completed` when all tasks complete, `PartialFailure` when any task fails with others completed.
- `gc batch status <id>` shows aggregate progress. `gc batch list` shows active batches.

**CLI commands:**

```
gc batch list                          # Show active batches
gc batch status <id>                   # Aggregate progress for a batch
gc batch retry <id>                    # Retry failed tasks in a batch
```

---

## 9. Messaging System

### 9.1 Messaging Interface

```go
type MessageBackend interface {
    Send(msg Message) error
    Get(id string) (Message, error)
    Inbox(agent string, filter MessageFilter) ([]Message, error)
    MarkRead(id string) error
    Delete(id string) error
    Subscribe(agent string, since time.Time) (<-chan Message, error)
}

type Message struct {
    ID        string
    From      string      // Agent name (sender)
    To        string      // Agent name or channel name
    Subject   string
    Body      string
    Channel   string      // Empty for direct messages
    CreatedAt time.Time
    Read      bool
    ReadAt    *time.Time
    ExpiresAt *time.Time  // Optional TTL — nil means no expiration
}

type MessageFilter struct {
    UnreadOnly bool
    FromAgent  string
    Channel    string
    Since      time.Time
    Before     time.Time
    SortBy     string  // "created" (default), "received"
    SortOrder  string  // "asc" (default), "desc"
}
```

### 9.2 Channel Semantics

- Direct messages: `To = "agent-name"`, `Channel = ""`
- Channel messages: `To = ""`, `Channel = "channel-name"`
- All agents subscribed to a channel receive the message
- Channels are configured in `[messaging] channels = [...]`

### 9.3 Delivery Guarantees

- **Ordering:** Messages are delivered in `CreatedAt` order within a channel or direct conversation. `Subscribe(agent, since)` replays messages created after `since`, then streams new ones.
- **Delivery:** At-least-once for direct messages (retried on backend failure). At-most-once for channel broadcasts (fire-and-forget to each subscriber).
- **Idempotency:** `Send()` with the same message ID is a no-op. Message IDs are generated by the sender.
- **TTL:** Messages with `ExpiresAt` set are automatically cleaned up by a periodic garbage collector. Expired messages are excluded from `Inbox()` results.
- **Backpressure:** If a subscriber's channel buffer fills, new messages are buffered in the backend and delivered on next poll. The backend never blocks `Send()`.

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

[strategy]
on_step_failure = "fail"   # "fail", "skip", "retry"
max_retries = 2
step_timeout = "30m"

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

```go
type WorkflowExecution struct {
    TemplateID    string
    Steps         map[string]*StepExecution
    Status        WorkflowStatus  // Running, Completed, Failed
    StartedAt     time.Time
    CompletedAt   time.Time
    FailureReason string
}

type StepExecution struct {
    ID          string
    Status      StepStatus  // Pending, Running, Completed, Failed, Skipped
    Agent       *AgentHandle
    AssignedAt  time.Time
    CompletedAt time.Time
    Result      *TaskResult
    Error       string
    RetryCount  int
}

type WorkflowStrategy struct {
    OnStepFailure string         // "fail", "skip", "retry"
    MaxRetries    int
    StepTimeout   time.Duration
}

func ExecuteWorkflow(ctx context.Context, template WorkflowTemplate, agents []AgentHandle, strategy WorkflowStrategy) (*WorkflowExecution, error) {
    exec := &WorkflowExecution{
        TemplateID: template.ID,
        StartedAt:  time.Now(),
        Steps:      make(map[string]*StepExecution),
        Status:     WorkflowRunning,
    }

    switch template.Type {
    case "workflow":
        return executeSequential(ctx, template.Steps, agents, strategy, exec)
    case "aspect":
        return executeParallel(ctx, template.Aspects, agents, strategy, exec)
    case "expansion":
        return expandAndExecute(ctx, template, agents, strategy, exec)
    }
    return nil, fmt.Errorf("unknown workflow type: %s", template.Type)
}

func executeSequential(ctx context.Context, steps []Step, agents []AgentHandle, strategy WorkflowStrategy, exec *WorkflowExecution) (*WorkflowExecution, error) {
    for _, step := range topologicalSort(steps) {
        // Wait for dependencies
        for _, dep := range step.Needs {
            depExec := exec.Steps[dep]
            if depExec != nil && depExec.Status == StepFailed {
                if strategy.OnStepFailure == "fail" {
                    exec.Status = WorkflowFailed
                    exec.FailureReason = fmt.Sprintf("dependency %s failed", dep)
                    return exec, fmt.Errorf("workflow failed: dependency %s failed", dep)
                }
            }
        }

        agent := findAvailable(agents)
        stepExec := &StepExecution{
            ID:         step.ID,
            Agent:      &agent,
            AssignedAt: time.Now(),
            Status:     StepRunning,
        }
        exec.Steps[step.ID] = stepExec

        // Execute with retry loop
        var lastErr error
        for attempt := 0; attempt <= strategy.MaxRetries; attempt++ {
            stepCtx, cancel := context.WithTimeout(ctx, strategy.StepTimeout)
            runtime.Assign(agent, stepToTask(step))
            result, err := runtime.WaitForResult(stepCtx, agent)
            cancel()

            if err == nil && result.Status != Failed {
                // Success
                stepExec.Status = StepCompleted
                stepExec.Result = &result
                stepExec.CompletedAt = time.Now()
                bus.Publish(Event{Type: EventWorkflowStep, Payload: stepExec})
                lastErr = nil
                break
            }

            lastErr = coalesceErr(err, result.Error)
            stepExec.RetryCount = attempt + 1

            if strategy.OnStepFailure != "retry" || attempt >= strategy.MaxRetries {
                break  // Don't retry if strategy is fail/skip, or retries exhausted
            }
            log.Info("retrying step", "step", step.ID, "attempt", attempt+2)
        }

        if lastErr != nil {
            stepExec.Status = StepFailed
            stepExec.Error = lastErr.Error()

            switch strategy.OnStepFailure {
            case "fail":
                exec.Status = WorkflowFailed
                exec.FailureReason = fmt.Sprintf("step %s failed: %s", step.ID, stepExec.Error)
                return exec, fmt.Errorf("workflow failed at step %s", step.ID)
            case "skip":
                stepExec.Status = StepSkipped
            case "retry":
                exec.Status = WorkflowFailed
                return exec, fmt.Errorf("step %s failed after %d retries", step.ID, strategy.MaxRetries)
            }
        }
    }

    exec.Status = WorkflowCompleted
    exec.CompletedAt = time.Now()
    return exec, nil
}

func executeParallel(ctx context.Context, aspects []Aspect, agents []AgentHandle, strategy WorkflowStrategy, exec *WorkflowExecution) (*WorkflowExecution, error) {
    var wg sync.WaitGroup
    var firstErr error
    var errOnce sync.Once

    for i, aspect := range aspects {
        agent := agents[i % len(agents)]
        stepExec := &StepExecution{
            ID:         aspect.ID,
            Agent:      &agent,
            AssignedAt: time.Now(),
            Status:     StepRunning,
        }
        exec.Steps[aspect.ID] = stepExec

        wg.Add(1)
        go func(a Aspect, h AgentHandle, se *StepExecution) {
            defer wg.Done()

            stepCtx, cancel := context.WithTimeout(ctx, strategy.StepTimeout)
            defer cancel()

            runtime.Assign(h, aspectToTask(a))
            result, err := runtime.WaitForResult(stepCtx, h)

            if err != nil || result.Status == Failed {
                se.Status = StepFailed
                se.Error = coalesce(err, result.Error)
                errOnce.Do(func() {
                    firstErr = fmt.Errorf("aspect %s failed: %s", a.ID, se.Error)
                })
                bus.Publish(Event{Type: EventWorkflowFailed, Payload: se})
            } else {
                se.Status = StepCompleted
                se.Result = &result
                se.CompletedAt = time.Now()
                bus.Publish(Event{Type: EventWorkflowStep, Payload: se})
            }
        }(aspect, agent, stepExec)
    }
    wg.Wait()

    if firstErr != nil {
        exec.Status = WorkflowFailed
        exec.FailureReason = firstErr.Error()
        return exec, firstErr
    }

    exec.Status = WorkflowCompleted
    exec.CompletedAt = time.Now()
    return exec, nil
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
gc agent restart <name> [--graceful]
gc agent logs <name> [--follow]

gc task list [--status=X]
gc task create <title>
gc task assign <task-id> <agent>
gc task show <task-id>
gc task retry <task-id>            # Re-queue a failed task

gc mail send <to> -s "Subject" -m "Body"
gc mail inbox [--all]
gc mail read <id>

gc workflow run <template>
gc workflow plan <template>        # Dry-run: print execution graph without running
gc workflow status
gc workflow list

gc activity [--follow]             # Real-time event feed
gc stats                           # Agent metrics

gc validate                        # Config validation with diagnostics
gc migrate [--dry-run]             # Generate gas-city.toml from Gas Town workspace
gc doctor                          # Health checks (extended from gt doctor)

gc test-adapter <name> [--config=c.toml]  # Run conformance suite against an adapter
                                          # Returns 0 on success, prints failure report

gc config show                     # Display resolved config
gc version
```

### 11.2 `gc init` Wizard

```
$ gc init

Welcome to Gas City SDK!

What kind of orchestration do you need?

  > Ralph         -- Single agent with task loop (simplest)
    Agent Teams   -- Coordinator + worker pool
    Gas Town      -- Full multi-project orchestration
    Custom        -- Build your own topology

Select [Ralph]:

Which coding agent do you use?

  > Claude Code   -- claude --dangerously-skip-permissions
    Codex         -- codex (OpenAI)
    Gemini CLI    -- gemini
    Other         -- Custom command

Select [Claude Code]:

Created gas-city.toml (Level 2 - Task Loop)
Run `gc start` to begin.
```

### 11.3 `gc level` Output

```
$ gc level

Workspace "my-project" is at Level 3 (Worker Pool)

  V Level 0: Agent Runtime
  V Level 1: Work Tracking
  V Level 2: Task Loop
  V Level 3: Worker Pool       <- you are here
  o Level 4: Messaging          -- add [messaging] section
  o Level 5: Workflows          -- add [workflows] section
  o Level 6: Health Monitoring  -- add a supervisor agent
  o Level 7: Multi-Project      -- add 2+ projects with project-scoped agents
```

---

## 12. Migration from Gas Town

### 12.1 Migration Algorithm

```go
func Migrate(townRoot string) (*MigrationReport, error) {
    report := &MigrationReport{
        SourceTown: townRoot,
    }
    cfg := &WorkspaceConfig{}

    // 0. Check for active sessions
    activeSessions := detectActiveSessions(townRoot)
    if len(activeSessions) > 0 {
        report.Warnings = append(report.Warnings,
            fmt.Sprintf("Found %d active tmux sessions. Run 'gt stop --all' first.",
                len(activeSessions)))
    }

    // 1. Read town identity
    town := readJSON(townRoot + "/mayor/town.json")
    cfg.Workspace.Name = town.Name

    // 2. Read rig registry -> projects
    rigs := readJSON(townRoot + "/mayor/rigs.json")
    for _, rig := range rigs {
        cfg.Projects[rig.Name] = ProjectConfig{Repo: rig.URL}
        report.Projects = append(report.Projects, ProjectMigration{
            GasTownRig: rig.Name, Status: "migrated",
        })
    }

    // 3. Read agent registries -> agents
    townAgents := readJSON(townRoot + "/settings/agents.json")
    for _, agent := range townAgents {
        converted, notes := convertAgent(agent)
        cfg.Agents = append(cfg.Agents, converted)
        report.Agents = append(report.Agents, AgentMigration{
            GasTownRole: agent.Role,
            GasCityRole: converted.Role,
            Status:      "migrated",
            Notes:       notes,
        })
    }

    // 4. Map hardcoded roles -> agent entries
    cfg.Agents = append(cfg.Agents, AgentEntry{Name: "coordinator", Role: "coordinator", Scope: "workspace"})
    cfg.Agents = append(cfg.Agents, AgentEntry{Name: "supervisor", Role: "supervisor", Scope: "workspace"})
    for _, rig := range rigs {
        cfg.Agents = append(cfg.Agents, AgentEntry{Name: "observer-" + rig.Name, Role: "observer", Scope: "project"})
        cfg.Agents = append(cfg.Agents, AgentEntry{Name: "integrator-" + rig.Name, Role: "integrator", Scope: "project"})
    }

    // 5. Read formulas -> workflows (with validation)
    formulas := readTOML(townRoot + "/.beads/formulas/*.toml")
    cfg.Workflows.TemplatesDir = "workflows/"
    for _, formula := range formulas {
        converted := convertFormula(formula) // {{rig}} -> {{project}}
        if err := validateWorkflowTemplate(converted); err != nil {
            report.Errors = append(report.Errors,
                fmt.Sprintf("Formula %s: invalid conversion - %s", formula.ID, err))
        }
        report.Formulas = append(report.Formulas, FormulaMigration{
            Source: formula.ID, Status: "migrated",
        })
    }

    // 6. Flag hooks for manual review
    hookFiles := scanHookFiles(townRoot)
    for _, hf := range hookFiles {
        report.Hooks = append(report.Hooks, HookMigration{
            Source: hf,
            Status: "requires_review",
            Notes:  "Claude hooks should be reviewed and migrated to [agents.hooks]",
        })
    }

    // 7. Set task and messaging backends
    cfg.Tasks = &TaskConfig{Backend: "beads"}
    cfg.Messaging = &MessagingConfig{Backend: "beads"}

    report.TargetConfig = cfg
    return report, nil
}
```

### 12.2 Migration Report

The `gc migrate --dry-run` command outputs a structured report:

```go
type MigrationReport struct {
    SourceTown   string
    TargetConfig *WorkspaceConfig
    Agents       []AgentMigration
    Projects     []ProjectMigration
    Formulas     []FormulaMigration
    Hooks        []HookMigration
    Warnings     []string  // Non-blocking issues
    Errors       []string  // Blocking issues requiring manual fix
}

func (r *MigrationReport) Print(w io.Writer) {
    // 1. Lists mapped agents (Gas Town Role -> Gas City Agent)
    // 2. Lists config files sourced (town.json, rigs.json, etc.)
    // 3. Flags "manual intervention needed" items (hooks, env vars)
    // 4. Shows the resulting gas-city.toml content
    // 5. Lists warnings and errors separately
}
```

### 12.3 Compatibility Guarantees

1. A Gas Town workspace without `gas-city.toml` continues to work with `gt` commands unchanged.
2. `gc migrate --dry-run` generates the equivalent config without modifying anything.
3. After migration, both `gt` and `gc` commands work (Gas Town compatibility mode persists).
4. Migration is additive: it creates `gas-city.toml` but doesn't modify existing Gas Town configs.

### 12.4 File-by-File Migration Map

| Gas Town File | Migration Target | Notes |
|--------------|-----------------|-------|
| `mayor/town.json` | `[workspace] name` | Direct mapping |
| `mayor/rigs.json` | `[projects.*]` | One entry per rig |
| `settings/agents.json` | `[[agents]]` entries | Agent configs with runtime_config |
| `settings/config.json` | Various sections | Town settings spread across config sections |
| `internal/config/roles/*.toml` | Built-in role behaviors | Not migrated -- hardcoded in SDK |
| `.beads/formulas/*.toml` | `workflows/*.toml` | Copy with variable renaming ({{rig}} -> {{project}}) |
| `~/.gt/hooks-base.json` | Not migrated | Claude Code hooks remain managed by `gt hooks` |
| `~/.gt/hooks-overrides/*.json` | `[agents.hooks]` | Flagged for manual migration |

---

## 13. Identity and Addressing

### 13.1 Agent Address Format

Gas City uses a hierarchical addressing scheme with two layers:

**Logical identity** (stable across restarts):
```
<workspace>/<agent-name>
<workspace>/<project>/<agent-name>
<workspace>/<project>/<agent-name>[<instance>]
```

**Runtime handle** (transient, changes on restart):
```
<handle-uuid>
```

Examples:
```
my-town/coordinator          # Workspace-scoped agent
my-town/gastown/observer     # Project-scoped agent
my-town/gastown/workers[2]   # Specific pool instance
```

### 13.2 Identity Resolution

```go
// AgentRegistry tracks logical identities and their current runtime handles.
type AgentRegistry struct {
    mu         sync.RWMutex
    identities map[string]*AgentHandle  // logical address -> current handle
    handles    map[string]string        // handle ID -> logical address
    stateDir   string                   // .gc/agents/ for persistence
}

func (r *AgentRegistry) Register(handle *AgentHandle) {
    r.mu.Lock()
    defer r.mu.Unlock()
    addr := handle.Identity.String()
    r.identities[addr] = handle
    r.handles[handle.ID] = addr
    r.persist(handle)  // Write to .gc/agents/<addr>.json for crash recovery
}

func (r *AgentRegistry) Resolve(address string) (*AgentHandle, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    parts := strings.Split(address, "/")
    switch len(parts) {
    case 2:
        // workspace/agent -- workspace-scoped agent
        return r.identities[address], nil
    case 3:
        // workspace/project/agent -- project-scoped agent
        // Check for instance syntax: "workers[2]"
        name, instance := parseInstance(parts[2])
        if instance >= 0 {
            return r.identities[fmt.Sprintf("%s/%s/%s[%d]", parts[0], parts[1], name, instance)], nil
        }
        return r.identities[address], nil
    default:
        return nil, fmt.Errorf("invalid address: %s", address)
    }
}
```

### 13.3 Backward Compatibility with Gas Town Addresses

Gas Town uses `gastown/polecats/Toast` format. The migration layer maps:
- `<rig>/polecats/<name>` -> `<workspace>/<project>/workers[<instance>]`
- `<rig>/crew/<name>` -> `<workspace>/<project>/<name>`
- `mayor` -> `<workspace>/coordinator`
- `deacon` -> `<workspace>/supervisor`

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
    events := h.bus.Subscribe(h)  // HookExecutor is a critical subscriber
    // Events delivered via OnEvent callback
}

func (h *HookExecutor) OnEvent(event Event) error {
    configs, ok := h.hooks[event.Type]
    if !ok {
        return nil
    }
    for _, cfg := range configs {
        if cfg.Critical {
            // Critical hooks run synchronously — failure propagates through event bus
            if err := h.executeHookSync(cfg, event); err != nil {
                return fmt.Errorf("critical hook %s failed: %w", cfg.Command, err)
            }
        } else {
            // Best-effort hooks run asynchronously
            go h.executeHook(cfg, event)
        }
    }
    return nil
}

func (h *HookExecutor) IsCritical() bool { return true }

func (h *HookExecutor) executeHookSync(cfg HookConfig, event Event) error {
    cmd := expandTemplate(cfg.Command, event)
    ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
    defer cancel()
    return exec.CommandContext(ctx, "sh", "-c", cmd).Run()
}

func (h *HookExecutor) executeHook(cfg HookConfig, event Event) {
    if err := h.executeHookSync(cfg, event); err != nil {
        log.Warn("hook failed", "event", event.Type, "error", err)
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
| Adapter registry | Registration, lookup, unknown adapter error, thread safety | None (pure logic) |
| Pool manager | Scale up/down, min/max bounds, idle timeout, scale-down races | RuntimeAdapter mock |
| Task loop | Cycle execution, clear context, stall handling, claim races | RuntimeAdapter + TaskBackend mocks |
| Startup sequencer | DAG resolution, parallel start, failure rollback, cycle detection | RuntimeAdapter mock |
| Shutdown sequencer | Reverse ordering, graceful timeout escalation, shuttingDown flag | RuntimeAdapter mock |
| Event bus | Publish/subscribe, critical vs optional tiers, replay, unsubscribe | None (pure concurrency) |
| Hook executor | Template expansion, timeout, error handling | EventBus (inject events) |
| Identity resolver | Address parsing, backward compatibility, persistence | None (pure logic) |
| Migration | Town->City config conversion, edge cases, report generation | Filesystem (test fixtures) |
| Workflow executor | Sequential/parallel execution, timeout, retry, failure strategies | RuntimeAdapter mock |
| Task backend | Claim atomicity, optimistic locking, status transitions | None (filesystem backend with temp dirs) |

### 15.2 Contract Tests

The `gc test-adapter` command runs this suite against any registered adapter. This allows custom adapter authors to certify compliance with the Gas City spec.

```go
func RunAdapterContractTests(t *testing.T, adapter RuntimeAdapter, cfg AgentConfig) {
    // === Happy path tests ===

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

    t.Run("Double Stop is idempotent (P1)", func(t *testing.T) {
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

    t.Run("Ping within timeout (P2)", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        waitForRunning(t, adapter, handle)
        result, err := adapter.Ping(handle)
        require.NoError(t, err)
        require.True(t, result.OK)
        require.Less(t, result.Latency, cfg.Health.PingTimeout)
    })

    t.Run("Attach matches SupportsAttach (P3)", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        if adapter.SupportsAttach() {
            require.NoError(t, adapter.Attach(handle))
            adapter.Detach(handle)
        } else {
            require.ErrorIs(t, adapter.Attach(handle), ErrNotSupported)
        }
    })

    t.Run("Concurrent access is safe (P5)", func(t *testing.T) {
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

    t.Run("Start creates unique handles (P6)", func(t *testing.T) {
        h1, _ := adapter.Start(ctx, cfg)
        h2, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(h1, true)
        defer adapter.Stop(h2, true)
        require.NotEqual(t, h1.ID, h2.ID)
    })

    // === Error path tests ===

    t.Run("Nudge on stopped agent returns error", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        adapter.Stop(handle, true)
        err := adapter.Nudge(handle, "should-fail")
        require.Error(t, err)
    })

    t.Run("SendInput on stopped agent returns error", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        adapter.Stop(handle, true)
        err := adapter.SendInput(handle, "should-fail")
        require.Error(t, err)
    })

    t.Run("GetState on stopped agent returns Stopped", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        adapter.Stop(handle, true)
        state, err := adapter.GetState(handle)
        require.NoError(t, err)
        require.Equal(t, StatusStopped, state.Status)
    })

    t.Run("WaitForResult returns after assign", func(t *testing.T) {
        handle, _ := adapter.Start(ctx, cfg)
        defer adapter.Stop(handle, true)
        waitForRunning(t, adapter, handle)

        task := TaskDescriptor{ID: "test-1", Title: "Test task", Description: "echo hello"}
        require.NoError(t, adapter.Assign(handle, task))

        ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()
        result, err := adapter.WaitForResult(ctx, handle)
        require.NoError(t, err)
        require.NotEmpty(t, result.TaskID)
    })
}
```

### 15.3 Integration Tests

| Test | What It Validates | Prerequisites |
|------|-------------------|---------------|
| Claude-code adapter roundtrip | Start -> Assign -> WaitForResult -> Stop | tmux, claude CLI |
| Subprocess adapter roundtrip | Start -> SendInput -> ReadOutput -> Stop | None |
| Config -> startup -> shutdown | Parse TOML -> start all -> verify -> stop all -> verify clean | claude CLI |
| Ralph smoke test | Single agent processes one task end-to-end via Claim() | claude CLI, task backend |
| Agent Teams smoke test | Coordinator dispatches task to worker pool | claude CLI, task backend |
| Migration test | Existing Gas Town workspace -> gc migrate -> verify config + report | Gas Town workspace fixture |
| Level progression | Level 0 -> 7 configs all start correctly | claude CLI |
| Startup rollback | Fail agent 3 of 5 -> verify agents 1-2 stopped | RuntimeAdapter mock |
| Claim concurrency | 5 agents race to Claim() same task -> only 1 wins | Filesystem backend |
| Event bus tiers | Critical subscriber blocks, optional doesn't | None |
| Workflow failure | Step 2 fails -> verify strategy (fail/skip/retry) | RuntimeAdapter mock |

### 15.4 Acceptance Criteria per Implementation Phase

| Phase | Acceptance Criteria |
|-------|-------------------|
| 1: Runtime Abstraction | All existing Gas Town tests pass with adapter wrapper. `gc agent start/stop/status` work. Contract tests pass for `claude-code` adapter. `WaitForResult()` returns `TaskResult`. |
| 2: Config Parser | `gc validate` accepts all 8 levels. `gc level` reports correct level. `gc init` generates valid configs for all 3 shapes. `depends_on` cycle detection works. Runtime auto-detection works. |
| 3: Task System | `gc task create/list/assign` work with filesystem backend. Task loop runs for Ralph shape. `Claim()` is atomic. Optimistic locking prevents stale updates. |
| 4: Pool Manager | Agent Teams shape starts. Pool scales between min and max. Scale-down respects mutex. Readiness probes used when available. |
| 5: Messaging | `gc mail send/inbox` work. Direct messages and channels deliver. TTL expiration works. Ordering guarantees hold. |
| 6: Workflows | `gc workflow run` executes sequential and parallel templates. Timeout, retry, skip strategies work. Failure produces `WorkflowExecution` report. |
| 7: Health Monitor | Supervisor detects stalled agents (with progress-aware detection). Automatic restart with cooldown. Shutdown flag prevents restart during shutdown. |
| 8: Full Gas Town | Level 7 config starts all agents with DAG ordering. Migration generates valid config + report. Equivalent behavior to current Gas Town. |

---

## 16. Implementation Phases

### Phase 1: Agent Runtime Extraction (2-3 weeks)

**Goal:** Extract the `claude-code` adapter from Gas Town's existing tmux code without changing behavior.

**Deliverables:**
- `internal/runtime/adapter.go` -- RuntimeAdapter interface + ReadinessProbe
- `internal/runtime/registry.go` -- Thread-safe adapter registry
- `internal/runtime/types.go` -- AgentHandle, AgentIdentity, AgentConfig, TaskResult
- `internal/runtime/claude_code/` -- Claude Code adapter wrapping `internal/tmux/`
- `internal/runtime/subprocess/` -- Generic subprocess adapter
- Contract test suite in `internal/runtime/contract_test.go`
- `gc agent start/stop/status` commands

**Acceptance:** All existing Gas Town tests pass. Contract tests (happy + error paths) pass for both adapters.

### Phase 2: Config Parser + CLI Foundation (1-2 weeks)

**Goal:** Parse `gas-city.toml`, detect levels, validate, and provide `gc init` + `gc level` + `gc validate`.

**Deliverables:**
- `internal/config/gas_city.go` -- TOML parser with level detection (monotonic guards)
- `internal/config/validate.go` -- Validation rules including dependency cycle detection
- `internal/config/autodetect.go` -- Runtime auto-detection
- `gc init`, `gc level`, `gc validate`, `gc config show` commands
- Test fixtures for all 8 levels

**Acceptance:** All validation rules enforced. Level detection correct with monotonic progression. Init wizard generates valid configs. Dependency cycles detected.

### Phase 3: Task System + Ralph Shape (2 weeks)

**Goal:** Task backend interface + filesystem backend + task loop = working Ralph.

**Deliverables:**
- `internal/tasks/backend.go` -- TaskBackend interface with Claim()
- `internal/tasks/filesystem/` -- Filesystem backend with flock-based Claim()
- `internal/tasks/beads/` -- Beads backend adapter
- `internal/loop/` -- Task loop controller using atomic Claim
- `gc task create/list/assign/show/retry` commands

**Acceptance:** Ralph shape processes tasks end-to-end. Claim() is atomic (concurrent test passes). Optimistic locking prevents stale updates.

### Phase 4: Pool Manager + Agent Teams Shape (2 weeks)

**Goal:** Pool management + coordinator dispatch = working Agent Teams.

**Deliverables:**
- `internal/pool/` -- Pool manager with synchronized scale-down
- Startup/shutdown sequencer with DAG ordering and rollback in `internal/orchestration/`
- Agent registry with persistence in `internal/registry/`
- Coordinator role behavior

**Acceptance:** Pool scales correctly with mutex protection. Startup rollback on failure works. Agent Teams shape dispatches tasks to workers.

### Phase 5: Messaging + Workflows (1-2 weeks)

**Goal:** Inter-agent messaging and workflow template execution.

**Deliverables:**
- `internal/messaging/backend.go` -- MessageBackend interface with delivery guarantees
- `internal/messaging/filesystem/` -- Filesystem backend
- `internal/workflows/` -- Workflow template parser and executor with strategy support
- `gc mail` and `gc workflow` commands

**Acceptance:** Direct messages and channels work with ordering. Sequential and parallel workflows execute with timeout/retry/skip strategies.

### Phase 6: Health Monitoring + Event Bus (1-2 weeks)

**Goal:** Supervisor patrol cycle + tiered event bus + lifecycle hooks.

**Deliverables:**
- `internal/events/bus.go` -- Tiered event bus (critical + optional subscribers)
- `internal/health/supervisor.go` -- Progress-aware supervisor patrol cycle
- `internal/hooks/executor.go` -- Lifecycle hook executor (critical subscriber)
- `gc activity` and `gc stats` commands

**Acceptance:** Stalled agents detected (progress-aware). Events never dropped for critical subscribers. Hooks fire. Shutdown flag prevents restart during shutdown.

### Phase 7: Migration + Gas Town Shape (2 weeks)

**Goal:** Full Gas Town shape + migration tooling with detailed reporting.

**Deliverables:**
- `internal/migration/` -- Gas Town -> Gas City migration with MigrationReport
- `gc migrate` command with `--dry-run` and structured report output
- Level 7 config with all role behaviors and DAG dependencies
- Observer, integrator, service role behaviors

**Acceptance:** `gc migrate --dry-run` generates correct config + report from Gas Town workspace. Report surfaces hooks/env-vars needing manual review. Gas Town shape starts all agents with DAG ordering.

### Phase 8: Additional Adapters + Conformance CLI (ongoing)

**Goal:** Codex, Gemini, Agent SDK, Docker adapters + `gc test-adapter`.

**Deliverables:**
- Per-adapter package in `internal/runtime/`
- Contract tests passing for each (happy + error paths)
- `gc test-adapter <name>` CLI command for ecosystem certification

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
forall n in [0,7]: C(n) is a subset of C(n+1)
```

**Proof:** Each level adds a new optional config section ([tasks], [messaging], etc.) without modifying the schema of lower-level sections. The config parser accepts any superset of a valid config. Formally:
- *Additive-only schema evolution:* Let `S(n)` be the set of valid config fields at level `n`. Level `n+1` adds fields but never removes or redefines fields from level `n`. Thus `S(n) ⊂ S(n+1)`.
- *Independent usefulness:* A config at level `n` is valid and functional without any level `n+1` features. The level detection function returns `n` (not `n+1`) when level `n+1` sections are absent. Every level enables at least one new CLI command or capability.
- *Backward compatibility:* A tool built for level `n` works correctly on any config at level `m >= n` — it ignores unrecognized sections via the TOML parser's permissive behavior.

### 18.2 Startup Ordering (DAG)

For agents `a` and `b` where `a` is in `b.depends_on`:

```
IsRunning(a) = true -> Start(b) is called
```

This is enforced by the topological group algorithm: agents in the same group have no inter-dependencies and start in parallel; groups execute sequentially.

**Cycle prevention:** The config validator rejects any `depends_on` graph containing cycles via DFS-based cycle detection at parse time.

### 18.3 Pool Bounds

For a pool with config `{min: M, max: N}`, at any time `t`:

```
M <= |running_instances(t)| <= N
```

**Invariant enforcement:** The PoolManager holds `mu` during all scale operations, serializing scale-up/down with task dispatch. It checks bounds before every scale operation:
- `scaleUp`: `if running < N then spawn(min(needed, N - running))`
- `scaleDown`: `if idle > M then kill(idle - M)` (with double-check after lock)

### 18.4 Event Bus Liveness

Every event published to the bus is delivered to every critical subscriber synchronously (blocking). Optional subscribers receive events asynchronously. Critical events (agent crashes, task failures) are guaranteed not to be dropped by using the critical tier.

### 18.5 Task Claim Atomicity

For any task `t` with `Status = Ready`, at most one agent can successfully call `Claim(t.id, agent)` and receive `true`. All other concurrent claims receive `false`. This is enforced by backend-specific atomic operations (SQL transactions for beads, flock for filesystem, label-then-verify for GitHub).

### 18.6 Startup Rollback Safety

If `StartWorkspace()` fails at agent `k` of `n`, agents `1..k-1` are stopped in reverse order. The workspace returns to a clean state with no running agents. Formally: if `Start(agent_k)` fails, then `forall i < k: IsRunning(agent_i) = false` after rollback completes.

---

## 19. Open Questions

1. **Out-of-process adapter protocol.** Go plugins are fragile. JSON-RPC over stdin/stdout (LSP-style) is more robust for custom adapters. Decision: implement JSON-RPC in Phase 8.

2. **Config hot-reload.** Can you add agents to a running workspace? v1: no (requires restart). v2: yes (via file watch + diff).

3. **Multi-runtime task serialization.** When results cross runtimes (Claude Code coordinator -> Codex worker), the `TaskResult` schema (Section 2.3) provides the bridge. Adapters normalize their output into this structure. Validate in Phase 4.

4. **State persistence across restarts.** Claude-code uses tmux session survival + `/resume`. Other runtimes need checkpoint files. The `AgentRegistry` persists logical identities to `.gc/agents/` for crash recovery. Define `StateCheckpoint` interface in Phase 3.

5. **`gc` vs `gt` CLI namespace.** Options: (a) `gc` is a separate binary, (b) `gt gc` subcommand, (c) `gt` detects Gas City mode and adjusts behavior. Recommendation: (a) separate binary for clarity.

---

## 20. Known Gaps (Vision Alignment)

The following gaps were identified by comparing this spec against `specs/gas-city-vision.md`. All 8 vision requirements are fully addressed in the spec's architecture, but these areas have weaker coverage — deliberate v1 scope limits that should be addressed in future phases.

### 20.1 Out-of-Process Plugin Protocol

**Vision requirement:** Plugins (extensibility + integration points)

**Gap:** v1 supports only in-process Go adapters and subprocess wrappers. Third-party plugin authors must either write Go code and link it into the binary, or wrap their tool as a CLI and use the `subprocess` adapter. There is no language-agnostic plugin protocol.

**Current mitigation:** The `RegisterAdapter()` API (Section 2.8, line 479) allows Go-level extension. The `subprocess` adapter (Section 2.7) wraps any CLI tool. These cover most use cases but are not equivalent to a true plugin system.

**Resolution plan:** JSON-RPC over stdin/stdout (LSP-style) protocol for out-of-process adapters, deferred to Phase 8 (see Open Question 1). This will enable adapters written in Python, TypeScript, Rust, etc.

### 20.2 No Pre-Event Hooks

**Vision requirement:** Hooks (lifecycle events)

**Gap:** All hooks defined in the TOML schema (Section 3.2, lines 596-602) are post-event: `on_start`, `on_stop`, `on_task_assign`, `on_task_complete`, `on_task_fail`, `on_stall`. There is no mechanism for pre-event hooks that can intercept and veto actions before they happen (e.g., "should I accept this task?", "should I start this agent?").

**Current mitigation:** Pre-event decision logic can be expressed in agent prompts (system prompt instructions) and coordination rules (`depends_on`, workflow templates). Hooks are for side effects (notifications, scripts), not policy decisions.

**Resolution plan:** Consider adding `before_task_assign` and `before_start` hooks in a future version. These would return a boolean (proceed/abort) and enable policy-based gating without modifying SDK source. However, this adds complexity to the event dispatch path and must be designed carefully to avoid blocking the control plane.

### 20.3 Thin Sandbox Implementation Detail

**Vision requirement:** Sandboxes (isolation modes)

**Gap:** The spec defines four isolation strategies (Section 3.2, lines 554-558): `none`, `worktree`, `directory`, `container`. However, only `worktree` has significant implementation detail via the existing Gas Town tmux adapter. The `directory` and `container` modes lack pseudocode, setup/teardown algorithms, and edge case documentation.

**Current mitigation:** `worktree` isolation is well-understood from Gas Town. `container` isolation maps to the Docker adapter (Section 2.4, line 310). `directory` is a straightforward temp-dir copy pattern. All three are standard patterns that implementers can follow.

**Resolution plan:** Add implementation sketches for `directory` (temp dir creation, file sync, cleanup) and `container` (Dockerfile generation, volume mounting, network isolation) during Phase 2 (Agent Teams) when isolation becomes critical for pooled workers.

### 20.4 No Formal Coordination Grammar

**Vision requirement:** Custom coordination rules

**Gap:** Complex multi-agent coordination protocols rely on a combination of TOML config fields (`depends_on`, `hooks`, `loop`), natural language prompts, and workflow templates. There is no formal grammar or DSL for expressing coordination rules like "if agent A produces output X, route to agent B; otherwise escalate to coordinator." Coordination logic beyond startup ordering and event hooks must be embedded in agent prompts.

**Current mitigation:** Workflow templates (Section 10) provide structured multi-step processes with dependency chains, retry strategies, and parallel execution. The messaging system (Section 9) enables agents to coordinate via direct messages and channels. The combination of workflows + messaging + prompts covers most coordination patterns.

**Resolution plan:** Evaluate whether a lightweight coordination DSL (e.g., state machine definitions in TOML) is needed based on Phase 3+ user feedback. The current prompt-based approach is more flexible but less verifiable. A DSL would enable static analysis of coordination correctness but risks over-engineering for v1.

### 20.5 No Role Definition Schema Validation

**Vision requirement:** Roles expressed externally

**Gap:** Role definitions are externalized as markdown files with Go template variables (Section 4.4, lines 1046-1060), but there is no schema or validation for these files. An invalid role prompt (e.g., referencing `{{.NonexistentVar}}`, missing required sections, or containing contradictory instructions) would only surface as runtime agent behavior failures, not as config-time errors.

**Current mitigation:** `gc validate` (Section 3.4) validates the TOML config structure but does not validate prompt content. Template variable expansion is checked at render time, which catches undefined variables but not semantic errors.

**Resolution plan:** Add optional role prompt linting to `gc validate` in a future version. This could check for: (a) undefined template variables, (b) required sections (e.g., "Responsibilities", "Coordination rules"), (c) references to configured channels/agents that exist in the workspace. This is a quality-of-life improvement, not a correctness requirement.

### 20.6 Beads Task Backend Not Specified

**Vision requirement:** Work tracking / task system

**Gap:** The beads task backend is referenced throughout the spec as a first-class option (Section 3.2 line 610, Section 8.2 line 1585, Section 4.2 line 782) but its implementation is not specified in this document. The beads system is an existing Gas Town artifact (Dolt-based structured data) and is assumed to be available.

**Current mitigation:** The `TaskBackend` interface (Section 8.1) fully defines the contract that any backend must satisfy. The filesystem backend (Section 8.2, lines 1592-1635) provides a zero-dependency reference implementation. The beads backend wraps an existing, working system.

**Resolution plan:** The beads backend implementation is a thin adapter layer over Gas Town's existing Dolt/beads infrastructure. It does not require new specification — it requires mapping the `TaskBackend` interface to existing beads SQL queries. This mapping is straightforward and will be documented during Phase 1 implementation as inline code comments.
