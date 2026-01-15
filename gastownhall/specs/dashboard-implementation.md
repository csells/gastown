# Gastown Dashboard: Complete End-to-End Implementation

## Overview

The **Gastown Dashboard** is a real-time web-based monitoring interface for the Gas Town multi-agent orchestration system. It provides visibility into convoys (work tracking units), merge queues, and active worker agents (polecats).

---

## 1. Architecture Stack

```
Browser (localhost:8080)
    ↓ HTMX auto-refresh (10s)
Go HTTP Server (internal/cmd/dashboard.go)
    ↓
ConvoyHandler (internal/web/handler.go)
    ↓
LiveConvoyFetcher (internal/web/fetcher.go)
    ↓ ↓ ↓
    │ │ └─► GitHub API (gh CLI) → Merge Queue
    │ └───► Tmux Sessions → Worker Activity
    └─────► Beads SQLite (.beads/beads.db) → Convoys & Issues
```

---

## 2. Front-End Implementation

**Location**: `internal/web/templates/convoy.html`

### Key Features:
- **Single-page application** with HTMX for auto-refresh
- **Dark theme** with monospace font
- **No authentication** (localhost only)
- **Auto-refresh every 10 seconds** via `hx-trigger="every 10s"`

### UI Components:

| Component | Purpose | Visualization |
|-----------|---------|---------------|
| **Convoy Table** | Lists open convoys with progress | Progress bars showing completed/total |
| **Merge Queue** | Shows PRs with CI/merge status | Color-coded rows (green/yellow/red) |
| **Polecat Workers** | Active worker agents | Real-time activity indicators |
| **Activity Dots** | Status visualization | Color-coded glowing dots with shadows |

### Activity Color Coding:
- 🟢 **Green**: < 2 minutes (active)
- 🟡 **Yellow**: 2-5 minutes (stale)
- 🔴 **Red**: > 5 minutes (stuck)
- ⚪ **Gray**: No activity data

### CSS Classes:
- `.convoy-table` - Main table styling
- `.work-status` - Badge showing convoy state (complete/active/stale/stuck/waiting)
- `.activity-dot` - Glowing indicator with shadow effects
- `.progress-bar` - Simple percentage fill visualization
- `.mq-*` - Merge queue row coloring (green/yellow/red)

---

## 3. Back-End Implementation

### Entry Point: `internal/cmd/dashboard.go`

```go
Command: gt dashboard --port 8080 --open
```

### Core Data Models (`internal/web/templates.go`):

```go
ConvoyData struct {
    Convoys    []ConvoyRow        // Open convoys with progress
    MergeQueue []MergeQueueRow    // PRs ready for merge
    Polecats   []PolecatRow       // Active worker agents
}

ConvoyRow struct {
    ID            string           // e.g., "hq-cv-abc"
    Title         string
    Status        string           // "open" or "closed"
    WorkStatus    string           // complete/active/stale/stuck/waiting
    Progress      string           // e.g., "2/5"
    Completed     int
    Total         int
    LastActivity  activity.Info    // Color-coded status
    TrackedIssues []TrackedIssue   // Expandable list
}

MergeQueueRow struct {
    Number     int
    Repo       string             // roxas, gastown
    Title      string
    URL        string
    CIStatus   string             // pass/fail/pending
    Mergeable  string             // ready/conflict/pending
    ColorClass string             // CSS styling
}

PolecatRow struct {
    Name         string            // Worker name (dag, nux)
    Rig          string            // rig name (roxas)
    SessionID    string            // tmux session
    LastActivity activity.Info
    StatusHint   string            // Last tmux output
}
```

### HTTP Handler (`internal/web/handler.go`):

```go
type ConvoyHandler struct {
    fetcher  ConvoyFetcher
    template *template.Template
}

func (h *ConvoyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    convoys := h.fetcher.FetchConvoys()       // Query beads
    mergeQueue := h.fetcher.FetchMergeQueue() // Query GitHub
    polecats := h.fetcher.FetchPolecats()     // Query tmux

    h.template.ExecuteTemplate(w, "convoy.html", data)
}
```

### HTTP Endpoints:

```
GET /
  - Fetches all convoys, merge queue, and polecat status
  - Returns rendered HTML template
  - Called every 10 seconds by HTMX client-side refresh
```

---

## 4. Data Flow & Sources

### Primary Data Pipeline:

```
1. Convoys & Issues
   .beads/beads.db (SQLite)
      ↓
   bd list --type=convoy --status=open --json
      ↓
   For each convoy:
     - Query dependencies: SELECT depends_on_id FROM dependencies
       WHERE issue_id = 'convoy-id' AND type = 'tracks'
     - Get issue details: bd show <issue-id> --json
     - Calculate progress: completed_count / total_tracked_count

2. Worker Activity
   Tmux sessions
      ↓
   tmux list-sessions -F "#{session_name}|#{session_activity}"
      ↓
   Parse session activity timestamps
      ↓
   Calculate age and color code

3. Merge Queue
   GitHub API
      ↓
   gh pr list --repo michaellady/roxas --state open --json
      ↓
   Parse CI status and mergeable state
      ↓
   Color code rows based on status
```

