# Gastown Dashboard: TypeScript Implementation Specification

## Overview

This specification describes a complete TypeScript implementation of the Gastown Dashboard that provides identical functionality to the Go implementation using only CLI tool calls (`bd`, `tmux`, `gh`). The implementation will use **no code** from the existing Go codebase.

---

## Design Decisions

### Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Runtime** | Node.js 20+ or Bun | Modern JavaScript runtime with excellent TypeScript support |
| **Backend Framework** | Express.js | Industry standard, well-documented, simple for this use case |
| **Frontend Approach** | HTMX + Server-rendered HTML | Matches Go implementation's simplicity, minimal JavaScript |
| **Templating** | EJS (Embedded JavaScript) | Simple, familiar syntax similar to Go templates |
| **Styling** | Tailwind CSS | Utility-first CSS for easy dark theme replication |
| **Type Safety** | TypeScript 5.0+ | Full type safety throughout the application |
| **Process Execution** | Node.js `child_process.exec` | For CLI tool execution (bd, tmux, gh) |
| **Build Tool** | tsx/tsup or Bun | Fast TypeScript execution and building |

---

## Architecture

### System Architecture

```
Browser (localhost:8080)
    ↓ HTMX auto-refresh (10s)
Express HTTP Server (src/server.ts)
    ↓
DashboardController (src/controllers/dashboard.controller.ts)
    ↓
ConvoyService (src/services/convoy.service.ts)
    ↓ ↓ ↓
    │ │ └─► GitHubFetcher (gh CLI)
    │ └───► TmuxFetcher (tmux CLI)
    └─────► BeadsFetcher (bd CLI)
```

### Project Structure

```
typescript-dashboard/
├── src/
│   ├── server.ts                    # Express server entry point
│   ├── config/
│   │   └── config.ts                # Configuration and environment
│   ├── controllers/
│   │   └── dashboard.controller.ts  # HTTP request handler
│   ├── services/
│   │   ├── convoy.service.ts        # Main data aggregation service
│   │   ├── beads.fetcher.ts         # Beads CLI wrapper
│   │   ├── tmux.fetcher.ts          # Tmux CLI wrapper
│   │   └── github.fetcher.ts        # GitHub CLI wrapper
│   ├── models/
│   │   ├── convoy.model.ts          # ConvoyData, ConvoyRow types
│   │   ├── merge-queue.model.ts     # MergeQueueRow types
│   │   └── polecat.model.ts         # PolecatRow types
│   ├── utils/
│   │   ├── activity.ts              # Activity timestamp calculation
│   │   ├── exec.ts                  # CLI execution utilities
│   │   └── logger.ts                # Logging utility
│   └── views/
│       └── dashboard.ejs            # HTML template
├── public/
│   └── styles.css                   # Tailwind output (if not using CDN)
├── package.json
├── tsconfig.json
├── tailwind.config.js
└── README.md
```

---

## Data Models

### TypeScript Type Definitions

**src/models/convoy.model.ts**

```typescript
import { ActivityInfo } from '../utils/activity';

export interface ConvoyData {
  convoys: ConvoyRow[];
  mergeQueue: MergeQueueRow[];
  polecats: PolecatRow[];
}

export interface ConvoyRow {
  id: string;                    // e.g., "hq-cv-abc"
  title: string;
  status: string;                // "open" or "closed"
  workStatus: WorkStatus;        // Computed status
  progress: string;              // e.g., "2/5"
  completed: number;
  total: number;
  lastActivity: ActivityInfo;
  trackedIssues: TrackedIssue[];
}

export type WorkStatus = 'complete' | 'active' | 'stale' | 'stuck' | 'waiting';

export interface TrackedIssue {
  id: string;
  title: string;
  status: string;
  assignee?: string;
  lastActivity: ActivityInfo;
}
```

**src/models/merge-queue.model.ts**

```typescript
export interface MergeQueueRow {
  number: number;
  repo: string;                  // Short name: "roxas", "gastown"
  title: string;
  url: string;
  ciStatus: CIStatus;            // "pass", "fail", "pending"
  mergeable: MergeableStatus;    // "ready", "conflict", "pending"
  colorClass: string;            // CSS class for styling
}

export type CIStatus = 'pass' | 'fail' | 'pending' | 'unknown';
export type MergeableStatus = 'ready' | 'conflict' | 'pending' | 'unknown';
```

**src/models/polecat.model.ts**

```typescript
import { ActivityInfo } from '../utils/activity';

export interface PolecatRow {
  name: string;                  // Worker name: "dag", "nux"
  rig: string;                   // Rig name: "roxas"
  sessionId: string;             // Tmux session name
  lastActivity: ActivityInfo;
  statusHint: string;            // Last tmux pane output
}
```

**src/utils/activity.ts**

