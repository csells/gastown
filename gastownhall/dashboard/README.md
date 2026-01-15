# Gas Town Dashboard

A real-time web-based monitoring interface for the Gas Town multi-agent orchestration system, implemented in TypeScript.

## Features

- 🎯 **Convoy Tracking** - Monitor work tracking units with progress indicators
- 📋 **Merge Queue** - View open PRs with CI status and merge readiness
- 🐱 **Polecat Workers** - Track active worker agents and their status
- 🔄 **Auto-refresh** - HTMX-powered 10-second auto-refresh
- 🌙 **Dark Theme** - Easy-on-the-eyes monospace interface
- 🔒 **Localhost Only** - No authentication needed, runs locally

## Prerequisites

The dashboard requires these CLI tools to be installed:

- **Node.js** 20+ (or Bun)
- **bd** - Beads issue tracking CLI
- **tmux** - Terminal multiplexer (3.0+)
- **gh** - GitHub CLI (2.0+)
- **sqlite3** - SQLite database CLI

Verify installation:

```bash
node --version
bd --version
tmux -V
gh --version
sqlite3 --version
```

## Installation

```bash
# Install dependencies
npm install

# Copy and configure environment variables
cp .env.example .env
```

## Configuration

Edit `.env` file:

```bash
# Server configuration
PORT=8080                          # Dashboard port
LOG_LEVEL=info                     # debug, info, warn, error

# Beads configuration
BEADS_DIR=.beads                   # Path to beads database

# GitHub repositories (comma-separated)
GITHUB_REPOS=owner/repo1,owner/repo2

# Optional: Town root for workspace detection
GT_TOWN_ROOT=/path/to/gt/mayor
```

## Usage

### Development Mode

```bash
npm run dev
```

Dashboard will be available at `http://localhost:8080` with hot-reload.

### Production Build

```bash
npm run build
npm start
```

### Other Commands

```bash
npm run type-check    # TypeScript type checking
npm run lint          # ESLint code linting
```

## How It Works

### Architecture

```
Browser (HTMX)
    ↓
Express Server
    ↓
ConvoyService
    ↓ ↓ ↓
    │ │ └─► GitHubFetcher (gh pr list)
    │ └───► TmuxFetcher (tmux list-sessions)
    └─────► BeadsFetcher (bd list/show)
```

### Data Sources

1. **Beads SQLite** - Convoy and issue tracking data
2. **Tmux Sessions** - Worker activity timestamps
3. **GitHub API** - Pull request status via gh CLI

### Activity Status

The dashboard color-codes activity based on time since last update:

- 🟢 **Green** - Active (< 2 minutes)
- 🟡 **Yellow** - Stale (2-5 minutes)
- 🔴 **Red** - Stuck (> 5 minutes)
- ⚪ **Gray** - Unknown (no activity data)

### Work Status

Convoys show work status based on progress and activity:

- **Complete** - All tracked issues closed
- **Active** - Work in progress (recent activity)
- **Stale** - Slow progress (2-5 min idle)
- **Stuck** - No recent progress (>5 min idle)
- **Waiting** - No activity data

## Project Structure

```
gastownhall/dashboard/
├── src/
│   ├── config/
│   │   └── config.ts              # Configuration & env validation
│   ├── controllers/
│   │   └── dashboard.controller.ts # HTTP request handler
│   ├── models/
│   │   └── convoy.model.ts        # TypeScript type definitions
│   ├── services/
│   │   ├── beads.fetcher.ts       # Beads CLI integration
│   │   ├── convoy.service.ts      # Data aggregation logic
│   │   ├── github.fetcher.ts      # GitHub CLI integration
│   │   └── tmux.fetcher.ts        # Tmux CLI integration
│   ├── utils/
│   │   ├── activity.ts            # Activity calculation
│   │   ├── exec.ts                # CLI execution wrapper
│   │   └── logger.ts              # Logging utility
│   ├── views/
│   │   └── dashboard.ejs          # HTML template
│   └── server.ts                  # Express server entry point
├── public/                        # Static assets (if needed)
├── dist/                          # Compiled JavaScript
├── package.json
├── tsconfig.json
└── .env                           # Environment configuration
```

## Troubleshooting

### Dashboard shows no convoys

- Verify `bd list --type=convoy --status=open --json` returns data
- Check BEADS_DIR environment variable points to correct .beads directory

### No worker activity shown

- Ensure tmux sessions are running with format `gt-<rig>-<worker>`
- Verify `tmux list-sessions` shows active sessions

### Merge queue is empty

- Configure GITHUB_REPOS environment variable
- Verify `gh auth status` shows you're authenticated
- Check `gh pr list --repo owner/repo --state open` returns PRs

### Port already in use

- Change PORT in .env file
- Or kill the process using the port: `lsof -ti:8080 | xargs kill`

## Development

### Adding New Features

1. **New Data Source**: Add fetcher in `src/services/`
2. **New Model**: Add types in `src/models/`
3. **Update Service**: Modify `ConvoyService` to aggregate new data
4. **Update Template**: Add display in `src/views/dashboard.ejs`

### Error Handling

All CLI operations use `throwOnError: false` to prevent single failures from breaking the entire dashboard. Errors are logged but don't crash the server.

## Security

- **No authentication** - Dashboard is designed for localhost access only
- **Input escaping** - All convoy IDs are escaped before SQL queries
- **Timeouts** - Requests timeout after 60 seconds to prevent hanging
- **CLI only** - No direct database access, uses CLI tools for safety

## Performance

- **Parallel fetching** - Convoys, PRs, and workers fetched simultaneously
- **Auto-refresh** - 10-second polling keeps data current
- **Graceful degradation** - Missing data doesn't break the UI
- **Timeout protection** - CLI calls timeout after 30 seconds

## License

This dashboard is part of the Gas Town project. See the main project for licensing information.

## Related Documentation

- [Dashboard Implementation Spec](../specs/typescript-dashboard-spec.md)
- [Go Implementation Analysis](../specs/dashboard-implementation.md)
- [Implementation Comparison](../specs/implementation-comparison.md)
