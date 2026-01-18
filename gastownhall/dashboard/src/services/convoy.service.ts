import {
  ConvoyData,
  ConvoyRow,
  TrackedIssue,
  MergeQueueRow,
  PolecatRow,
  RigRow,
  RigDetails,
  CrewRow,
  AgentMail,
  MailRow,
  PeekOutput,
  BeadRow
} from '../models/convoy.model';
import { BeadsFetcher } from './beads.fetcher';
import { TmuxFetcher } from './tmux.fetcher';
import { GitHubFetcher } from './github.fetcher';
import { RigFetcher } from './rig.fetcher';
import { CrewFetcher } from './crew.fetcher';
import { MailFetcher, MailMessage } from './mail.fetcher';
import { PeekFetcher } from './peek.fetcher';
import { ActivityCalculator, ActivityInfo } from '../utils/activity';
import { logger } from '../utils/logger';

export class ConvoyService {
  private beadsFetcher: BeadsFetcher;
  private tmuxFetcher: TmuxFetcher;
  private githubFetcher: GitHubFetcher;
  private rigFetcher: RigFetcher;
  private crewFetcher: CrewFetcher;
  private mailFetcher: MailFetcher;
  private peekFetcher: PeekFetcher;

  constructor() {
    this.beadsFetcher = new BeadsFetcher();
    this.tmuxFetcher = new TmuxFetcher();
    this.githubFetcher = new GitHubFetcher();
    this.rigFetcher = new RigFetcher();
    this.crewFetcher = new CrewFetcher();
    this.mailFetcher = new MailFetcher();
    this.peekFetcher = new PeekFetcher();
  }

  /**
   * Fetch all dashboard data (main view)
   */
  async fetchDashboardData(): Promise<ConvoyData> {
    logger.info('Fetching dashboard data');

    const [convoys, mergeQueue, rigs, townBeads] = await Promise.all([
      this.fetchConvoys(),
      this.fetchMergeQueue(),
      this.fetchRigs(),
      this.fetchTownBeads()
    ]);

    return { convoys, mergeQueue, rigs, townBeads };
  }

  /**
   * Fetch rig-specific details (when a rig is selected)
   */
  async fetchRigDetails(rigName: string): Promise<RigDetails> {
    logger.info(`Fetching details for rig: ${rigName}`);

    // Get rig path for beads directory
    const rigPath = await this.rigFetcher.getRigPath(rigName);

    // Get list of polecats for this rig (need names for mail fetching)
    const polecatSessions = await this.tmuxFetcher.getPolecatSessions();
    const rigPolecats = polecatSessions.filter(s => s.rig === rigName);
    const polecatNames = rigPolecats.map(p => p.worker);

    // Get crew for this rig (need names for peek fetching)
    const crewList = await this.crewFetcher.fetchCrewForRig(rigName);
    const crewNames = crewList.filter(c => c.hasSession).map(c => c.name);

    // Fetch all data in parallel
    const [crew, mail, peek, beads] = await Promise.all([
      Promise.resolve(crewList.map(c => this.transformCrewRow(c))),
      this.fetchRigMail(rigName, polecatNames),
      this.peekFetcher.fetchAllPeekForRig(rigName, polecatNames, crewNames),
      rigPath ? this.fetchRigBeads(rigPath) : Promise.resolve([])
    ]);

    // Transform polecats
    const polecats: PolecatRow[] = rigPolecats.map(session => ({
      name: session.worker,
      rig: session.rig,
      sessionId: session.name,
      lastActivity: ActivityCalculator.calculate(session.activityTimestamp),
      statusHint: session.lastOutput
    }));

    return {
      name: rigName,
      crew,
      polecats,
      mail,
      peek: peek.map(p => ({
        worker: p.worker,
        workerType: p.workerType,
        output: p.output,
        timestamp: p.timestamp
      })),
      beads
    };
  }

  /**
   * Fetch rigs list
   */
  private async fetchRigs(): Promise<RigRow[]> {
    const rigs = await this.rigFetcher.fetchRigs();
    return rigs.map(rig => ({
      name: rig.name,
      polecatCount: rig.polecatCount,
      crewCount: rig.crewCount,
      agents: rig.agents
    }));
  }

  /**
   * Fetch town-level beads (hq-* prefix)
   */
  private async fetchTownBeads(): Promise<BeadRow[]> {
    const beads = await this.beadsFetcher.fetchTownBeads();
    return beads.map(bead => ({
      id: bead.id,
      title: bead.title,
      status: bead.status,
      priority: bead.priority ?? 2,
      issueType: bead.issue_type ?? 'task',
      labels: bead.labels ?? []
    }));
  }

  /**
   * Fetch rig-level beads
   */
  private async fetchRigBeads(rigPath: string): Promise<BeadRow[]> {
    const beads = await this.beadsFetcher.fetchRigBeads(rigPath);
    return beads.map(bead => ({
      id: bead.id,
      title: bead.title,
      status: bead.status,
      priority: bead.priority ?? 2,
      issueType: bead.issue_type ?? 'task',
      labels: bead.labels ?? []
    }));
  }

  /**
   * Fetch mail for all agents in a rig
   */
  private async fetchRigMail(rigName: string, polecatNames: string[]): Promise<AgentMail[]> {
    const agentMails = await this.mailFetcher.fetchAllMailForRig(rigName, polecatNames);

    return agentMails.map(am => ({
      agent: am.agent,
      agentType: am.agentType,
      messages: am.messages.map(msg => this.transformMailMessage(msg))
    }));
  }

  private transformMailMessage(msg: MailMessage): MailRow {
    const timestamp = new Date(msg.timestamp);
    return {
      id: msg.id,
      from: msg.from,
      to: msg.to,
      subject: msg.subject,
      timestamp,
      read: msg.read,
      priority: msg.priority || 'normal',
      type: msg.type || 'notification',
      lastActivity: ActivityCalculator.calculate(timestamp)
    };
  }

  private transformCrewRow(crew: { name: string; rig: string; branch: string; hasSession: boolean; gitClean: boolean }): CrewRow {
    return {
      name: crew.name,
      rig: crew.rig,
      branch: crew.branch,
      hasSession: crew.hasSession,
      gitClean: crew.gitClean
    };
  }

  /**
   * Fetch and aggregate convoy data
   */
  private async fetchConvoys(): Promise<ConvoyRow[]> {
    const convoys = await this.beadsFetcher.fetchConvoys();
    const convoyRows: ConvoyRow[] = [];

    for (const convoy of convoys) {
      const row = await this.buildConvoyRow(convoy);
      convoyRows.push(row);
    }

    return convoyRows;
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
    return this.githubFetcher.fetchMergeQueue();
  }
}