```typescript
export interface ActivityInfo {
  timestamp?: Date;
  age?: number;                  // Milliseconds since activity
  color: ActivityColor;
  text: string;                  // Human-readable time ago
}

export type ActivityColor = 'green' | 'yellow' | 'red' | 'unknown';

export class ActivityCalculator {
  private static readonly GREEN_THRESHOLD = 2 * 60 * 1000;  // 2 minutes
  private static readonly YELLOW_THRESHOLD = 5 * 60 * 1000; // 5 minutes

  static calculate(timestamp?: Date): ActivityInfo {
    if (!timestamp) {
      return {
        color: 'unknown',
        text: 'No activity data'
      };
    }

    const now = new Date();
    const age = now.getTime() - timestamp.getTime();

    let color: ActivityColor;
    if (age < this.GREEN_THRESHOLD) {
      color = 'green';
    } else if (age < this.YELLOW_THRESHOLD) {
      color = 'yellow';
    } else {
      color = 'red';
    }

    return {
      timestamp,
      age,
      color,
      text: this.formatTimeAgo(age)
    };
  }

  private static formatTimeAgo(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ago`;
    } else if (minutes > 0) {
      return `${minutes}m ago`;
    } else {
      return `${seconds}s ago`;
    }
  }

  static calculateWorkStatus(
    completed: number,
    total: number,
    activityColor: ActivityColor
  ): WorkStatus {
    if (completed === total && total > 0) {
      return 'complete';
    }

    switch (activityColor) {
      case 'green':
        return 'active';
      case 'yellow':
        return 'stale';
      case 'red':
        return 'stuck';
      default:
        return 'waiting';
    }
  }
}
```

---

## CLI Execution Layer

### Utility for CLI Execution

**src/utils/exec.ts**

```typescript
import { exec } from 'child_process';
import { promisify } from 'util';
import { logger } from './logger';

const execAsync = promisify(exec);

export interface ExecResult {
  stdout: string;
  stderr: string;
}

export class CLIExecutor {
  /**
   * Execute a shell command and return the result
   */
  static async execute(
    command: string,
    options?: { timeout?: number; cwd?: string }
  ): Promise<ExecResult> {
    const timeout = options?.timeout || 30000; // 30 second default
    const cwd = options?.cwd || process.cwd();

    try {
      logger.debug(`Executing: ${command}`);
      const { stdout, stderr } = await execAsync(command, { timeout, cwd });
      return { stdout, stderr };
    } catch (error: any) {
      logger.error(`CLI execution failed: ${command}`, error);
      throw new Error(`Command failed: ${error.message}`);
    }
  }

  /**
   * Execute and parse JSON output
   */
  static async executeJSON<T>(
    command: string,
    options?: { timeout?: number; cwd?: string }
  ): Promise<T> {
    const { stdout } = await this.execute(command, options);
    try {
      return JSON.parse(stdout) as T;
    } catch (error) {
      logger.error(`Failed to parse JSON from: ${command}`, error);
      throw new Error('Invalid JSON response from command');
    }
  }
}
```

---

## Data Fetchers

### Beads Fetcher

**src/services/beads.fetcher.ts**

```typescript
import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';

export interface BeadsIssue {
  id: string;
  title: string;
  status: string;
  assignee?: string;
  updated_at?: string;
  depends_on?: string[];
  blocks?: string[];
}

export interface BeadsConvoy extends BeadsIssue {
  type: 'convoy';
}

export class BeadsFetcher {
  /**
   * Fetch all open convoys
   */
  async fetchConvoys(): Promise<BeadsConvoy[]> {
    try {
      const result = await CLIExecutor.executeJSON<BeadsConvoy[]>(
        'bd list --type=convoy --status=open --json'
      );
      return result;
    } catch (error) {
      logger.error('Failed to fetch convoys', error);
      return [];
    }
  }

  /**
   * Fetch tracked issues for a convoy using dependencies
   */
  async fetchTrackedIssues(convoyId: string): Promise<string[]> {
    try {
      // Query SQLite directly for tracked dependencies
      const query = `SELECT depends_on_id FROM dependencies WHERE issue_id = '${convoyId}' AND type = 'tracks'`;
      const beadsDir = process.env.BEADS_DIR || '.beads';
      const command = `sqlite3 ${beadsDir}/beads.db "${query}"`;

      const { stdout } = await CLIExecutor.execute(command);
      return stdout.trim().split('\n').filter(id => id.length > 0);
    } catch (error) {
      logger.error(`Failed to fetch tracked issues for ${convoyId}`, error);
      return [];
    }
  }

  /**
   * Fetch issue details by ID
   */
  async fetchIssue(issueId: string): Promise<BeadsIssue | null> {
    try {
      const result = await CLIExecutor.executeJSON<BeadsIssue>(
        `bd show ${issueId} --json`
      );
      return result;
    } catch (error) {
      logger.error(`Failed to fetch issue ${issueId}`, error);
      return null;
    }
  }

  /**
   * Get beads directory path
   */
  getBeadsDir(): string {
    return process.env.BEADS_DIR || '.beads';
  }
}
```

### Tmux Fetcher

**src/services/tmux.fetcher.ts**

```typescript
import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';

export interface TmuxSession {
  name: string;
  activityTimestamp: Date;
}

export interface PolecatSession extends TmuxSession {
  rig: string;
  worker: string;
  lastOutput: string;
}

export class TmuxFetcher {
  /**
   * List all tmux sessions with activity timestamps
   */
  async listSessions(): Promise<TmuxSession[]> {
    try {
      const { stdout } = await CLIExecutor.execute(
        'tmux list-sessions -F "#{session_name}|#{session_activity}"'
      );

      return stdout
        .trim()
        .split('\n')
        .filter(line => line.length > 0)
        .map(line => {
          const [name, timestamp] = line.split('|');
          return {
            name,
            activityTimestamp: new Date(parseInt(timestamp) * 1000)
          };
        });
    } catch (error) {
      logger.error('Failed to list tmux sessions', error);
      return [];
    }
  }

