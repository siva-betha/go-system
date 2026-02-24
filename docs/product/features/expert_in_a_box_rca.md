# Feature Deep Dive: "Expert in a Box" (Automated RCA)

The "Expert in a Box" is the platform's premier diagnostic feature, designed to reduce MTTR (Mean Time to Repair) by automatically identifying the physical cause of process excursions.

## 1. How it Works
When a process violation is detected (e.g., SPC alert, scrap precursor), the RCA service kicks in:
1.  **Excursion Trigger**: High-priority alert marks a specific wafer/lot.
2.  **Multivariate Comparison**: The system compares the anomalous OES trace against the "Golden Fleet Average."
3.  **Feature Importance (XAI)**: Using SHAP values, it identifies which sensors or chemical species (peaks) deviated most.
4.  **Heuristic Mapping**: Maps the data deviation to a physical component using a pre-trained probability model.

## 2. Probable Cause Ranking
The result is a ranked list of causes delivered to the engineer:
- **RF Matching Network (82%)**: "Bias reflected power spike at Step 3 correlates with electrode wear."
- **Gas Panel Drift (15%)**: "OES Nitrogen lines elevated; potential MFC calibration issue."

## 3. Recommended Actions
The system doesn't just "alert"—it "consults":
> **Recommendation**: "Inspect the RF match capacitors for signs of arcing. Replace the 486nm filter if intensity remains <30% of Golden value."

## 4. Institutional Knowledge Loop
Every time an engineer resolves a ticket, they can "Confirm" or "Correct" the AI diagnosis. This feedback loop ensures the "Expert" grows smarter with every fab incident.
