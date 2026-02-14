# Gas City SDK — Research Notes

> **Pipeline:** Spec-Forge Stage 1 (Research)
> **Date:** 2026-02-14
> **Source repo:** https://github.com/steveyegge/gastown
> **Codebase:** Go 1.24.2, ~396 CLI subcommands, 50+ internal packages, 235+ test files

---

## 1. Project Overview

Gas Town is a **multi-agent orchestration system** for coordinating multiple AI coding agents (Claude Code, Codex, Gemini, OpenCode, Cursor, Amp) with persistent work tracking via git-backed beads. The **Gas City SDK** is the next evolution: an **orchestration-builder toolkit** that extracts Gas Town's hardcoded 7-role architecture into a configurable, progressive-capability SDK.

### 1.1 Current Architecture (Gas Town)

- **7 hardcoded roles**: Mayor, Deacon, Witness, Refinery, Polecat, Crew, Dog (+ Boot sub-role)
- **tmux-coupled**: All agent sessions run in tmux panes; I/O goes through `tmux send-keys` / `capture-pane`
- **Beads-backed**: Work items stored as structured data in Dolt SQL Server (git-compatible version-controlled DB)
- **Two-level beads**: Town-level (`hq-*` prefix) for coordination + Rig-level (project prefix) for implementation
- **Directory-based identity**: Agent role detected from CWD path (`detectRole()` in `internal/cmd/role.go`)

### 1.2 Gas City Vision (from `specs/gas-city-vision.md`)

Six requirements from Steve Yegge's blog post:
1. Orchestration-builder toolkit (not just an orchestrator)
2. Kit for building different "town shapes" / topologies
3. Create your own roles, teams, coordination rules, worker instructions
4. Roles expressed in dynamic format external to code
5. Supports wiring in sandboxes, plugins, and hooks
6. Full configurability surface

### 1.3 Existing Draft Spec (`specs/gas-city-spec-initial-draft.md`)

A 1150-line draft spec already exists covering:
- Agent Runtime abstraction (the foundational design principle)
- Terminology mapping (Gas Town → SDK terms)
- Progressive capability model (Levels 0-7)
- Configuration schema (`gas-city.toml`)
- Three milestone shapes: Ralph, Agent Teams, Gas Town
- Custom roles and topologies
- Migration strategy (5 phases)
- File-by-file migration map
- Testing strategy

---

## 2. Commands

Gas Town exposes two primary CLIs: `gt` (orchestration) and `bd` (beads data operations). The `gt` binary has ~396 AddCommand registrations across `internal/cmd/`, built with Cobra. Below are the key commands organized by domain.

- Code: `internal/cmd/*.go` — All cobra command definitions
- Code: `cmd/gt/main.go` — Entry point calling `cmd.Execute()`

### 2.1 Primary CLI: `gt`

Built with Cobra. ~396 AddCommand registrations across `internal/cmd/`. Binary name: `gt`.

#### Workspace Management
| Command | Description | Source |
|---------|-------------|--------|
| `gt install <path>` | Initialize workspace | `internal/cmd/install.go` |
| `gt install --git` | With git init | `internal/cmd/install.go` |
| `gt doctor` | Health check (20+ checks) | `internal/cmd/doctor.go` |
| `gt doctor --fix` | Auto-repair issues | `internal/cmd/doctor.go` |

#### Rig Management
| Command | Description | Source |
|---------|-------------|--------|
| `gt rig add <name> <url>` | Add project | `internal/cmd/rig.go` |
| `gt rig list` | List projects | `internal/cmd/rig.go` |
| `gt rig remove <name>` | Remove project | `internal/cmd/rig.go` |
| `gt rig stop <name>` | Stop rig agents | `internal/cmd/rig_stop.go` |

#### Agent Operations
| Command | Description | Source |
|---------|-------------|--------|
| `gt agents` | List active agents | `internal/cmd/agents.go` |
| `gt sling <bead-id> <rig>` | Assign work to agent | `internal/cmd/sling.go` |
| `gt sling <bead-id> <rig> --agent <alias>` | Override runtime for this sling | `internal/cmd/sling.go` |
| `gt mayor attach` | Start Mayor session | `internal/cmd/mayor.go` |
| `gt mayor start --agent <alias>` | Start Mayor with specific agent | `internal/cmd/mayor.go` |
| `gt prime` | Context recovery | `internal/cmd/prime.go` |
| `gt peek <agent>` | Check agent health | `internal/cmd/peek.go` |
| `gt nudge <agent> "msg"` | Send message to agent | `internal/cmd/nudge.go` |

