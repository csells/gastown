# Gas City SDK — Ranked Ideas

> **Pipeline:** Spec-Forge Stage 2 (Ideation)
> **Date:** 2026-02-14
> **Input:** research-notes.md (656 lines), gas-city-spec-initial-draft.md (1150 lines)

---

## Top 5 Ideas

### 1. Config Scaffolding with Progressive Discovery (`gc init` + `gc level`)

**The idea:** An interactive `gc init` wizard that generates the right `gas-city.toml` based on user intent, combined with a `gc level` command that shows your current capability level and what the next level would unlock. Together they make the progressive capability model tangible and discoverable.

**How it works:**

`gc init` asks: "What do you want to build?" and offers the three milestone shapes plus "Custom":
- **Ralph** → generates a Level 2 config (single agent + task loop)
- **Agent Teams** → generates a Level 3 config (coordinator + worker pool)
- **Gas Town** → generates a Level 7 config (full orchestration)
- **Custom** → walks through each level asking what to enable

Then `gc level` reads the existing config and reports:
```
Your workspace "my-project" is at Level 3 (Worker Pool).

Enabled: Agent Runtime, Work Tracking, Task Loop, Worker Agents
Next level: Inter-Agent Messaging
  Add [messaging] section to gas-city.toml to enable.
```

**Why this is #1:** The biggest barrier to any SDK is "how do I start?" The draft spec presents 8 capability levels and 3 milestone shapes — that's a lot of surface area for a new user to absorb. Without scaffolding, users face a blank TOML file and a long spec document. With it, they get a working config in 30 seconds and a clear growth path. The progressive model is Gas City's differentiator over other orchestration tools; if users can't discover it naturally, it fails.

**How users will perceive it:** Immediately productive. The init wizard eliminates the "stare at an empty config file" problem. The level command gives a gamification-like progression that encourages exploration. Users who start with Ralph and grow to Agent Teams will feel guided, not overwhelmed.

**Implementation:** Moderate complexity. The wizard is a Bubbletea TUI (Gas Town already uses Bubbletea extensively). `gc level` is a config parser that counts which sections are present and maps to the level hierarchy. Both are pure logic — no runtime dependencies.

**Rationale:** Every successful SDK (Docker, Kubernetes, Terraform) has strong scaffolding. The progressive model is only as good as users' ability to navigate it. This idea makes the model work in practice, not just on paper.

---

### 2. Runtime Adapter Contract Test Harness

**The idea:** A standardized `gc test-adapter <name>` command and Go test suite (`TestAdapterContract(t, adapter)`) that validates any runtime adapter — built-in or custom — against the full RuntimeAdapter interface contract.

**How it works:**

The test suite verifies 12 properties:
1. `Start()` returns a valid handle without error
2. `IsRunning()` returns true within `ping_timeout` after Start
3. `Stop(graceful=true)` makes `IsRunning()` return false
4. `Stop()` on an already-stopped agent is idempotent (no error)
5. `Nudge()` delivers a message (verified via `CaptureOutput()`)
6. `SendInput()` delivers text (verified via `ReadOutput()`)
7. `Ping()` returns within `ping_timeout`
8. `SupportsAttach()` matches `Attach()` behavior (true → succeeds, false → returns ErrNotSupported)
9. `CaptureOutput()` returns a readable stream after Start
10. `GetState()` returns a valid state object
11. `Stop(graceful=false)` releases all OS resources within `kill_cooldown`
12. Concurrent calls to `Nudge()` and `CaptureOutput()` are thread-safe

The CLI version (`gc test-adapter claude-code`) runs a real instance and reports pass/fail for each property. The Go test version is importable for custom adapter development.

**Why this is #2:** The Agent Runtime abstraction is Gas City's foundational design principle. If it doesn't hold — if adapters have subtly different behaviors — everything built on top breaks in unpredictable ways. Contract tests are the formal mechanism that ensures the abstraction isn't just conceptual but enforced. Without them, you end up with "works on Claude Code, broken on Codex" bugs that erode trust.

**How users will perceive it:** Custom adapter authors get a "certification" path. Built-in adapter quality is visibly high. The contract tests also serve as executable documentation of the runtime interface semantics.

**Implementation:** Medium complexity. The test fixture needs a configurable "echo agent" for each runtime (a minimal agent that echoes input to output). For the `claude-code` adapter, the existing tmux integration tests provide a foundation. The test runner itself is straightforward Go test infrastructure.

**Rationale:** The draft spec already defines 4 adapter correctness properties (Section 13.3). This idea operationalizes those properties into runnable tests, which is the difference between a design principle and an engineering guarantee.

---

### 3. Unified Event Bus for All Agent Activity

**The idea:** A single internal event bus that captures every significant event in the system — agent lifecycle transitions, task state changes, health check results, messages sent, workflow step completions — with a consistent schema and pluggable consumers.