### Beads Issue Structure:

```go
type Issue struct {
    ID        string
    Title     string
    Status    string  // "open", "in_progress", "closed"
    Assignee  string  // e.g., "roxas/polecats/dag"
    UpdatedAt string  // RFC3339 timestamp
    DependsOn []string
    Blocks    []string
}
```

### Convoy Tracking Relations:
- Convoys use `tracks` dependency type (non-blocking)
- Each tracked issue can be in different rig's beads (external:prefix:id)
- Progress calculated as: `completed_count / total_tracked_count`

---

## 5. LiveConvoyFetcher Implementation

**Location**: `internal/web/fetcher.go` (756 lines)

The `LiveConvoyFetcher` implements the data aggregation pipeline:

### FetchConvoys()
- Queries `bd list --type=convoy --status=open --json`
- For each convoy, queries tracked issues from SQLite dependencies
- Fetches issue details with `bd show`
- Gets worker activity from tmux sessions
- Calculates activity age and work status
- Returns aggregated `[]ConvoyRow`

### FetchMergeQueue()
- Queries hardcoded repos (michaellady/roxas, michaellady/gastown)
- Uses GitHub API via `gh pr list`
- Evaluates CI status from statusCheckRollup
- Determines merge readiness from mergeable field
- Colors rows based on pass/fail/pending status
- Returns `[]MergeQueueRow`

### FetchPolecats()
- Lists all tmux sessions matching `gt-*-*` pattern
- Filters for worker sessions (excludes witness, mayor, deacon, boot)
- Gets session activity timestamps
- Captures last non-empty line from tmux panes for status hints
- Special handling for "refinery" workers (shows PR count)
- Returns `[]PolecatRow`

---

## 6. Work Status Calculation

### Status Logic:

```go
func calculateWorkStatus(completed, total int, activityColor string) string {
    if completed == total && total > 0 {
        return "complete"  // All tracked issues done
    }

    switch activityColor {
    case "green":   return "active"   // Recent activity
    case "yellow":  return "stale"    // 2-5 min idle
    case "red":     return "stuck"    // > 5 min idle
    default:        return "waiting"  // No activity data
    }
}
```

### Activity Timestamp Logic (`internal/activity/activity.go`):

| Time Since Last Activity | Status | CSS Class |
|-------------------------|--------|-----------|
| < 2 minutes | Active | activity-green |
| 2-5 minutes | Stale | activity-yellow |
| > 5 minutes | Stuck | activity-red |
| No data | Unknown | activity-unknown |

---

## 7. Authentication & Security

### Security Model: **No authentication**

- Runs on **localhost only** (`http://localhost:8080`)
- No middleware or auth guards
- Access implicit to shell users with workspace access
- Beads database accessed via CLI, inheriting filesystem permissions
- Port is configurable but intended for local development/monitoring

### Suitable for:
- Local development monitoring
- Trusted internal networks
- CI/build environments

### Not suitable for:
- Public internet (requires reverse-proxy auth)

---

## 8. Configuration & Setup

### Installation & Startup:

```bash
# Start dashboard with default port (8080)
gt dashboard --port 8080 --open

# Custom port
gt dashboard --port 3000
```

### Configuration Files:

**1. Workspace Root** (`~/gt/mayor/town.json`):
- Marks workspace root for dashboard detection
- Contains town name and configuration

**2. Beads Config** (`~/gt/.beads/config.yaml`):
```yaml
sync-branch: beads-sync
external_projects:
  beads: ../../../beads/mayor/rig  # Cross-rig dependencies
```

**3. Rig Config** (`~/gt/rigs/<rigname>/settings/config.json`):
- Runtime configuration per rig
- Agent provider settings

### Environment Variables:

```bash
GT_TOWN_ROOT      # Set by polecat sessions for fallback workspace detection
BEADS_DIR         # Explicit beads database location
```

### Dependencies:

- **beads (bd)**: CLI for issue tracking queries
- **tmux**: For worker session activity tracking
- **gh**: GitHub CLI for PR merge queue status
- **sqlite3**: For querying beads database (bd wrapper uses it)
- **Go 1.24+**: For HTTP server and template execution

### Server Configuration (`internal/cmd/dashboard.go`):

```go
server := &http.Server{
    Addr:              fmt.Sprintf(":%d", dashboardPort), // Default 8080
    Handler:           handler,
    ReadHeaderTimeout: 10 * time.Second,
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      60 * time.Second,
    IdleTimeout:       120 * time.Second,
}
```

---

## 9. Key Technical Details

### Template Functions (`templates.go`):
- `activityClass()` - Maps activity.Info to CSS class
- `statusClass()` - Maps convoy status to CSS class
- `workStatusClass()` - Maps work status to CSS class
- `progressPercent()` - Calculates percentage for progress bar