#### Convoy (Work Tracking)
| Command | Description | Source |
|---------|-------------|--------|
| `gt convoy create <name> [issues...]` | Create convoy | `internal/cmd/convoy.go` |
| `gt convoy list [--all] [--status=X]` | List convoys | `internal/cmd/convoy.go` |
| `gt convoy show [id]` | Show convoy details | `internal/cmd/convoy.go` |
| `gt convoy add <convoy-id> <issue-id...>` | Add issues to convoy | `internal/cmd/convoy.go` |
| `gt convoy refresh <id>` | Force refresh | `internal/cmd/convoy.go` |

#### Mail
| Command | Description | Source |
|---------|-------------|--------|
| `gt mail inbox [--all]` | Check messages | `internal/cmd/mail.go` |
| `gt mail read <id>` | Read message (supports numeric index) | `internal/cmd/mail.go` |
| `gt mail send <addr> -s "Subj" -m "Body"` | Send message | `internal/cmd/mail.go` |
| `gt mail send --human -s "..."` | Send to overseer | `internal/cmd/mail.go` |
| `gt mail delete <id...>` | Delete messages | `internal/cmd/mail.go` |
| `gt mail reply <id> "msg"` | Reply to message | `internal/cmd/mail.go` |
| `gt mail check --inject` | Check and inject mail into session | `internal/cmd/mail.go` |

#### Session Management
| Command | Description | Source |
|---------|-------------|--------|
| `gt handoff` | Request session cycle | `internal/cmd/handoff.go` |
| `gt handoff --shutdown` | Terminate (polecats) | `internal/cmd/handoff.go` |
| `gt done` | Signal work completion (self-cleaning polecat) | `internal/cmd/done.go` |
| `gt session stop <rig>/<agent>` | Stop specific session | `internal/cmd/session.go` |
| `gt seance` | List discoverable predecessor sessions | `internal/cmd/seance.go` |
| `gt seance --talk <id>` | Talk to predecessor | `internal/cmd/seance.go` |

#### Merge Queue
| Command | Description | Source |
|---------|-------------|--------|
| `gt mq list [rig]` | Show merge queue | `internal/cmd/mq.go` |
| `gt mq next [rig]` | Show highest-priority MR | `internal/cmd/mq.go` |
| `gt mq submit` | Submit branch to MQ | `internal/cmd/mq.go` |
| `gt mq status <id>` | MR status | `internal/cmd/mq.go` |
| `gt mq retry <id>` | Retry failed MR | `internal/cmd/mq.go` |
| `gt mq reject <id>` | Reject MR | `internal/cmd/mq.go` |
| `gt mq integration create <epic-id>` | Create integration branch | `internal/cmd/mq.go` |
| `gt mq integration land <epic-id>` | Land integration branch | `internal/cmd/mq.go` |

#### Hooks Management
| Command | Description | Source |
|---------|-------------|--------|
| `gt hooks sync [--dry-run]` | Regenerate Claude settings | `internal/cmd/hooks.go` |
| `gt hooks diff` | Show what sync would change | `internal/cmd/hooks.go` |
| `gt hooks base [--show]` | Edit/view base config | `internal/cmd/hooks.go` |
| `gt hooks override <target>` | Edit role overrides | `internal/cmd/hooks.go` |
| `gt hooks list [--json]` | List managed settings | `internal/cmd/hooks.go` |
| `gt hooks scan [--verbose]` | Scan for existing hooks | `internal/cmd/hooks.go` |
| `gt hooks init [--dry-run]` | Bootstrap from existing | `internal/cmd/hooks.go` |
| `gt hooks registry` | Browse hook registry | `internal/cmd/hooks.go` |
| `gt hooks install <id>` | Install hook from registry | `internal/cmd/hooks.go` |

#### Configuration
| Command | Description | Source |
|---------|-------------|--------|
| `gt config agent list [--json]` | List all agents | `internal/cmd/config.go` |
| `gt config agent get <name>` | Show agent config | `internal/cmd/config.go` |
| `gt config agent set <name> <cmd>` | Create/update agent | `internal/cmd/config.go` |
| `gt config agent remove <name>` | Remove agent | `internal/cmd/config.go` |
| `gt config default-agent [name]` | Get/set default agent | `internal/cmd/config.go` |
| `gt config show` | View all config | `internal/cmd/config.go` |

#### Escalation
| Command | Description | Source |
|---------|-------------|--------|
| `gt escalate "topic"` | Escalate issue | `internal/cmd/escalate.go` |
| `gt escalate -s CRITICAL "msg"` | With severity | `internal/cmd/escalate.go` |
| `gt escalate ack <bead-id>` | Acknowledge escalation | `internal/cmd/escalate.go` |