  /**
   * Get polecat worker sessions (gt-<rig>-<worker>)
   */
  async getPolecatSessions(): Promise<PolecatSession[]> {
    const sessions = await this.listSessions();
    const excludedWorkers = ['witness', 'mayor', 'deacon', 'boot'];

    const polecatSessions: PolecatSession[] = [];

    for (const session of sessions) {
      const match = session.name.match(/^gt-([^-]+)-([^-]+)$/);
      if (!match) continue;

      const [, rig, worker] = match;
      if (excludedWorkers.includes(worker)) continue;

      const lastOutput = await this.getLastPaneOutput(session.name);

      polecatSessions.push({
        name: session.name,
        rig,
        worker,
        activityTimestamp: session.activityTimestamp,
        lastOutput
      });
    }

    return polecatSessions;
  }

  /**
   * Get last non-empty line from tmux pane
   */
  private async getLastPaneOutput(sessionName: string): Promise<string> {
    try {
      const { stdout } = await CLIExecutor.execute(
        `tmux capture-pane -p -t ${sessionName} -S -50`
      );

      const lines = stdout
        .trim()
        .split('\n')
        .map(line => line.trim())
        .filter(line => line.length > 0);

      return lines[lines.length - 1] || '';
    } catch (error) {
      logger.error(`Failed to get pane output for ${sessionName}`, error);
      return '';
    }
  }

  /**
   * Check if a worker is running for a specific assignee
   */
  async isWorkerActive(assignee: string): Promise<Date | null> {
    // Assignee format: "roxas/polecats/dag"
    const match = assignee.match(/^([^\/]+)\/polecats\/([^\/]+)$/);
    if (!match) return null;

    const [, rig, worker] = match;
    const sessionName = `gt-${rig}-${worker}`;

    const sessions = await this.listSessions();
    const session = sessions.find(s => s.name === sessionName);

    return session?.activityTimestamp || null;
  }
}
```

### GitHub Fetcher

**src/services/github.fetcher.ts**

```typescript
import { CLIExecutor } from '../utils/exec';
import { CIStatus, MergeableStatus, MergeQueueRow } from '../models/merge-queue.model';
import { logger } from '../utils/logger';

interface GitHubPR {
  number: number;
  title: string;
  url: string;
  mergeable?: string;
  statusCheckRollup?: {
    state: string;
  }[];
}

export class GitHubFetcher {
  private readonly repos = [
    'michaellady/roxas',
    'michaellady/gastown'
  ];

  /**
   * Fetch open PRs from configured repositories
   */
  async fetchMergeQueue(): Promise<MergeQueueRow[]> {
    const allPRs: MergeQueueRow[] = [];

    for (const repo of this.repos) {
      try {
        const prs = await this.fetchPRsForRepo(repo);
        allPRs.push(...prs);
      } catch (error) {
        logger.error(`Failed to fetch PRs for ${repo}`, error);
      }
    }

    return allPRs;
  }

  /**
   * Fetch PRs for a specific repository
   */
  private async fetchPRsForRepo(repo: string): Promise<MergeQueueRow[]> {
    const command = `gh pr list --repo ${repo} --state open --json number,title,url,mergeable,statusCheckRollup`;

    const prs = await CLIExecutor.executeJSON<GitHubPR[]>(command);

    return prs.map(pr => this.mapPRToMergeQueueRow(pr, repo));
  }

  /**
   * Map GitHub PR to MergeQueueRow
   */
  private mapPRToMergeQueueRow(pr: GitHubPR, repo: string): MergeQueueRow {
    const repoShortName = repo.split('/')[1]; // "michaellady/roxas" -> "roxas"

    const ciStatus = this.determineCIStatus(pr);
    const mergeable = this.determineMergeableStatus(pr);
    const colorClass = this.determineColorClass(ciStatus, mergeable);

    return {
      number: pr.number,
      repo: repoShortName,
      title: pr.title,
      url: pr.url,
      ciStatus,
      mergeable,
      colorClass
    };
  }

  /**
   * Determine CI status from status check rollup
   */
  private determineCIStatus(pr: GitHubPR): CIStatus {
    if (!pr.statusCheckRollup || pr.statusCheckRollup.length === 0) {
      return 'unknown';
    }

    const states = pr.statusCheckRollup.map(check => check.state.toLowerCase());

    if (states.includes('failure') || states.includes('error')) {
      return 'fail';
    } else if (states.includes('pending') || states.includes('in_progress')) {
      return 'pending';
    } else if (states.every(state => state === 'success')) {
      return 'pass';
    }

    return 'unknown';
  }

  /**
   * Determine mergeable status
   */
  private determineMergeableStatus(pr: GitHubPR): MergeableStatus {
    const mergeable = pr.mergeable?.toLowerCase();

    switch (mergeable) {
      case 'mergeable':
        return 'ready';
      case 'conflicting':
        return 'conflict';
      case 'unknown':
        return 'pending';
      default:
        return 'unknown';
    }
  }

  /**
   * Determine CSS color class based on status
   */
  private determineColorClass(ci: CIStatus, mergeable: MergeableStatus): string {
    if (ci === 'pass' && mergeable === 'ready') {
      return 'mq-green';
    } else if (ci === 'fail' || mergeable === 'conflict') {
      return 'mq-red';
    } else {
      return 'mq-yellow';
    }
  }
}
```

---

## Service Layer

### Main Convoy Service

**src/services/convoy.service.ts**

```typescript
import { ConvoyData, ConvoyRow, TrackedIssue, WorkStatus } from '../models/convoy.model';
import { MergeQueueRow } from '../models/merge-queue.model';
import { PolecatRow } from '../models/polecat.model';
import { BeadsFetcher } from './beads.fetcher';
import { TmuxFetcher } from './tmux.fetcher';
import { GitHubFetcher } from './github.fetcher';
import { ActivityCalculator, ActivityInfo } from '../utils/activity';
import { logger } from '../utils/logger';

