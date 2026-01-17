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
        'tmux list-sessions -F "#{session_name}|#{session_activity}"',
        { throwOnError: false }
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

      let lastOutput: string;

      // Special handling for refinery workers - show PR count
      if (worker === 'refinery') {
        lastOutput = await this.getRefineryStatus(session.name);
      } else {
        lastOutput = await this.getLastPaneOutput(session.name);
      }

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
        `tmux capture-pane -p -t ${sessionName} -S -50`,
        { throwOnError: false }
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
   * Get refinery worker status (shows PR count)
   */
  private async getRefineryStatus(sessionName: string): Promise<string> {
    try {
      const output = await this.getLastPaneOutput(sessionName);

      // Try to extract PR count from output
      // Refinery workers typically show something like "Processing PR #123"
      const prMatch = output.match(/PR\s*#?(\d+)/i);
      if (prMatch) {
        return `Working on PR #${prMatch[1]}`;
      }

      // If no PR number found, just return the raw output
      return output || 'Idle';
    } catch (error) {
      logger.error(`Failed to get refinery status for ${sessionName}`, error);
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
