import numpy as np
import pickle
import zstandard as zstd
from datetime import datetime
from pathlib import Path
from influxdb_client import InfluxDBClient, Point
from influxdb_client.client.write_api import SYNCHRONOUS
import logging

class OESDataManager:
    def __init__(self, config):
        self.config = config
        self.influx_client = InfluxDBClient(
            url=config['influxdb_url'],
            token=config['influxdb_token'],
            org=config['influxdb_org']
        )
        self.compressor = zstd.ZstdCompressor(level=3)
        self.decompressor = zstd.ZstdDecompressor()
        self.data_path = Path(config.get('oes_data_path', './data/oes'))
        self.data_path.mkdir(parents=True, exist_ok=True)
        self.logger = logging.getLogger(__name__)

    def store_spectrum(self, chamber_id, wafer_id, recipe_step, timestamp, wavelengths, intensities):
        spectrum_data = {
            'wavelengths': wavelengths.tolist() if isinstance(wavelengths, np.ndarray) else wavelengths,
            'intensities': intensities.tolist() if isinstance(intensities, np.ndarray) else intensities
        }
        compressed = self.compressor.compress(pickle.dumps(spectrum_data))
        
        # Directory structure: chamber/YYYY/MM/DD
        rel_path = f"{chamber_id}/{timestamp.strftime('%Y/%m/%d')}"
        abs_dir = self.data_path / rel_path
        abs_dir.mkdir(parents=True, exist_ok=True)
        
        filename = f"{timestamp.strftime('%H_%M_%S_%f')}.oes.zst"
        full_path = abs_dir / filename
        
        with open(full_path, 'wb') as f:
            f.write(compressed)
            
        # InfluxDB metadata
        point = Point("oes_spectra") \
            .tag("chamber_id", chamber_id) \
            .tag("wafer_id", wafer_id) \
            .tag("recipe_step", recipe_step) \
            .field("file_path", str(full_path)) \
            .time(timestamp)
            
        write_api = self.influx_client.write_api(write_options=SYNCHRONOUS)
        write_api.write(bucket=self.config['influxdb_bucket'], record=point)
        
        return str(full_path)
