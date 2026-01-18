# Gastownhall

A TypeScript-based dashboard implementation for monitoring the Gas Town multi-agent orchestration system.

## Overview

Gastownhall is a standalone TypeScript dashboard that provides real-time monitoring of Gas Town convoys, pull requests, rigs, and worker agents. It's implemented entirely using CLI tools (`gt`, `bd`, `tmux`, `gh`) with no dependencies on the main gastown Go implementation.

**Architecture Principle**: The dashboard MUST NOT reimplement any Gas Town functionality. All data comes from existing CLIs. See [CLAUDE.md](CLAUDE.md) for architecture guidelines.

## Project Structure

```
gastownhall/
├── specs/                          # Implementation specifications
│   ├── dashboard-implementation.md # Analysis of Go implementation
│   ├── typescript-dashboard-spec.md # TypeScript implementation spec
│   └── implementation-comparison.md # Feature parity analysis
│
└── dashboard/                      # TypeScript dashboard implementation
    ├── src/
    │   ├── config/                # Configuration & environment
    │   ├── controllers/           # HTTP request handlers
    │   ├── models/                # TypeScript type definitions
    │   ├── services/              # CLI wrappers (fetchers)
    │   │   ├── beads.fetcher.ts   # bd list, bd show, bd dep list
    │   │   ├── convoy.service.ts  # Data aggregation
    │   │   ├── crew.fetcher.ts    # gt crew list
    │   │   ├── github.fetcher.ts  # gh pr list
    │   │   ├── mail.fetcher.ts    # gt mail inbox
    │   │   ├── peek.fetcher.ts    # gt peek
    │   │   ├── rig.fetcher.ts     # gt rig list
    │   │   └── tmux.fetcher.ts    # tmux list-sessions
    │   ├── utils/                 # Utilities (activity, exec, logger)
    │   └── views/                 # EJS templates
    │       ├── dashboard.ejs      # Main dashboard
    │       └── partials/          # HTMX partial templates
    │           └── rig-details.ejs
    ├── package.json
    ├── tsconfig.json
    ├── CLAUDE.md                  # Architecture guidelines
    └── README.md                  # Dashboard-specific documentation
```

## Components

### 1. Specifications (`specs/`)

Documentation and analysis of the dashboard implementation:

- **dashboard-implementation.md**: Complete end-to-end analysis of the original Go dashboard implementation, including architecture, data flow, and features.

- **typescript-dashboard-spec.md**: Comprehensive specification for the TypeScript implementation, including architecture, code examples, configuration, and deployment strategies.

- **implementation-comparison.md**: Feature parity matrix and gap analysis comparing TypeScript spec against Go implementation.

### 2. Dashboard (`dashboard/`)

Production-ready TypeScript dashboard implementation:

**Key Features:**
- 🎯 Real-time convoy tracking with progress indicators
- 📋 Merge queue monitoring with CI status
- 🏗️ Rig-centric view with clickable rig list
- 👥 Crew workspace monitoring (branch, session, git status)
- 🐱 Per-rig polecat worker activity tracking
- 📬 Agent mail inbox (witness + polecats, labeled by agent)
- 👁️ Peek output (recent terminal output per worker)
- 📊 Two-level beads: town-level (hq-*) and rig-level
- 🎨 Light/Dark/System theme cycling icon
- 🔄 Auto-refresh every 10 seconds (HTMX)
- 🌙 Terminal-inspired monospace interface
- 🔒 Localhost-only (no authentication)

**Technology Stack:**
- **Runtime**: Node.js 20+ or Bun
- **Backend**: Express.js with TypeScript
- **Frontend**: EJS templates + HTMX
- **Styling**: CSS with theme variables
- **CLI Integration**: gt, bd, tmux, gh

## Getting Started

### Prerequisites

```bash
# Install required CLI tools
npm install -g typescript tsx

# Verify CLI dependencies
gt --version    # Gas Town orchestrator
bd --version    # Beads issue tracking
tmux -V         # Terminal multiplexer
gh --version    # GitHub CLI
```

### Quick Start

```bash
# Navigate to dashboard
cd gastownhall/dashboard

# Install dependencies
npm install

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Run in development mode
npm run dev

# Dashboard available at http://localhost:8080
```

### Configuration

Edit `dashboard/.env`:

```bash
PORT=8080                              # Dashboard port
LOG_LEVEL=info                         # Logging level
BEADS_DIR=.beads                       # Path to beads database
GT_TOWN_ROOT=~/gt                      # Gas Town workspace root
GITHUB_REPOS=owner/repo1,owner/repo2   # Repos to monitor
```

See `dashboard/.env.example` for complete documentation.

## Features

### Real-Time Monitoring