#### Molecule/Formula
| Command | Description | Source |
|---------|-------------|--------|
| `gt mol status` | Check hooked work | `internal/cmd/mol.go` |
| `gt mol current` | What to work on next | `internal/cmd/mol.go` |
| `gt mol step done <step>` | Complete molecule step | `internal/cmd/mol.go` |
| `gt mol squash` | Squash attached molecule | `internal/cmd/mol.go` |
| `gt mol burn` | Burn attached molecule | `internal/cmd/mol.go` |
| `gt formula run <name>` | Run convoy formula | `internal/cmd/formula.go` |
| `gt formula show <name>` | Show formula details | `internal/cmd/formula.go` |

#### Other Commands
| Command | Description | Source |
|---------|-------------|--------|
| `gt start [--agent <alias>]` | Start all town agents | `internal/cmd/start.go` |
| `gt stop --all` | Kill all sessions | `internal/cmd/stop.go` |
| `gt dashboard --port 8080` | Start web dashboard | `internal/cmd/dashboard.go` |
| `gt crew add <name> --rig <rig>` | Create crew workspace | `internal/cmd/crew.go` |
| `gt crew at <name> [--agent <alias>]` | Enter crew workspace | `internal/cmd/crew.go` |
| `gt orphans kill` | Clean up orphaned processes | `internal/cmd/orphans.go` |
| `gt deacon health-check <agent>` | Send health ping | `internal/cmd/deacon.go` |
| `gt deacon health-state` | Show health state | `internal/cmd/deacon.go` |
| `gt polecat nuke <name>` | Clean up polecat | `internal/cmd/polecat.go` |
| `gt worktree <rig>` | Create cross-rig worktree | `internal/cmd/worktree.go` |
| `gt costs record` | Record session costs | `internal/cmd/costs.go` |
| `gt stale [--json] [--quiet]` | Check binary staleness | `internal/cmd/stale.go` |
| `gt completion <shell>` | Shell completions | `internal/cmd/completion.go` |
| `gt version` | Show version info | `internal/cmd/version.go` |

### 2.2 Beads CLI: `bd`