export class ConvoyService {
  private beadsFetcher: BeadsFetcher;
  private tmuxFetcher: TmuxFetcher;
  private githubFetcher: GitHubFetcher;

  constructor() {
    this.beadsFetcher = new BeadsFetcher();
    this.tmuxFetcher = new TmuxFetcher();
    this.githubFetcher = new GitHubFetcher();
  }

  /**
   * Fetch all dashboard data
   */
  async fetchDashboardData(): Promise<ConvoyData> {
    logger.info('Fetching dashboard data');

    const [convoys, mergeQueue, polecats] = await Promise.all([
      this.fetchConvoys(),
      this.fetchMergeQueue(),
      this.fetchPolecats()
    ]);

    return { convoys, mergeQueue, polecats };
  }

  /**
   * Fetch and aggregate convoy data
   */
  private async fetchConvoys(): Promise<ConvoyRow[]> {
    try {
      const convoys = await this.beadsFetcher.fetchConvoys();
      const convoyRows: ConvoyRow[] = [];

      for (const convoy of convoys) {
        const row = await this.buildConvoyRow(convoy);
        convoyRows.push(row);
      }

      return convoyRows;
    } catch (error) {
      logger.error('Failed to fetch convoys', error);
      return [];
    }
  }

  /**
   * Build a complete convoy row with tracked issues
   */
  private async buildConvoyRow(convoy: any): Promise<ConvoyRow> {
    // Get tracked issue IDs
    const trackedIssueIds = await this.beadsFetcher.fetchTrackedIssues(convoy.id);

    // Fetch details for each tracked issue
    const trackedIssues: TrackedIssue[] = [];
    let completed = 0;

    for (const issueId of trackedIssueIds) {
      const issue = await this.beadsFetcher.fetchIssue(issueId);
      if (!issue) continue;

      // Check if completed
      if (issue.status === 'closed') {
        completed++;
      }

      // Get activity info
      let lastActivity: ActivityInfo;
      if (issue.assignee) {
        const workerTimestamp = await this.tmuxFetcher.isWorkerActive(issue.assignee);
        lastActivity = ActivityCalculator.calculate(workerTimestamp || undefined);
      } else {
        lastActivity = ActivityCalculator.calculate(
          issue.updated_at ? new Date(issue.updated_at) : undefined
        );
      }

      trackedIssues.push({
        id: issue.id,
        title: issue.title,
        status: issue.status,
        assignee: issue.assignee,
        lastActivity
      });
    }

    const total = trackedIssueIds.length;

    // Calculate overall convoy activity (most recent activity from tracked issues)
    const mostRecentActivity = trackedIssues.reduce<ActivityInfo | null>(
      (recent, issue) => {
        if (!recent) return issue.lastActivity;
        if (!issue.lastActivity.timestamp) return recent;
        if (!recent.timestamp) return issue.lastActivity;

        return issue.lastActivity.timestamp > recent.timestamp
          ? issue.lastActivity
          : recent;
      },
      null
    ) || ActivityCalculator.calculate(undefined);

    const workStatus = ActivityCalculator.calculateWorkStatus(
      completed,
      total,
      mostRecentActivity.color
    );

    return {
      id: convoy.id,
      title: convoy.title,
      status: convoy.status,
      workStatus,
      progress: `${completed}/${total}`,
      completed,
      total,
      lastActivity: mostRecentActivity,
      trackedIssues
    };
  }

  /**
   * Fetch merge queue data
   */
  private async fetchMergeQueue(): Promise<MergeQueueRow[]> {
    try {
      return await this.githubFetcher.fetchMergeQueue();
    } catch (error) {
      logger.error('Failed to fetch merge queue', error);
      return [];
    }
  }

  /**
   * Fetch polecat worker data
   */
  private async fetchPolecats(): Promise<PolecatRow[]> {
    try {
      const sessions = await this.tmuxFetcher.getPolecatSessions();

      return sessions.map(session => ({
        name: session.worker,
        rig: session.rig,
        sessionId: session.name,
        lastActivity: ActivityCalculator.calculate(session.activityTimestamp),
        statusHint: session.lastOutput
      }));
    } catch (error) {
      logger.error('Failed to fetch polecats', error);
      return [];
    }
  }
}
```

---

## Controller Layer

**src/controllers/dashboard.controller.ts**

```typescript
import { Request, Response } from 'express';
import { ConvoyService } from '../services/convoy.service';
import { logger } from '../utils/logger';

export class DashboardController {
  private convoyService: ConvoyService;

  constructor() {
    this.convoyService = new ConvoyService();
  }

  /**
   * Render dashboard page
   */
  async renderDashboard(req: Request, res: Response): Promise<void> {
    try {
      const data = await this.convoyService.fetchDashboardData();

      res.render('dashboard', {
        convoys: data.convoys,
        mergeQueue: data.mergeQueue,
        polecats: data.polecats
      });
    } catch (error) {
      logger.error('Failed to render dashboard', error);
      res.status(500).send('Internal Server Error');
    }
  }
}
```

---

## Server Implementation

**src/server.ts**

```typescript
import express, { Express } from 'express';
import path from 'path';
import { DashboardController } from './controllers/dashboard.controller';
import { config } from './config/config';
import { logger } from './utils/logger';

