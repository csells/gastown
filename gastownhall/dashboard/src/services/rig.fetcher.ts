import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';
import { config } from '../config/config';

export interface RigInfo {
  name: string;
  polecatCount: number;
  crewCount: number;
  agents: string[];
  path?: string;
}

export class RigFetcher {
  /**
   * Fetch all rigs by parsing text output from `gt rig list`
   * Note: gt rig list has no --json flag, so we parse text
   */
  async fetchRigs(): Promise<RigInfo[]> {
    const cwd = config.gtTownRoot;

    const { stdout } = await CLIExecutor.execute('gt rig list', {
      cwd,
      throwOnError: false
    });

    return this.parseRigListOutput(stdout);
  }

  /**
   * Get the filesystem path for a rig
   */
  async getRigPath(rigName: string): Promise<string | null> {
    const cwd = config.gtTownRoot;

    // Use gt rig show to get rig details including path
    const { stdout } = await CLIExecutor.execute(`gt rig show ${rigName}`, {
      cwd,
      throwOnError: false
    });

    // Parse path from output - look for "Path:" line
    const pathMatch = stdout.match(/Path:\s*(.+)/);
    if (pathMatch) {
      return pathMatch[1].trim();
    }

    // Fallback: assume standard location
    return `${cwd}/${rigName}`;
  }

  /**
   * Parse the text output from `gt rig list`
   * Format:
   *   rigname
   *     Polecats: N  Crew: M
   *     Agents: refinery, witness
   */
  private parseRigListOutput(output: string): RigInfo[] {
    const rigs: RigInfo[] = [];
    const lines = output.split('\n');

    let currentRig: Partial<RigInfo> | null = null;

    for (const line of lines) {
      // Skip header lines and empty lines
      if (line.startsWith('Rigs in') || line.startsWith('No rigs') || line.trim() === '') {
        if (currentRig?.name) {
          rigs.push(this.completeRig(currentRig));
        }
        currentRig = null;
        continue;
      }

      // Rig name line (2 spaces indent, no colon)
      const rigNameMatch = line.match(/^  ([a-zA-Z0-9_-]+)$/);
      if (rigNameMatch) {
        if (currentRig?.name) {
          rigs.push(this.completeRig(currentRig));
        }
        currentRig = {
          name: rigNameMatch[1],
          polecatCount: 0,
          crewCount: 0,
          agents: []
        };
        continue;
      }

      // Polecats/Crew line
      const countsMatch = line.match(/Polecats:\s*(\d+)\s+Crew:\s*(\d+)/);
      if (countsMatch && currentRig) {
        currentRig.polecatCount = parseInt(countsMatch[1], 10);
        currentRig.crewCount = parseInt(countsMatch[2], 10);
        continue;
      }

      // Agents line (may or may not have brackets)
      const agentsMatch = line.match(/Agents:\s*\[?([^\]]*)\]?/);
      if (agentsMatch && currentRig) {
        currentRig.agents = agentsMatch[1]
          .split(',')
          .map(a => a.trim())
          .filter(a => a.length > 0);
      }
    }

    // Don't forget the last rig
    if (currentRig?.name) {
      rigs.push(this.completeRig(currentRig));
    }

    return rigs;
  }

  private completeRig(partial: Partial<RigInfo>): RigInfo {
    return {
      name: partial.name || '',
      polecatCount: partial.polecatCount || 0,
      crewCount: partial.crewCount || 0,
      agents: partial.agents || []
    };
  }
}
