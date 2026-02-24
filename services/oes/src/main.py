import asyncio
import logging
import signal
import yaml
import os
from datetime import datetime
from src.endpoint_detection import OESEndpointDetector
from src.data_manager import OESDataManager

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("OESMain")

class OESOrchestrator:
    def __init__(self, config_path):
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.data_manager = OESDataManager(self.config.get('storage', {}))
        self.detector = OESEndpointDetector(self.config)
        self.running = True

    async def run_ingestion(self):
        logger.info("Starting OES data ingestion via Kafka")
        # Placeholder for Kafka consumer loop
        while self.running:
            await asyncio.sleep(10)

    async def run(self):
        logger.info("OES Service started")
        await self.run_ingestion()

    def stop(self):
        self.running = False
        logger.info("Stopping OES Service")

async def main():
    config_file = os.getenv("OES_CONFIG", "config/oes_config.yaml")
    orchestrator = OESOrchestrator(config_file)
    
    loop = asyncio.get_running_loop()
    for sig in (signal.SIGTERM, signal.SIGINT):
        loop.add_signal_handler(sig, lambda: orchestrator.stop())
    
    await orchestrator.run()

if __name__ == "__main__":
    asyncio.run(main())
