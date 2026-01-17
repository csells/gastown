import { exec } from 'child_process';
import { promisify } from 'util';
import { logger } from './logger';

const execAsync = promisify(exec);

export interface ExecResult {
  stdout: string;
  stderr: string;
}

export interface ExecOptions {
  timeout?: number;
  cwd?: string;
  throwOnError?: boolean;
}

export class CLIExecutor {
  /**
   * Execute a shell command and return the result
   */
  static async execute(
    command: string,
    options?: ExecOptions
  ): Promise<ExecResult> {
    const {
      throwOnError = true,
      timeout = 30000,
      cwd = process.cwd()
    } = options || {};

    try {
      logger.debug(`Executing: ${command}`);
      const { stdout, stderr } = await execAsync(command, { timeout, cwd });
      return { stdout, stderr };
    } catch (error: any) {
      logger.error(`CLI execution failed: ${command}`, error);

      if (throwOnError) {
        throw new Error(`Command failed: ${error.message}`);
      }

      return { stdout: '', stderr: error.message };
    }
  }

  /**
   * Execute and parse JSON output
   */
  static async executeJSON<T>(
    command: string,
    options?: ExecOptions
  ): Promise<T> {
    const { stdout } = await this.execute(command, options);
    try {
      return JSON.parse(stdout) as T;
    } catch (error) {
      logger.error(`Failed to parse JSON from: ${command}`, error);
      throw new Error('Invalid JSON response from command');
    }
  }
}
