# Feature: Recipe Impact Predictor

AI-powered tool that predicts the impact of recipe changes before running them.

## User Story
As a process engineer, I want to simulate how changing a parameter (e.g., increasing CF4 flow by 10%) will affect endpoint time, uniformity, and OES signatures before I actually run the wafer.

## Implementation Details

### ML Engine Strategy
- **Training Data**: Historical recipe variations (DOE - Design of Experiments) and their corresponding outcomes.
- **Model Architecture**: A multi-head neural network that predicts both scalar values (endpoint time, etch rate) and vector data (OES spectra).
- **Calibrated Physics**: Combines data-driven models with physics-based plasma equations to ensure predictions stay within physical bounds.

### Backend: Prediction API
```python
# services/analytics/predictor.py

class RecipeImpactPredictor:
    def __init__(self, model_service):
        self.model = model_service

    async def predict_impact(self, base_recipe_id: str, 
                             parameter_tweaks: dict):
        """
        Simulate the effect of recipe changes
        """
        base_recipe = await self.db.recipes.find_one({"id": base_recipe_id})
        target_parameters = {**base_recipe['parameters'], **parameter_tweaks}
        
        # Get baseline performance
        baseline_outcomes = await self.get_historical_outcomes(base_recipe_id)
        
        # Run inference
        prediction = await self.model.predict(
            features=target_parameters,
            context={'chamber_id': base_recipe['chamber_id']}
        )
        
        return {
            'predicted_endpoint_time': prediction['endpoint_time'],
            'confidence_interval': prediction['confidence'],
            'oes_shift': self.calculate_spectral_shift(
                base_recipe['baseline_oes'], 
                prediction['predicted_oes']
            ),
            'risk_score': self.calculate_risk(prediction)
        }
```

## Killer Differentiator: Virtual R&D
- **"What-If" Analysis**: Instantly see how a pressure drop will affect selectivity without wasting a $5,000 test wafer.
- **Sensitivity Analysis**: Identifies which parameters are the most sensitive for a given recipe, helping engineers tighten controls where they matter most.
- **Optimal Suggestion**: Not just predicting impact, but suggesting the "closest" change to achieve a desired target outcome (e.g., "reduce endpoint time by 5s while maintaining uniformity").