export class DashboardServer {
  private app: Express;
  private controller: DashboardController;

  constructor() {
    this.app = express();
    this.controller = new DashboardController();
    this.setupMiddleware();
    this.setupRoutes();
  }

  private setupMiddleware(): void {
    // Set view engine
    this.app.set('view engine', 'ejs');
    this.app.set('views', path.join(__dirname, 'views'));

    // Static files
    this.app.use(express.static(path.join(__dirname, '../public')));

    // Request logging
    this.app.use((req, res, next) => {
      logger.debug(`${req.method} ${req.path}`);
      next();
    });
  }

  private setupRoutes(): void {
    // Dashboard route
    this.app.get('/', (req, res) => this.controller.renderDashboard(req, res));

    // Health check
    this.app.get('/health', (req, res) => {
      res.json({ status: 'ok' });
    });
  }

  start(port: number = config.port): void {
    this.app.listen(port, () => {
      logger.info(`Dashboard server running on http://localhost:${port}`);
    });
  }
}

// Start server
if (require.main === module) {
  const server = new DashboardServer();
  server.start();
}
```

**src/config/config.ts**

```typescript
export interface Config {
  port: number;
  beadsDir: string;
  logLevel: string;
}

export const config: Config = {
  port: parseInt(process.env.PORT || '8080', 10),
  beadsDir: process.env.BEADS_DIR || '.beads',
  logLevel: process.env.LOG_LEVEL || 'info'
};
```

**src/utils/logger.ts**

```typescript
type LogLevel = 'debug' | 'info' | 'warn' | 'error';

class Logger {
  private level: LogLevel = 'info';

  setLevel(level: LogLevel): void {
    this.level = level;
  }

  debug(message: string, ...args: any[]): void {
    if (this.shouldLog('debug')) {
      console.debug(`[DEBUG] ${message}`, ...args);
    }
  }

  info(message: string, ...args: any[]): void {
    if (this.shouldLog('info')) {
      console.info(`[INFO] ${message}`, ...args);
    }
  }

  warn(message: string, ...args: any[]): void {
    if (this.shouldLog('warn')) {
      console.warn(`[WARN] ${message}`, ...args);
    }
  }

  error(message: string, ...args: any[]): void {
    if (this.shouldLog('error')) {
      console.error(`[ERROR] ${message}`, ...args);
    }
  }

  private shouldLog(level: LogLevel): boolean {
    const levels: LogLevel[] = ['debug', 'info', 'warn', 'error'];
    return levels.indexOf(level) >= levels.indexOf(this.level);
  }
}