- **Convoys**: Work tracking units with progress bars and completion status
- **Merge Queue**: Open PRs with CI status and merge readiness indicators
- **Town Beads**: HQ-level beads (hq-* prefix) from ~/gt/.beads/
- **Rigs**: Clickable list showing polecat count, crew count, and active agents

### Rig Details (HTMX-loaded on selection)

- **Crew Workspaces**: Name, branch, session status, git clean/dirty
- **Polecat Workers**: Worker name, session ID, activity status, status hint
- **Agent Mail**: Inbox for witness + all polecats, labeled by agent
- **Peek Output**: Recent terminal output (collapsible per worker)
- **Rig Beads**: Issues from `<rig>/mayor/rig/.beads/`

### Activity Tracking

Color-coded status based on time since last activity:
- 🟢 **Green** - Active (< 2 minutes)
- 🟡 **Yellow** - Stale (2-5 minutes)
- 🔴 **Red** - Stuck (> 5 minutes)
- ⚪ **Gray** - Unknown (no data)

### Work Status

Convoys show work status based on progress:
- **Complete** - All tracked issues closed
- **Active** - Recent activity
- **Stale** - 2-5 min idle
- **Stuck** - > 5 min idle
- **Waiting** - No activity data

### Theme Modes

Three theme options via cycling icon (16x16 in header with tooltip):
- ☀️ **Light Mode** - Clean, bright interface
- 🌙 **Dark Mode** - Terminal-style dark theme
- 💻 **System Mode** - Follows OS preference (default)

Click the icon to cycle through modes. Preference persists in localStorage.

## Architecture

### Data Flow

```
Browser (HTMX Auto-refresh)
    ↓
Express HTTP Server
    ├── GET /         → Dashboard (convoys, merge queue, rigs, town beads)
    └── GET /rig/:name → Rig Details (HTMX partial)
    ↓
ConvoyService (Data Aggregation)
    ↓
┌──────────────────────────────────────────────────────┐
│                    Fetchers                          │
├──────────────────────────────────────────────────────┤
│ RigFetcher     → gt rig list                         │
│ CrewFetcher    → gt crew list --rig=<name> --json    │
│ MailFetcher    → gt mail inbox <identity> --json     │
│ PeekFetcher    → gt peek <rig>/<worker>              │
│ BeadsFetcher   → bd list, bd show, bd dep list       │
│ TmuxFetcher    → tmux list-sessions                  │
│ GitHubFetcher  → gh pr list                          │
└──────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **CLI-Only Integration (CRITICAL)**: Uses only CLI tools (`gt`, `bd`, `gh`, `tmux`) for data access. The dashboard MUST NOT reimplement any Gas Town functionality - CLIs are the source of truth.
2. **Rig-Centric Architecture**: Rigs are the primary organizational unit; rig-specific data loads via HTMX partials
3. **TypeScript Throughout**: Full type safety across all layers
4. **Service Layer Pattern**: Clean separation between data fetching and business logic
5. **HTMX for Simplicity**: Server-rendered HTML with minimal JavaScript, partial updates for rig details
6. **Theme Variables**: CSS custom properties for easy theme switching
7. **Graceful Degradation**: Missing data doesn't break the UI, CLI failures return empty arrays

## Implementation Highlights

### Improvements Over Specification

The implementation includes several enhancements:

1. **Rig-Centric Architecture**: Primary organizational unit with HTMX partial loading
2. **CLI-Only Integration**: All data from `gt`, `bd`, `gh`, `tmux` - no direct DB access
3. **Two-Level Beads**: Town-level (hq-*) and rig-level beads support
4. **Agent Mail Aggregation**: Combined inbox view for witness + all polecats
5. **Peek Output**: Terminal output capture via `gt peek`
6. **Refinery Worker Handling**: Special status display for refinery workers
7. **Server Timeouts**: Proper timeout configuration (60s/120s/10s)
8. **Configurable Repos**: GitHub repos via environment variable
9. **Environment Validation**: Configuration validation on startup
10. **Non-Fatal Errors**: CLI failures return empty arrays, don't crash
11. **Theme Cycling Icon**: Compact 16x16 icon with tooltip (replaced 3-button overlay)

### Security & Robustness

- **No Authentication**: Designed for localhost access only
- **Input Validation**: Environment variables validated on startup
- **Timeout Protection**: CLI calls timeout after 30 seconds
- **Error Isolation**: Failed fetchers don't block other data
- **Graceful Fallbacks**: Missing data shows empty states

## Development

### Available Commands

```bash
npm run dev        # Development mode with hot-reload
npm run build      # Production build
npm start          # Run production build
npm run type-check # TypeScript validation
npm run lint       # ESLint code linting
```

### Project Structure

```
src/
├── config/config.ts              # Validated configuration (includes GT_TOWN_ROOT)
├── controllers/
│   └── dashboard.controller.ts   # HTTP handlers (/, /rig/:name, /health)
├── models/
│   └── convoy.model.ts           # Type definitions (Rig, Crew, Mail, Bead, etc.)
├── services/
│   ├── beads.fetcher.ts          # bd list, bd show, bd dep list
│   ├── convoy.service.ts         # Data aggregation
│   ├── crew.fetcher.ts           # gt crew list --rig=<name> --json
│   ├── github.fetcher.ts         # gh pr list
│   ├── mail.fetcher.ts           # gt mail inbox <identity> --json
│   ├── peek.fetcher.ts           # gt peek <rig>/<worker>
│   ├── rig.fetcher.ts            # gt rig list (text parsing)
│   └── tmux.fetcher.ts           # tmux list-sessions
├── utils/
│   ├── activity.ts               # Activity calculations
│   ├── exec.ts                   # CLI execution wrapper
│   └── logger.ts                 # Logging utility
├── views/
│   ├── dashboard.ejs             # Main dashboard template
│   └── partials/
│       └── rig-details.ejs       # HTMX partial for rig details
└── server.ts                     # Express server with routes
```

### Adding Features

**Important**: Read [CLAUDE.md](CLAUDE.md) before adding features. The dashboard must use CLI tools only.

1. **Check CLIs First**: Does `gt` or `bd` have a command for this? Use it. If not, consider adding to `gt`/`bd` first.
2. **New Data Source**: Add fetcher in `src/services/` that shells out to CLI
3. **New Model**: Add types in `src/models/convoy.model.ts`
4. **Update Service**: Modify `ConvoyService` to call fetcher
5. **Update Controller**: Pass data to view
6. **Update Template**: Add display in `dashboard.ejs` or create partial in `views/partials/`

## Documentation

### Specification Files

- **dashboard-implementation.md** (17KB): Complete analysis of the Go implementation
- **typescript-dashboard-spec.md** (58KB): Full TypeScript implementation specification
- **implementation-comparison.md** (22KB): Gap analysis and feature parity matrix

### Dashboard Documentation

- **dashboard/README.md**: Dashboard-specific setup and usage guide
- **dashboard/.env.example**: Comprehensive environment variable documentation

## Performance

- **Parallel Fetching**: Convoys, PRs, and workers fetched simultaneously
- **Auto-Refresh**: 10-second polling keeps data current
- **Graceful Degradation**: Missing data doesn't break the UI
- **Timeout Protection**: CLI calls timeout after 30 seconds
- **Smooth Transitions**: Theme changes animate with 0.3s ease

## Troubleshooting

### Dashboard shows no data

```bash
# Check beads database
ls -la ${BEADS_DIR}/beads.db

