export interface ActivityInfo {
  timestamp?: Date;
  age?: number;
  color: ActivityColor;
  text: string;
}

export type ActivityColor = 'green' | 'yellow' | 'red' | 'unknown';
export type WorkStatus = 'complete' | 'active' | 'stale' | 'stuck' | 'waiting';

export class ActivityCalculator {
  private static readonly GREEN_THRESHOLD = 2 * 60 * 1000;  // 2 minutes
  private static readonly YELLOW_THRESHOLD = 5 * 60 * 1000; // 5 minutes

  static calculate(timestamp?: Date): ActivityInfo {
    if (!timestamp) {
      return {
        color: 'unknown',
        text: 'No activity data'
      };
    }

    const now = new Date();
    const age = now.getTime() - timestamp.getTime();

    let color: ActivityColor;
    if (age < this.GREEN_THRESHOLD) {
      color = 'green';
    } else if (age < this.YELLOW_THRESHOLD) {
      color = 'yellow';
    } else {
      color = 'red';
    }

    return {
      timestamp,
      age,
      color,
      text: this.formatTimeAgo(age)
    };
  }

  private static formatTimeAgo(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ago`;
    } else if (minutes > 0) {
      return `${minutes}m ago`;
    } else {
      return `${seconds}s ago`;
    }
  }

  static calculateWorkStatus(
    completed: number,
    total: number,
    activityColor: ActivityColor
  ): WorkStatus {
    if (completed === total && total > 0) {
      return 'complete';
    }

    switch (activityColor) {
      case 'green':
        return 'active';
      case 'yellow':
        return 'stale';
      case 'red':
        return 'stuck';
      default:
        return 'waiting';
    }
  }
}
