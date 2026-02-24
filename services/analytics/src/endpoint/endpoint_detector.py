import numpy as np
import pandas as pd
from datetime import datetime
import json
import ruptures as rpt
from scipy import signal
import logging
from kafka import KafkaConsumer, KafkaProducer
from typing import Dict, List, Optional, Tuple

class EndpointDetector:
    def __init__(self, config):
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Buffers for each chamber/symbol
        self.buffers = {}
        self.buffer_size = 1000

    async def start(self):
        self.logger.info("Starting real-time Endpoint Detector via Kafka")
        # In a real setup, this would loop infinitely consuming from Kafka
        # consumer = KafkaConsumer(self.config['kafka_topic'], ...)
        pass

    def detect_change_points(self, values: np.ndarray) -> List[int]:
        if len(values) < 50: return []
        signal_data = values.reshape(-1, 1)
        algo = rpt.Pelt(model="rbf").fit(signal_data)
        result = algo.predict(pen=10)
        return [cp for cp in result if cp < len(values)]

    def detect_outliers(self, values: np.ndarray, threshold: float = 3.5) -> List[int]:
        median = np.median(values)
        mad = np.median(np.abs(values - median))
        if mad == 0: return []
        z_scores = 0.6745 * (values - median) / mad
        return np.where(np.abs(z_scores) > threshold)[0].tolist()

    def detect_level_shifts(self, values: np.ndarray, window: int = 20) -> List[Dict]:
        if len(values) < 2 * window: return []
        shifts = []
        for i in range(window, len(values) - window):
            before = np.mean(values[i-window:i])
            after = np.mean(values[i:i+window])
            if abs(after - before) > 2 * np.std(values[i-window:i+window]):
                shifts.append({'index': i, 'delta': float(after - before)})
        return shifts
