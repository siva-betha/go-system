# Feature: Predictive Chamber Health Index (CHI)

## Value Proposition
- **Reduce Downtime**: Predict failures 24-48 hours in advance, allowing for scheduled maintenance.
- **Cost Savings**: Extend the life of consumable parts by replacing them based on actual wear rather than conservative schedules.
- **Yield Protection**: Detect chamber degradation before it results in wafer scrap.

## Technical Implementation
- **Data Source**: OES spectral trends, RF power stability, DC bias drift, and gas flow variance.
- **Algorithm**: LSTM-Autoencoders for sequence anomaly detection + Survival Analysis (Cox Proportional Hazards) for Remaining Useful Life (RUL) estimation.
- **Output**: A "Health Score" (0-100) and "Days to Maintenance" prediction.

## User Experience
- Dedicated dashboard showing CHI for all tool chambers.
- Visual heatmaps identifying which specific chemical species or hardware parameters are driving the health score down.