Separate binary from [steveyegge/beads](https://github.com/steveyegge/beads). Git-backed issue tracking.

| Command | Description |
|---------|-------------|
| `bd ready` | Work with no blockers |
| `bd list --status=X` | List by status |
| `bd show <id>` | Show bead details |
| `bd create --title="..." --type=X` | Create bead |
| `bd update <id> --status=X` | Update status |
| `bd close <id>` | Close bead |
| `bd dep add <child> <parent>` | Add dependency |
| `bd formula list` | List formulas |
| `bd cook <formula>` | Compile formula to proto |
| `bd mol pour <proto>` | Create persistent molecule |
| `bd mol wisp <proto>` | Create ephemeral molecule |
| `bd mol list` | List active molecules |
| `bd activity --follow` | Real-time event feed |
| `bd audit --actor=X` | Attribution audit |
| `bd stats --actor=X` | Agent statistics |

---

## 3. Config

Gas Town uses 15+ JSON config files spread across town, rig, and agent directories. Configuration is layered: town defaults → rig overrides → agent-specific. The hooks system adds another layer for Claude Code settings.

- Code: `internal/config/types.go` — All config struct definitions
- Code: `internal/config/loader.go` — Config loading and path resolution

### 3.1 File Locations and Schemas

| File | Location | Schema | Source |
|------|----------|--------|--------|
| Town identity | `mayor/town.json` | `TownConfig` | `internal/config/types.go:14` |
| Mayor config | `mayor/config.json` | `MayorConfig` | `internal/config/types.go:24` |
| Town settings | `settings/config.json` | `TownSettings` | `internal/config/types.go:38` |
| Rig identity | `<rig>/config.json` | `RigConfig` | `internal/config/types.go` |
| Rig settings | `<rig>/settings/config.json` | `RigSettings` | `internal/config/types.go` |
| Agent registry (town) | `settings/agents.json` | `AgentRegistry` | `internal/config/agents.go` |
| Agent registry (rig) | `<rig>/settings/agents.json` | `AgentRegistry` | `internal/config/agents.go` |
| Rig registry | `mayor/rigs.json` | `RigsConfig` | `internal/config/types.go` |
| Daemon patrol | `mayor/daemon.json` | `DaemonPatrolConfig` | `internal/config/types.go:190` |
| Messaging config | `config/messaging.json` | `MessagingConfig` | `internal/config/types.go` |
| Escalation config | `settings/escalation.json` | `EscalationConfig` | `internal/config/types.go` |
| Accounts config | `mayor/accounts.json` | `AccountsConfig` | `internal/config/types.go` |
| Overseer config | (town root) | `OverseerConfig` | `internal/config/overseer.go` |
| Hooks base | `~/.gt/hooks-base.json` | Claude settings JSON | `internal/cmd/hooks.go` |
| Hooks overrides | `~/.gt/hooks-overrides/*.json` | Claude settings JSON | `internal/cmd/hooks.go` |
| Dolt state | `daemon/dolt-state.json` | Process state (pid, port) | `internal/doltserver/` |
| Beads routes | `.beads/routes.jsonl` | JSONL prefix routing | `internal/beads/` |
| Beads redirect | `.beads/redirect` | Path to canonical beads | `internal/beads/` |

### 3.2 Runtime Configuration

```go
// RuntimeConfig — from internal/config/types.go
type RuntimeConfig struct {
    Command        string            `json:"command"`
    Args           []string          `json:"args,omitempty"`
    Env            map[string]string `json:"env,omitempty"`
    ResumeFlag     string            `json:"resume_flag,omitempty"`     // e.g., "--resume"
    ResumeStyle    string            `json:"resume_style,omitempty"`    // "flag" or "subcommand"
    NonInteractive *NonInteractiveConfig `json:"non_interactive,omitempty"`
}
```

**Built-in agent presets**: `claude`, `gemini`, `codex`, `cursor`, `auggie`, `amp`, `opencode`

**Resolution order**: rig-level agents → town-level agents → built-in presets

### 3.3 Role Definitions (Hardcoded)

From `internal/config/roles.go`:
- `AllRoles()` returns `["mayor", "deacon", "dog", "witness", "refinery", "polecat", "crew"]`
- Embedded TOML files in `internal/config/roles/*.toml` define per-role:
  - Session patterns
  - Health check config
  - Environment variables
  - Nudge prompts

### 3.4 Role Templates (Go Templates)

From `internal/templates/roles/`:
- `mayor.md.tmpl`, `polecat.md.tmpl`, `witness.md.tmpl`, `refinery.md.tmpl`
- `deacon.md.tmpl`, `boot.md.tmpl`, `dog.md.tmpl`, `crew.md.tmpl`
- Injected via `gt prime` at SessionStart

---

## 4. Env Vars

Gas Town injects 40+ environment variables into agent tmux sessions via `config.AgentEnv()`. Variables control identity, role detection, beads routing, and debug behavior. Some are set per-session; others are global.

- Code: `internal/config/env.go` — `AgentEnv()` function and all variable definitions
- Code: `internal/cmd/role.go` — Variables used for role detection

### 4.1 Core Agent Variables (set via `config.AgentEnv()`)

| Variable | Purpose | Set For |
|----------|---------|---------|
| `GT_ROLE` | Agent role type | All agents |
| `GT_ROOT` | Town root directory | All agents |
| `GT_RIG` | Rig name | Rig-level agents |
| `GT_POLECAT` | Polecat worker name | Polecats only |
| `GT_CREW` | Crew worker name | Crew only |
| `GT_AGENT` | Agent runtime alias being used | All agents |
| `GT_SESSION` / `GT_SESSION_ID` | Session identifier | All agents |
| `GT_TOWN` / `GT_TOWN_ROOT` | Override town root detection | Manual |
| `GT_THEME` | CLI output theme (`dark`/`light`/`auto`) | All |
| `GT_ACCOUNT` | Claude Code account | All |
| `BD_ACTOR` | Agent identity for attribution | All agents |
| `BD_BRANCH` | Per-polecat Dolt branch for write isolation | Polecats |
| `BD_DOLT_AUTO_COMMIT` | Automatic Dolt commits | Polecats |
| `BD_LOG` | Beads debug logging | Debug |
| `BEADS_DIR` | Beads database location | All agents |
| `BEADS_AGENT_NAME` | Agent name for beads ops | Polecats, Crew |
| `GIT_AUTHOR_NAME` | Commit attribution (= BD_ACTOR) | All agents |
| `GIT_AUTHOR_EMAIL` | Owner email from git config | All agents |

### 4.2 Internal / Debug Variables

| Variable | Purpose |
|----------|---------|
| `GT_DEBUG` | Debug logging |
| `GT_STALE_WARNED` | Stale binary warning shown |
| `GT_NUKE_ACKNOWLEDGED` | Polecat nuke confirmation |
| `GT_COMMAND` | Current command being executed |
| `GT_CWD` | Working directory override |
| `GT_BRANCH` | Branch override |
| `GT_ISSUE` | Issue ID context |
| `GT_INTEGRATION_LAND` | Integration branch landing mode |
| `GT_ROLE_HOME` | Role home directory |
| `GT_DEACON` | Deacon session reference |
| `GT_MAYOR` | Mayor session reference |
| `GT_WITNESS` | Witness session reference |
| `GT_REFINERY` | Refinery session reference |
| `GT_REFINERY_WORKER` | Refinery worker mode |
| `GT_CREW_PATH` | Crew workspace path |
| `GT_POLECAT_PATH` | Polecat workspace path |
| `BD_DEBUG_ROUTING` | Debug beads routing |
| `CLAUDE_CODE_USE_BEDROCK` | Use AWS Bedrock for Claude |
| `CLAUDE_CONFIG_DIR` | Custom Claude config directory |
| `CLAUDE_RUNTIME_CONFIG_DIR` | Custom Claude settings directory |
| `CLAUDE_SESSION_ID` | Claude session identifier |
| `OPENCODE_PERMISSION` | OpenCode autonomous mode |
| `SKIP_UPDATE_CHECK` | Skip `make install` update check |

### 4.3 Test Variables

| Variable | Purpose |
|----------|---------|
| `GT_TEST_ATTACHED_MOLECULE_LOG` | Test molecule logging |
| `GT_TEST_NO_NUDGE` | Disable nudging in tests |
| `GT_TEST_NUDGE_LOG` | Log nudge operations |
| `GT_TEST_SKIP_HOOK_VERIFY` | Skip hook verification in tests |

---

## 5. Internal Architecture

### 5.1 Package Overview (~50 packages in `internal/`)

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `cmd` | CLI command registration (Cobra) | All `*Cmd` cobra commands |
| `config` | Config types and serialization | `TownConfig`, `RigConfig`, `TownSettings`, `RuntimeConfig` |
| `constants` | Hardcoded role/directory constants | `RoleMayor`, `DirPolecats`, session prefixes |
| `tmux` | tmux subprocess wrapper | `Tmux` struct with all tmux operations |
| `session` | Agent identity and session management | `AgentIdentity`, `ParseAddress()`, session name functions |
| `runtime` | Agent runtime resolution | `ResolveRuntime()`, `RuntimeConfig` |
| `polecat` | Polecat lifecycle management | `Manager`, `Spawn()`, `Nuke()` |
| `convoy` | Work batch tracking (SQLite) | `ConvoyStore`, `Convoy`, `ConvoyIssue` |
| `mail` | Agent mailbox system | `Mailbox` (legacy JSONL + beads-backed) |
| `beads` | Beads integration | `ResolveBeadsDir()`, routing, agent IDs |
| `hooks` | Claude Code hooks management | Hook base/override merging, sync |
| `workspace` | Town workspace operations | Install, directory structure |
| `rig` | Project container management | Rig add/remove, worktree creation |
| `crew` | Crew workspace management | Clone creation, identity |
| `mayor` | Mayor orchestration | Convoy dispatch, agent spawning |
| `deacon` | Background supervisor | Patrol cycles, health monitoring |
| `daemon` | Background daemon process | Heartbeat, Dolt server management |
| `witness` | Per-rig polecat monitor | Nudging, stall detection, cleanup |
| `refinery` | Merge queue processor | MR processing, conflict handling |
| `formula` | Formula/molecule management | Cook, pour, wisp operations |
| `swarm` | Multi-agent coordination | Worker pool management |
| `plugin` | Plugin system (80% complete) | Plugin scanning, gate evaluation |
| `web` | Web dashboard (Go + embedded JS) | HTTP API, SSE events |
| `tui` | Terminal UI (Bubbletea) | Interactive TUI components |
| `feed` | Activity event feed | Real-time event streaming |
| `escalation` | Escalation routing | Severity-based routing, acknowledgment |
| `dog` | Deacon helper workers | Dog dispatch, Boot watchdog |
| `doltserver` | Dolt SQL Server management | Server lifecycle, database routing |
| `git` | Git operations wrapper | Worktree, branch, merge operations |
| `lock` | File locking | Cross-process synchronization |
| `mq` | Merge queue message handling | MQ protocol messages |
| `state` | Agent state management | State persistence |
| `templates` | Go template rendering | Role templates, message templates |
| `style` | CLI output styling (Lipgloss) | Themed output rendering |
| `util` | Shared utilities | Path helpers, string operations |

### 5.2 Key Architectural Patterns

**tmux Coupling**: All agent interaction flows through `internal/tmux/tmux.go`:
- `NewSession()` / `NewSessionWithCommand()` / `NewSessionWithCommandAndEnv()`
- `SendKeys()` with literal mode + debounce + separate Enter
- `CapturePane()` for output reading
- `HasSession()` / `IsAgentRunning()` for health checks
- Per-session nudge locks (channel-based semaphores, 30s timeout)
- Session name validation regex: `^[a-zA-Z0-9_-]+$`

**Beads as Control Plane**: No separate orchestrator DB. Molecule steps ARE beads issues. State transitions are Dolt commits. Each polecat gets its own Dolt branch (`BD_BRANCH`) for write isolation.

**Directory-Based Identity Detection**: `detectRole()` in `internal/cmd/role.go` maps CWD to role via hardcoded directory patterns. `getRoleHome()` has per-role path logic.

**Hooks-Based Context Injection**: Full role context (~300-500 lines per role) injected by `gt prime` via Claude Code's SessionStart hook. Only `~/gt/CLAUDE.md` exists on disk as a minimal identity anchor.

### 5.3 Storage Layer

**Dolt SQL Server** (one per town):
- Port 3307, managed by daemon
- Data: `~/gt/.dolt-data/` with subdirectory per rig
- Each rig is a separate database (`USE hq`, `USE gastown`, `USE beads`)
- Per-polecat branches for write isolation
- Daemon monitors on heartbeat, auto-restarts on crash

**Beads Routing** (`routes.jsonl`):
- Maps ID prefix → rig location (relative to town root)
- Example: `{"prefix":"gt-","path":"gastown/mayor/rig"}`
- `ResolveBeadsDir()` follows redirect chains (max depth 3)

### 5.4 Dependencies (from `go.mod`)

| Dependency | Purpose |
|------------|---------|
| `spf13/cobra` | CLI framework |
| `charmbracelet/bubbletea` | Terminal UI |
| `charmbracelet/lipgloss` | CLI styling |
| `charmbracelet/bubbles` | UI components |
| `charmbracelet/glamour` | Markdown rendering |
| `BurntSushi/toml` | TOML parsing (formulas, roles) |
| `go-rod/rod` | Browser automation (dashboard) |
| `gofrs/flock` | Cross-process file locking |
| `google/uuid` | UUID generation |
| `golang.org/x/term` | Terminal detection |

---

## 6. Gotchas

The following are critical gotchas, edge cases, and undocumented behaviors discovered during codebase analysis. These directly inform Gas City SDK design decisions.

- Code: `internal/tmux/tmux.go` — Central tmux coupling point
- PRs: #1418 — Recent refactor moving Claude settings out of customer repos

### 6.1 tmux Coupling is Deep

The coupling to tmux is pervasive — not just in `internal/tmux/` but throughout:
- `internal/cmd/start.go`: Hardcoded parallel startup sequence (Mayor+Deacon first, then rig agents)
- `internal/cmd/role.go`: `detectRole()` with hardcoded directory-to-role mapping
- `internal/polecat/manager.go`: Worktree creation tied to tmux session lifecycle
- `internal/session/names.go`: Per-role session name generators (`MayorSessionName()`, `WitnessSessionName()`)
- `internal/deacon/`: Health checks via `capture-pane` + prompt detection
- Nudge delivery uses literal mode + debounce + separate Enter key for reliability

**Gotcha**: Issue #1382 explicitly requests headless/sandboxed polecat support. Current architecture assumes agents stay alive in tmux panes. One-shot CLI agents (e.g., `claude -p`, `codex exec --sandbox`) fail because the session dies on completion.

### 6.2 Hardcoded Role Constants

`internal/constants/constants.go` defines:
- `RoleMayor = "mayor"`, `RoleWitness = "witness"`, `RolePolecat = "polecat"`, etc.
- Directory names: `DirMayor`, `DirPolecats`, `DirCrew`, `DirRefinery`, `DirWitness`
- Session prefixes: `gt-`, `hq-`

These are referenced in 50+ files. Migration to data-driven roles requires touching all of them.

### 6.3 Identity System is String-Based

`internal/session/identity.go`:
- `AgentIdentity` struct with hardcoded `Role` field
- `ParseAddress()` with per-role switch statement
- Address format: `<rig>/<role>/<name>` (e.g., `gastown/polecats/Toast`)
- Session name format: `gt-<rig>-<role>[-<name>]` (e.g., `gt-gastown-Toast`)

### 6.4 Polecat Lifecycle is Self-Cleaning

Critical design decision: Polecats are responsible for their own cleanup via `gt done`:
- No idle state exists — polecats are Working, Stalled, or Zombie
- `gt done` pushes branch, submits to MQ, requests self-nuke, exits immediately
- Witness handles stalled/zombie cleanup, not normal lifecycle
- Three layers: Identity (permanent) → Sandbox (ephemeral per assignment) → Session (ephemeral per step)

### 6.5 Merge Queue Complexity

The Refinery/MQ system is substantial:
- Integration branch support with templates (`integration/{title}`)
- Auto-landing when all children closed
- Conflict resolution: `assign_back` or `auto_rebase`
- Configurable test/lint/build commands per rig
- Squash merge for cleaner history

### 6.6 Mail System Has Two Backends

`internal/mail/mailbox.go` supports:
- **Legacy JSONL mode**: File-based (`inbox.jsonl`), used by crew workers
- **Beads mode**: Backed by Dolt SQL, used by polecats and infrastructure agents
- Lock acquisition via `gofrs/flock` for JSONL mode

### 6.7 Plugin System is 80% Complete

`internal/plugin/` and `docs/design/plugin-system.md`:
- Plugins defined as markdown files with TOML frontmatter
- Gate types: cooldown, cron, condition, event, manual
- Executed by Dog workers dispatched by Deacon
- State tracked as wisps (ephemeral beads)
- No plugins in production use yet

### 6.8 Config Resolution is Layered

Multiple overlapping config systems:
- Claude settings: CWD → parent dirs → `~/.claude/settings.json` → `--settings` flag
- Agent resolution: rig-level → town-level → built-in presets
- Rig settings: wisp layer → rig identity bead labels → town defaults → system defaults
- Hooks: base → role override → rig+role override

### 6.9 Dashboard Exists (Web UI)

`internal/web/` contains a Go web server with embedded JavaScript:
- `gt dashboard --port 8080`
- Real-time agent status via SSE
- Convoy progress tracking
- Configuration management
- Uses `go-rod/rod` for browser automation

### 6.10 Formula System is Mature

35+ TOML formulas in `.beads/formulas/`:
- Three types: `workflow` (sequential), `expansion` (macro), `aspect` (parallel)
- Composition: `extends`, `compose.aspects`, `compose.expand`
- Variables with template substitution (`{{feature}}`)
- Step dependencies (`needs = ["other-step"]`)
- **Critical distinction**: `gt sling` for workflow formulas, `gt formula run` for convoy formulas

### 6.11 Cost Tracking Exists

- `gt costs record` command (triggered by Claude Code Stop hook)
- Cost data stored as beads
- Per-session cost attribution to agent identity

### 6.12 Known Active Bugs (from GitHub Issues)

| Issue | Description | Priority |
|-------|-------------|----------|
| #1387 | Claude uses NBSP in prompt, gastown expects regular space | P1 |
| #1379 | Polecats nuked before `gt done` — lost work | P1 |
| #1381 | Orphaned mol-polecat-work blocks parent bead closure | P2 |
| #1380 | IN_PROGRESS beads assigned to dead polecats not detected | P2 |
| #1378 | Polecats stall at permission prompts with no escalation | P2 |
| #1382 | Support headless/sandboxed polecat agents (no tmux) | Needs triage |
| #1344 | Dockerfile and docker-compose requested | P3 |

---

## 7. Distribution

### 7.1 Installation Methods

1. **Homebrew**: `brew install gastown`
2. **npm**: `npm install -g @gastown/gt` (wrapper downloads native binary)
3. **Go install**: `go install github.com/steveyegge/gastown/cmd/gt@latest`
4. **From source**: `make install` (installs to `~/.local/bin/`)

### 7.2 npm Package (`npm-package/`)

- Package name: `@gastown/gt`
- Wrapper script (`bin/gt.js`) downloads/executes native binary for platform
- PostInstall hook downloads appropriate binary from GitHub Releases
- Supports: darwin, linux, win32 (x64, arm64)

### 7.3 Release Process

- GoReleaser for multi-platform binary builds
- Version injected via ldflags: `Version`, `Commit`, `BuildTime`, `BuiltProperly`
- GitHub Actions CI: Linux, Windows, integration tests
- Semantic versioning (currently v0.5.0)

---

## 8. Testing

- 235+ test files (`*_test.go`)
- `go test ./...` runs all tests
- golangci-lint for static analysis (errcheck, gosec, misspell)
- Migration test suite in `scripts/migration-test/`
- Integration tests require tmux and Claude Code

---

## 9. Key Source Code Locations for Gas City Migration

| What | File | Line |
|------|------|------|
| All roles hardcoded | `internal/config/roles.go` | `AllRoles()` |
| Role TOML definitions | `internal/config/roles/*.toml` | 7 files |
| Role constants | `internal/constants/constants.go` | `RoleMayor`, etc. |
| Identity parsing | `internal/session/identity.go` | `AgentIdentity`, `ParseAddress()` |
| Session name generation | `internal/session/names.go` | `MayorSessionName()`, etc. |
| Role detection from CWD | `internal/cmd/role.go` | `detectRole()`, `getRoleHome()` |
| Hardcoded startup sequence | `internal/cmd/start.go` | Mayor+Deacon first, then rig agents |
| Per-role patrol switch | `internal/cmd/patrol_new.go` | deacon/witness/refinery |
| tmux wrapper | `internal/tmux/tmux.go` | All tmux operations |
| Agent env vars | `internal/config/env.go` | `AgentEnv()` |
| Runtime config | `internal/config/types.go` | `RuntimeConfig` struct |
| Agent registry | `internal/config/agents.go` | Built-in presets + custom agents |
| Polecat manager | `internal/polecat/manager.go` | Spawn, nuke, worktree management |
| Convoy store | `internal/convoy/` | SQLite-based convoy tracking |
| Mail system | `internal/mail/mailbox.go` | Dual-backend (JSONL + beads) |
| Plugin system | `internal/plugin/` | Plugin scanning, gates, dispatch |
| Web dashboard | `internal/web/` | Go HTTP server + embedded JS |
| Role templates | `internal/templates/roles/*.md.tmpl` | 8 role templates |

---

## 10. Sources

All findings grounded in specific source code locations, documentation, and issues.

- Code: `internal/cmd/*.go` — 396+ CLI command registrations (Cobra)
- Code: `internal/config/types.go` — All config structs (TownConfig, RigConfig, TownSettings, RuntimeConfig)
- Code: `internal/config/roles.go` — Hardcoded 7-role registry (`AllRoles()`)
- Code: `internal/config/env.go` — Agent environment variable injection
- Code: `internal/config/agents.go` — Built-in agent presets and registry
- Code: `internal/constants/constants.go` — Role, directory, and session prefix constants
- Code: `internal/tmux/tmux.go` — tmux subprocess wrapper (session mgmt, nudging, capture)
- Code: `internal/session/identity.go` — Agent identity parsing and address generation
- Code: `internal/session/names.go` — Per-role session name generators
- Code: `internal/cmd/role.go` — Directory-based role detection (`detectRole()`)
- Code: `internal/cmd/start.go` — Hardcoded startup sequence
- Code: `internal/polecat/manager.go` — Polecat lifecycle (spawn, nuke, worktree)
- Code: `internal/mail/mailbox.go` — Dual-backend mail system (JSONL + beads)
- Code: `internal/convoy/` — SQLite-based convoy tracking
- Code: `internal/plugin/` — Plugin system (80% complete)
- Code: `internal/web/` — Web dashboard (Go HTTP + embedded JS)
- Code: `internal/templates/roles/*.md.tmpl` — 8 role template files
- Code: `cmd/gt/main.go` — CLI entry point
- Code: `go.mod` — Go 1.24.2 with Cobra, Bubbletea, Lipgloss, Rod, Flock dependencies
- Code: `Makefile` — Build targets with ldflags for version injection
- Code: `.beads/formulas/` — 35+ TOML workflow formula definitions
- Posts: Steve Yegge, "Stevey's Birthday Blog" — Gas City vision (6 requirements)
- PRs: #1418 — Refactor: move Claude settings from customer repos to parent dirs via --settings
- PRs: #1419 — Feat: deterministic merge pipeline with quality gates
- PRs: #1422 — Test: add containerized install and daemon tests
- Code: `/README.md` — Project overview, architecture, quick start
- Code: `/specs/gas-city-vision.md` — Gas City vision (from blog post)
- Code: `/specs/gas-city-spec-initial-draft.md` — 1150-line initial draft spec
- Code: `/docs/glossary.md` — Complete terminology (MEOW, GUPP, NDI)
- Code: `/docs/overview.md` — Role taxonomy, convoy tracking, directory structure
- Code: `/docs/reference.md` — Complete CLI and configuration reference
- Code: `/docs/design/architecture.md` — Two-level beads, agent taxonomy, Dolt storage
- Code: `/docs/design/plugin-system.md` — Plugin architecture with gate types
- Code: `/docs/concepts/polecat-lifecycle.md` — Three-layer lifecycle model
- Code: `/docs/HOOKS.md` — Claude Code hooks management system
- Code: `/docs/why-these-features.md` — Enterprise feature rationale
- Code: `/CONTRIBUTING.md` — Developer workflow, PR conventions
- Code: `/RELEASING.md` — GoReleaser + npm release process
- Code: `/CHANGELOG.md` — Version history (v0.5.0 current)
- Code: GitHub Issue #1382 — Headless/sandboxed polecat agents request
- Code: GitHub Issue #1344 — Docker/docker-compose request
- Code: GitHub Issue #1387 — NBSP bug in Claude prompts (P1)
