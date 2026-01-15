import { logger } from '../utils/logger';

export interface Config {
  port: number;
  beadsDir: string;
  logLevel: string;
  githubRepos: string[];
  gtTownRoot?: string;
}

function validateConfig(): Config {
  const port = parseInt(process.env.PORT || '8080', 10);
  if (isNaN(port) || port < 1 || port > 65535) {
    throw new Error(`Invalid PORT: ${process.env.PORT}`);
  }

  const beadsDir = process.env.BEADS_DIR || '.beads';
  const logLevel = process.env.LOG_LEVEL || 'info';

  if (!['debug', 'info', 'warn', 'error'].includes(logLevel)) {
    throw new Error(`Invalid LOG_LEVEL: ${logLevel}. Must be one of: debug, info, warn, error`);
  }

  // Parse GitHub repos from comma-separated env var
  const githubRepos = process.env.GITHUB_REPOS
    ? process.env.GITHUB_REPOS.split(',').map(r => r.trim()).filter(r => r.length > 0)
    : [];

  const gtTownRoot = process.env.GT_TOWN_ROOT;

  return {
    port,
    beadsDir,
    logLevel,
    githubRepos,
    gtTownRoot
  };
}

export const config = validateConfig();

// Set logger level based on config
logger.setLevel(config.logLevel as any);
