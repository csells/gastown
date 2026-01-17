import { ConvoyData, ConvoyRow, TrackedIssue, MergeQueueRow, PolecatRow } from '../models/convoy.model';
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
