# Feature: Defect-to-Process Correlation Engine

Correlate inline defect inspection data with real-time process signatures.

## User Story
As a yield engineer, I want to know which process steps and parameters are causing specific defect types.

## Implementation Details

### Data Fusion Strategy
- **Inspection Data Ingest**: Automatically pulls KLARF/inspection files from KLA-Tencor or Hitachi tools.
- **Spatiotemporal Alignment**: Uses wafer serial numbers and timestamps to align physical defect coordinates (X, Y) with the exact time segment in the OES trace.
- **Signature Extraction**: Extracts OES features (peaks, ratios, slopes) for the "hot zones" on the wafer map.

### Backend: Correlation Engine
```python
# services/yield/correlation.py

class DefectCorrelator:
    def find_root_cause(self, wafer_id: str):
        """
        Identify process steps that correlate with defect patterns
        """
        defect_map = self.inspection_api.get_wafer_map(wafer_id)
        process_data = self.influx.get_process_data(wafer_id)
        
        # Spatial-to-temporal mapping
        # Maps circular scan patterns to time-series segments
        correlations = spatial_temporal_join(defect_map, process_data)
        
        return {
            'top_correlated_steps': correlations.rank_by_importance(),
            'probable_chemistry_cause': self.analyze_oes_shifts(correlations),
            'prediction_confidence': correlations.confidence_score
        }
```

## Killer Differentiator: Yield-Aware Monitoring
- **Automated Root Cause**: No more guessing if a yield drop was caused by the etch tool or the previous litho step—the data alignment proves it.
- **Defect Fingerprinting**: Build a library of "Process Signatures" for common defects (e.g., "The Pressure Spike Signature for Arcing Defects").
- **Closed-Loop Yield**: Automatically tightens control limits for a recipe if it starts producing high defect counts, even if it's still within the original "engineering" spec.
