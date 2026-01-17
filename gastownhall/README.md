# Gastownhall

A TypeScript-based dashboard implementation for monitoring the Gas Town multi-agent orchestration system.

## Overview

Gastownhall is a standalone TypeScript dashboard that provides real-time monitoring of Gas Town convoys, pull requests, and worker agents. It's implemented entirely using CLI tools (`bd`, `tmux`, `gh`) with no dependencies on the main gastown Go implementation.

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
    │   ├── services/              # Data fetchers & aggregation
    │   ├── utils/                 # Utilities (activity, exec, logger)
    │   └── views/                 # EJS templates
    ├── package.json
    ├── tsconfig.json
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
- 🐱 Polecat worker activity tracking
- 🎨 Light/Dark/System theme modes with switcher
- 🔄 Auto-refresh every 10 seconds (HTMX)
- 🌙 Terminal-inspired monospace interface
- 🔒 Localhost-only (no authentication)

**Technology Stack:**
- **Runtime**: Node.js 20+ or Bun
- **Backend**: Express.js with TypeScript
- **Frontend**: EJS templates + HTMX
- **Styling**: CSS with theme variables
- **CLI Integration**: bd, tmux, gh

## Getting Started

### Prerequisites

```bash
# Install required CLI tools
npm install -g typescript tsx

# Verify CLI dependencies
bd --version    # Beads issue tracking
tmux -V         # Terminal multiplexer
gh --version    # GitHub CLI
sqlite3 --version
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
GITHUB_REPOS=owner/repo1,owner/repo2   # Repos to monitor
```

See `dashboard/.env.example` for complete documentation.

## Features

### Real-Time Monitoring

- **Convoys**: Work tracking units with progress bars and completion status
- **Merge Queue**: Open PRs with CI status and merge readiness indicators
- **Polecat Workers**: Active agents with activity timestamps and status hints

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

Three theme options with persistent preference:
- ☀️ **Light Mode** - Clean, bright interface
- 🌙 **Dark Mode** - Terminal-style dark theme
- 💻 **System Mode** - Follows OS preference (default)

## Architecture

### Data Flow

```
Browser (HTMX Auto-refresh)
    ↓
Express HTTP Server
    ↓
ConvoyService (Data Aggregation)
    ↓ ↓ ↓
    │ │ └─► GitHubFetcher → gh pr list
    │ └───► TmuxFetcher → tmux list-sessions
    └─────► BeadsFetcher → bd list/show
```

### Key Design Decisions

1. **CLI-Only Integration**: Uses only CLI tools for data access, no direct Go code dependencies
2. **TypeScript Throughout**: Full type safety across all layers
3. **Service Layer Pattern**: Clean separation between data fetching and business logic
4. **HTMX for Simplicity**: Server-rendered HTML with minimal JavaScript
5. **Theme Variables**: CSS custom properties for easy theme switching
6. **Graceful Degradation**: Missing data doesn't break the UI

## Implementation Highlights

### Improvements Over Specification

The implementation includes several enhancements:

1. **Input Escaping**: SQL queries escape single quotes for safety
2. **Refinery Worker Handling**: Special status display for refinery workers
3. **Server Timeouts**: Proper timeout configuration (60s/120s/10s)
4. **Configurable Repos**: GitHub repos via environment variable
5. **Environment Validation**: Configuration validation on startup
6. **Non-Fatal Errors**: CLI failures don't crash the dashboard
7. **Theme Switcher**: Light/Dark/System modes with persistent preference (added post-spec)

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
├── config/config.ts              # Validated configuration
├── controllers/
│   └── dashboard.controller.ts   # HTTP handlers
├── models/
│   └── convoy.model.ts           # Type definitions
├── services/
│   ├── beads.fetcher.ts          # Beads CLI wrapper
│   ├── tmux.fetcher.ts           # Tmux CLI wrapper
│   ├── github.fetcher.ts         # GitHub CLI wrapper
│   └── convoy.service.ts         # Data aggregation
├── utils/
│   ├── activity.ts               # Activity calculations
│   ├── exec.ts                   # CLI execution
│   └── logger.ts                 # Logging utility
├── views/
│   └── dashboard.ejs             # HTML template
└── server.ts                     # Express server
```

### Adding Features

1. **New Data Source**: Add fetcher in `src/services/`
2. **New Model**: Add types in `src/models/`
3. **Update Service**: Modify `ConvoyService` to aggregate data
4. **Update Template**: Add display in `dashboard.ejs`

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
