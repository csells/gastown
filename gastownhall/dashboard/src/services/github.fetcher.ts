import { CLIExecutor } from '../utils/exec';
import { CIStatus, MergeableStatus, MergeQueueRow } from '../models/convoy.model';
import { logger } from '../utils/logger';
import { config } from '../config/config';

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
  /**
   * Fetch open PRs from configured repositories
   */
  async fetchMergeQueue(): Promise<MergeQueueRow[]> {
    const repos = config.githubRepos;

    if (repos.length === 0) {
      logger.warn('No GitHub repos configured. Set GITHUB_REPOS environment variable.');
      return [];
    }

    const allPRs: MergeQueueRow[] = [];

    for (const repo of repos) {
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

    try {
      const prs = await CLIExecutor.executeJSON<GitHubPR[]>(command, { throwOnError: false });
      return prs.map(pr => this.mapPRToMergeQueueRow(pr, repo));
    } catch (error) {
      logger.error(`Failed to fetch PRs for ${repo}`, error);
      return [];
    }
  }

  /**
   * Map GitHub PR to MergeQueueRow
   */
  private mapPRToMergeQueueRow(pr: GitHubPR, repo: string): MergeQueueRow {
    const repoShortName = repo.split('/')[1] || repo;

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
