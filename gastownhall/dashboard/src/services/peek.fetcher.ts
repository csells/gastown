import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';
import { config } from '../config/config';

export interface PeekOutput {
  worker: string;
  workerType: 'polecat' | 'crew';
  output: string;
  timestamp: Date;
}

export class PeekFetcher {
  /**
   * Fetch recent terminal output for a worker
   */
  async fetchPeekOutput(rigName: string, workerName: string, lines: number = 20): Promise<PeekOutput | null> {
    const cwd = config.gtTownRoot;

    const { stdout, stderr } = await CLIExecutor.execute(
      `gt peek ${rigName}/${workerName} -n ${lines}`,
      { cwd, throwOnError: false }
    );

    if (stderr && stderr.includes('no such session')) {
      return null;
    }

    return {
      worker: workerName,
      workerType: 'polecat',
      output: stdout.trim(),
      timestamp: new Date()
    };
  }

  /**
   * Fetch peek output for a crew member
   */
  async fetchCrewPeekOutput(rigName: string, crewName: string, lines: number = 20): Promise<PeekOutput | null> {
    const cwd = config.gtTownRoot;

    const { stdout, stderr } = await CLIExecutor.execute(
      `gt peek ${rigName}/crew/${crewName} -n ${lines}`,
      { cwd, throwOnError: false }
    );

    if (stderr && stderr.includes('no such session')) {
      return null;
    }

    return {
      worker: crewName,
      workerType: 'crew',
      output: stdout.trim(),
      timestamp: new Date()
    };
  }

  /**
   * Fetch peek output for all workers in a rig
   */
  async fetchAllPeekForRig(rigName: string, polecatNames: string[], crewNames: string[] = []): Promise<PeekOutput[]> {
    const results: PeekOutput[] = [];

    // Fetch polecat peek outputs in parallel
    const polecatPromises = polecatNames.map(name => this.fetchPeekOutput(rigName, name));
    const polecatResults = await Promise.all(polecatPromises);
    results.push(...polecatResults.filter((r): r is PeekOutput => r !== null));

    // Fetch crew peek outputs in parallel
    const crewPromises = crewNames.map(name => this.fetchCrewPeekOutput(rigName, name));
    const crewResults = await Promise.all(crewPromises);
    results.push(...crewResults.filter((r): r is PeekOutput => r !== null));

    return results;
  }
}