# Test beads CLI
bd list --type=convoy --status=open --json
```

### No worker activity

```bash
# Check tmux sessions
tmux list-sessions

# Verify session format: gt-<rig>-<worker>
```

### Merge queue empty

```bash
# Configure GitHub repos in .env
GITHUB_REPOS=owner/repo1,owner/repo2

# Verify GitHub authentication
gh auth status

# Test PR listing
gh pr list --repo owner/repo --state open
```

## Contributing

When making changes to the dashboard:

1. Update TypeScript code in `src/`
2. Test with `npm run dev`
3. Run type checking: `npm run type-check`
4. Run linting: `npm run lint`
5. Update specifications if architecture changes
6. Update README if features change

## License

This project is part of the Gas Town ecosystem. See the main Gas Town repository for licensing information.

## Related Documentation

- [CLAUDE.md](CLAUDE.md) - Architecture guidelines (CLI-only integration)
- [Dashboard Implementation Analysis](specs/dashboard-implementation.md)
- [TypeScript Implementation Spec](specs/typescript-dashboard-spec.md)
- [Implementation Comparison](specs/implementation-comparison.md)
- [Dashboard README](dashboard/README.md)

## Version History

- **v1.0.0** - Initial TypeScript implementation
  - Express.js backend with TypeScript
  - EJS templates with HTMX auto-refresh
  - Three data fetchers (Beads, Tmux, GitHub)
  - Dark theme monospace interface
  - Comprehensive documentation

- **v1.1.0** - Theme switcher addition
  - Light/Dark/System theme modes
  - Persistent theme preference in localStorage
  - Smooth transitions between themes
  - OS preference detection for system mode

- **v1.2.0** - Rig-centric view
  - Clickable rig list with summary counts
  - HTMX-powered rig detail loading
  - Per-rig crew workspace display (branch, session, git status)
  - Per-rig polecat workers (moved from top-level)
  - Agent mail inbox (witness + polecats, labeled by agent)
  - Peek output (recent terminal output per worker)
  - Two-level beads: town-level (hq-*) and rig-level
  - Theme switcher changed to compact cycling icon (16x16)
  - 4 new fetchers: rig, crew, mail, peek
  - CLAUDE.md architecture guidelines
  - CLI-only integration audit (no direct DB access)
