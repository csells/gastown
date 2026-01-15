import { ActivityInfo, WorkStatus } from '../utils/activity';

export interface ConvoyData {
  convoys: ConvoyRow[];
  mergeQueue: MergeQueueRow[];
  polecats: PolecatRow[];
}

export interface ConvoyRow {
  id: string;
  title: string;
  status: string;
  workStatus: WorkStatus;
  progress: string;
  completed: number;
  total: number;
  lastActivity: ActivityInfo;
  trackedIssues: TrackedIssue[];
}

export interface TrackedIssue {
  id: string;
  title: string;
  status: string;
  assignee?: string;
  lastActivity: ActivityInfo;
}

export interface MergeQueueRow {
  number: number;
  repo: string;
  title: string;
  url: string;
  ciStatus: CIStatus;
  mergeable: MergeableStatus;
  colorClass: string;
}

export type CIStatus = 'pass' | 'fail' | 'pending' | 'unknown';
export type MergeableStatus = 'ready' | 'conflict' | 'pending' | 'unknown';

export interface PolecatRow {
  name: string;
  rig: string;
  sessionId: string;
  lastActivity: ActivityInfo;
  statusHint: string;
}
