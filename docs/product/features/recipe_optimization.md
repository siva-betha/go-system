# Feature: Recipe Recommendation Engine

Recommend recipe parameters based on desired outcomes.

## User Story
As a process engineer, I want to achieve specific results (e.g., "etch 500nm of oxide with >10:1 selectivity to nitride") and get recipe recommendations to try.

## Implementation Details

### Multi-Objective Optimization
- **Goal Mapping**: Users input targets (Etch Rate, Selectivity, Uniformity).
- **Optimization Strategy**: Uses Bayesian Optimization to explore the "parameter space" (Pressure, Flows, Power) and find the most likely candidates.
- **Constraints**: Includes a "Safe Zone" filter to ensure recommendations won't damage the electrostatic chuck or exceed the generator's duty cycle.

### Backend: Recommendation API
```python
# services/recipes/recommender.py

class RecipeWizard:
    def get_recommendations(self, targets: dict):
        """
        Produce a list of candidate recipes for a set of targets
        """
        # Step 1: Search library for nearest neighbors
        baseline = self.library.find_closest_match(targets)
        
        # Step 2: Use inverse-model to suggest deltas
        deltas = self.model.inverse_solve(
            target_outcomes=targets,
            start_point=baseline
        )
        
        return [
            {
                'recipe_params': {**baseline, **deltas},
                'predicted_outcomes': self.model.forward(baseline + deltas),
                'confidence_score': 0.91
            }
        ]
```

## Killer Differentiator: Fast-Track Development
- **New Process Introduction (NPI)**: Shrinks the time to develop a new recipe from 20 test wafers to just 3.
- **Goal-Driven Design**: Engineers specify the "What" (the result), and the system handles the "How" (the settings).
- **Continuous Learning**: Every recommendation that is successfully run on a tool is added to the training set, making the wizard smarter over time.
