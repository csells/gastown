import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';
import { config } from '../config/config';

export interface CrewInfo {
  name: string;
  rig: string;
  branch: string;
  path: string;
  hasSession: boolean;
  gitClean: boolean;
}

interface RawCrewItem {
  name: string;
  rig: string;
  branch: string;
  path: string;
  has_session: boolean;
  git_clean: boolean;
}

export class CrewFetcher {
  /**
   * Fetch crew workspaces for a specific rig
   */
  async fetchCrewForRig(rigName: string): Promise<CrewInfo[]> {
    const cwd = config.gtTownRoot;

    const result = await CLIExecutor.executeJSON<RawCrewItem[]>(
      `gt crew list --rig=${rigName} --json`,
      { cwd, throwOnError: false }
    );

    return this.transformCrewItems(result);
  }

  /**
   * Fetch all crew workspaces across all rigs
   */
  async fetchAllCrew(): Promise<CrewInfo[]> {
    const cwd = config.gtTownRoot;

    const result = await CLIExecutor.executeJSON<RawCrewItem[]>(
      'gt crew list --all --json',
      { cwd, throwOnError: false }
    );

    return this.transformCrewItems(result);
  }

  private transformCrewItems(items: RawCrewItem[] | null): CrewInfo[] {
    if (!items) return [];

    return items.map(item => ({
      name: item.name,
      rig: item.rig,
      branch: item.branch || 'main',
      path: item.path,
      hasSession: item.has_session ?? false,
      gitClean: item.git_clean ?? true
    }));
  }
}