**How it works:**

Every subsystem emits events to the bus:
```
{timestamp, event_type, agent, project, payload}
```

Event types:
- `agent.started`, `agent.stopped`, `agent.stalled`, `agent.crashed`
- `task.created`, `task.assigned`, `task.completed`, `task.failed`
- `health.ping_ok`, `health.ping_timeout`, `health.agent_restarted`
- `message.sent`, `message.delivered`
- `workflow.step_started`, `workflow.step_completed`
- `pool.scaled_up`, `pool.scaled_down`

Consumers can subscribe to the bus:
- **CLI consumer**: Powers `gc activity --follow` (real-time feed)
- **Log consumer**: Structured JSON logs to file
- **Dashboard consumer**: Powers the web dashboard (future)
- **Hook consumer**: Triggers user-defined shell commands on specific events
- **Metrics consumer**: Aggregates event counts for `gc stats`

**Why this is #3:** Gas Town already has `bd activity --follow` for beads events and the deacon patrol for health monitoring, but these are separate, ad-hoc systems. A unified event bus is the observation equivalent of the Agent Runtime abstraction for execution. It provides a single source of truth for "what's happening" across all agents, all runtimes, all projects. Without it, building dashboards, alerting, and operational tooling requires stitching together multiple data sources.

**How users will perceive it:** Transparency. Users can observe everything the system does. `gc activity --follow` becomes the universal debugging tool. Custom consumers enable integration with existing monitoring infrastructure (Datadog, PagerDuty, Slack webhooks) without core SDK changes.

**Implementation:** Medium complexity. The event bus itself is a Go channel with fan-out to registered consumers. The schema is a flat struct with a discriminated union payload. Gas Town's existing `internal/feed/` package provides a starting point. The main work is instrumenting all subsystems to emit events.

**Rationale:** Observability is a cross-cutting concern that's expensive to retrofit. Building it as a first-class primitive from the start means every new feature automatically becomes observable. This is the difference between "working" and "operable."

---

### 4. Agent Lifecycle Hooks (User-Defined Shell Commands)

**The idea:** Extend Gas Town's existing Claude Code hooks pattern to all agent runtimes via `gas-city.toml`. Users define shell commands that trigger on agent lifecycle events (start, stop, assign, complete, fail, stall) without writing Go code or plugins.

**How it works:**

In `gas-city.toml`:
```toml
[hooks]
on_agent_start = "echo 'Agent {{.Agent}} started' >> /tmp/gas-city.log"
on_agent_stop = "echo 'Agent {{.Agent}} stopped' >> /tmp/gas-city.log"
on_task_complete = "./scripts/notify-slack.sh {{.Agent}} {{.Task}}"
on_agent_stall = "./scripts/page-oncall.sh {{.Agent}} {{.Project}}"
on_task_fail = "./scripts/create-incident.sh {{.Task}} {{.Error}}"

# Per-agent hooks override global ones
[[agents]]
name = "workers"
[agents.hooks]
on_task_complete = "./scripts/deploy-preview.sh {{.Task}}"
```

Template variables provide context: `.Agent`, `.Project`, `.Task`, `.State`, `.Error`, `.Timestamp`. Hooks run asynchronously by default with configurable timeout.

**Why this is #4:** Gas Town already proves that hooks are a powerful extensibility mechanism — the entire Claude Code integration works through SessionStart/Stop hooks. Generalizing this to all lifecycle events across all runtimes gives Gas City composability with existing infrastructure without requiring users to write Go plugins. "On deploy completion, update the Jira ticket." "On agent stall, page the team." These integrations are what make an SDK production-ready.

**How users will perceive it:** Powerful and familiar. Shell hooks are the most accessible extensibility pattern in software — every developer knows how to write a shell script. This lowers the barrier to customization from "write a Go plugin" to "write a bash script."

