# Feature: Custom Dashboard Builder

Drag-and-drop dashboard builder for custom views.

## User Story
As a process engineer, I want to build my own dashboard with exactly the charts and metrics I need for my specific process area.

## Key Capabilities
- **Widget Library**: Pre-built widgets for OES traces, SPC charts, tool health gauges, and 3D waterfall plots.
- **Data Binding**: Simple drag-and-drop selection of data sources (Chambers, Recipes, Lots).
- **Templating**: Share custom-built dashboards with the whole team or save them as fab-wide standards.

## Technical Implementation

### Dynamic Layout Engine
Uses a grid-based layout system (like `react-grid-layout`) where widget configurations are stored as JSON.

```json
{
  "dashboard_id": "oxide_health_01",
  "widgets": [
    {
      "id": "w1",
      "type": "Gauge",
      "source": "ch4.rf_stability",
      "grid": { "x": 0, "y": 0, "w": 4, "h": 2 }
    },
    {
      "id": "w2",
      "type": "OESTraceOverlay",
      "source": "recipe.STI_ETCH_GOLDEN",
      "grid": { "x": 4, "y": 0, "w": 8, "h": 4 }
    }
  ]
}
```

## Killer Differentiator: Personal Productivity
- **Tailored to the Persona**: A maintenance tech needs different data than a yield manager—this tool lets both build what they need.
- **Real-time Updates**: Every widget is natively linked to the live data stream for zero-latency monitoring.
- **Global Deployment**: "Standard" dashboards can be pushed to every monitor on the fab floor to ensure everyone is looking at the same KPIs.
