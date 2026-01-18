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
   * Fetch tracked issues for a convoy using bd dep list
   */
  async fetchTrackedIssues(convoyId: string): Promise<string[]> {
    interface DepListItem {
      id: string;
      type: string;
    }

    const result = await CLIExecutor.executeJSON<DepListItem[]>(
      `bd dep list ${convoyId} --direction=down -t tracks --json`,
      { throwOnError: false }
    );

    return (result || []).map(item => item.id);
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

  /**
   * Fetch town-level beads (hq-* prefix) from ~/gt/.beads/
   */
  async fetchTownBeads(): Promise<RawBead[]> {
    const cwd = config.gtTownRoot;

    const result = await CLIExecutor.executeJSON<RawBead[]>(
      'bd list --json',
      { cwd, throwOnError: false }
    );

    // Filter to only hq-* beads (town-level)
    return (result || []).filter(bead => bead.id.startsWith('hq-'));
  }

  /**
   * Fetch rig-level beads from a specific rig's beads directory
   */
  async fetchRigBeads(rigPath: string): Promise<RawBead[]> {
    const beadsDir = `${rigPath}/mayor/rig/.beads`;

    // Check if the beads directory exists
    const { stdout: checkOutput } = await CLIExecutor.execute(
      `test -d "${beadsDir}" && echo "exists"`,
      { throwOnError: false }
    );

    if (!checkOutput.includes('exists')) {
      logger.debug(`No rig beads directory at ${beadsDir}`);
      return [];
    }

    const result = await CLIExecutor.executeJSON<RawBead[]>(
      `BEADS_DIR="${beadsDir}" bd list --json`,
      { throwOnError: false }
    );

    return result || [];
  }
}

interface RawBead {
  id: string;
  title: string;
  status: string;
  priority?: number;
  issue_type?: string;
  labels?: string[];
}