**Implementation:** Low-medium complexity. The event bus (idea #3) provides the trigger mechanism. The hook executor is a subprocess runner with template expansion, timeout management, and error handling. Gas Town's existing `internal/hooks/` package provides patterns for hook management.

**Rationale:** The single most requested pattern in orchestration tools is "when X happens, do Y." Lifecycle hooks are the simplest, most composable answer. They also serve as the foundation for more complex integrations without the SDK needing to know about Slack, Jira, PagerDuty, etc.

---

### 5. Migration Validator (`gc migrate --dry-run`)

**The idea:** A `gc migrate` command that reads an existing Gas Town workspace (parsing `mayor/town.json`, `settings/config.json`, rig configs, agent registries, hooks, and formulas) and generates the equivalent `gas-city.toml`, with `--dry-run` mode that shows exactly what would change without modifying anything.

**How it works:**

```bash
# Show what would be generated
gc migrate --dry-run

# Output:
# Migration Report for Gas Town workspace at ~/gt
#
# Detected configuration:
#   Town: "my-town"
#   Rigs: gastown, beads (2 projects)
#   Agents: claude (default), codex (rig-level override on gastown)
#   Formulas: 12 workflow, 3 expansion, 2 aspect
#
# Generated gas-city.toml would be at Level 7 (Full Orchestration):
#   [workspace] name = "my-town"
#   [projects.gastown] repo = "https://github.com/steveyegge/gastown"
#   [projects.beads] repo = "https://github.com/steveyegge/beads"
#   ... (full TOML preview)
#
# Incompatibilities:
#   - Custom hooks in ~/.gt/hooks-overrides/polecat-gastown.json
#     → Migrate manually to [agents.hooks] section
#   - 3 formulas use Gas Town-specific variables ({{rig}})
#     → Will be mapped to {{project}} automatically
#
# Run `gc migrate` (without --dry-run) to generate gas-city.toml

gc migrate  # Actually generates the file
```

The migrator reads all 15+ config files documented in research-notes.md Section 3.1, maps them to the Gas City config schema, and flags anything that can't be automatically converted.

**Why this is #5:** Gas Town has real users with real workspaces. If migration is painful, scary, or lossy, they won't adopt Gas City. A dry-run migration tool that shows exactly what would happen makes adoption risk-free. Users can see the generated config, compare it to their existing setup, and gain confidence before committing. This is the critical bridge between Gas Town and Gas City.

**How users will perceive it:** Safe and respectful. The dry-run mode says "we won't change anything until you're ready." The detailed report builds trust by showing the migration tool understands their existing configuration. Flagging incompatibilities upfront prevents unpleasant surprises.

**Implementation:** Medium-high complexity. Requires parsing all existing Gas Town config formats (JSON configs, TOML roles, agent registries, hooks). The mapping logic follows the file-by-file migration map in the draft spec (Section 8.2). The output formatter needs to generate valid TOML and a human-readable report.

**Rationale:** Adoption friction kills SDKs. The biggest risk for Gas City is that existing Gas Town users say "my setup works, why change?" The migration validator answers that question by showing the path is safe, reversible, and clearly beneficial.

---

## Next Best 10

### 6. Standardized Task Result Schema

**The idea:** Define a standard JSON schema for task results that all runtimes emit upon task completion. This enables cross-runtime coordination where a Claude Code coordinator can understand results from a Codex worker without adapter-specific parsing.

**Schema:**
```json
{
  "task_id": "string",
  "status": "completed | failed | partial",
  "agent": "string",
  "runtime": "string",
  "started_at": "ISO-8601",
  "completed_at": "ISO-8601",
  "outputs": {
    "files_changed": ["path"],
    "branch": "string",
    "commit": "sha",
    "pr_url": "string"
  },
  "metrics": {
    "tokens_used": 0,
    "duration_seconds": 0,
    "cost_usd": 0.0
  },
  "error": "string | null"
}
```

**Rationale:** The draft spec's Open Question #2 asks "how does task result serialization work across runtimes?" This answers it. Without a standard result schema, coordinators need runtime-specific parsing logic, which defeats the purpose of the Agent Runtime abstraction. With it, a coordinator can dispatch to any runtime and understand the result identically.

---

### 7. Structured Logging with Agent Attribution

**The idea:** All log entries are tagged with the originating agent, project, and runtime. `gc logs` provides unified, filterable log access across all agents.

```bash
gc logs                           # All agents, last hour
gc logs --agent=coder             # Specific agent
gc logs --project=gastown         # All agents on a project
gc logs --level=error             # Errors only
gc logs --since=30m --follow      # Streaming tail
```

**Rationale:** Debugging multi-agent systems is hard because output interleaves. Gas Town partially solves this with per-agent tmux panes, but that doesn't work for non-tmux runtimes. Structured logging with attribution is the universal debugging tool that works across all runtimes and scales to any number of agents.

---

### 8. Config Validation with Progressive Diagnostics

**The idea:** `gc validate` goes beyond "valid/invalid" to provide progressive diagnostics: what level your config achieves, which sections are incomplete, and specific suggestions for improvement.

```bash
gc validate

# Output:
# gas-city.toml is valid at Level 3 (Worker Pool)
#
# Suggestions:
#   [messaging] section not present → add to enable Level 4 (Inter-Agent Messaging)
#   agents.workers.pool.max = 5 → consider max = 3 for development (resource usage)
#   agents.lead.health not configured → add for production readiness
#
# Warnings:
#   agents.workers.runtime = "claude-code" but tmux not detected on system
#   → Consider runtime = "subprocess" or install tmux
```

**Rationale:** Config files are the primary user interface for Gas City. Rich validation feedback turns the config file into a guided experience rather than a trial-and-error process. This is especially important for the progressive model where users may not realize they're one section away from the next level.

---

### 9. Config Composition via Includes

**The idea:** Allow `gas-city.toml` to include other TOML files for modular configuration at scale.

```toml
[workspace]
name = "my-town"
includes = ["agents/coordinators.toml", "agents/workers.toml", "workflows/*.toml"]
```

**Rationale:** Gas Town already has 15+ config files. At scale (many projects, many agent types, many workflows), a single `gas-city.toml` becomes unwieldy. Includes let teams own their agent definitions while the workspace owner controls the overall shape. This is the standard pattern in Terraform (modules), Docker Compose (extends), and Kubernetes (Kustomize).

---

### 10. Agent Sandbox Isolation Levels

**The idea:** Configurable isolation per agent: `none`, `worktree`, `directory`, `container`. Each level provides stronger isolation with corresponding trade-offs.

```toml
[[agents]]
name = "workers"
isolation = "worktree"    # git worktree per worker (current Gas Town behavior)
# isolation = "directory"  # full directory copy
# isolation = "container"  # Docker container (strongest)
# isolation = "none"       # shared workspace (simplest)
```

**Rationale:** This directly addresses GitHub Issue #1382 (headless/sandboxed polecats). The current system only supports git worktrees. Container isolation enables CI-like environments, while `none` enables simple single-agent setups. Making isolation a configuration choice rather than an architectural assumption increases the range of deployment scenarios Gas City supports.

---

### 11. Runtime Adapter Auto-Detection

**The idea:** When an agent's `runtime` field is omitted, Gas City detects what's available on the system and picks the best match.

Detection order: claude CLI → tmux (for claude-code adapter) → codex CLI → gemini CLI → subprocess fallback.

```toml
[[agents]]
name = "coder"
# runtime omitted — auto-detected
```

Output: `gc start` reports which runtime was selected: "Agent 'coder' using runtime 'claude-code' (auto-detected)."

**Rationale:** Reduces friction for new users who don't know (or care) about runtime adapters. They just want to start an agent. Auto-detection makes the simplest config (`name + nothing else`) work out of the box. More experienced users override with explicit runtime selection.

---

### 12. Custom Health Check Extensions

**The idea:** Beyond the built-in `Ping()`, users define custom health checks as shell commands or Go functions.

```toml
[[agents]]
name = "coder"
[agents.health]
ping_timeout = "30s"
custom_checks = [
  { name = "branch-clean", command = "git -C {{.WorkDir}} status --porcelain | wc -l", expect = "0" },
  { name = "no-stale-locks", command = "find {{.WorkDir}} -name '*.lock' -mmin +30 | wc -l", expect = "0" }
]
```

**Rationale:** Generic ping only tells you "agent is alive." Production orchestration needs domain-specific health: "is the agent's workspace clean?", "are there stale locks?", "is the agent making progress?" Custom health checks bridge generic liveness with application-specific readiness without modifying core code.

---

### 13. Dry-Run Mode for Workflows

**The idea:** `gc workflow run --dry-run <template>` shows the execution plan without running anything: which steps would execute, in what order, which agents would be assigned, and estimated duration.

**Rationale:** Complex workflow templates with dependencies can have non-obvious execution paths. Dry-run lets users validate their workflow design before committing agent resources. It's the equivalent of `terraform plan` — see what would happen before it happens. This is especially valuable for workflow templates with parallel aspects or conditional steps.

---

### 14. Agent Output Streaming Protocol

**The idea:** A standardized streaming protocol for agent output that works across all runtimes, enabling unified observation tools.

The protocol defines:
- A stream format (newline-delimited JSON with timestamp, agent, stream type [stdout/stderr/event])
- A multiplexer that combines streams from multiple agents
- A `gc watch` command that displays multiplexed output with agent attribution

**Rationale:** Currently, observing agent output requires `tmux attach` (tmux-only). For `agent-sdk` or `subprocess` runtimes, there's no observation mechanism. A standardized streaming protocol makes all agents observable regardless of runtime, which is essential for debugging multi-agent coordination issues.

---

### 15. Config Schema Versioning

**The idea:** Add `schema_version` to the config format to enable backward-compatible evolution.

```toml
[workspace]
name = "my-project"
schema_version = 1  # Required
```

The parser validates against the declared version. `gc validate` warns if the schema version is outdated and offers migration suggestions. Future config changes increment the version with documented migration paths.

**Rationale:** Config formats evolve. Without explicit versioning, users discover breaking changes at runtime. With it, the tooling can provide clear migration guidance when the schema changes. This is standard practice (Docker Compose, GitHub Actions, OpenAPI) and costs almost nothing to implement upfront but saves significant pain later.
