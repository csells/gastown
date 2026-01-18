import { CLIExecutor } from '../utils/exec';
import { logger } from '../utils/logger';
import { config } from '../config/config';

export interface MailMessage {
  id: string;
  from: string;
  to: string;
  subject: string;
  body: string;
  timestamp: string;
  read: boolean;
  priority: string;
  type: string;
}

export interface AgentMail {
  agent: string;
  agentType: 'witness' | 'polecat' | 'crew';
  messages: MailMessage[];
}

export class MailFetcher {
  /**
   * Fetch mail for a specific agent identity
   */
  async fetchMailForAgent(identity: string): Promise<MailMessage[]> {
    const cwd = config.gtTownRoot;

    const result = await CLIExecutor.executeJSON<MailMessage[]>(
      `gt mail inbox ${identity} --json`,
      { cwd, throwOnError: false }
    );

    return result || [];
  }

  /**
   * Fetch all mail for a rig (witness + all polecats)
   */
  async fetchAllMailForRig(rigName: string, polecatNames: string[]): Promise<AgentMail[]> {
    const results: AgentMail[] = [];

    // Fetch witness inbox
    const witnessMessages = await this.fetchMailForAgent(`${rigName}/witness`);
    results.push({
      agent: 'witness',
      agentType: 'witness',
      messages: witnessMessages
    });

    // Fetch each polecat's inbox in parallel
    const polecatPromises = polecatNames.map(async (name) => {
      const messages = await this.fetchMailForAgent(`${rigName}/${name}`);
      return {
        agent: name,
        agentType: 'polecat' as const,
        messages
      };
    });

    const polecatResults = await Promise.all(polecatPromises);
    results.push(...polecatResults);

    // Filter out agents with no messages to reduce noise
    return results.filter(r => r.messages.length > 0);
  }

  /**
   * Fetch mayor's inbox (town-level)
   */
  async fetchMayorMail(): Promise<MailMessage[]> {
    return this.fetchMailForAgent('mayor/');
  }
}
