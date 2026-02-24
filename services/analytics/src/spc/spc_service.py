import numpy as np
import pandas as pd
from datetime import datetime, timedelta
import asyncio
from typing import Dict, List, Optional, Tuple
import logging
from dataclasses import dataclass

@dataclass
class SPCRules:
    """Western Electric / Nelson Rules for control chart interpretation"""
    rule1_violation: bool  # Point beyond ±3σ
    rule2_violation: bool  # 9 points on same side of center
    rule3_violation: bool  # 6 points increasing/decreasing
    rule4_violation: bool  # 14 points alternating up/down
    rule5_violation: bool  # 2 of 3 points beyond ±2σ
    rule6_violation: bool  # 4 of 5 points beyond ±1σ
    rule7_violation: bool  # 15 points within ±1σ
    rule8_violation: bool  # 8 points beyond ±1σ either side

class SPCService:
    def __init__(self, influx_client, postgres_client, config):
        self.influx = influx_client
        self.postgres = postgres_client
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Control chart constants
        self.D3 = {2:0, 3:0, 4:0, 5:0, 6:0, 7:0.076, 8:0.136, 9:0.184, 10:0.223}
        self.D4 = {2:3.267, 3:2.574, 4:2.282, 5:2.114, 6:2.004, 7:1.924, 
                   8:1.864, 9:1.816, 10:1.777}
        self.A2 = {2:1.880, 3:1.023, 4:0.729, 5:0.577, 6:0.483, 7:0.419, 
                   8:0.373, 9:0.337, 10:0.308}
    
    async def run_scheduled_check(self):
        """Main entry point - runs periodically to check all active chambers"""
        self.logger.info("Starting scheduled SPC check")
        
        # Placeholder for fetching active chambers and analyzing them
        # In a real scenario, this would iterate through chambers and symbols
        pass

    def xbar_chart(self, subgroups: np.ndarray) -> Dict:
        subgroup_means = np.mean(subgroups, axis=1)
        grand_mean = np.mean(subgroup_means)
        subgroup_ranges = np.ptp(subgroups, axis=1)
        mean_range = np.mean(subgroup_ranges)
        
        n = subgroups.shape[1]
        A2 = self.A2.get(n, 0.729)
        UCL = grand_mean + A2 * mean_range
        LCL = grand_mean - A2 * mean_range
        
        violations = []
        for i, mean in enumerate(subgroup_means):
            if mean > UCL or mean < LCL:
                violations.append({'subgroup': i, 'value': float(mean), 'limit': 'UCL' if mean > UCL else 'LCL'})
        
        return {
            'grand_mean': float(grand_mean),
            'UCL': float(UCL),
            'LCL': float(LCL),
            'subgroup_means': subgroup_means.tolist(),
            'violations': violations
        }

    def r_chart(self, subgroups: np.ndarray) -> Dict:
        subgroup_ranges = np.ptp(subgroups, axis=1)
        mean_range = np.mean(subgroup_ranges)
        n = subgroups.shape[1]
        D3 = self.D3.get(n, 0)
        D4 = self.D4.get(n, 2.114)
        
        UCL = D4 * mean_range
        LCL = D3 * mean_range
        
        violations = []
        for i, r in enumerate(subgroup_ranges):
            if r > UCL or r < LCL:
                violations.append({'subgroup': i, 'value': float(r), 'limit': 'UCL' if r > UCL else 'LCL'})
        
        return {
            'mean_range': float(mean_range),
            'UCL': float(UCL),
            'LCL': float(LCL),
            'violations': violations
        }

    def nelson_rules(self, values: np.ndarray) -> SPCRules:
        n = len(values)
        mean = np.mean(values)
        sigma = np.std(values)
        if sigma == 0: sigma = 1e-9
        z = (values - mean) / sigma
        
        rules = SPCRules(
            rule1_violation=any(abs(z) > 3),
            rule2_violation=False, rule3_violation=False, rule4_violation=False,
            rule5_violation=False, rule6_violation=False, rule7_violation=False, rule8_violation=False
        )
        # Rule 2: 9 points on same side
        for i in range(n-8):
            if all(z[i:i+9] > 0) or all(z[i:i+9] < 0):
                rules.rule2_violation = True; break
        # Rule 3: 6 points in a row increasing/decreasing
        diffs = np.diff(z)
        for i in range(n-5):
            if all(diffs[i:i+5] > 0) or all(diffs[i:i+5] < 0):
                rules.rule3_violation = True; break
        return rules
