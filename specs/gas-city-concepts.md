# Gas City Concepts: The Irreducible Core

> Gas Town has ~45 concepts across 8 layers. Gas City needs the smallest set
> of ideas such that removing any one means Gas Town can't be rebuilt, and
> adding any new one is unnecessary because it's derivable from the core.
> Ralph, Claude Code Agent Teams, AND Gas Town must all be buildable on this core.

---

## Table of Contents

1. [The Reduction Test](#1-the-reduction-test)
2. [The Five Core Primitives](#2-the-five-core-primitives)
3. [The Four Derived Mechanisms](#3-the-four-derived-mechanisms)
4. [The Core Data Flows](#4-the-core-data-flows)
5. [Implementation Pluggability](#5-implementation-pluggability)
6. [Four Configs, One SDK](#6-four-configs-one-sdk)
7. [What Got Cut (and Why It's OK)](#7-what-got-cut-and-why-its-ok)
8. [Vision Requirements Cross-Reference](#8-vision-requirements-cross-reference)

*Note: §-references (e.g., §2.2, §10) point to sections in `gas-city-spec.md`, not this document.*

---

## 1. The Reduction Test

Every concept in this document passed two tests:

**Necessity Test:** Remove this concept. Can Gas Town still be built?
If yes, it's not core. If no, it stays.

**Irreducibility Test:** Can this concept be expressed as a combination of
other core concepts? If yes, it's derived (built from core). If no, it's
a primitive.

These tests produce a sharp boundary:

- **Core primitives** are SDK infrastructure. They exist as Go interfaces and
  engines. Removing any one breaks all four example configs.
- **Derived mechanisms** are built from primitives via configuration. They
  provide real value but introduce no new irreducible ideas.
- **Everything else** is either deployment-specific (tmux, Dolt, git
  worktrees), a configured role (witness, refinery, deacon), or
  design/future work (federation, Mol Mall).

### The Substrate Layering Principle

The 5+4 concepts form a strict layer cake. Each layer builds only on the
layers below it — nothing in a lower layer references anything higher.

```
Layer 4: Dispatch/Coordination  — Sling + Convoys + Health Patrol
Layer 3: Workflow Engine        — Formulas + Molecules + Plugins
Layer 2: Messaging              — Mail + Nudge + Protocol Messages
Layer 1: Rich Semantics         — Hooks, dependencies, labels, pool management, template rendering
Layer 0: Core Interfaces        — Agent Protocol + Task Store + Event Bus + Config + Prompt Templates
```

**Five invariants that must hold:**
1. No layer may import from a higher layer
2. Each layer exposes a clean public API consumed by the layer above
3. Layer 0 has zero Gas City-specific logic (pure interfaces)
4. Removing any layer above 0 leaves the layers below fully functional
5. Side effects (I/O, process spawning) are confined to Layer 0 implementations

---

## 2. The Five Core Primitives

These are Layer 0-1 in the Gas City spec. Everything else is built FROM them.

### Primitive 1: Agent Protocol

**What it is:** The uniform interface to start, stop, prompt, observe, and
health-check agents regardless of their implementation.

**Gas Town origin:** tmux sessions + `send-keys` + PID detection + `AgentEnv()`.

**Gas City interface (§2.2):**
```
Start / Stop / Restart / IsRunning    — lifecycle
SendPrompt                            — work delivery
ReadOutput / GetState                 — observation
Ping                                  — health
```

**Why irreducible:** Without an agent protocol, there's nothing to
orchestrate. You can't start workers, deliver prompts, or detect stalls.
Every higher concept assumes agents exist and can be controlled.

**What it absorbs from Gas Town:**
- **Agent Identity** — `AgentIdentity` struct (workspace/project/name/instance)
  is part of the protocol, not a separate concept. Identity supports a
  **dual attribution model**: agent identity (`BD_ACTOR`) for executor
  attribution and debugging, and owner identity (`GIT_AUTHOR_EMAIL`) for
  human credit and compliance. Agents execute; humans own.
- **Session Management** — `SessionConfig` (tmux patterns, work dirs) is
  protocol configuration, not a standalone system.
- **Name Pool** — `PoolConfig` (themes, allocation, overflow) is pool
  management within the protocol.
- **Polecat/Crew/Dog distinction** — All are agents with different config
  (ephemeral flag, pool settings, scope). The protocol doesn't know or care.
  Ephemeral agents (polecats) follow a **three-layer architecture**: (1)
  Identity is permanent (agent bead, work history — survives across
  sessions; full CV/skill derivation is future work, see §7 Ledger Export), (2) Sandbox is per-assignment (worktree + branch — created per
  issue, destroyed on completion), (3) Session is per-step (Claude instance +
  context window — may restart within an assignment). There is **no idle
  state** — ephemeral agents are spawned WITH work and destroyed WHEN done.
- **Handoff/Seance** — Session resume is a provider capability (`ResumeConfig`),
  not a separate mechanism. Handoff enables indefinitely long workflows
  despite finite context windows — each new session resumes at the current
  molecule step.
- **Session Checkpoints** — Crash recovery via the `Adopter` interface, which
  is an optional protocol extension.

**Removal test:** Remove Agent Protocol → can't start agents, can't send
prompts, can't detect liveness. Nothing works. **Fails for all four configs.**

---

### Primitive 2: Task Store (Beads)

**What it is:** The universal persistence substrate for all domain state.
CRUD on work units with hooks, dependencies, labels, and concurrent access.

**Gas Town origin:** Dolt SQL tables + `bd` CLI + beads routing + redirects.

**Gas City interface (§10):**
```
Create / Get / Update / Close         — CRUD on beads
Hook / Unhook / Pin                   — atomic work claiming
GetHook(agentID) → *Bead             — query hooked work (health patrol recovery)
AddDependency / GetDependencies       — relationship graph
Query (by status, labels, assignee)   — fast state queries
```

**Why irreducible:** Without a task store, there is no tracked work. No
assignments, no status, no history, no dependencies. Agents could run but
couldn't coordinate — they'd have nothing to claim, nothing to close,
nothing to query.

**Interface requirement — concurrent write safety:** Any Task Store backend
must handle N agents writing simultaneously. Gas Town uses Dolt's
branch-per-agent for write isolation (tested clean with 50 concurrent
writers). Other backends need equivalent strategies (row-level locking,
optimistic concurrency, serializable transactions). This is a functional
requirement of the Task Store interface, not just a deployment detail.

**What it absorbs from Gas Town:**
- **Hook System** — `Hook()` / `Unhook()` are task store operations. The
  agent bead's `hook_bead` field is a task store column, not a separate
  system.
- **Labels-as-State** — Labels are a property of beads. The pattern of using
  labels for fast queries (`needs-rebase`, `gt:merge-request`) is a usage
  convention, not a separate concept.
- **Dependencies** — `AddDependency()` with typed relationships (`blocks`,
  `tracks`, `parent`) is a task store operation.
- **Beads Routing** — Multi-project prefix dispatch (gt-*, bd-*, hq-*) is
  internal task store routing, not a separate mechanism.
- **KRC (Key Record Chronicle)** — TTL-based pruning of ephemeral data is
  a task store lifecycle policy, not a separate system.
- **Beads Redirects** — Shared database access via redirect files is a
  deployment concern of the task store backend.

**Removal test:** Remove Task Store → can't track work, can't assign issues,
can't record state. Agents run in a void. **Fails for all four configs.**

---

### Primitive 3: Event Bus

**What it is:** Append-only, timestamped log of all system activity with
pub/sub for cross-component observation.

**Gas Town origin:** `.events.jsonl` + `flock` + `bd activity --follow`.

**Gas City interface (§8):**
```
Publish(event)                        — any component emits events
Subscribe(filter) → channel           — observe events in real time
Query(timeRange, filters)             — historical lookup
```

**Why irreducible:** Without events, components can't observe each other.
Health monitoring can't detect stalls. Plugins can't trigger on conditions.
The daemon can't provide transparency. You'd have to poll beads for every
state change — which is what events eliminate.

**Interface requirement — concurrent append safety:** Multiple agents publish
events simultaneously. Any Event Bus implementation must guarantee concurrent
append safety (Gas Town uses `flock`; alternatives include write-ahead logs
or message broker guarantees).

**What it absorbs from Gas Town:**
- **Activity feed** — `bd activity --follow` is a subscriber to the event bus.
- **Feed curator** — The daemon's activity-based wake is an event subscriber.
- **Transparency/websocket** — The daemon's websocket streams events. The
  websocket is a delivery mechanism for the event bus, not a separate concept.

**Removal test:** Remove Event Bus → health monitoring is blind, plugins
can't trigger, no activity feed, no transparency. System works but can't
self-heal or self-observe. **Fails for gastown.toml** (health patrol and
plugins require events).

---

### Primitive 4: Config

**What it is:** Declarative TOML configuration with progressive activation.
Config determines which subsystems exist, which agents run, and how they
behave. The SDK is constant; config grows.

**Gas Town origin:** `town.json` + `config.json` + TOML role files +
property layers (wisp/rig/town/system).

**Gas City interface (§3):**
```
Parse(toml) → WorkspaceConfig         — load config
Resolve(overrides...)                  — multi-layer resolution
Detect() → CapabilityLevel            — what's active (0-8)
Validate() → []error                  — consistency checks
```

**Why irreducible:** Without config, behavior is hardcoded. You can't have
"zero hardcoded roles" (the core Gas City principle) without a config system
that defines roles externally. Config is what makes an SDK an SDK instead of
an application.

**What it absorbs from Gas Town:**
- **Property layers** — Config resolution order (wisp/rig/town/system) is
  the config system's override mechanism.
- **Progressive capability model** — Levels 0-8 are determined by which
  config sections are present. Each level is independently useful and
  deployable — you don't need level 8 to get value from level 1. This IS
  the config system's activation logic.
- **Claude Code hooks config** — Hook configuration per role is part of the
  agent config, not a separate system.
- **Rig config** — Per-project identity, prefixes, and remotes are `[projects]`
  sections in the workspace config.

**Removal test:** Remove Config → roles are hardcoded, subsystems can't be
toggled, no progressive capability. You have Gas Town, not Gas City.
**Fails for the SDK concept itself.**

---

### Primitive 5: Prompt Templates

**What it is:** Go templates in Markdown that render role-specific
instructions for agents. Templates define WHAT agents do; the SDK defines
HOW agents work.

**Gas Town origin:** `BuildStartupPrompt()` + role-specific TOML/Markdown
files + GUPP context injection.

**Gas City interface (§6):**
```
roles/<name>.md.tmpl                  — Markdown Go template files
roles/<name>.toml                     — Operational settings per role
Template variables: .Agent, .Task, .Workspace, .Config
```

**Why irreducible:** Without prompt templates, agents have no instructions.
You can start an agent (Agent Protocol), give it work (Task Store), and
observe it (Event Bus), but it doesn't know what to DO. The template is
the behavioral specification — the thing that makes a "witness" different
from a "refinery" or a "worker."

**What it absorbs from Gas Town:**
- **Role definitions** — TOML files defining operational settings per role
  are the structured half of prompt templates.
- **GUPP injection** — The propulsion principle is rendered into startup
  prompts via templates. GUPP is a design philosophy; its manifestation is
  template content.
- **Formula instructions** — Each formula step has an `instructions` field
  that is a mini-template rendered at execution time.

**Removal test:** Remove Prompt Templates → agents start but have no
instructions. You'd have to hardcode every role's behavior in Go, which
violates "zero hardcoded roles." **Fails for ccat.toml and gastown.toml**
(which define coordinator and worker roles via templates).

---

### Primitive Summary

| # | Primitive | One-line definition | Spec section |
|---|-----------|---------------------|--------------|
| 1 | Agent Protocol | Start/stop/prompt/observe agents | §2 |
| 2 | Task Store | CRUD + Hook + Dependencies + Labels | §10 |
| 3 | Event Bus | Pub/sub for all state changes | §8 |
| 4 | Config | TOML parsing + progressive activation | §3-4 |
| 5 | Prompt Templates | Go templates rendering role instructions | §6 |

**The primitive invariant:** These five have no circular dependencies.
Agent Protocol and Event Bus depend on nothing. Task Store depends on
Config (for `data_dir`) and Event Bus (for change notifications). Config
depends on nothing. Prompt Templates depend on Config (for template
variables). The dependency graph is a DAG. All five sit at Layer 0 as
core interfaces; intra-layer dependencies (e.g., Task Store → Config,
Task Store → Event Bus) are permitted within Layer 0.

---

## 3. The Four Derived Mechanisms

These are built from primitives. They provide real capability but introduce
no new irreducible ideas.

### Mechanism 6: Messaging (Mail + Nudge)

**Built from:**
- Task Store — mail messages are stored as beads (`type=message`)
- Agent Protocol — nudge is `SendPrompt()` on a running agent
- Config — routing rules, channel membership

**What it provides:** Inter-agent communication via two complementary modes:
- **Mail** — async, persistent, survives restarts (stored in task store)
- **Nudge** — sync, immediate, fire-and-forget (delivered via agent protocol)

**Gas Town features it replaces:**
- Mail system (`gt mail send/inbox/check`)
- Nudge system (tmux `send-keys` with 500ms debounce)
- Protocol Messages (typed mail: `POLECAT_DONE`, `MERGE_READY`, etc.)
- Groups, Queues, Channels (beads-native messaging primitives)
- Escalation System (severity-based routing: low→labels only,
  medium→mail to overseer, high→mail + external notification,
  critical→all channels + immediate attention)

**Implementation note — signal rate-limiting:** Any `SendPrompt`
implementation needs rate-limiting (Gas Town uses 500ms debounce) to prevent
rapid-fire nudges from overwhelming an agent. This is not tmux-specific —
it's a general concern for any immediate-delivery signal mechanism.

**Why it's derived, not core:** Mail is literally "beads with type=message."
Nudge is literally "call SendPrompt on a running agent." Protocol messages
are typed mail with handler dispatch. Groups/Queues/Channels are query
patterns over message beads. No new primitive is needed — just composition.

**Derivation proof:** `Mail.Send(to, msg)` = `TaskStore.Create(bead{type: "message", to: to, body: msg})`.
`Nudge(agent, text)` = `AgentProtocol.SendPrompt(handle, Prompt{Text: text})`.

---

### Mechanism 7: Formulas and Molecules

**Built from:**
- Config — formulas are TOML template files discovered via config
- Task Store — molecules are root beads with child step beads
- Prompt Templates — step instructions rendered at execution time

**What it provides:** Reusable multi-step workflow patterns.
- **Formula** — static TOML template defining steps (4 types: workflow,
  convoy, expansion for issue decomposition, aspect for cross-cutting concerns)
- **Molecule** — instantiated formula: root bead + step child beads.
  Three-phase lifecycle: Formula (static template) → Protomolecule (formula
  bound to issue but not yet crystallized — important for transactional
  safety) → Molecule (fully instantiated with root + step beads)
- **Step navigation** — `mol next` / `bd close --continue` advances steps
- **Formula health-checking** — SHA256 checksums track installed formulas.
  `CheckFormulaHealth()` detects: ok, outdated, modified, missing, new,
  untracked. Safe updates preserve user-modified formulas while upgrading
  system defaults

**Gas Town features it replaces:**
- Formulas (4 types: workflow, convoy, expansion, aspect)
- Molecules (formula → protomolecule → molecule lifecycle)
- Wisps (ephemeral molecules — just molecules with a TTL/auto-delete flag)
- Formula Resolution (rig → town → embedded → Mol Mall discovery order)
- Plugins (markdown + TOML frontmatter + gate conditions). Plugins have a
  full lifecycle: **gate evaluation** (5 types: cooldown/cron/condition/event/manual)
  runs in a periodic evaluation loop, **wisp execution** creates an ephemeral
  bead and dispatches it to a pool agent via nudge, and **tracking** handles
  labels, digest, and failure notifications

**Why it's derived, not core:** A molecule is a parent bead with child beads
linked by `parent` dependencies. Step navigation is "close current child,
open next child." Formulas are TOML files parsed by the config system.
Plugins are formulas with gate conditions evaluated against the event bus.
All of this composes from Task Store + Config + Event Bus.

**Derivation proof:** `Molecule.Bond(issue, formula)` =
`TaskStore.Create(rootBead)` + `for step in formula.steps: TaskStore.Create(stepBead); TaskStore.AddDependency(stepBead, rootBead, "parent")`.

---

### Mechanism 8: Dispatch (Sling)

**Built from:**
- All five primitives + Messaging + Formulas/Molecules
- Specifically: Formula selection → Molecule instantiation → Hook to agent →
  Nudge agent

**What it provides:** Single-command work dispatch that encapsulates the full
assignment flow.

**Gas Town features it replaces:**
- Sling (`gt sling <issue> <rig>`)
- Convoys (batch tracking via `tracks` dependencies + observers)
- Swarms (multi-agent parallel work + integration branches)
- Reactive feeding (convoy observer dispatches next ready issue)

**Why it's derived, not core:** Sling is a composed operation:
1. Find/spawn agent (Agent Protocol + Config pool settings)
2. Select formula (Config resolution order)
3. Create molecule (Task Store + Formula)
4. Hook molecule to agent (Task Store)
5. Nudge agent (Messaging)
6. Log event (Event Bus)

No step requires a new primitive. Convoys are tracking beads with `tracks`
dependencies — pure Task Store. Convoys use a **redundant observation
pattern**: Witness, Refinery, and Daemon all independently check convoy
completion on every relevant event. All checks are idempotent — running
them multiple times is safe. This is NDI in practice: multiple independent
observers converge on correct state. Swarms are convoys with shared
branches — deployment-specific (Git), not a new concept.

**Derivation proof:** `Sling(issue, rig)` = `agent := Pool.GetOrSpawn()` +
`mol := Formula.Bond(issue)` + `TaskStore.Hook(agent, mol)` +
`Messaging.Nudge(agent, "work assigned")` + `EventBus.Publish("sling", ...)`.

---

### Mechanism 9: Health Patrol

**Built from:**
- Agent Protocol — `Ping()` for liveness, `Restart()` for recovery
- Event Bus — stall events, patrol events, activity monitoring
- Config — thresholds (`stuck_threshold`, `ping_timeout`, `consecutive_failures`)
- Task Store — stale hook detection (work-on-hook + dead agent)

**What it provides:** Self-healing through continuous monitoring, stall
detection, and automatic recovery. Includes a **self-cleaning model** for
ephemeral agents: (1) agent finishes → signals DONE to lifecycle monitor,
(2) monitor verifies cleanup (checks all remotes before delete), (3) sends
MERGE_READY to merge processor, (4) after merge confirmation, nukes agent
sandbox. Safety verification at each step prevents data loss (uses
Messaging (6) for inter-agent coordination).

**Design principle — deterministic shutdown:** Agent shutdown operations
use deterministic code (Go state machines), NOT AI agents. This prevents
the "AI trying to shut down another AI" problem. Gas Town's "Dance Dogs"
are lightweight goroutines with `Warrant` structs and explicit state
transitions — mechanical, not intelligent. Some infrastructure operations
must be deterministic to be safe.

**Gas Town features it replaces:**
- Daemon heartbeat cycle (15 checks per cycle)
- Watchdog chain (daemon → boot → deacon → witnesses)
- Boot triage (fresh-context intelligent assessment)
- Deacon patrol (continuous Claude-driven monitoring)
- Witness zombie detection (cross-reference beads + tmux)
- Proactive crash detection (work-on-hook + dead session)

**Why it's derived, not core:** Health monitoring is "ping agents
periodically, compare against thresholds from config, restart if stalled,
publish events." Every verb in that sentence uses an existing primitive.
The watchdog chain topology (daemon → boot → deacon → witnesses) is a
specific configuration of agents monitoring other agents — pure config.

**Derivation proof:** `Patrol()` = `for agent in agents: result := AgentProtocol.Ping(agent); if result.Latency > Config.StuckThreshold: EventBus.Publish("stall", agent); AgentProtocol.Restart(agent)`.

---

### Mechanism Summary

| # | Mechanism | Built from | What it enables |
|---|-----------|-----------|-----------------|
| 6 | Messaging | TaskStore + AgentProtocol + Config | Inter-agent communication |
| 7 | Formulas/Molecules | Config + TaskStore + Templates | Multi-step workflows |
| 8 | Dispatch (Sling) | All primitives + Mechanisms 6-7 | Unified work assignment |
| 9 | Health Patrol | AgentProtocol + EventBus + Config + TaskStore | Self-healing |

**The mechanism invariant:** Every mechanism can be expressed as a function
over the five primitives. No mechanism requires introducing a 6th primitive.

---

## 4. The Core Data Flows

Four canonical flows showing how primitives compose into system behavior.

### Flow 1: Single Task (hello-world.toml level)

The simplest possible flow. One agent, one task, no messaging, no loop.

```
User                    TaskStore           Agent Protocol
  │                        │                      │
  ├─ bd create "fix bug" ──►                      │
  │                  [bead created]                │
  │                        │                      │
  │                        ├── Hook(agent, bead) ──►
  │                  [agent claims work]           │
  │                        │                      │
  │                        │    [agent executes]   │
  │                        │                      │
  │                        ◄── Close(bead) ───────┤
  │                  [bead closed]                 │
```

**Primitives used:** Agent Protocol, Task Store, Config.
**Mechanisms used:** None. This flow predates all derived mechanisms.

> **ralph.toml adds `[agents.loop]`:** after closing a bead, the agent
> loops back to poll for the next one, each time with a clean context.

---

### Flow 2: Multi-Step Workflow (ccat.toml level)

Formula → Molecule → step-by-step execution.

```
Coordinator         TaskStore          Formula Engine       Worker
    │                  │                    │                  │
    ├─ Sling(issue) ──►│                    │                  │
    │            [select formula] ─────────►│                  │
    │                  │◄─ Bond(issue) ─────┤                  │
    │            [root + step beads]        │                  │
    │                  │                    │                  │
    │                  ├─── Hook(worker, step1) ──────────────►│
    │                  │                                [execute step1]
    │                  │◄──── Close(step1, --continue) ───────┤
    │            [auto-advance]                                │
    │                  ├─── Hook(worker, step2) ──────────────►│
    │                  │                                [execute step2]
    │                  │◄──── Close(step2) ────────────────────┤
    │            [molecule complete]                            │
```

**Primitives used:** Agent Protocol, Task Store, Config, Prompt Templates.
**Mechanisms used:** Formulas/Molecules (7), Dispatch (8).

---

### Flow 3: Multi-Agent Dispatch (gastown.toml level)

Sling → spawn → hook → nudge → execute → done → merge.

```
Sling       Pool        TaskStore      Worker       Messaging     Witness
  │          │              │             │             │            │
  ├─ spawn ──►              │             │             │            │
  │     [allocate name]     │             │             │            │
  │          ├─ Start() ────┼─────────────►             │            │
  │          │              │        [agent live]       │            │
  │          │              │             │             │            │
  ├─ bond ───┼──────────────►             │             │            │
  │          │        [molecule created]  │             │            │
  │          │              │             │             │            │
  ├─ hook ───┼──────────────►             │             │            │
  │          │        [work claimed]      │             │            │
  │          │              │             │             │            │
  ├─ nudge ──┼──────────────┼─────────────┼─── Nudge ──►            │
  │          │              │             │     [wake]  │            │
  │          │              │        [execute work]     │            │
  │          │              │             │             │            │
  │          │              │             ├── Mail ─────┼── DONE ───►│
  │          │              │             │             │    [verify]│
  │          │              │             │             │            │
  │          │              │             │    ◄────────┼── MERGE ───┤
  │          │              │             │             │   READY    │
```

**Primitives used:** All five.
**Mechanisms used:** All four (Messaging, Formulas, Dispatch, Health Patrol
runs concurrently monitoring the workers).

---

### Flow 4: Self-Healing (health patrol)

Patrol → detect stall → restart → re-hook work.

```
Health Patrol      Agent Protocol       TaskStore         Event Bus
     │                  │                  │                 │
     ├─── Ping(agent) ──►                  │                 │
     │             [no response]           │                 │
     │                  │                  │                 │
     ├─── Ping(agent) ──►                  │                 │
     │             [timeout × 3]           │                 │
     │                  │                  │                 │
     ├──────────────────┼──────────────────┼── Publish ─────►│
     │                  │                  │          [stall event]
     │                  │                  │                 │
     ├─── Restart ──────►                  │                 │
     │             [new session]           │                 │
     │                  │                  │                 │
     │                  ├── GetHook ───────►                 │
     │                  │            [work still hooked]     │
     │                  │                  │                 │
     │                  ├── SendPrompt ────►                 │
     │                  │            [resume work]           │
     │                  │                  │                 │
     ├──────────────────┼──────────────────┼── Publish ─────►│
     │                  │                  │        [recovery event]
```

**Primitives used:** Agent Protocol, Task Store, Event Bus, Config (thresholds).
**Mechanisms used:** Health Patrol (9).

---

## 5. Implementation Pluggability

Each primitive has a pluggable implementation. The SDK provides interfaces;
deployments provide backends.

| Primitive | Gas Town Implementation | Alternative Implementations |
|-----------|------------------------|----------------------------|
| **Agent Protocol** | tmux + Claude Code | Agent SDK, subprocess, remote SSH |
| **Task Store** | Dolt SQL (branch-per-agent) | SQLite, filesystem, Postgres, in-memory |
| **Event Bus** | JSONL file + flock | In-process channels, Redis pub/sub, NATS |
| **Config** | TOML files + property layers | Environment variables, API-served config |
| **Prompt Templates** | Go `text/template` in Markdown | Any template engine (Handlebars, Jinja2) |

**The pluggability invariant:** Swapping any implementation preserves all
derived mechanisms. If you replace Dolt with SQLite, mail still works
(it's still beads). If you replace tmux with subprocess, nudge still works
(it's still `SendPrompt`). The mechanisms are defined in terms of primitive
interfaces, not implementations.

**Gas City's built-in providers (v1):**

| Primitive | Default Provider | Shipped Alternatives |
|-----------|-----------------|---------------------|
| Agent Protocol | `claude`, `codex`, `gemini` (all tmux-based) | `subprocess` (stdin/stdout) |
| Task Store | `beads` (Dolt/SQLite) | `filesystem` (JSONL files) |
| Event Bus | JSONL + flock | (single impl, extensible) |
| Config | TOML parser | (single impl, standard) |
| Prompt Templates | Go `text/template` | (single impl, standard) |

---

## 6. Four Configs, One SDK

The four example configs shipped with Gas City are different selections
from the same 5+4 concepts. They demonstrate progressive capability —
not different SDKs, but different activation levels. Each level is
independently useful and deployable on its own.

### hello-world.toml — Single Agent, Single Session (Primitives 1-2, 4 only)

```
Primitives active:  Agent Protocol, Task Store, Config
Mechanisms active:  None
Capability level:   1 (workspace + agent + tasks)
```

What it uses:
- **Agent Protocol** — one agent, auto-detected provider
- **Task Store** — beads backend for tracking work
- **Config** — workspace name, agent definition (`[workspace]` + `[[agents]]` + `[tasks]`)

What it skips:
- No loop — user creates beads, agent works them in a single session
- No Prompt Templates (single agent, no role differentiation needed)
- No Event Bus (no monitoring, no subscriptions)
- No Messaging (one agent, no one to message)
- No Formulas (flat tasks, no multi-step workflows)
- No Dispatch (no pool, no spawning)
- No Health Patrol (no daemon, no monitoring)

### ralph.toml — Single Agent + Loop (Adds `[agents.loop]`)

```
Primitives active:  Agent Protocol, Task Store, Config, (no templates needed)
Mechanisms active:  None
Capability level:   2 (agent + tasks + loop)
```

What it adds over hello-world:
- **Loop** — `[agents.loop]` config turns the agent into a polling loop.
  Each bead is a separate run through the agent with a **clean context**.
  The loop is what distinguishes ralph from hello-world: the agent
  continuously polls for ready beads, claims one, executes it in a fresh
  session, then loops back.
- **Event Bus** — not configured (no `[daemon]`), but available passively

What it skips:
- No Prompt Templates (single agent, no role differentiation needed)
- No Messaging (one agent, no one to message)
- No Formulas (flat tasks, no multi-step workflows)
- No Dispatch (no pool, no spawning, just a poll loop)
- No Health Patrol (no daemon, no monitoring)

### ccat.toml — Coordinator + Workers / Claude Code Agent Teams (Adds Mechanisms 6, 8)

```
Primitives active:  All five (Agent Protocol, Task Store, Event Bus, Config, Templates)
Mechanisms active:  Messaging (6), Dispatch (8)
Capability level:   4 (multiple agents + messaging)
```

What it adds over ralph:
- **Prompt Templates** — coordinator and worker roles need different instructions
- **Event Bus** — coordinator observes worker state
- **Messaging** — coordinator sends work, workers report completion
- **Dispatch** — pool auto-spawn, work distribution

What it skips:
- No Formulas (coordinator manages workflow logic directly)
- No Health Patrol (no daemon, coordinator handles its own pool)

### gastown.toml — Full Orchestration (All Primitives + All Mechanisms)

```
Primitives active:  All five
Mechanisms active:  All four (Messaging, Formulas, Dispatch, Health Patrol)
Capability level:   8 (full orchestration)
```

What it adds over ccat:
- **Formulas/Molecules** — structured multi-step workflows per role
- **Health Patrol** — daemon + watchdog chain for self-healing
- **Plugins** — automated maintenance tasks with gate conditions
- Multiple projects, 8 roles, cross-project coordination

### The Progressive Proof

Each config is a strict superset of the previous:

```
hello     = {AgentProtocol, TaskStore, Config}
ralph     = hello + {Config: agents.loop (clean context per bead)}
ccat      = ralph + {EventBus, Templates, Messaging, Dispatch}
gastown   = ccat + {Formulas, HealthPatrol}
```

No config requires a concept that isn't in the 5+4 set. No config requires
a concept that can't be derived from the five primitives. Each config is
independently useful — hello-world runs real work, not just a demo stub.

---

## 7. What Got Cut (and Why It's OK)

Every concept from `gas-town-concepts.md` is accounted for below. Each is
either absorbed into a primitive, absorbed into a mechanism, identified as
deployment-specific, or identified as future/design work.

### Absorbed into Core Primitives

| Gas Town Concept | Absorbed Into | How |
|------------------|--------------|-----|
| Agent Identity | Agent Protocol | `AgentIdentity` struct (workspace/project/name/instance) |
| Session Management | Agent Protocol | `SessionConfig` for tmux patterns and work dirs |
| Name Pool | Agent Protocol | `PoolConfig` with themes and allocation rules |
| Polecat Manager | Agent Protocol | Ephemeral pool agents with `ephemeral=true` + `PoolConfig` |
| Crew | Agent Protocol | Persistent agents with `ephemeral=false` + `isolation=none` |
| Dog Pool | Agent Protocol | Helper agents with specific role config |
| Handoff/Seance | Agent Protocol | `ResumeConfig` + `Adopter` interface |
| Session Checkpoints | Agent Protocol | `Adopter` interface for crash recovery |
| Hook System | Task Store | `Hook()` / `Unhook()` operations on beads |
| Labels-as-State | Task Store | Labels are a bead property for fast queries |
| Dependencies | Task Store | `AddDependency()` with typed relationships |
| Beads Routing | Task Store | Multi-project prefix dispatch (internal routing) |
| KRC | Task Store | TTL-based pruning as a lifecycle policy |
| Activity Feed | Event Bus | A subscriber that streams events |
| Config Property Layers | Config | Override resolution order (wisp/rig/town/system) |
| Progressive Capability | Config | Level detection from present config sections |
| Role Definitions | Prompt Templates | `roles/*.toml` + `roles/*.md.tmpl` files |
| GUPP (propulsion) | Prompt Templates | Rendered into startup prompts + `auto_execute` config flag |

### Absorbed into Derived Mechanisms

| Gas Town Concept | Absorbed Into | How |
|------------------|--------------|-----|
| Mail | Messaging (6) | Messages stored as beads in the task store |
| Nudge | Messaging (6) | `SendPrompt()` on a running agent |
| Protocol Messages | Messaging (6) | Typed mail with handler dispatch |
| Groups/Queues/Channels | Messaging (6) | Query patterns over message beads |
| Escalation System | Messaging (6) | Severity-based routing via mail + config rules |
| Formulas | Formulas/Molecules (7) | TOML templates parsed by config system |
| Molecules | Formulas/Molecules (7) | Root bead + child step beads in task store |
| Wisps | Formulas/Molecules (7) | Molecules with ephemeral/auto-delete flag |
| Plugins | Formulas/Molecules (7) + Health Patrol (9) + Config (4) | Gate-evaluated (cooldown/cron/condition/event/manual) formulas with ephemeral wisp execution and pool dispatch. Full lifecycle: gate evaluation loop → wisp bead creation → pool dispatch + nudge → tracking via labels/digest/failure notifications. CLI: `gc plugin {list, run, disable, enable, status}` |
| Formula Resolution | Formulas/Molecules (7) | Config resolution order for formula discovery |
| Sling | Dispatch (8) | Composed: spawn + bond + hook + nudge + log |
| Convoys | Dispatch (8) | Tracking beads with `tracks` dependencies |
| Swarms | Dispatch (8) | Convoys + integration branches (Git-specific) |
| Reactive Feeding | Dispatch (8) | Convoy observer dispatches next ready issue |
| Merge Queue | Dispatch (8) | Role-specific workflow (refinery formula) |
| Daemon | Health Patrol (9) | Go process running the patrol loop |
| Watchdog Chain | Health Patrol (9) | Configured hierarchy of monitoring agents |
| Boot Triage | Health Patrol (9) | Fresh-context health assessment agent |
| Witness Zombie Detection | Health Patrol (9) | Cross-reference beads + agent liveness |
| Proactive Crash Detection | Health Patrol (9) | Work-on-hook + dead agent → auto-restart |

### Deployment-Specific (Not in SDK)

| Gas Town Concept | Why It's Deployment-Specific |
|------------------|------------------------------|
| tmux | One possible execution substrate for Agent Protocol |
| Dolt SQL | One possible storage backend for Task Store |
| Git Worktrees | One possible sandbox strategy (SandboxProvider) |
| Bare Repo (`.repo.git`) | Git-specific shared object optimization |
| Filesystem Conventions (`~/gt/`) | One possible directory layout |
| Beads Redirects | Dolt/worktree-specific shared DB access |
| Branch-per-Polecat | Dolt-specific write isolation strategy |
| Integration Branches | Git-specific multi-agent landing strategy. Three-layer safety guardrails: (1) formula/role instructions tell agents which branch to target, (2) pre-push hook validates branch targeting, (3) only the merge processor's code path can push to main |
| Claude Code Hooks | One possible agent tool integration |
| `tmux send-keys` debounce | tmux-specific nudge implementation detail |
| Dolt server health | Dolt-specific infrastructure monitoring |
| Cross-rig worktree paths | Filesystem-specific identity preservation |

### Configured Roles (Not Concepts)

| Gas Town Role | What It Is in Gas City |
|---------------|----------------------|
| Mayor | A workspace-scoped agent with a coordinator role template |
| Deacon | A workspace-scoped agent with a patrol role template |
| Witness | A project-scoped agent with a lifecycle-monitor role template |
| Refinery | A project-scoped agent with a merge-processor role template |
| Polecat | An ephemeral, pool-managed agent with a worker role template |
| Crew Member | A persistent agent with `isolation=none` |
| Boot | An ephemeral agent spawned for health triage |
| Dog | A helper agent managed by the deacon role |

These are all just `[[agents]]` entries with different `role`, `scope`,
`ephemeral`, and `pool` settings. The SDK doesn't know about any of them.

### Design/Future Work (Not in v1)

| Gas Town Concept | Status |
|------------------|--------|
| Federation (HOP) | Future: multi-workspace coordination protocol |
| Mol Mall | Future: formula registry/marketplace |
| Ledger Export | Future: three data planes (Operational for live queries, Ledger for audit/compliance, Design for cross-workspace ideas) with trigger-based transitions (bead closure→L2 compressed record, convoy completion→L2 summary, refinery merge→L2 outcome, design decision→L3 full-fidelity reasoning capture). CV and skill derivation built on the Ledger plane |
| Data Fidelity Levels | Future: L0 ephemeral noise (KRC auto-prunes), L1 operational state (fast access), L2 compressed completion records (audit trail), L3 full-fidelity ground truth (skill derivation, CV building) |
| `gt doctor` | Tooling: diagnostic command, not a core concept |

### Cross-Cutting Principles (Design Philosophy, Not Concepts)

| Principle | How It Manifests |
|-----------|-----------------|
| **GUPP** (Propulsion) | `auto_execute=true` in Config + rendered into Prompt Templates |
| **MEOW** (Molecular Work) | Formulas/Molecules mechanism (7) |
| **NDI** (Nondeterministic Idempotence) | Redundant observers in Health Patrol (9) + idempotent Task Store ops |
| **ZFC** (Zero-File Compliance) | Task Store as single source of truth — no state files |

These principles inform design decisions but are not themselves concepts
that need SDK representation. They emerge from correct use of the primitives.

---

## 8. Vision Requirements Cross-Reference

Every requirement from `gas-city-vision.md` is addressed by the 5+4 core.

| # | Vision Requirement | Addressed By |
|---|-------------------|--------------|
| 1 | Orchestration-builder toolkit (not just an orchestrator) | Config (4) + Prompt Templates (5) — behavior is user-supplied |
| 2 | Next step beyond Gas Town | All 5+4 — Gas Town is one configuration of the SDK |
| 3 | Multiple architectures/topologies | Config (4) — hello-world, ralph, ccat (Claude Code Agent Teams), gastown are different topologies |
| 4 | Progressive capability model | Config (4) — levels 0-8 from config section presence; each level independently useful |
| 5 | Reasonable defaults, configure what you need | Config (4) — defaults table in §3.3 |
| 6 | Create your own roles, teams, coordination rules | Prompt Templates (5) + Config (4) |
| 7 | Roles external to code | Prompt Templates (5) — `roles/*.md.tmpl` files |
| 8 | Sandboxes, plugins, hooks (explicit extensibility) | Agent Protocol (1) sandbox providers + Formulas (7) plugins |
| 9 | Same capabilities as GT, configurable not hardcoded | All 5+4 — every GT concept maps to config + primitives |
| 10 | Higher concepts built on lower concepts | Substrate Layering: Primitives (L0-1) → Mechanisms (L2-4) |
| 11 | Transparent via daemon websocket | Event Bus (3) — websocket is a delivery mechanism for events |
| 12 | Uniform agent abstraction | Agent Protocol (1) — provider interface decouples from impl |

---

## Appendix: The Completeness Argument

**Claim:** The 5 primitives + 4 mechanisms are necessary and sufficient to
build Gas Town, Claude Code Agent Teams, or Ralph.

**Necessity (removal test):**
- Remove Agent Protocol → can't run agents (all configs fail)
- Remove Task Store → can't track work (all configs fail)
- Remove Event Bus → can't observe/self-heal (gastown fails)
- Remove Config → can't define roles externally (SDK concept fails)
- Remove Prompt Templates → can't differentiate roles (ccat, gastown fail)

**Sufficiency (construction test):**
- Every Gas Town concept in the 45-concept inventory maps to either a
  primitive, a mechanism, a deployment detail, or a configured role.
- No Gas Town concept requires an idea outside the 5+4 set.
- The four example configs exercise all 5+4 concepts at different levels.

**Irreducibility (no concept is derivable from the others):**
- Agent Protocol can't be built from Task Store + Event Bus (you need
  process control)
- Task Store can't be built from Agent Protocol + Event Bus (you need
  persistence)
- Event Bus can't be built from Task Store (you need real-time pub/sub,
  not just persistence)
- Config can't be built from the others (you need declarative activation)
- Prompt Templates can't be built from Config alone (you need template
  rendering with agent/task context)

The five primitives are a basis. The four mechanisms are derived. Together
they span the full space of Gas Town's functionality.