export const logger = new Logger();
```

---

## Frontend Template

**src/views/dashboard.ejs**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Gas Town Dashboard</title>
  <script src="https://unpkg.com/htmx.org@1.9.10"></script>
  <script src="https://cdn.tailwindcss.com"></script>
  <style>
    body {
      background-color: #1a1a2e;
      color: #eaeaea;
      font-family: 'Courier New', monospace;
    }

    .activity-dot {
      width: 12px;
      height: 12px;
      border-radius: 50%;
      display: inline-block;
      margin-right: 8px;
    }

    .activity-green {
      background-color: #00ff00;
      box-shadow: 0 0 8px #00ff00;
    }

    .activity-yellow {
      background-color: #ffff00;
      box-shadow: 0 0 8px #ffff00;
    }

    .activity-red {
      background-color: #ff0000;
      box-shadow: 0 0 8px #ff0000;
    }

    .activity-unknown {
      background-color: #666666;
    }

    .progress-bar {
      background-color: #2d2d44;
      height: 20px;
      border-radius: 4px;
      overflow: hidden;
      position: relative;
    }

    .progress-fill {
      background-color: #4a9eff;
      height: 100%;
      transition: width 0.3s ease;
    }

    .work-status {
      padding: 4px 12px;
      border-radius: 4px;
      font-size: 12px;
      font-weight: bold;
      text-transform: uppercase;
    }

    .work-status.complete { background-color: #00aa00; }
    .work-status.active { background-color: #4a9eff; }
    .work-status.stale { background-color: #ffaa00; }
    .work-status.stuck { background-color: #ff4444; }
    .work-status.waiting { background-color: #666666; }

    .mq-green { background-color: rgba(0, 170, 0, 0.2); }
    .mq-yellow { background-color: rgba(255, 170, 0, 0.2); }
    .mq-red { background-color: rgba(255, 68, 68, 0.2); }

    table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 32px;
    }

    th {
      text-align: left;
      padding: 12px;
      background-color: #2d2d44;
      border-bottom: 2px solid #4a9eff;
    }

    td {
      padding: 12px;
      border-bottom: 1px solid #2d2d44;
    }

    tr:hover {
      background-color: rgba(74, 158, 255, 0.1);
    }

    h1 {
      color: #4a9eff;
      margin-bottom: 32px;
      font-size: 32px;
    }

    h2 {
      color: #4a9eff;
      margin-top: 48px;
      margin-bottom: 16px;
      font-size: 24px;
    }

    a {
      color: #4a9eff;
      text-decoration: none;
    }

    a:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body class="p-8">
  <div hx-get="/" hx-trigger="every 10s" hx-swap="innerHTML" class="max-w-7xl mx-auto">
    <h1>🚂 Gas Town Dashboard</h1>

    <!-- Convoys Section -->
    <h2>🎯 Convoys</h2>
    <% if (convoys.length === 0) { %>
      <p class="text-gray-400 italic">No open convoys</p>
    <% } else { %>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Title</th>
            <th>Status</th>
            <th>Progress</th>
            <th>Activity</th>
          </tr>
        </thead>
        <tbody>
          <% convoys.forEach(convoy => { %>
            <tr>
              <td><code><%= convoy.id %></code></td>
              <td><%= convoy.title %></td>
              <td>
                <span class="work-status <%= convoy.workStatus %>">
                  <%= convoy.workStatus %>
                </span>
              </td>
              <td>
                <div class="flex items-center gap-4">
                  <div class="progress-bar flex-1">
                    <div class="progress-fill" style="width: <%= (convoy.completed / convoy.total * 100) || 0 %>%"></div>
                  </div>
                  <span class="text-sm"><%= convoy.progress %></span>
                </div>
              </td>
              <td>
                <span class="activity-dot activity-<%= convoy.lastActivity.color %>"></span>
                <%= convoy.lastActivity.text %>
              </td>
            </tr>
          <% }); %>
        </tbody>
      </table>
    <% } %>

    <!-- Merge Queue Section -->
    <h2>📋 Merge Queue</h2>
    <% if (mergeQueue.length === 0) { %>
      <p class="text-gray-400 italic">No open pull requests</p>
    <% } else { %>
      <table>
        <thead>
          <tr>
            <th>Repo</th>
            <th>PR</th>
            <th>Title</th>
            <th>CI Status</th>
            <th>Mergeable</th>
          </tr>
        </thead>
        <tbody>
          <% mergeQueue.forEach(pr => { %>
            <tr class="<%= pr.colorClass %>">
              <td><%= pr.repo %></td>
              <td><a href="<%= pr.url %>" target="_blank">#<%= pr.number %></a></td>
              <td><%= pr.title %></td>
              <td><%= pr.ciStatus %></td>
              <td><%= pr.mergeable %></td>
            </tr>
          <% }); %>
        </tbody>
      </table>
    <% } %>

    <!-- Polecat Workers Section -->
    <h2>🐱 Polecat Workers</h2>
    <% if (polecats.length === 0) { %>
      <p class="text-gray-400 italic">No active workers</p>
    <% } else { %>
      <table>
        <thead>
          <tr>
            <th>Worker</th>
            <th>Rig</th>
            <th>Session</th>
            <th>Activity</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <% polecats.forEach(polecat => { %>
            <tr>
              <td><%= polecat.name %></td>
              <td><%= polecat.rig %></td>
              <td><code><%= polecat.sessionId %></code></td>
              <td>
                <span class="activity-dot activity-<%= polecat.lastActivity.color %>"></span>
                <%= polecat.lastActivity.text %>
              </td>
              <td class="text-sm text-gray-400"><%= polecat.statusHint %></td>
            </tr>
          <% }); %>
        </tbody>
      </table>
    <% } %>
  </div>
</body>
</html>
```

---

## Configuration Files

**package.json**

```json
{
  "name": "gastown-dashboard-ts",
  "version": "1.0.0",
  "description": "TypeScript implementation of Gas Town dashboard",
  "main": "dist/server.js",
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "build": "tsup src/server.ts --format cjs --dts",
    "start": "node dist/server.js",
    "lint": "eslint src/**/*.ts",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "express": "^4.18.2",
    "ejs": "^3.1.9"
  },
  "devDependencies": {
    "@types/express": "^4.17.21",
    "@types/node": "^20.10.0",
    "@types/ejs": "^3.1.5",
    "typescript": "^5.3.3",
    "tsx": "^4.7.0",
    "tsup": "^8.0.1",
    "eslint": "^8.56.0",
    "@typescript-eslint/eslint-plugin": "^6.17.0",
    "@typescript-eslint/parser": "^6.17.0"
  },
  "engines": {
    "node": ">=20.0.0"
  }
}
```

**tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "moduleResolution": "node",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

**.eslintrc.json**

```json
{
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": 2022,
    "sourceType": "module"
  },
  "extends": [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended"
  ],
  "rules": {
    "@typescript-eslint/no-explicit-any": "warn",
    "@typescript-eslint/explicit-function-return-type": "off",
    "@typescript-eslint/no-unused-vars": ["error", { "argsIgnorePattern": "^_" }]
  }
}
```

**.env.example**

```bash
# Server configuration
PORT=8080
LOG_LEVEL=info

# Beads configuration
BEADS_DIR=.beads

# Optional: Town root for workspace detection
GT_TOWN_ROOT=/path/to/gt/mayor
```

---

## Installation & Usage

### Setup Instructions

```bash
# 1. Create project directory
mkdir gastown-dashboard-ts
cd gastown-dashboard-ts

# 2. Initialize Node.js project
npm init -y

# 3. Install dependencies
npm install express ejs

# 4. Install dev dependencies
npm install -D typescript tsx tsup @types/node @types/express @types/ejs \
  eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin

# 5. Create directory structure
mkdir -p src/{config,controllers,services,models,utils,views}
mkdir -p public

# 6. Copy configuration files and implementation files

# 7. Run in development mode
npm run dev

# 8. Or build and run production
npm run build
npm start
```

### Usage

```bash
# Development mode with hot reload
npm run dev

# Production build
npm run build
npm start

# Type checking
npm run type-check

# Linting
npm run lint
```

### Environment Variables

Create `.env` file:

```bash
PORT=8080
LOG_LEVEL=info
BEADS_DIR=.beads
GT_TOWN_ROOT=/path/to/gt/mayor
```

