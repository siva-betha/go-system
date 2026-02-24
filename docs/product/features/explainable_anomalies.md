# Feature: Automated Root Cause Analysis

Automatically analyze excursions and identify probable root causes.

## User Story
As a process engineer, I want to know why a chamber drifted without spending hours analyzing data.

## Implementation Details

### Automated Diagnostic Engine
- **Excursion Trigger**: Fired automatically when a run violates SPC (Statistical Process Control) limits.
- **Comparative Analysis**: Compares the "Anomalous Run" against the "Fleet Average" and "Recent Golden Run."
- **Feature Importance**: Uses SHAP values to rank which sensors contributed most to the anomaly.

### Backend: RCA Service
```python
# services/analytics/rca.py

class RootCauseAnalyzer:
    async def analyze_incident(self, incident_id: str):
        """
        Diagnose the physical cause of a process excursion
        """
        # Step 1: Gather multi-source telemetry
        full_context = await self.aggregator.get_all_context(incident_id)
        
        # Step 2: Run diagnostic models
        causes = self.classifier.get_probable_causes(full_context)
        
        return {
            'probable_causes': [
                {'factor': 'RF Matching Network', 'probability': 0.82},
                {'factor': 'Backside Helium Leak', 'probability': 0.15}
            ],
            'recommended_action': "Check RF match capacitors for signs of arcing.",
            'historical_precedent': self.find_similar_past_events(full_context)
        }
```

## Killer Differentiator: The "Expert in a Box"
- **Reduces MTTR (Mean Time To Repair)**: Pinpoints the specific valve, generator, or gas line that is at fault, skipping hours of manual testing.
- **Institutional Knowledge Retention**: Learns from resolved tickets—if a pressure spike was caused by a specific sensor failure once, the AI will remember and suggest it next time.
- **Explainable AI**: Doesn't just give a "black box" answer; provides the evidence (charts, OES peaks) that led to the conclusion.
