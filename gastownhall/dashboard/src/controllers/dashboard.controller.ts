import { Request, Response } from 'express';
import { ConvoyService } from '../services/convoy.service';
import { logger } from '../utils/logger';

export class DashboardController {
  private convoyService: ConvoyService;

  constructor() {
    this.convoyService = new ConvoyService();
  }

  /**
   * Render dashboard page
   */
  async renderDashboard(req: Request, res: Response): Promise<void> {
    try {
      const data = await this.convoyService.fetchDashboardData();

      res.render('dashboard', {
        convoys: data.convoys,
        mergeQueue: data.mergeQueue,
        polecats: data.polecats
      });
    } catch (error) {
      logger.error('Failed to render dashboard', error);
      res.status(500).send('Internal Server Error');
    }
  }
}
