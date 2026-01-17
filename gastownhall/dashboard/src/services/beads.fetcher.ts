import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';
import { config } from '../config/config';

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
        'bd list --type=convoy --status=open --json',
        { throwOnError: false }
      );
      return result || [];
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
      // Escape single quotes in convoy ID for SQLite
      const escapedId = convoyId.replace(/'/g, "''");
      const query = `SELECT depends_on_id FROM dependencies WHERE issue_id = '${escapedId}' AND type = 'tracks'`;
      const beadsDir = config.beadsDir;
      const command = `sqlite3 ${beadsDir}/beads.db "${query}"`;

      const { stdout } = await CLIExecutor.execute(command, { throwOnError: false });
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
        `bd show ${issueId} --json`,
        { throwOnError: false }
      );
      return result;
    } catch (error) {
      logger.error(`Failed to fetch issue ${issueId}`, error);
      return null;
    }
  }
}
