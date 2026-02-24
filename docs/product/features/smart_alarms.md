# Feature: Scrap Prevention Alerts

Real-time alerts when process signatures indicate impending scrap.

## User Story
As a process engineer, I want to be alerted immediately when a chamber is about to produce scrap, so I can abort the run.

## Implementation Details

### Real-Time Signature Detection
- **Pattern Matching**: Constantly compares the live data stream against a library of "Scrap Signatures" (e.g., specific OES dropouts, RF reflected power spikes).
- **Early Exit Logic**: If a 99% match for a scrap event is detected within the first 10 seconds of a step, the tool sends an abort signal.

### Backend: Alerting Engine
```python
# services/alerts/scrap_prevention.py

class ScrapAlertManager:
    async def monitor_live_stream(self, stream_data):
        """
        Check live OES/PLC data for scrap precursors
        """
        match_score = self.matcher.calculate_similarity(
            stream_data, 
            SignatureLibrary.SCRAP_PATTERNS
        )
        
        if match_score > 0.95:
            await self.trigger_emergency_alert(
                type="SCRAP_RISK",
                severity="CRITICAL",
                action_required="IMMEDIATE_ABORT"
            )
```

## Killer Differentiator: Stop the Bleeding
- **Zero-Latency Response**: Alerts are triggered at the edge (near the tool) using optimized C++ inference engines, ensuring response times <20ms.
- **Prevented-Value Tracker**: Shows management a real-time dollar value of scrap prevented by the system (e.g., "$450k of material saved this month").
- **Multi-Channel Escalation**: If the primary engineer doesn't acknowledge within 60 seconds, the alert is automatically escalated to the shift manager via pager and SMS.
