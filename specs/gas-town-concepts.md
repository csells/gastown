# Gas Town Concepts: From Core Atoms to Orchestration

> A comprehensive analysis of Gas Town's layered architecture, documenting every
> feature, what it's built on, what unique value it provides, and how it works
> through the system. The goal is to identify which concepts are **core
> intrinsics** and which are **derived/configured** behaviors — informing the
> Gas City SDK design.

---

## Table of Contents

1. [Layering Overview](#layering-overview)
2. [Layer 0: External Infrastructure](#layer-0-external-infrastructure)
3. [Layer 1: Core Primitives](#layer-1-core-primitives)
4. [Layer 2: Work Management](#layer-2-work-management)
5. [Layer 3: Messaging](#layer-3-messaging)
6. [Layer 4: Workflow Templates](#layer-4-workflow-templates)
7. [Layer 5: Agent Lifecycle](#layer-5-agent-lifecycle)
8. [Layer 6: Coordination](#layer-6-coordination)
9. [Layer 7: Orchestration](#layer-7-orchestration)
10. [Cross-Cutting Principles](#cross-cutting-principles)
11. [Gas City Implications](#gas-city-implications)
12. [Appendix: Document Inventory](#appendix-document-inventory)

---

## Layering Overview

Gas Town is built as a strict layer cake. Each layer builds on the ones below
it. Nothing in a lower layer references anything in a higher layer. This is the
**Substrate Layering Principle** that Gas City formalizes.

```
┌─────────────────────────────────────────────────────────┐
│  Layer 7: ORCHESTRATION                                 │
│  Daemon, Deacon, Witness, Refinery, Mayor, Watchdog     │
├─────────────────────────────────────────────────────────┤
│  Layer 6: COORDINATION                                  │
│  Sling, Convoys, Swarms, Merge Queue, Integration       │
├─────────────────────────────────────────────────────────┤
│  Layer 5: AGENT LIFECYCLE                               │
│  Polecat Manager, Name Pool, Worktrees, Sessions        │
├─────────────────────────────────────────────────────────┤
│  Layer 4: WORKFLOW TEMPLATES                            │
│  Formulas, Molecules, Wisps, Plugins                    │
├─────────────────────────────────────────────────────────┤
│  Layer 3: MESSAGING                                     │
│  Mail, Nudge, Protocol Messages, Channels               │
├─────────────────────────────────────────────────────────┤
│  Layer 2: WORK MANAGEMENT                               │
│  Hooks, Labels-as-State, Routing, Redirects, Deps       │
├─────────────────────────────────────────────────────────┤
│  Layer 1: CORE PRIMITIVES                               │
│  Beads, Agent Identity, Events, Config, Sessions        │
├─────────────────────────────────────────────────────────┤
│  Layer 0: EXTERNAL INFRASTRUCTURE                       │
│  tmux, Dolt SQL, Git/Worktrees, Filesystem              │
└─────────────────────────────────────────────────────────┘
```

**Key invariant:** Each layer depends ONLY on the layers below it. A concept at
Layer N may use concepts from Layer 0..N-1 but never from Layer N+1 or above.

---

## Layer 0: External Infrastructure

These are external systems that Gas Town does NOT own. They provide the raw
substrate that everything else is built on. Gas City must treat these as
pluggable — the specific choices here are Gas Town implementation decisions,
not intrinsic to the orchestration model.

### tmux — Process Execution Substrate

**What it is:** Terminal multiplexer providing named sessions, panes, and
process isolation.

**What Gas Town uses it for:**
- **Agent sessions:** Each agent (polecat, witness, refinery, deacon) runs in
  a named tmux session (e.g. `gt-gastown-toast`, `gt-deacon`)
- **Nudging:** Synchronous immediate message delivery via `tmux send-keys`
- **Process detection:** `IsAgentRunning()` checks if tmux pane PID is alive
- **Environment injection:** Sets `BD_ACTOR`, `GT_ROLE`, etc. per session

**Key implementation details:**
- Session naming convention: `gt-<rig>-<agent>` or `gt-<role>`
- Nudge uses `send-keys` with literal mode and separate Enter keystroke
- 500ms debounce on nudges to prevent rapid-fire delivery
- `NewSession()` / `KillSession()` for lifecycle management

**Unique value:** Provides OS-level process isolation, named addressability,
and a side-channel communication mechanism (send-keys) that operates outside
the agent's stdin/stdout.

**Gas City implication:** tmux is one possible "execution substrate." Gas City
abstracts this as the Agent Protocol — any substrate (tmux, Docker, Agent SDK,
direct process) that can start/stop agents, detect liveness, and deliver
signals.

### Dolt SQL — Storage Substrate

**What it is:** Git-for-data SQL database with branch/merge semantics.

**What Gas Town uses it for:**
- **Only storage backend** — all beads, mail, issues, events stored in Dolt
- **SQL server mode** — one Dolt SQL server per town on port 3307
- **Branch-per-polecat** — concurrent write isolation (50 writers tested clean)
- **Three data planes:**
  - **Operational** (Dolt server) — live state, fast queries
  - **Ledger** (permanent record) — audit trail, compliance
  - **Design** (global ideas) — cross-town concept storage

**Key implementation details:**
- Multi-table SQL schema: `issues`, `mail`, `channels`, etc.
- Agent writes go to named branches, merged back on completion
- Server health checked every 30 seconds by daemon
- Dolt commit on every state change for audit trail

**Unique value:** Provides SQL query power with git-like branching for write
concurrency. Eliminates the need for file-based state or external databases.

**Gas City implication:** Dolt is Gas Town's specific storage choice. Gas City
should define a **Task Store** abstraction — any backend that supports CRUD on
work units, dependency graphs, labels, and concurrent writes.

### Git / Worktrees — Code Management Substrate

**What it is:** Distributed version control with worktree support for parallel
checkouts.

**What Gas Town uses it for:**
- **Bare repo** (`.repo.git`) — shared git objects across all worktrees
- **Worktrees per agent** — each polecat/crew/refinery gets own checkout
- **Branch-per-polecat** — unique timestamped branches: `polecat/<name>/<issue>@<timestamp>`
- **Multi-remote verification** — checks all remotes before cleanup

**Key implementation details:**
- Mayor's clone is canonical — bare repo created from it
- Worktree creation: `git worktree add` from bare repo
- Cross-rig worktrees: `~/gt/beads/crew/gastown-joe/` preserves identity
- Stale branch cleanup by daemon periodic patrol

**Unique value:** Enables parallel code work by multiple agents without
repository contention. Each agent has full isolation.

**Gas City implication:** Workspace management is a concern of the agent
lifecycle system. Gas City should provide a **Workspace** abstraction that
handles agent-specific isolated work areas.

### Filesystem — Directory Conventions

**What it is:** The physical layout of the Gas Town installation.

**What Gas Town uses it for:**

```
~/gt/                           Town root
├── .beads/                     Town-level beads (hq-* prefix, mail)
├── .events.jsonl               Event log
├── mayor/                      Mayor config + clone
│   └── town.json
├── deacon/                     Deacon daemon
│   ├── heartbeat.txt
│   └── dogs/                   Deacon helpers
│       └── boot/
└── <rig>/                      Project container
    ├── config.json             Rig identity
    ├── .beads/ → redirect      Symlink to canonical beads
    ├── .repo.git/              Bare repo (shared by worktrees)
    ├── .dolt-data/             Dolt SQL data
    ├── mayor/rig/              Mayor's clone (canonical beads)
    ├── refinery/rig/           Worktree on main
    ├── witness/                No clone (monitors only)
    ├── crew/                   Persistent human workspaces
    └── polecats/               Ephemeral polecat worktrees
```

**Unique value:** Convention-over-configuration directory structure that encodes
role, rig membership, and identity through path position.

**Gas City implication:** Directory layout is a deployment concern. Gas City
should support but not mandate specific filesystem layouts.

---

## Layer 1: Core Primitives

These are the foundational abstractions that Gas Town builds directly on top of
Layer 0 infrastructure. They are the **true atoms** of the system.

> **Note on scope:** This document covers implemented features and documented
> design concepts. Some features exist as design specs but are not yet
> implemented (marked with "[Design]"). Features that are fully implemented
> and in production are unmarked.

### Beads — Atomic Work Units

**Built on:** Dolt SQL (storage), Filesystem (routing config)

**What it is:** A bead is the universal unit of tracked work. Every task, issue,
message, convoy, molecule step, and agent record is a bead. The `bd` CLI is the
primary interface to beads.

**Core data model (from `internal/beads/issue.go`):**
```go
type Issue struct {
    ID          string   // Prefixed identifier (gt-abc, hq-xyz, bd-123)
    Title       string
    Status      string   // open, in_progress, closed
    Labels      []string // Cached state signals
    Priority    int
    Assignee    string   // Current agent working on it
    HookBead    string   // The bead hooked to this agent bead
    AgentState  string   // Agent-specific state
    CreatedBy   string   // BD_ACTOR of creator
    Description string   // May contain structured JSON fields
}
```

**Key operations:**
- `bd create` — Create a new bead
- `bd show` / `bd list` — Query beads
- `bd edit` — Modify a bead
- `bd close` — Close a bead (mark work complete)
- `bd dep add` — Add dependency relationships
- `bd audit` — Query work history by actor

**Unique value:** Beads are the **single source of truth** for all domain state.
There are no state files, no in-memory caches, no secondary databases. If it's
not in a bead, it doesn't exist. This is ZFC (Zero-File Compliance) in practice.

**How it works through the system:** When work enters the system (via `gt sling`,
`bd create`, or convoy), a bead is created. The bead tracks the work through
assignment (hook), execution (status transitions, labels), completion (close),
and merge (refinery processing). Every role reads and writes beads — they are
the shared language of the entire system.

### Agent Identity — Attribution Model

**Built on:** Filesystem (directory position), Environment variables

**What it is:** Every Gas Town agent has a unique identity expressed as a path:
`<rig>/<role>/<name>`. This identity is embedded in everything the agent does.

**Identity format (BD_ACTOR):**

| Role | Format | Example |
|------|--------|---------|
| Mayor | `mayor` | `mayor` |
| Deacon | `deacon` | `deacon` |
| Witness | `<rig>/witness` | `gastown/witness` |
| Refinery | `<rig>/refinery` | `gastown/refinery` |
| Crew | `<rig>/crew/<name>` | `gastown/crew/joe` |
| Polecat | `<rig>/polecats/<name>` | `gastown/polecats/toast` |
| Dog | `deacon/dogs/<name>` | `deacon/dogs/boot` |

**How identity propagates:**
- **Git commits:** `GIT_AUTHOR_NAME=gastown/polecats/toast`
- **Beads records:** `created_by: gastown/crew/joe`
- **Event logs:** `actor: gastown/polecats/toast`
- **Environment:** `BD_ACTOR`, `GT_ROLE`, `GT_RIG`, `GT_POLECAT`/`GT_CREW`

**Unique value:** Every action in the system is permanently attributed to a
specific agent. This enables work history (CVs), capability-based routing,
model A/B testing, compliance auditing, and debugging. Identity is preserved
across ephemeral sessions — a polecat's name persists even as its tmux sessions
and worktrees come and go.

**Agent vs Owner distinction:**
- `BD_ACTOR` = local executor attribution (debugging)
- `GIT_AUTHOR_EMAIL` = global human identity (CV credit)
- Agents execute; humans own.

### Events — Activity Log

**Built on:** Filesystem (JSONL file), cross-process flock

**What it is:** A timestamped, append-only log of all system activity stored as
JSONL at `{town}/.events.jsonl`.

**Event structure (from `internal/events/`):**
```go
type Event struct {
    Timestamp  time.Time
    Source     string    // Component that generated event
    Type       string    // Event type (sling, patrol.muted, agent.started, etc.)
    Actor      string    // BD_ACTOR of the agent
    Payload    any       // Event-specific data
    Visibility string    // Who can see this event
}
```

**Key event types:**
- `sling` — Work dispatched to agent
- `agent.started` / `agent.stopped` — Lifecycle transitions
- `patrol.muted` — Patrol cycle suppressed
- `mode.degraded` — System entered degraded mode
- `merge.completed` / `merge.failed` — Refinery results

**Unique value:** Provides a complete, chronological record of everything that
happened in the town. Used for real-time activity feeds (`bd activity --follow`),
debugging, and the daemon's feed curator for activity-based wake.

**Implementation detail:** Uses `flock` for cross-process write safety — multiple
agents can append events concurrently without corruption.

### Config System — Role Definitions and Settings

**Built on:** Filesystem (TOML/JSON files)

**What it is:** Declarative configuration for roles, rigs, and runtime behavior.

**Configuration layers:**
1. **Town config** (`mayor/town.json`) — Town-level identity and settings
2. **Rig config** (`<rig>/config.json`) — Per-rig identity, prefixes, remotes
3. **Role definitions** (TOML with resolution order) — Instructions per role
4. **Settings** (per-rig) — Merge queue config, patrol intervals, etc.
5. **Runtime config** — Agent presets, model selection
6. **Daemon config** (`mayor/daemon.json`) — Patrol filtering, heartbeat settings

**Property Layers (multi-level config resolution):**
Gas Town uses a four-layer property system where earlier layers override later
ones (see `docs/design/property-layers.md`):
1. **Wisp layer** (transient, town-local) — Temporary local overrides
2. **Rig layer** (persistent, project-scoped) — Per-rig settings
3. **Town layer** (persistent, town-scoped) — Cross-rig defaults
4. **System layer** (embedded defaults) — Built-in fallbacks

**Key structures:**
- `AgentEnv()` — Centralized function generating all environment variables for
  any role. Ensures consistent identity propagation across all spawn paths.
- Formula discovery and resolution order
- Hook configuration and management (see Claude Code Hooks below)

**Claude Code Hooks (`docs/HOOKS.md`):**
Gas Town manages Claude Code's `.claude/settings.json` files for each agent,
providing role-specific hook configuration:
- `hooks-base.json` — Shared base config for all agents
- Per-role/per-rig overrides (e.g., `crew.json`, `gastown__crew.json`)
- Merge strategy: base → role → rig+role
- Default hooks: `SessionStart→gt prime`, `UserPromptSubmit→gt mail check`,
  `Stop→gt costs record`
- Commands: `gt hooks sync`, `gt hooks diff`, `gt hooks override`

**Unique value:** Single source for all behavioral configuration. The
`AgentEnv()` function ensures that no matter how an agent is spawned (by daemon,
by sling, by manual start), it gets the same identity and configuration. The
property layers enable local customization without forking global defaults.

### Session Management — Agent Startup

**Built on:** tmux (sessions), Config (environment), Beads (agent beads)

**What it is:** The lifecycle management of agent tmux sessions — creation,
startup prompt injection, and beacon identification.

**Key components:**
- `SessionConfig` — Parameters for session creation
- `BeaconConfig` — Session identification (how the system finds this session)
- `StartSession()` — Creates tmux session, sets environment, injects startup prompt
- `BuildStartupPrompt()` — Constructs the GUPP-compliant initial prompt

**Startup flow:**
1. Create tmux session with unique name
2. Inject all environment variables via `AgentEnv()`
3. Send startup prompt that includes role instructions
4. Agent begins execution per GUPP (check hook → execute immediately)

**Unique value:** Standardized agent boot process that ensures every agent starts
with correct identity, context, and instructions. The startup prompt embeds the
propulsion principle — agents execute immediately upon finding work.

---

## Layer 2: Work Management

These concepts build on beads to provide work claiming, state tracking, and
multi-rig coordination.

### Hook System — Atomic Work Claiming

**Built on:** Beads (agent bead's `hook_bead` field)

**What it is:** The mechanism by which an agent claims exclusive ownership of a
piece of work. Each agent has an "agent bead" — a persistent bead tracking the
agent's identity and current assignment. The agent bead has a `hook_bead` field
that points to the issue bead currently assigned to that agent.

**How it works:**
1. `gt sling` assigns work by setting the agent bead's `hook_bead` to the issue ID
2. The issue's `assignee` is also set to the agent's identity
3. Agent checks `gt mol status` to find what's on its hook
4. When done, agent calls `gt done` which clears the hook

**Key properties:**
- **Atomic:** Setting hook_bead is a single Dolt write — no race conditions
- **Observable:** Any role can check any agent's hook via beads query
- **Authoritative:** The hook_bead field IS the assignment — not a copy, not a cache

**Unique value:** Provides a single, unambiguous answer to "what is this agent
working on?" No state files, no lock files, no coordination protocols. Just read
the bead.

### Labels-as-State — Fast State Queries

**Built on:** Beads (labels array), Events (state transitions)

**What it is:** A pattern where the current state of a bead is encoded in its
labels for fast querying, while the full event history captures all transitions.

**How it works:**
- Events capture **what happened** (historical, append-only)
- Labels cache **what's true now** (queryable, updatable)
- Example: When a merge fails with conflicts, the event logs the failure details
  and the label `needs-rebase` is applied to the bead for fast filtering

**Key labels in practice:**
- `gt:merge-request` — Bead is in the merge queue
- `needs-rebase` — Merge conflict detected
- `needs-fix` — Tests/build failed
- `needs-retry` — Transient failure (push failed)
- `patrol-muted` — Temporarily suppress patrol attention

**Unique value:** Enables O(1) state queries ("show me all beads needing rebase")
while maintaining complete history. Labels are the fast path; events are the audit trail.

### Beads Routing — Multi-Rig Prefix Dispatch

**Built on:** Beads (ID prefixes), Filesystem (`routes.jsonl`)

**What it is:** A system for directing bead operations to the correct rig's
database based on the bead ID prefix.

**How it works:**
- Every bead has a prefixed ID: `gt-abc`, `bd-123`, `hq-xyz`
- `routes.jsonl` maps prefixes to rig directories
- `beads.ExtractPrefix()` parses the prefix from an ID
- `beads.GetRigNameForPrefix()` resolves to the target rig

**Routing rules:**
- `hq-*` → Town-level beads (`.beads/` at town root)
- `gt-*` → Gastown rig beads
- `bd-*` → Beads rig beads
- Custom prefixes for custom rigs

**Unique value:** Enables a single `bd` command to work transparently across
multiple rigs. The user doesn't need to know which database holds a bead — the
prefix tells the system.

### Beads Redirects — Shared Database Access

**Built on:** Filesystem (`.beads/redirect` files)

**What it is:** A mechanism that lets worktrees (polecats, crew, refinery) share
the rig's canonical beads database without copying or syncing.

**How it works:**
- Each rig has ONE canonical beads database at `mayor/rig/.beads/`
- Worktrees contain a `.beads/redirect` file pointing to the canonical location
- `beads.ResolveRedirect()` follows the chain (max depth 3 for safety)
- All agents in a rig read/write the same database

**Unique value:** Zero sync overhead. All agents see the same state immediately.
No eventual consistency, no replication lag. Combined with Dolt's branch-per-polecat
for write isolation, this gives both consistency and concurrency.

### Dependencies — Bead Relationship Graph

**Built on:** Beads (dependency table in Dolt)

**What it is:** Typed relationships between beads that express blocking, tracking,
and parent-child relationships.

**Dependency types:**
- `blocks` — Bead A must close before Bead B can start
- `tracks` — Convoy A monitors Bead B (non-blocking observation)
- `parent` — Bead A is a step within Molecule B

**Key operations:**
- `bd dep add <from> <to> --type=<type>` — Create relationship
- `bd dep list <id> --direction=up/down --type=<type>` — Query graph
- `bd ready` — Find beads with no open blockers

**Unique value:** Enables work ordering (blocking), batch tracking (convoys),
and workflow decomposition (molecules) all through a single unified graph.

### Key Record Chronicle (KRC) — Ephemeral Data Lifecycle

**Built on:** Beads (operational data), Dolt (storage)

**What it is:** Lifecycle management for Level 0 ephemeral operational data.
KRC provides configurable TTLs and auto-pruning to prevent the operational data
plane from growing without bound (see `internal/krc/`).

**How it works:**
- Configurable TTLs per bead type (wisps, patrol records, etc.)
- Auto-pruning of expired operational noise
- "Decay" model — ephemeral data automatically ages out
- Important data promoted to Ledger (Level 2) before expiry

**Unique value:** Prevents data bloat in the operational plane. Without KRC,
patrol wisps and temporary state records would accumulate indefinitely, degrading
query performance and consuming storage.

### Session Checkpoints — Crash Recovery

**Built on:** Beads (agent state), Filesystem (checkpoint files)

**What it is:** Persistence of agent session state to enable crash recovery. When
an agent crashes mid-task, checkpoints allow the replacement session to resume
from the last known good state (see `internal/checkpoint/`).

**How it works:**
- Checkpoint file (`.polecat-checkpoint.json`) records current step, progress
- On crash detection (daemon/witness), new session loads checkpoint
- `gt prime` detects crash recovery vs normal startup vs post-handoff states
- Session states: `normal`, `post-handoff`, `crash-recovery`, `autonomous`

**Unique value:** Enables long-running workflows to survive agent crashes without
losing progress. Critical for the "Towers of Hanoi" demo proving arbitrary-length
sequential workflow durability.

---

## Layer 3: Messaging

Two complementary communication mechanisms built on different substrates.

### Mail — Async Persistent Messages

**Built on:** Beads (messages stored as beads with `type=message`), Dolt SQL

**What it is:** An asynchronous, persistent messaging system where messages are
beads. Every message is a permanent record with full attribution.

**Message structure:**
```
From:    gastown/witness       (BD_ACTOR of sender)
To:      gastown/refinery      (target role/agent)
Subject: MERGE_READY           (message type)
Body:    { JSON payload }      (structured data)
Channel: gastown               (rig scope)
```

**Key operations:**
- `gt mail send` — Send a message
- `gt mail inbox` — Check for messages
- `gt mail check` — Process pending messages (called via hooks)

**Unique value:** Messages are beads, so they inherit all bead properties:
attribution, persistence, queryability, audit trail. Messages survive agent
restarts, crashes, and even town restarts. Nothing is ever lost.

### Nudge — Sync Immediate Delivery

**Built on:** tmux (send-keys)

**What it is:** A synchronous, immediate signal delivery mechanism that uses
tmux's `send-keys` to inject text directly into an agent's terminal session.

**How it works:**
1. Sender calls `NudgeSession(sessionName, message)`
2. System uses `tmux send-keys -t <session> -l <message>` + Enter
3. The text appears as if typed into the agent's terminal
4. Agent's Claude instance processes it as input

**Key properties:**
- **500ms debounce** — Prevents rapid-fire nudges from overwhelming agent
- **Literal mode** — Uses `-l` flag to prevent key interpretation
- **Separate Enter** — Enter keystroke sent separately for reliability
- **Fire-and-forget** — No delivery confirmation (tmux guarantees delivery if session exists)

**Unique value:** Provides immediate wake-up capability. While mail requires the
agent to poll its inbox, nudge delivers instantly. Used when latency matters:
waking refinery for immediate merge processing, notifying polecats of merge
failures, triggering daemon patrol cycles.

### Protocol Messages — Typed Mail

**Built on:** Mail (transport), Protocol handlers (dispatch)

**What it is:** A layer on top of mail that defines typed, structured messages
with explicit handler registration. The protocol system distinguishes between
"not a protocol message" and "protocol message but no handler registered."

**Message types (from `internal/protocol/`):**

| Type | Sender | Receiver | Purpose |
|------|--------|----------|---------|
| `POLECAT_DONE` | Polecat | Witness | Work completed (exit types: COMPLETED, ESCALATED, DEFERRED, PHASE_COMPLETE) |
| `MERGE_READY` | Witness | Refinery | Branch verified, ready for merge queue |
| `MERGED` | Refinery | Witness | Merge succeeded |
| `MERGE_FAILED` | Refinery | Witness | Merge failed (with failure type) |
| `REWORK_REQUEST` | Refinery | Polecat | Branch needs fixing |
| `HELP` | Any | Escalation target | Agent needs assistance |
| `HANDOFF` | Agent | Next agent | Work transfer |
| `LIFECYCLE:Shutdown` | Daemon | Agents | Graceful shutdown request |

**Handler interfaces:**
```go
type WitnessHandler interface {
    HandleMerged(payload *MergedPayload) error
    HandleMergeFailed(payload *MergeFailedPayload) error
    HandleReworkRequest(payload *ReworkRequestPayload) error
}

type RefineryHandler interface {
    HandleMergeReady(payload *MergeReadyPayload) error
}
```

**How it flows (polecat completion example):**
1. Polecat finishes work → sends `POLECAT_DONE` mail to Witness
2. Witness `HandlePolecatDone()` verifies cleanup, sends `MERGE_READY` to Refinery
3. Refinery `HandleMergeReady()` queues branch, nudges refinery session
4. Refinery merges → sends `MERGED` to Witness
5. Witness `HandleMerged()` verifies commit on main, nukes polecat worktree

**Unique value:** Type-safe inter-role communication with explicit contracts.
Each role knows exactly what messages it sends and receives. The handler registry
pattern allows roles to register only for message types they care about.

### Beads-Native Messaging Primitives — Groups, Queues, Channels

**Built on:** Beads (Dolt tables), Mail

**What it is:** Three messaging primitives stored as beads that provide
different communication patterns (see `docs/beads-native-messaging.md`).
All use the `hq-` prefix as town-level entities spanning rigs.

**Groups (`gt:group`):**
Named collections of addresses for mail distribution. Send to a group →
message delivered to all members. Used for role-based addressing (e.g.,
"all witnesses").

**Queues (`gt:queue`):**
Work queues where messages can be claimed by workers. Unlike groups (fan-out),
queues provide exclusive claiming — each message is processed by exactly one
consumer. Used for load-balanced work distribution.

**Channels (`gt:channel`):**
Pub/sub broadcast streams with message retention. Each rig has a default
channel matching its name. Channels retain messages so late-subscribing agents
can catch up.

**Implementation:** All three types are implemented in `internal/beads/`
(`beads_group.go`, `beads_queue.go`, `beads_channel.go`) with full test
coverage.

**Unique value:** Three complementary communication patterns (fan-out,
exclusive-claim, pub/sub) all backed by the same beads substrate. No external
message brokers needed.

### Escalation System — Severity-Based Alerting

**Built on:** Mail (transport), Beads (escalation beads), Config (routing rules)

**What it is:** A unified system for routing high-severity alerts through
multiple channels (beads, email, SMS) based on config-driven rules (see
`docs/design/escalation-system.md`).

**Components:**
- `gt escalate` command — Single entry point for all escalation needs
- Escalation bead type — Persistent record of escalation
- `settings/escalation.json` — Config-driven routing by severity
- Stale escalation detection — Alerts for unacknowledged escalations

**Severity levels** determine routing:
- Low severity → Bead labels only
- Medium severity → Mail to overseer
- High severity → Mail + external notification (email/SMS)
- Critical → All channels + immediate attention required

**Unique value:** Config-driven escalation routing that ensures the right human
gets the right alert through the right channel. Prevents alert fatigue (low
severity stays in beads) while ensuring critical issues get immediate attention.

---

## Layer 4: Workflow Templates

Reusable patterns for structuring multi-step work.

### Formulas — Static Workflow Templates

**Built on:** Filesystem (TOML files at `.beads/formulas/`), Go embed

**What it is:** TOML templates that define multi-step workflows. Formulas are
static definitions — they don't execute until instantiated as molecules.

**Formula format:**
```toml
[formula]
name = "polecat-work"
type = "workflow"        # workflow | convoy | expansion | aspect
description = "Standard polecat work cycle"

[[formula.steps]]
name = "capture"
title = "Capture the issue"
instructions = "Read the issue and plan approach"

[[formula.steps]]
name = "execute"
title = "Execute the plan"
instructions = "Write code, run tests"

[[formula.steps]]
name = "deliver"
title = "Deliver the work"
instructions = "Create branch, push, signal done"
```

**Four formula types:**
1. **workflow** — Multi-step agent work sequence
2. **convoy** — Batch tracking template
3. **expansion** — Issue decomposition template
4. **aspect** — Cross-cutting concern template (e.g., testing)

**Key implementation details:**
- Formulas embedded in the `gt` binary via Go's `embed` package
- Canonical source at `.beads/formulas/` with hash-based health checking
- `InstalledRecord` tracks SHA256 checksums to detect user modifications
- `ProvisionFormulas()` installs without overwriting user customizations
- `CheckFormulaHealth()` reports: ok, outdated, modified, missing, new, untracked
- `UpdateFormulas()` safely updates only non-user-modified formulas

**Unique value:** Provides standardized, repeatable work patterns. Different
situations use different formulas. Formulas can be customized per-rig without
losing the ability to receive upstream updates.

### Molecules — Instantiated Workflows

**Built on:** Formulas (templates), Beads (bead-per-step)

**What it is:** A molecule is an instantiated formula — a root bead connected to
child step beads forming a directed acyclic graph. The name comes from Vonnegut's
*Cat's Cradle* — ice-nine crystallizing from a seed.

**Lifecycle: Formula → Protomolecule → Molecule**
1. **Formula** — Static TOML template (ice-nine seed)
2. **Protomolecule** — Formula bound to an issue but not yet crystallized
3. **Molecule** — Fully instantiated: root bead + step beads in Dolt

**How molecules work:**
- `gt mol bond <issue> [formula]` — Crystallize formula into molecule
- `gt mol status` — Show current step
- `gt mol next` / `gt mol step <name>` — Navigate between steps
- `bd close --continue` — Close current step, advance to next (propulsion)

**Step navigation enables GUPP:**
When an agent closes a step with `--continue`, the next step is automatically
opened and hooked. The agent doesn't stop to ask "what next?" — the molecule
tells it. This is how multi-step work maintains propulsion.

**Unique value:** Transforms flat issue lists into navigable multi-step workflows
with clear progression. Each step is a bead, so it has full attribution, timing,
and state tracking. Roll-up status (how far through the molecule) is automatic.

### Wisps — Ephemeral Molecules

**Built on:** Molecules (instantiation), Beads (ephemeral lifecycle)

**What it is:** A wisp is a molecule that is destroyed after execution. Used for
patrol cycles, plugin executions, and other recurring activities that shouldn't
accumulate permanent records.

**How wisps differ from molecules:**
- Molecules persist — they're the permanent record of work
- Wisps are deleted after the run completes
- Wisps are used for system-level recurring tasks (patrols, health checks)
- State is tracked on the ledger, not the operational plane

**Examples of wisp usage:**
- `mol-deacon-patrol` — Deacon's patrol cycle
- `mol-witness-patrol` — Witness's monitoring cycle
- Plugin execution molecules

**Unique value:** Prevents system-level housekeeping from cluttering the work
record. Patrol cycles happen every few minutes — without wisps, the beads
database would be overwhelmed with operational noise.

### Plugins — Gate-Controlled Automation

**Built on:** Formulas (execution template), Wisps (ephemeral execution),
Filesystem (plugin files)

**What it is:** Markdown files with TOML frontmatter that define automated
actions triggered by gate conditions. Plugins extend Gas Town behavior without
modifying core code.

**Plugin format:**
```markdown
---
[plugin]
name = "stale-branch-cleanup"
gate = "cron"
schedule = "0 */6 * * *"  # Every 6 hours

[plugin.config]
max_age_days = 7
---

# Stale Branch Cleanup

Find and remove branches older than {{max_age_days}} days...
```

**Gate types:**
| Gate | Trigger | Example |
|------|---------|---------|
| `cooldown` | Time since last run | "Run at most once per hour" |
| `cron` | Cron schedule | "Every day at 2am" |
| `condition` | Beads query result | "When >5 open merge requests" |
| `event` | Event bus signal | "When a polecat dies" |
| `manual` | Human trigger | "Run when I say so" |

**Plugin locations:**
- Town-level: `~/gt/.beads/plugins/`
- Rig-level: `~/gt/<rig>/.beads/plugins/`

**Execution model:**
- Plugins are dispatched as dogs (Deacon helpers)
- Each execution creates a wisp for tracking
- State persisted on ledger for cooldown enforcement

**Unique value:** Extensibility without code changes. Users can add custom
automation (stale branch cleanup, notification integrations, custom health
checks) by dropping markdown files into the plugins directory.

### Formula Resolution — Discovery and Precedence

**Built on:** Formulas (template files), Config (resolution order), Filesystem

**What it is:** The system that determines which formula version an agent gets
when multiple copies exist in different locations (see `docs/formula-resolution.md`).

**Resolution order (first match wins):**
1. Rig-level formulas (`.beads/formulas/` in rig directory)
2. Town-level formulas (`.beads/formulas/` at town root)
3. Embedded formulas (compiled into the `gt` binary)
4. [Future] Mol Mall (remote formula registry)

**Key design goals:**
- Predictable resolution with clear precedence rules
- Local customization without forking system defaults
- Project-specific formulas committed for collaborators
- Architecture ready for Mol Mall remote formula installation

**Unique value:** Enables per-rig customization of workflows while maintaining
a fallback chain to system defaults. A rig can override the polecat work formula
without affecting other rigs.

### [Design] Mol Mall — Formula Registry

**Built on:** Formulas, Federation (future)

**What it is:** A planned central registry/marketplace for sharing formulas
across towns. Enables ecosystem-level workflow distribution.

**Planned components:**
- Registry API for publishing/discovering formulas
- Capability tagging for formula matching
- Trusted publisher verification
- Federated discovery across towns

**Unique value:** Transforms formulas from local files into a shareable ecosystem.
Teams can publish proven workflows that other teams discover and install.

---

## Layer 5: Agent Lifecycle

The systems that create, manage, and destroy agents.

### Polecat Manager — Ephemeral Worker Lifecycle

**Built on:** Git/Worktrees (sandbox), tmux (session), Beads (agent bead),
Config (identity), Name Pool

**What it is:** The system that creates and destroys polecat workers. Polecats
are the workhorses of Gas Town — ephemeral agents spawned with specific work
and destroyed when done.

**Core operations:**
- `Manager.Add()` — Create git worktree from bare repo, allocate name
- `Manager.Remove()` — Kill tmux session, nuke worktree, clean beads
- `SessionManager.Start()` — Launch Claude agent with GUPP context
- `SessionManager.Stop()` — Graceful or force termination

**Polecat three-layer architecture:**
1. **Identity (permanent):** Agent bead, CV chain, work history — survives across sessions
2. **Sandbox (ephemeral per assignment):** Git worktree + branch — created per issue, destroyed on completion
3. **Session (ephemeral per step):** Claude instance + context window — may restart within an assignment

**Polecat states (NO idle state):**
- `Working` — Active tmux session, work on hook
- `Stalled` — Stuck or waiting, needs attention
- `Zombie` — Dead session but worktree still exists

**Key design choice:** There is NO idle state. Polecats are spawned WITH work
and destroyed WHEN done. The name pool manages reusable names, but the polecat
itself is ephemeral. This prevents resource waste and simplifies lifecycle
management.

**Self-cleaning model:**
1. Polecat finishes work → calls `gt done`
2. `gt done` pushes branch, signals `POLECAT_DONE` to Witness
3. Witness verifies cleanup, sends `MERGE_READY` to Refinery
4. After merge confirmation, Witness nukes the polecat worktree

**Unique value:** Fully automated worker lifecycle with no manual cleanup
required. Polecats are like serverless functions — spawn, execute, destroy.

### Name Pool — Reusable Identity Slots

**Built on:** Filesystem (themed name lists), Beads (agent beads)

**What it is:** A pool of themed names (mad-max, minerals, wasteland themes)
that are recycled across polecat incarnations.

**How it works:**
- Pool of ~50 themed names per rig
- `AllocateName()` picks an available name
- `Release()` returns name to the pool
- Overflow uses `rigname-N` format (numeric suffix)
- `InUse` is derived from filesystem — ZFC compliant

**Key distinction:** Names are slots, not polecats. The name "toast" might be
used by 100 different polecat incarnations over time. Each incarnation gets its
own agent bead, worktree, and session — but the name provides continuity for
human readability.

**Unique value:** Human-friendly naming without identity confusion. You can say
"toast is working on gt-abc" instead of "polecat-a7f3b2c1-d4e5 is working on
gt-abc."

### Worktree Management — Isolated Work Areas

**Built on:** Git (worktrees), Filesystem (directory conventions)

**What it is:** The system for creating, repairing, and cleaning up git worktrees
that serve as agent sandboxes.

**Key operations:**
- Create worktree from bare repo with unique branch
- `RepairWorktree()` — Recover from stale/corrupt state
- `CleanupStaleBranches()` — Remove orphaned branches
- Branch naming: `polecat/<name>/<issue>@<timestamp>`

**Cross-rig worktrees:**
When an agent from rig A needs to work in rig B:
```
~/gt/beads/crew/gastown-joe/    # joe from gastown working in beads rig
```
Identity is preserved: commits still attributed to `gastown/crew/joe`.

**Unique value:** Complete code isolation per agent. No merge conflicts between
concurrent workers because each has its own branch and worktree.

### Handoff and Seance — Session Continuity

**Built on:** tmux (session management), Mail (handoff messages), Beads
(agent bead state), Molecules (pinned workflow)

**Handoff** (`gt handoff`): When an agent's context window fills up or it gets
stuck, it can self-initiate a handoff. This:
1. Sends handoff mail with current state summary
2. Respawns the agent with a fresh Claude session
3. Pinned molecule survives across sessions — new session resumes at current step
4. `gt prime` detects post-handoff state and outputs "HANDOFF COMPLETE" warning

**Seance** (`gt seance`): Allows an agent to query the conversation history of
a predecessor session. When you need context from a previous incarnation of the
same polecat, seance retrieves it — even across account boundaries.

**Session state detection:**
- `gt prime --state` outputs current state: `normal`, `post-handoff`,
  `crash-recovery`, or `autonomous`
- Different startup behaviors per state (e.g., crash recovery loads checkpoint)

**Unique value:** Enables indefinitely long work sequences despite finite context
windows. An agent can work on a 1000-step molecule by handing off every ~50 steps,
with each new session resuming exactly where the previous one left off.

### Crew — Persistent Human-Managed Workers

**Built on:** Git/Worktrees (own clone), Beads (agent bead), Config (identity)

**What it is:** Long-lived agents with persistent clones, managed directly by
humans rather than the Witness. Crew members are the interactive counterpart to
polecats.

**Key differences from polecats:**
| Aspect | Crew | Polecat |
|--------|------|---------|
| **Lifecycle** | Persistent (user controls) | Transient (Witness controls) |
| **Monitoring** | None | Witness watches, nudges, recycles |
| **Work assignment** | Human-directed or self-assigned | Slung via `gt sling` |
| **Git state** | Pushes to main directly | Works on branch, Refinery merges |
| **Cleanup** | Manual | Automatic on completion |
| **Identity** | `<rig>/crew/<name>` | `<rig>/polecats/<name>` |

**When to use Crew:**
- Exploratory work, long-running projects, work requiring human judgment
- Tasks where you want direct control over the agent

**Unique value:** Provides a persistent workspace for interactive, human-directed
work. Crew members accumulate context across sessions and can be used for
exploratory or supervisory tasks that don't fit the "sling → execute → nuke"
polecat model.

### Dog Pool — Deacon Helpers

**Built on:** tmux (sessions), Deacon (owner), Beads (state tracking)

**What it is:** Dogs are NOT workers — they are the Deacon's helpers for
infrastructure tasks. The dog pool architecture (see `docs/design/dog-pool-architecture.md`)
distinguishes between two types:

**Helper Dogs:** Persistent Claude agents managed by the Deacon for recurring
infrastructure tasks (e.g., Boot for health triage).

**Dance Dogs:** Lightweight goroutines (not full Claude sessions) used for
shutdown dances — the safe termination of stuck agents. These use a deterministic
state machine with `Warrant` structs and state files at `~/gt/deacon/dogs/active/`.

**Key design distinction:** Dance Dogs are NOT AI agents — they're Go code that
follows a mechanical shutdown protocol. This prevents the "AI trying to shut
down another AI" problem.

**Unique value:** Separates infrastructure concerns (dogs) from project work
(polecats/crew). The shutdown dance state machine provides safe, deterministic
agent termination.

---

## Layer 6: Coordination

Higher-level patterns for dispatching, tracking, and landing work.

### Sling — Unified Work Dispatch

**Built on:** Formulas (workflow selection), Molecules (instantiation),
Hooks (assignment), Nudge (activation), Polecat Manager (spawning)

**What it is:** The single command that orchestrates the entire work dispatch
flow: select a formula, instantiate a molecule, hook it to an agent, and nudge
the agent to begin.

**The sling flow:**
```
gt sling <issue> <rig> [--formula=<name>]
    │
    ├── 1. Find or spawn an available polecat
    ├── 2. Select formula (explicit or auto-detect)
    ├── 3. Create molecule (bond formula to issue)
    ├── 4. Set hook_bead on agent bead
    ├── 5. Set assignee on issue bead
    ├── 6. Nudge agent's tmux session
    └── 7. Log sling event
```

**Key features:**
- Auto-spawns polecats if none available
- Auto-selects formula based on issue type
- Creates convoy automatically if multiple issues slung
- `--no-boot` flag for reactive feeding (skip daemon wake)

**Unique value:** Single entry point for all work dispatch. Encapsulates the
entire complexity of agent spawning, work claiming, and activation into one
command. This is the primary API surface for getting work done.

### Convoys — Batch Work Tracking

**Built on:** Beads (convoy bead + `tracks` dependency type), Convoy observer

**What it is:** A persistent tracking bead that groups related issues across
rigs and monitors their collective progress.

**Convoy lifecycle:**
1. `gt convoy create "Feature X" gt-abc gt-def` — Create convoy tracking issues
2. Issues are linked via `tracks` dependency (non-blocking)
3. Redundant observers check completion on each issue close
4. When all tracked issues close → convoy auto-closes ("lands")

**Redundant observation pattern:**
- Witness checks convoys when a polecat completes
- Refinery checks convoys when a merge succeeds
- Daemon checks convoys during zombie cleanup
- All checks are idempotent — safe to run multiple times

**Reactive feeding:**
When an issue in a convoy completes, the observer checks for the next ready
issue and dispatches it via `gt sling`. This makes convoy progression
**event-driven** instead of polling-based.

**Convoy vs Swarm:**
- **Convoy** = tracking container (which issues belong together)
- **Swarm** = the set of workers currently assigned to convoy issues

**Unique value:** Cross-rig work tracking with automatic completion detection
and reactive work feeding. You create a convoy, sling the first issue, and the
system automatically feeds subsequent issues as previous ones complete.

### Swarm — Multi-Agent Coordination

**Built on:** Beads (epic bead + task beads), Polecat Manager (workers),
Integration Branches (shared branch)

**What it is:** Coordinates multiple polecats working on related tasks that
must be merged together. Think of it as a convoy where the workers share an
integration branch.

**Core types:**
```go
type Swarm struct {
    ID          string       // Matches beads epic ID
    EpicID      string       // Beads epic tracking work
    BaseCommit  string       // Git SHA all workers branch from
    Integration string       // Shared integration branch name
    State       SwarmState   // Created, Active, Merging, Landed, Failed, Canceled
    Workers     []string     // Polecat names
    Tasks       []SwarmTask  // Individual work items
}

type SwarmTask struct {
    IssueID  string
    Title    string
    Assignee string
    Branch   string
    State    TaskState  // Pending, Assigned, InProgress, Review, Merged, Failed
}
```

**Key property:** ALL state is derived from beads — no in-memory cache. The
swarm manager is **stateless**, loading state fresh from beads on every query.

**Unique value:** Enables parallel multi-agent work on features that span
multiple files/concerns but must land atomically. The integration branch
collects all workers' contributions.

### Merge Queue / Refinery — Quality Gate

**Built on:** Beads (merge request beads), Git (merge operations), Mail
(status notifications), Worktrees (refinery workspace)

**What it is:** A per-rig automated merge processor that tests, builds, and
merges polecat branches into the target branch.

**Merge request lifecycle:**
```
Open → InProgress → Closed (merged | rejected | conflict | superseded)
         ↓
       Open (failure → reassign to worker)
```

**Priority scoring algorithm:**
- Base priority from issue
- Age penalty (exponential — older MRs get priority)
- Retry count penalty
- Convoy boost (issues in convoys are prioritized)

**Failure handling:**
| Failure Type | Label Applied | Action |
|---|---|---|
| `conflict` | `needs-rebase` | Assign back to worker |
| `tests_fail` | `needs-fix` | Assign back to worker |
| `build_fail` | `needs-fix` | Assign back to worker |
| `flaky_test` | `needs-fix` | May retry |
| `push_fail` | `needs-retry` | Retry automatically |
| `fetch_fail` | — | Infrastructure issue |

**Unique value:** Automated quality gate that prevents broken code from reaching
main. The priority scoring ensures critical/old work is processed first. Failure
categorization enables appropriate automated responses.

### Integration Branches — Epic-Scoped Landing

**Built on:** Git (branches), Formulas (branch naming templates), Refinery
(merge processing)

**What it is:** Shared branches where multiple agents' work accumulates before
being landed atomically on the target branch.

**How it works:**
1. Swarm creates integration branch from base commit
2. Each polecat works on its own sub-branch off the integration branch
3. Refinery merges sub-branches into the integration branch
4. When all tasks complete, Refinery lands the integration branch on main

**Safety guardrails (three layers):**
1. **Formula/role instructions** — Agents told which branch to target
2. **Pre-push hook** — Validates branch targeting before push
3. **Authorized code path** — Only refinery's merge path can push to main

**Build pipeline configuration:**
Integration branches use a configurable 5-command build pipeline for verification
before landing.

**Unique value:** Enables atomic landing of multi-agent features. Either all the
work lands together or none of it does. Prevents partial features from reaching
production.

---

## Layer 7: Orchestration

The highest-level systems that coordinate all lower layers.

### Witness — Per-Rig Polecat Lifecycle Manager

**Built on:** Polecat Manager (lifecycle), Mail (protocol messages), Beads
(agent beads), tmux (session monitoring), Convoy observer

**What it is:** A persistent per-rig agent that monitors all polecats in its
rig, handles completion signals, detects zombies, and coordinates with the
refinery.

**Key responsibilities:**
1. **Handle `POLECAT_DONE`** — Process completion signals
   - Exit types: COMPLETED, ESCALATED, DEFERRED, PHASE_COMPLETE
   - Auto-nuke if clean, create cleanup wisp if dirty
   - Send `MERGE_READY` to Refinery
2. **Handle `MERGED`** — Verify commit on main, nuke polecat
3. **Handle `MERGE_FAILED`** — Notify polecat of rejection
4. **Zombie detection** — Cross-reference agent beads with tmux sessions
   - Session-dead: tmux gone but worktree exists
   - Agent-dead: tmux alive but Claude process died
5. **Convoy notification** — Check convoys on issue completion

**Safety verification:**
- `AutoNukeIfClean()` checks `cleanup_status` from agent bead before deletion
- `verifyCommitOnMain()` checks ALL git remotes before nuke
- Only nukes after merge is verified on permanent branch

**Unique value:** Provides the safety net for polecat lifecycle. Without the
witness, dead polecats would leave orphaned worktrees and branches. The witness
ensures every polecat is properly cleaned up.

### Refinery — Per-Rig Merge Queue Processor

**Built on:** Merge Queue (MR processing), Mail (protocol messages), Git
(merge operations), Beads (MR tracking)

**What it is:** A persistent per-rig agent (Claude Engineer) that processes
the merge queue — testing, building, and merging polecat branches.

**How it works:**
1. Receives `MERGE_READY` from Witness
2. Queries beads for open MRs with `gt:merge-request` label
3. Scores and prioritizes MRs
4. For each MR: fetch branch → merge → test → build → push
5. Sends `MERGED` or `MERGE_FAILED` back to Witness

**Session:** `gt-<rig>-refinery` — single persistent tmux session per rig,
nudged awake when new work arrives.

**Unique value:** Automated, prioritized merge processing with quality
verification. Polecats just push branches — the refinery handles the rest.

### Deacon — Town-Level Patrol Coordinator

**Built on:** tmux (persistent session), Molecules/Wisps (patrol cycles),
Mail (inbox processing)

**What it is:** A persistent town-level agent that runs continuous patrol
cycles, processing mail and executing health checks.

**Patrol loop (runs forever):**
1. `gt deacon heartbeat` — Update heartbeat file
2. Check `gt hook` — Execute if work is hooked
3. Otherwise create/execute patrol wisp (`mol-deacon-patrol`)
4. `await-signal` for next cycle trigger
5. NEVER exits voluntarily

**Key behaviors:**
- Auto-respawn hook (PATCH-010) for resilience
- Heartbeat file at `deacon/heartbeat.txt` checked by daemon
- Coordinates witness and refinery spawning via patrol
- Runs periodic cleanup tasks

**Unique value:** Provides continuous, intelligent patrol of the entire town.
While the daemon does basic health checks, the deacon (as a Claude agent) can
make nuanced decisions about system state.

### Boot — Intelligent Triage Agent

**Built on:** Deacon (target of triage), Beads (state queries)

**What it is:** A fresh-context Claude agent spawned by the daemon on each
heartbeat cycle to provide intelligent triage of the Deacon's state.

**Boot decision matrix:**
| Deacon State | Boot Action |
|---|---|
| Running, responsive | No action needed |
| Running, stuck | Nudge deacon session |
| Running, truly stuck | Interrupt and restart |
| Dead | Restart deacon |

**Key design:** Boot is spawned FRESH each tick — no accumulated context, no
bias from previous assessments. This prevents the triage agent from developing
blind spots.

**Unique value:** Intelligent health assessment without accumulated state
blindness. Each boot invocation sees the system with fresh eyes.

### Daemon — Background Recovery Safety Net

**Built on:** All lower layers (monitors everything)

**What it is:** A Go process (not a Claude agent) that runs in the background
providing the ultimate safety net for the entire town.

**Heartbeat cycle (every 3 minutes):**
1. Ensure Dolt server running
2. Ensure Deacon running (restart if dead)
3. Boot triage (intelligent stuck detection)
4. Direct Deacon heartbeat check (belt-and-suspenders)
5. Ensure Witnesses running (per-rig)
6. Ensure Refineries running (per-rig)
7. Ensure Mayor running
8. Trigger pending spawns (bootstrap nudges)
9. Process lifecycle requests
10. Check GUPP violations
11. Check orphaned work
12. Polecat session health checks
13. Cleanup orphaned processes
14. Cleanup town service beads
15. Prune stale branches

**Key subsystems:**
- **Feed curator** — Activity-based wake from `bd activity --follow`
- **Convoy watcher** — Event-driven convoy completion
- **Dolt server** — Manages SQL server lifecycle with 30s health checks
- **Shutdown coordination** — Uses flock to prevent daemon fighting shutdown

**Proactive crash detection:**
- Detects: work-on-hook + dead tmux session
- Auto-restarts crashed polecats
- Tracks mass death events for anomaly detection

**Unique value:** The daemon is the last line of defense. It's not a Claude
agent — it's a Go binary that can't get confused, can't run out of context,
and can't be manipulated. If everything else fails, the daemon will detect the
failure and restart the failed component.

### Watchdog Chain — Three-Tier Monitoring

**Built on:** Daemon (base process), Boot (triage), Deacon (patrol), Witnesses

**What it is:** A layered monitoring architecture where each tier watches the
one below it.

```
Daemon (Go process, 3-min heartbeat)
  └── Boot (fresh Claude agent, per-tick triage)
       └── Deacon (persistent Claude agent, continuous patrol)
            └── Witnesses (per-rig Claude agents, polecat monitoring)
```

**Why three tiers?**
- **Daemon** — Can't get confused (Go code), but can't make nuanced decisions
- **Boot** — Fresh perspective each time, but expensive to spawn
- **Deacon** — Deep context, but might get stuck
- **Witnesses** — Rig-specific knowledge, but only see their own rig

**Fallback chain:**
If the deacon is dead → daemon restarts it.
If witnesses are dead → daemon restarts them.
If polecats are dead → daemon detects and restarts them directly.

**Unique value:** Defense in depth. No single point of failure in the monitoring
chain. Each tier covers the failure modes of the tier below it.

### Mayor — Global Coordinator

**Built on:** Beads (town-level beads), Config (town.json)

**What it is:** The town-level coordinator responsible for cross-rig
orchestration and global state management.

**Responsibilities:**
- Canonical clone for each rig (`mayor/rig/` directory)
- Town-level beads management (`hq-*` prefix)
- Cross-rig coordination decisions
- Global configuration

**Unique value:** Provides the town-wide view that rig-level roles (witness,
refinery) can't see. Decisions that affect multiple rigs are the mayor's domain.

### [Design] Federation — Multi-Town Coordination

**Built on:** Beads (cross-workspace references), Identity (cross-town
attribution), Config (remote workspace registration)

**What it is:** Multi-workspace coordination using the Highway Operations
Protocol (HOP). Enables distributed engineering across multiple Gas Town
instances (see `docs/design/federation.md`).

**Key components:**
- **Entity Model** (Level 1-3) — Progressive identity detail
- **URI Scheme** — `hop://` and `beads://` for cross-workspace addressing
- **Cross-Workspace Identity** — Attribution preserved across town boundaries
- **Dolt Remotes** — `origin` and `local` remote configurations

**How it works:**
```bash
# Register remote workspace
gt remote add partner hop://partner.com/their-project

# Query across workspaces
bd list --remote=partner --tag=integration
```

**Unique value:** Enables enterprise-scale coordination spanning multiple
repositories, teams, and organizations. Visibility is unified despite separate
repos.

### Ledger Export — Data Plane Transitions

**Built on:** Beads (all three data planes), Events (triggers), Dolt (storage)

**What it is:** The mechanism for transitioning data between Operational (Level
0/1) and Ledger (Level 2/3) planes (see `docs/design/ledger-export-triggers.md`).

**Export triggers:**
| Trigger | Level | Description |
|---------|-------|-------------|
| Bead Closure | L2 | Compressed completion record |
| Convoy Completion | L2 | Batch work summary |
| Refinery Merge | L2 | Merge outcome record |
| Milestone/Sprint Boundary | L2 | Periodic snapshot |
| Design Decision | L3 | Full-fidelity reasoning capture |
| Novel Problem Resolution | L3 | Learnable problem-solving record |
| Cross-Rig Coordination | L3 | Cross-project interaction record |
| Explicit Full-Fidelity Flag | L3 | Manual "this matters" signal |

**Fidelity levels:**
- Level 0: Ephemeral operational noise (KRC manages lifecycle)
- Level 1: Operational state (live queries, fast access)
- Level 2: Compressed completion records (audit trail)
- Level 3: Full-fidelity ground truth (skill derivation, CV building)

**How CV/skills are derived:**
Work evidence flows from operational data through export triggers to the
Ledger. Skills are derived by querying the Ledger:
- `.go` files touched → Go skill
- Issue tags → domain skills
- Commit patterns → activity types

**Unique value:** Separates hot operational data from permanent audit records.
Enables long-term learning (skill derivation, CV) without operational data
bloat.

### gt doctor — Health Diagnostics

**Built on:** All subsystems (checks each one)

**What it is:** A diagnostic command that checks the health of a Gas Town
installation and can auto-repair common issues.

**Usage:**
```bash
gt doctor           # Check health
gt doctor --fix     # Auto-repair common issues
gt doctor --verbose # Detailed diagnostics
```

**Unique value:** Single entry point for diagnosing and fixing installation
problems. Essential for operational support.

---

## Cross-Cutting Principles

These principles are not layers — they are design philosophies that permeate
every layer of the system.

### GUPP — Gas Town Universal Propulsion Principle

> **If you find work on your hook, YOU RUN IT.**

**What it means:** Agents execute immediately upon finding work. No
announcements, no waiting for confirmation, no asking permission. The hook
having work IS the assignment.

**Why it exists:** Gas Town is a steam engine. Agents are pistons. If a piston
stops to ask "should I push?" the engine stalls. Humans may be AFK for hours —
the system must run autonomously.

**How it manifests:**
- Startup protocol: Check hook → execute immediately
- Molecule navigation: `bd close --continue` auto-advances to next step
- Sling: Sets hook + nudges in one atomic operation
- No idle state for polecats — spawned with work, destroyed when done

**Failure mode prevented:** Agent starts → announces itself → waits for human →
human is AFK → work sits idle → system stalls.

### MEOW — Molecular Expression of Work

**What it means:** Large goals decompose into trackable atomic units (beads)
organized into structured workflows (molecules).

**How it manifests:**
- Epics → features → tasks → steps (recursive decomposition)
- Each level is a bead with its own status, assignee, and history
- Roll-up status is automatic from child bead states
- Cross-project references via bead dependencies

### NDI — Nondeterministic Idempotence

**What it means:** Individual operations may produce different results each
time, but the system converges to useful outcomes through redundant observation
and idempotent checks.

**How it manifests:**
- Convoy observers: Witness + Refinery + Daemon all check completion
- All checks are idempotent — running them multiple times is safe
- Zombie detection by both Daemon and Witness
- Daemon heartbeat: belt-and-suspenders redundancy

### ZFC — Zero-File Compliance

**What it means:** State is derived from observable sources, never stored in
standalone files.

**Observable sources:**
- tmux sessions → running state
- Beads (Dolt) → work assignments, issue status
- Filesystem → polecat directories exist or don't
- Git → branches, commits, worktrees

**How it manifests:**
- `loadFromBeads()` derives polecat state from agent bead + tmux + filesystem
- Name pool `InUse` is derived from filesystem, never persisted
- Swarm manager loads ALL state from beads on every query
- Merge queue queries beads for open MRs, never caches locally

**Why it matters:** No stale state, no sync bugs, no recovery needed after
crashes. If you want to know what's happening, query the sources of truth.

---

## Gas City Implications

Based on this analysis, here is how Gas Town concepts map to Gas City's
abstraction layers.

### What MUST be core (intrinsic to the SDK):

| Concept | Why it's core | Gas City equivalent |
|---------|---------------|---------------------|
| Agent Protocol | Everything depends on starting/stopping/signaling agents | §2 Agent Protocol |
| Task Store (Beads) | Universal state substrate — all higher concepts build on it | §10 Task System |
| Agent Identity | Attribution permeates every operation | §18 Identity/Addressing |
| Event Bus | Cross-component communication substrate | §8 Event Bus |
| Hook System | Atomic work claiming is foundational to GUPP | Part of Task System |
| Dependencies | Blocking/tracking relationships enable coordination | Part of Task System |

### What SHOULD be configurable (built on core):

| Concept | Built from | Gas City equivalent |
|---------|-----------|---------------------|
| Messaging (Mail + Nudge + Groups/Queues) | Task Store + Agent Protocol signals | §15 Messaging |
| Formulas/Molecules | Task Store + Config | §11-12 Formulas/Molecules |
| Convoys | Task Store + Dependencies + Config | §13 Convoys |
| Plugins | Formulas + Event Bus + Config | §16 Plugin System |
| Sling (Dispatch) | All of the above | §14 Sling |
| Merge Queue | Task Store + Agent Protocol + Config | Via Formulas |
| Health Monitoring | Event Bus + Agent Protocol | §9 Health Monitoring |
| Escalation | Messaging + Config | §9 or §15 (alerting channel) |
| Handoff/Seance | Agent Protocol + Task Store | §2 Agent Protocol (Resume) |
| Data Lifecycle (KRC) | Task Store | §10 Task System (TTL/pruning) |
| Ledger Export | Task Store + Event Bus | §10 Task System (archival) |
| Federation | Task Store + Identity + Config | Future: Network/Discovery layer |
| Formula Registry (Mol Mall) | Formulas + Config | §11 (registry capability) |

### What is deployment-specific (not in SDK):

| Concept | Why it's deployment-specific |
|---------|----------------------------|
| tmux sessions | One possible execution substrate |
| Dolt SQL | One possible storage backend |
| Git worktrees | One possible workspace strategy |
| Directory conventions | One possible filesystem layout |
| Specific role names (witness, refinery) | Configurable via role system |
| Watchdog chain topology | One possible monitoring strategy |
| Property layer filesystem layout | One possible config resolution strategy |
| Claude Code hooks integration | One possible agent tool integration |

### The Substrate Layering Principle (from Gas City spec §1.1):

```
Layer 0: Infrastructure    → Agent Protocol + Task Store + Event Bus
Layer 1: Task/Agents       → Task lifecycle + agent registry + hooks
Layer 2: Messaging         → Mail + signals (replaces nudge abstraction)
Layer 3: Workflow Engine   → Formulas + Molecules + Plugins
Layer 4: Dispatch/Coord    → Sling + Convoys + health monitoring
Layer 5: Controller        → Workspace controller (replaces daemon)
```

**Five invariants that must hold:**
1. No layer may import from a higher layer
2. Each layer exposes a clean public API consumed by the layer above
3. Layer 0 has zero Gas City-specific logic
4. Removing any layer above 0 leaves the layers below fully functional
5. Side effects (I/O, process spawning) are confined to Layer 0 implementations

### Key insight for Gas City:

Gas Town's power comes from the **beads substrate** — everything is a bead,
everything has attribution, everything is queryable. Gas City must preserve
this property through its Task Store abstraction. The specific implementation
(Dolt, SQLite, Postgres, file-based) is irrelevant — what matters is that ALL
domain state lives in ONE place with consistent identity, dependencies, and
event history.

The second critical insight is the **messaging duality**: async persistent
messages (mail) + sync immediate signals (nudge). Gas City abstracts these as
the Messaging system (§15) with different delivery modes, decoupled from the
specific transport (tmux send-keys, WebSocket, pipe, etc.).

The third insight is **progressive capability**: Gas Town hardcodes all
capabilities. Gas City makes each capability layer opt-in via configuration.
A minimal Gas City deployment needs only Layer 0 (run agents, store tasks).
Adding messaging, workflows, dispatch, and orchestration are all additive
capabilities configured on top of the core.

The fourth insight is **data lifecycle management**: Gas Town's three data
planes (Operational, Ledger, Design) with KRC pruning and export triggers
show that a production orchestration system must actively manage data
lifecycle. Gas City should define data lifecycle policies in the Task Store
abstraction — TTLs for ephemeral data, export triggers for permanent records,
and fidelity levels for determining how much detail to preserve.

The fifth insight is **session continuity**: Handoff and Seance demonstrate
that agent sessions are ephemeral but work is not. Gas City must ensure that
agent protocol implementations support clean session cycling with state
preservation, enabling indefinitely long workflows despite finite context
windows.

---

## Appendix: Document Inventory

This analysis drew from the following sources:

**Documentation (docs/):**
- `overview.md`, `glossary.md`, `reference.md`, `why-these-features.md`
- `HOOKS.md`, `INSTALLING.md`, `beads-native-messaging.md`, `formula-resolution.md`
- `concepts/`: convoy, molecules, polecat-lifecycle, propulsion-principle, identity, integration-branches
- `design/`: architecture, mail-protocol, watchdog-chain, plugin-system, dolt-storage, operational-state, escalation-system, escalation, federation, dog-pool-architecture, ledger-export-triggers, hooks-registry-design, property-layers, convoy-lifecycle, at-spike-report, witness-at-team-lead

**Specifications (specs/):**
- `gas-city-spec.md` (v0.11.0, 2473 lines)
- `gas-city-vision.md`

**Source code (internal/):**
- `beads/`, `tmux/`, `session/`, `events/`, `config/`
- `polecat/`, `witness/`, `deacon/`, `daemon/`, `convoy/`, `refinery/`
- `formula/`, `mail/`, `protocol/`, `swarm/`, `dog/`
- `checkpoint/`, `krc/`, `crew/`, `doctor/`, `connection/`