### Error Handling:
- **Non-fatal**: Merge queue and polecats failures don't block convoy display
- **Graceful degradation**: Missing data shows as "N/A" or empty state
- **Fallback activity**: If no assignee data, checks running polecat sessions

### External Data Sources:

**1. Tmux Sessions** (for worker activity):
```bash
tmux list-sessions -F "#{session_name}|#{session_activity}"
# Parses: "gt-roxas-dag|1704312345" (unix timestamp)
```

**2. GitHub API** (for merge queue):
```bash
gh pr list --repo michaellady/roxas --state open --json number,title,url,mergeable,statusCheckRollup
```

**3. Beads SQLite** (for dependencies):
```sql
SELECT depends_on_id FROM dependencies
WHERE issue_id = 'convoy-id' AND type = 'tracks'
```

---

## 10. Complete Request Flow

```
User opens browser → http://localhost:8080
    ↓
HTMX auto-refresh (GET /) every 10s
    ↓
ConvoyHandler.ServeHTTP()
    ↓
LiveConvoyFetcher aggregates data:
    │
    ├─► Query Beads SQLite → Convoys + Issues
    │   └─► Calculate progress, activity status
    │
    ├─► Query Tmux → Worker sessions
    │   └─► Calculate activity timestamps
    │
    └─► Query GitHub → Open PRs
        └─► Calculate CI/merge status
    ↓
Render convoy.html template with data
    ↓
Return HTML to browser
    ↓
Display with color-coded status indicators
    ↓
Wait 10 seconds, repeat
```

---

## 11. Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Browser (localhost:8080)                  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  convoy.html (Template + HTMX)                       │   │
│  │  - Auto-refresh every 10s                            │   │
│  │  - Displays convoys, merge queue, polecats           │   │
│  │  - Styling: Dark theme, activity indicators          │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │ GET / (every 10s)                  │
│                         ▼                                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Go HTTP Server (:8080)                          │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ConvoyHandler                                       │   │
│  │ - ServeHTTP(w, r) → Fetch data → Render template   │   │
│  └────────────────┬────────────────────────────────────┘   │
└───────────────────┼──────────────────────────────────────────┘
                    │
        ┌───────────┴────────────┬──────────────────┐
        ▼                        ▼                  ▼
   ┌─────────────┐        ┌────────────┐    ┌──────────────┐
   │  Fetcher    │        │  Beads CLI │    │  GitHub CLI  │
   │             │        │            │    │              │
   │ Fetch       │        │ bd list    │    │ gh pr list   │
   │ Convoys()   │──────► │ bd show    │    │              │
   │             │        │ sqlite3    │    │ Get PR status│
   │ Fetch       │        │            │    │ CI/mergeable │
   │ Merge       │        │ .beads/    │    │              │
   │ Queue()     │        │ beads.db   │    │              │
   │             │        └────────────┘    └──────────────┘
   │ Fetch       │
   │ Polecats()  │        ┌──────────────────────┐
   │             │        │   Tmux Sessions      │
   │             │───────►│                      │
   └─────────────┘        │ tmux list-sessions   │
                          │ Get session activity │
                          │ timestamps           │
                          │                      │
                          └──────────────────────┘
```

---

## 12. File Reference

| File Path | Purpose | Lines |
|-----------|---------|-------|
| `internal/cmd/dashboard.go` | CLI command entry point | - |
| `internal/web/handler.go` | HTTP request handler | - |
| `internal/web/fetcher.go` | Data aggregation pipeline | 756 |
| `internal/web/templates.go` | Data models and template functions | - |
| `internal/web/templates/convoy.html` | HTML template with HTMX | - |
| `internal/activity/activity.go` | Activity timestamp calculation | - |
| `internal/beads/beads.go` | Beads issue structure | - |

---

## 13. Summary

The Gastown Dashboard provides **real-time visibility** into multi-agent workflows with:

- **Minimal overhead**: On-demand data fetching with 10-second refresh
- **Multiple data sources**: Beads (SQLite), Tmux, GitHub API
- **Visual status indicators**: Color-coded activity tracking
- **Progress monitoring**: Convoy completion percentages
- **Merge queue awareness**: PR status at a glance
- **Worker tracking**: Active agent sessions and status

### Key Implementation Points:

1. **No database**: All data fetched on-demand from external sources
2. **CLI orchestration**: Uses bd, tmux, and gh CLI tools
3. **HTMX-powered**: Client-side auto-refresh without JavaScript complexity
4. **Localhost security**: No auth needed, relies on network isolation
5. **Graceful degradation**: Missing data doesn't break the UI
6. **Activity-based status**: Color-coded indicators based on time thresholds

All data aggregation happens in `LiveConvoyFetcher` (internal/web/fetcher.go:756 lines), which orchestrates CLI calls to `bd`, `tmux`, and `gh` to build a comprehensive view of the system state.
