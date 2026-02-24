# Feature: Automated Shift Reports

Generate automated shift summary reports for handover between shifts.

## User Story
As a shift engineer, I want to quickly see what happened during the previous shift and what needs attention.

## Implementation Details

### Auto-Aggregation Engine
- **Event Summary**: Counts and categorizes all alarms, scrap events, and tool downtimes that occurred in a specific 8 or 12-hour window.
- **KPI Snapshots**: Captures yield and utilization trends for the top tool fleets.

### Backend: Reporting Service
```python
# services/reports/shift_handover.py

class ShiftReportGenerator:
    def generate(self, shift_times: tuple):
        """
        Compile shift summary from multi-source data
        """
        events = self.db.events.fetch_range(*shift_times)
        yield_data = self.influx.get_fleet_yield(*shift_times)
        
        report = {
            'total_runs': len(events.filter(type='RUN')),
            'critical_excursions': events.filter(severity='CRITICAL'),
            'top_unhealthy_tools': self.health_engine.calculate_bottom_fleet(yield_data),
            'open_tickets': self.collaboration.get_unresolved_mentions()
        }
        
        return self.pdf_engine.render('shift_summary.html', data=report)
```

## Killer Differentiator: Seamless Handovers
- **Eliminates Manual Logging**: Engineers spend their time fixing tools, not writing Word docs.
- **Facts over Feelings**: Handover is based on actual process data snapshots, not just verbal summaries.
- **Executive Visibility**: Reports can be automatically cc'd to management, ensuring everyone is aligned on the fab's status.
