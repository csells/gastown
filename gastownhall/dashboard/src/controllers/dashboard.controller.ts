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
  async renderDashboard(_req: Request, res: Response): Promise<void> {
    const data = await this.convoyService.fetchDashboardData();

    res.render('dashboard', {
      convoys: data.convoys,
      mergeQueue: data.mergeQueue,
      rigs: data.rigs,
      townBeads: data.townBeads
    });
  }

  /**
   * Render rig details partial (for HTMX)
   */
  async renderRigDetails(req: Request, res: Response): Promise<void> {
    const rigName = req.params.name;

    if (!rigName) {
      res.status(400).send('Rig name required');
      return;
    }

    const details = await this.convoyService.fetchRigDetails(rigName);

    res.render('partials/rig-details', {
      rig: details
    });
  }
}
