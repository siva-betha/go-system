# Feature: Chamber Fleet Performance Dashboard

Executive dashboard showing performance of all chambers across the fab.

## User Story
As a fab manager, I want to see at a glance which chambers are performing well, which are drifting, and where to focus resources.

## Visual Interface
┌─────────────────────────────────────────────────────────────┐
│                   FLEET PERFORMANCE DASHBOARD               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Heat Map: Chamber Health by Location                │  │
│  │  ┌────────────────────────────────────────────────┐   │  │
│  │  │  [Fab Layout with color-coded chambers]        │   │  │
│  │  │  Green = Healthy                                │   │  │
│  │  │  Yellow = Drifting                              │   │  │
│  │  │  Red = Needs attention                           │   │  │
│  │  └────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │ Chamber 1   │ Chamber 2   │ Chamber 3   │ Chamber 4   │  │
│  │ Health: 98% │ Health: 87% │ Health: 92% │ Health: 65% │  │
│  │ Trend:  ▲   │ Trend:  ▼   │ Trend:  →   │ Trend:  ▼   │  │
│  │ Runs: 1245  │ Runs: 987   │ Runs: 1567  │ Runs: 234   │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Top Performers vs Bottom Performers                  │  │
│  │  ┌────────────────────────────────────────────────┐   │  │
│  │  │  Best: Chamber 3 (92%)  Worst: Chamber 4 (65%)   │  │
│  │  │  Key difference: 200 RF hours since cleaning    │   │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘

## Implementation Details

### KPI Scoring Engine
Each chamber is assigned a "Health Score" (0-100) based on:
- **Stability**: Standard deviation of endpoint times for the last 50 runs.
- **Drift**: Distance from the "Golden" fleet baseline.
- **Utilization**: % of time the tool is in "Production" vs "Idle" or "Alarm".

## Killer Differentiator: Portfolio-Level Insights
- **Standardization**: Quickly see if a new process rollout is failing in one specific module but succeeding in others.
- **Resource Optimization**: Tells service teams exactly which tool to go to first for the highest yield impact.
- **Global Benchmarking**: (For multi-site customers) Compare performance between a fab in Singapore vs a fab in Arizona.
