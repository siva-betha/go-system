import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
from scipy import stats
from datetime import datetime
import json
import logging
from typing import Dict, List, Optional, Tuple

class GoldenFingerprint:
    def __init__(self, chamber_id: str):
        self.chamber_id = chamber_id
        self.mean = {}
        self.std = {}
        self.quantiles = {}
        self.model = None
        self.scaler = None

class GoldenFingerprintService:
    def __init__(self, influx_client, postgres_client):
        self.influx = influx_client
        self.postgres = postgres_client
        self.logger = logging.getLogger(__name__)

    def create_statistical_profile(self, chamber_id: str, df: pd.DataFrame) -> GoldenFingerprint:
        fingerprint = GoldenFingerprint(chamber_id)
        for column in df.columns:
            if column in ['timestamp', 'machine_id', 'chamber_id']: continue
            values = df[column].dropna()
            if len(values) == 0: continue
            
            fingerprint.mean[column] = float(values.mean())
            fingerprint.std[column] = float(values.std())
            fingerprint.quantiles[column] = {
                'q1': float(values.quantile(0.25)),
                'median': float(values.median()),
                'q3': float(values.quantile(0.75))
            }
        return fingerprint

    def train_isolation_forest(self, fingerprint: GoldenFingerprint, df: pd.DataFrame):
        feature_cols = [col for col in df.columns if col not in ['timestamp', 'machine_id', 'chamber_id']]
        X = df[feature_cols].fillna(method='ffill').values
        
        scaler = StandardScaler()
        X_scaled = scaler.fit_transform(X)
        
        model = IsolationForest(contamination=0.05, random_state=42)
        model.fit(X_scaled)
        
        fingerprint.model = model
        fingerprint.scaler = scaler

    def compare(self, fingerprint: GoldenFingerprint, df: pd.DataFrame) -> Dict:
        results = {'chamber_id': fingerprint.chamber_id, 'health_score': 1.0, 'anomalies': []}
        
        if fingerprint.model and fingerprint.scaler:
            feature_cols = [col for col in df.columns if col in fingerprint.mean.keys()]
            X = df[feature_cols].fillna(method='ffill').values
            X_scaled = fingerprint.scaler.transform(X)
            
            preds = fingerprint.model.predict(X_scaled)
            anomaly_count = np.sum(preds == -1)
            results['health_score'] = float(1.0 - (anomaly_count / len(X)))
            results['anomaly_percentage'] = float(anomaly_count / len(X))
            
        return results
