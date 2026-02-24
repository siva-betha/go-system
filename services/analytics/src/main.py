import asyncio
import logging
import signal
from datetime import datetime
import yaml
import os

from src.spc.spc_service import SPCService
from src.golden.golden_fingerprint import GoldenFingerprintService
from src.endpoint.endpoint_detector import EndpointDetector
from src.common.influx_client import get_influx_client
from src.common.postgres_client import get_postgres_client

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("AnalyticsMain")

class AnalyticsOrchestrator:
    def __init__(self, config_path):
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.influx = get_influx_client(self.config.get('storage', {}).get('influxdb', {}))
        self.postgres = get_postgres_client(self.config.get('storage', {}).get('postgres', {}))
        
        self.spc = SPCService(self.influx, self.postgres, self.config.get('spc', {}))
        self.golden = GoldenFingerprintService(self.influx, self.postgres)
        self.endpoint = EndpointDetector(self.config.get('endpoint', {}))
        
        self.running = True

    async def run_scheduled_tasks(self):
        logger.info("Starting scheduled analytics tasks (SPC/Golden)")
        while self.running:
            try:
                # Every 5 minutes run SPC
                await self.spc.run_scheduled_check()
                # Run daily golden fingerprint recalibration (placeholder)
                await asyncio.sleep(300) 
            except Exception as e:
                logger.error(f"Error in scheduler: {e}")
                await asyncio.sleep(60)

    async def run(self):
        logger.info("Initializing Analytics Microservice")
        
        tasks = [
            asyncio.create_task(self.endpoint.start()),
            asyncio.create_task(self.run_scheduled_tasks())
        ]
        
        await asyncio.gather(*tasks)

    def stop(self):
        self.running = False
        logger.info("Stopping Analytics Services")

async def main():
    config_file = os.getenv("ANALYTICS_CONFIG", "config/analytics.yaml")
    orchestrator = AnalyticsOrchestrator(config_file)
    
    loop = asyncio.get_running_loop()
    for sig in (signal.SIGTERM, signal.SIGINT):
        loop.add_signal_handler(sig, lambda: orchestrator.stop())
    
    await orchestrator.run()

if __name__ == "__main__":
    asyncio.run(main())