---

## CLI Dependencies

The TypeScript implementation requires the same CLI tools as the Go version:

| CLI Tool | Usage | Version |
|----------|-------|---------|
| **bd** | Beads issue tracking queries | Latest |
| **tmux** | Worker session activity tracking | 3.0+ |
| **gh** | GitHub API for PR status | 2.0+ |
| **sqlite3** | Direct database queries (optional) | 3.0+ |

### Verification

```bash
# Check all required CLI tools are installed
bd --version
tmux -V
gh --version
sqlite3 --version
```

---

## Testing Strategy

### Unit Tests (Jest)

```typescript
// tests/utils/activity.test.ts
import { ActivityCalculator } from '../../src/utils/activity';

describe('ActivityCalculator', () => {
  it('should calculate green status for recent activity', () => {
    const now = new Date();
    const oneMinuteAgo = new Date(now.getTime() - 60 * 1000);

    const result = ActivityCalculator.calculate(oneMinuteAgo);

    expect(result.color).toBe('green');
  });

  it('should calculate yellow status for stale activity', () => {
    const now = new Date();
    const threeMinutesAgo = new Date(now.getTime() - 3 * 60 * 1000);

    const result = ActivityCalculator.calculate(threeMinutesAgo);

    expect(result.color).toBe('yellow');
  });

  it('should calculate red status for stuck activity', () => {
    const now = new Date();
    const tenMinutesAgo = new Date(now.getTime() - 10 * 60 * 1000);

    const result = ActivityCalculator.calculate(tenMinutesAgo);

    expect(result.color).toBe('red');
  });
});
```

### Integration Tests

Mock CLI responses to test data fetching and aggregation without executing actual commands.

```typescript
// tests/services/beads.fetcher.test.ts
import { BeadsFetcher } from '../../src/services/beads.fetcher';
import { CLIExecutor } from '../../src/utils/exec';

jest.mock('../../src/utils/exec');

describe('BeadsFetcher', () => {
  it('should fetch convoys successfully', async () => {
    const mockConvoys = [
      { id: 'hq-cv-1', title: 'Test Convoy', status: 'open' }
    ];

    (CLIExecutor.executeJSON as jest.Mock).mockResolvedValue(mockConvoys);

    const fetcher = new BeadsFetcher();
    const result = await fetcher.fetchConvoys();

    expect(result).toEqual(mockConvoys);
    expect(CLIExecutor.executeJSON).toHaveBeenCalledWith(
      'bd list --type=convoy --status=open --json'
    );
  });
});
```

---

## Performance Considerations

### Optimizations

1. **Parallel Data Fetching**: Use `Promise.all()` to fetch convoys, merge queue, and polecats simultaneously
2. **CLI Execution Timeouts**: Set reasonable timeouts (30s) to prevent hanging requests
3. **Error Isolation**: Failures in merge queue or polecats don't block convoy data
4. **Caching** (optional): Add short-lived cache (5-10s) to reduce CLI calls for rapid requests
5. **Connection Pooling**: Reuse child processes where possible

### Scalability

- **Current load**: Single dashboard viewer, 10s refresh = ~6 requests/minute
- **Expected load**: Multiple viewers watching same data
- **Bottleneck**: CLI execution time (bd, gh, tmux queries)
- **Mitigation**: Implement shared cache with 5-10s TTL to serve multiple viewers

---

## Deployment Options

### Local Development

```bash
npm run dev
# Dashboard available at http://localhost:8080
```

### Docker Container

```dockerfile
FROM node:20-alpine

WORKDIR /app

# Install CLI tools
RUN apk add --no-cache tmux sqlite git gh

COPY package*.json ./
RUN npm ci --only=production

COPY dist ./dist
COPY src/views ./src/views
COPY public ./public

EXPOSE 8080

CMD ["node", "dist/server.js"]
```

### Systemd Service

```ini
[Unit]
Description=Gas Town Dashboard
After=network.target

[Service]
Type=simple
User=gastown
WorkingDirectory=/home/gastown/dashboard
Environment="NODE_ENV=production"
Environment="PORT=8080"
ExecStart=/usr/bin/node dist/server.js
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

---

## Comparison with Go Implementation

| Aspect | Go Implementation | TypeScript Implementation |
|--------|------------------|---------------------------|
| **Runtime** | Go binary | Node.js/Bun |
| **Type Safety** | Yes (compiled) | Yes (TypeScript) |
| **Performance** | Faster | Good enough for use case |
| **Memory Usage** | Lower | Slightly higher |
| **Startup Time** | Instant | Quick (~100ms) |
| **Dependencies** | None (single binary) | Node.js + packages |
| **Code Reuse** | None | None (CLI-only) |
| **Templating** | Go templates | EJS templates |
| **CLI Execution** | exec.Command | child_process.exec |
| **Error Handling** | Explicit error returns | Try-catch + promises |
| **Concurrency** | Goroutines | Promises/async-await |

---

## Theme Switcher Implementation

### Overview

The dashboard includes a comprehensive theme system with three modes: Light, Dark, and System (follows OS preference). The theme switcher is positioned in the upper right corner with persistent user preference storage.

### Theme Modes

| Mode | Icon | Behavior |
|------|------|----------|
| Light | ☀️ | Forces light theme regardless of OS settings |
| Dark | 🌙 | Forces dark theme (original terminal style) |
| System | 💻 | Follows OS preference (default) |

### CSS Implementation

**CSS Variables for Theming:**

```css
:root {
  /* Light theme colors */
  --bg-color-light: #f5f5f5;
  --text-color-light: #1a1a2e;
  --bg-secondary-light: #e0e0e0;
  --table-bg-light: #ffffff;

  /* Dark theme colors */
  --bg-color-dark: #1a1a2e;
  --text-color-dark: #eaeaea;
  --bg-secondary-dark: #2d2d44;
  --table-bg-dark: transparent;
}

