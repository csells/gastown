import express, { Express } from 'express';
import path from 'path';
import timeout from 'connect-timeout';
import { DashboardController } from './controllers/dashboard.controller';
import { config } from './config/config';
import { logger } from './utils/logger';

export class DashboardServer {
  private app: Express;
  private controller: DashboardController;

  constructor() {
    this.app = express();
    this.controller = new DashboardController();
    this.setupMiddleware();
    this.setupRoutes();
  }

  private setupMiddleware(): void {
    // Request timeout (60 seconds)
    this.app.use(timeout('60s'));

    // Set view engine
    this.app.set('view engine', 'ejs');
    this.app.set('views', path.join(__dirname, 'views'));

    // Static files
    this.app.use(express.static(path.join(__dirname, '../public')));

    // Request logging
    this.app.use((req, _res, next) => {
      logger.debug(`${req.method} ${req.path}`);
      next();
    });

    // Timeout error handler
    this.app.use((req, _res, next) => {
      if (!req.timedout) next();
    });
  }

  private setupRoutes(): void {
    // Dashboard route
    this.app.get('/', (req, res) => this.controller.renderDashboard(req, res));

    // Rig details route (for HTMX)
    this.app.get('/rig/:name', (req, res) => this.controller.renderRigDetails(req, res));

    // Health check
    this.app.get('/health', (_req, res) => {
      res.json({ status: 'ok' });
    });
  }

  start(port: number = config.port): void {
    const server = this.app.listen(port, () => {
      logger.info(`Dashboard server running on http://localhost:${port}`);
    });

    // Set server timeouts (matching Go implementation)
    server.timeout = 60000;         // 60 seconds
    server.keepAliveTimeout = 120000; // 120 seconds
    server.headersTimeout = 10000;    // 10 seconds
  }
}

// Start server
if (require.main === module) {
  const server = new DashboardServer();
  server.start();
}
