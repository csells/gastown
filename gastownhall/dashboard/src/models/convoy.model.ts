import { ActivityInfo, WorkStatus } from '../utils/activity';

// Main dashboard data structure
export interface ConvoyData {
  convoys: ConvoyRow[];
  mergeQueue: MergeQueueRow[];
  rigs: RigRow[];
  townBeads: BeadRow[];
}

// Rig detail data (loaded when rig is selected)
export interface RigDetails {
  name: string;
  crew: CrewRow[];
  polecats: PolecatRow[];
  mail: AgentMail[];
  peek: PeekOutput[];
  beads: BeadRow[];
}

// Convoy interfaces (existing)
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

// Merge Queue interfaces (existing)
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

// Polecat interface (existing, now used in RigDetails)
export interface PolecatRow {
  name: string;
  rig: string;
  sessionId: string;
  lastActivity: ActivityInfo;
  statusHint: string;
}

// Rig interfaces (new)
export interface RigRow {
  name: string;
  polecatCount: number;
  crewCount: number;
  agents: string[];
}

// Crew interfaces (new)
export interface CrewRow {
  name: string;
  rig: string;
  branch: string;
  hasSession: boolean;
  gitClean: boolean;
}

// Mail interfaces (new)
export interface AgentMail {
  agent: string;
  agentType: 'witness' | 'polecat' | 'crew';
  messages: MailRow[];
}

export interface MailRow {
  id: string;
  from: string;
  to: string;
  subject: string;
  timestamp: Date;
  read: boolean;
  priority: string;
  type: string;
  lastActivity: ActivityInfo;
}

// Peek output interface (new)
export interface PeekOutput {
  worker: string;
  workerType: 'polecat' | 'crew';
  output: string;
  timestamp: Date;
}

// Bead interface (new)
export interface BeadRow {
  id: string;
  title: string;
  status: string;
  priority: number;
  issueType: string;
  labels: string[];
}