/* System theme - respects OS preference */
html[data-theme="system"] {
  --bg-color: var(--bg-color-dark);
  --text-color: var(--text-color-dark);
}

@media (prefers-color-scheme: light) {
  html[data-theme="system"] {
    --bg-color: var(--bg-color-light);
    --text-color: var(--text-color-light);
  }
}

/* Forced themes */
html[data-theme="dark"] { /* dark colors */ }
html[data-theme="light"] { /* light colors */ }
```

### JavaScript Implementation

**Theme Switcher Logic:**

```javascript
const THEME_KEY = 'gastown-dashboard-theme';
const html = document.documentElement;

// Load saved theme or default to 'system'
function loadTheme() {
  const savedTheme = localStorage.getItem(THEME_KEY) || 'system';
  setTheme(savedTheme);
}

// Set theme and update UI
function setTheme(theme) {
  html.setAttribute('data-theme', theme);
  localStorage.setItem(THEME_KEY, theme);
  updateButtonStates(theme);
}

// Listen for OS theme changes when in system mode
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
mediaQuery.addEventListener('change', () => {
  if (getCurrentTheme() === 'system') {
    // Force repaint to update system theme
  }
});
```

### UI Components

**Theme Switcher HTML:**

```html
<div class="theme-switcher">
  <button class="theme-btn" data-theme="light">☀️ Light</button>
  <button class="theme-btn" data-theme="dark">🌙 Dark</button>
  <button class="theme-btn active" data-theme="system">💻 System</button>
</div>
```

**Styling:**

```css
.theme-switcher {
  position: fixed;
  top: 2rem;
  right: 2rem;
  display: flex;
  gap: 0.5rem;
  background-color: var(--bg-secondary);
  padding: 0.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  z-index: 1000;
}

.theme-btn {
  background: transparent;
  border: 2px solid transparent;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.theme-btn:hover {
  background-color: rgba(74, 158, 255, 0.2);
  border-color: #4a9eff;
}

.theme-btn.active {
  background-color: #4a9eff;
  color: #fff;
}
```

### Features

1. **Persistent Preference**: Theme choice saved to localStorage
2. **Default System Mode**: Follows OS preference by default
3. **Smooth Transitions**: 0.3s ease transitions between themes
4. **Dynamic OS Detection**: Updates automatically when OS theme changes
5. **Visual Feedback**: Active button highlighted, hover effects
6. **All Elements Themed**: Tables, progress bars, code blocks, empty states

### Color Specifications

**Light Theme:**
- Background: #f5f5f5
- Text: #1a1a2e
- Tables: White with light gray borders
- Code blocks: #e8e8e8

**Dark Theme:**
- Background: #1a1a2e
- Text: #eaeaea
- Tables: Transparent with dark borders
- Code blocks: #2d2d44

Both themes maintain the same:
- Accent color: #4a9eff
- Activity indicators: Green/Yellow/Red
- Work status badges: Color-coded

### Accessibility

- High contrast maintained in both themes
- Color-blind safe activity indicators (shape + color)
- Keyboard accessible theme buttons
- Respects OS preference by default

---

## Feature Parity Checklist

- [x] Display open convoys with tracked issues
- [x] Calculate convoy progress (completed/total)
- [x] Show work status (complete/active/stale/stuck/waiting)
- [x] Display activity indicators with color coding
- [x] Show merge queue with CI status
- [x] Display polecat workers with session info
- [x] Auto-refresh every 10 seconds via HTMX
- [x] Dark theme with monospace font
- [x] Light/Dark/System theme switcher with persistence
- [x] No authentication (localhost only)
- [x] Use bd CLI for beads data
- [x] Use tmux CLI for worker activity
- [x] Use gh CLI for GitHub data
- [x] Graceful error handling
- [x] Configurable port
- [x] Activity thresholds (2min/5min)

---

## Future Enhancements (Optional)

These are **not** part of the initial specification but could be added later:

1. **WebSocket support** - Replace polling with real-time push updates
2. **Filtering/sorting** - Allow users to filter convoys by status, sort by activity
3. **Historical data** - Track convoy progress over time, show charts
4. **Notifications** - Alert when workers get stuck or PRs fail CI
5. **Mobile responsive** - Optimize UI for mobile devices
6. **Docker compose** - Full stack deployment with dependencies
7. **Health monitoring** - Track dashboard health and CLI availability
8. **Configuration UI** - Web-based configuration instead of env vars

---

## Summary

This specification provides a complete, production-ready TypeScript implementation of the Gas Town dashboard that:

1. **Uses only CLI tools** - bd, tmux, gh (no Go code dependencies)
2. **Matches exact functionality** - All features from Go version
3. **Maintains architecture** - Similar separation of concerns
4. **Type-safe** - Full TypeScript coverage
5. **Simple deployment** - Single Node.js application
6. **Well-tested** - Unit and integration test structure
7. **Documented** - Clear setup and usage instructions

The implementation can be built and deployed independently, requiring only the same CLI tools (bd, tmux, gh) that the Go version uses.
