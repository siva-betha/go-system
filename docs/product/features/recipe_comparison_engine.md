# Feature: Recipe Comparison Engine

Create a comprehensive recipe comparison tool that allows engineers to analyze differences between multiple recipes visually and statistically.

## User Story
As a process engineer, I want to compare 2-10 recipes side-by-side to understand how parameter changes affect OES signatures, endpoint times, and process outcomes.

## Visual Interface
┌─────────────────────────────────────────────────────────────┐
│                    RECIPE COMPARISON DASHBOARD               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Recipe Cards (select 2-10 for comparison)           │  │
│  │  [Recipe A] [Recipe B] [Recipe C] [Add Recipe...]    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Parameter Table (color-coded differences)            │  │
│  │  ┌─────────────┬─────────┬─────────┬─────────┐      │  │
│  │  │ Parameter   │ Recipe A│ Recipe B│ % Diff  │      │  │
│  │  ├─────────────┼─────────┼─────────┼─────────┤      │  │
│  │  │ CF4 Flow    │ 100     │ 120     │ +20%    │      │  │
│  │  │ O2 Flow     │ 20      │ 15      │ -25%    │      │  │
│  │  │ Pressure    │ 50      │ 45      │ -10%    │      │  │
│  │  │ Power       │ 1000    │ 1100    │ +10%    │      │  │
│  │  └─────────────┴─────────┴─────────┴─────────┘      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  OES Trace Overlay                                    │  │
│  │  [Legend: Recipe A ____ Recipe B ____ Recipe C ---]   │  │
│  │  ┌────────────────────────────────────────────────┐   │  │
│  │  │                                                │   │  │
│  │  │   Intensity                                    │   │  │
│  │  │        ╱╲     ╱──╲     ╱──╲                   │   │  │
│  │  │       ╱  ╲   ╱    ╲   ╱    ╲                  │   │  │
│  │  │      ╱    ╲ ╱      ╲ ╱      ╲                 │   │  │
│  │  │     ╱      ╲        ╲        ╲                │   │  │
│  │  │    ╱        ╲        ╲        ╲               │   │  │
│  │  └────────────────────────────────────────────────┘   │  │
│  │    200     400     600     800     1000    Wavelength│  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Endpoint Time Comparison                             │  │
│  │  ┌────────────────────────────────────────────────┐   │  │
│  │  │  Recipe A: ████████████████████ 45.2s         │   │  │
│  │  │  Recipe B: ████████████████████████ 58.7s     │   │  │
│  │  │  Recipe C: ████████████████ 38.1s             │   │  │
│  │  └────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Statistical Summary                                  │  │
│  │  • Mean endpoint time: 47.3s ± 8.5s                  │  │
│  │  • Fastest recipe: Recipe C (38.1s)                  │  │
│  │  • Slowest recipe: Recipe B (58.7s)                  │  │
│  │  • Statistically different? Yes (p < 0.05)           │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘

## Implementation

### Backend: Recipe Comparison API
```python
# services/recipes/comparison.py

class RecipeComparisonEngine:
    def __init__(self, influx_client, postgres_client):
        self.influx = influx_client
        self.postgres = postgres_client
        
    async def compare_recipes(self, recipe_ids: List[str], 
                              metrics: List[str] = None):
        """
        Compare multiple recipes across all parameters and outcomes
        """
        # Load recipe parameters from PostgreSQL
        recipes = await self.load_recipe_parameters(recipe_ids)
        
        # Load OES data for each recipe from InfluxDB
        oes_data = await self.load_oes_data(recipe_ids)
        
        # Calculate comparison metrics
        comparison = {
            'recipes': recipes,
            'parameters': self.compare_parameters(recipes),
            'oes': self.compare_oes_traces(oes_data),
            'endpoints': self.compare_endpoints(oes_data),
            'statistics': self.calculate_statistics(recipes, oes_data),
            'similarity_matrix': self.calculate_similarity(recipes, oes_data)
        }
        
        return comparison
    
    def compare_parameters(self, recipes):
        """Compare recipe parameters with color-coded differences"""
        # Get all parameter names
        all_params = set()
        for recipe in recipes:
            all_params.update(recipe['parameters'].keys())
        
        # Calculate baseline (e.g., first recipe or median)
        baseline = recipes[0]['parameters']
        
        comparison = []
        for param in sorted(all_params):
            row = {'parameter': param}
            values = []
            
            for recipe in recipes:
                value = recipe['parameters'].get(param, 0)
                values.append(value)
                
                # Calculate difference from baseline
                baseline_value = baseline.get(param, 0)
                if baseline_value != 0:
                    pct_diff = ((value - baseline_value) / baseline_value) * 100
                else:
                    pct_diff = 0
                
                row[recipe['name']] = {
                    'value': value,
                    'pct_diff': pct_diff,
                    'color': self.get_color_for_diff(pct_diff)
                }
            
            # Calculate statistics
            row['mean'] = np.mean(values)
            row['std'] = np.std(values)
            row['min'] = np.min(values)
            row['max'] = np.max(values)
            
            comparison.append(row)
        
        return comparison

    def get_color_for_diff(self, pct_diff):
        """Return color code based on percentage difference"""
        if abs(pct_diff) < 5:
            return 'green'
        elif abs(pct_diff) < 15:
            return 'yellow'
        elif abs(pct_diff) < 30:
            return 'orange'
        else:
            return 'red'
```

### Frontend: Recipe Comparison Component
```jsx
// frontend/components/recipes/RecipeComparison.jsx
import { useState, useEffect } from 'react';
import { Grid, Card, Table, Badge } from '@mui/material';
import Plot from 'react-plotly.js';

export default function RecipeComparison({ recipeIds }) {
  const [comparison, setComparison] = useState(null);
  const [selectedWavelengths, setSelectedWavelengths] = useState([685.6, 288.1]);
  
  useEffect(() => {
    fetchComparison();
  }, [recipeIds]);
  
  const fetchComparison = async () => {
    const res = await fetch('/api/recipes/compare', {
      method: 'POST',
      body: JSON.stringify({ recipe_ids: recipeIds })
    });
    const data = await res.json();
    setComparison(data);
  };
  
  if (!comparison) return <div>Loading...</div>;
  
  return (
    <div className="recipe-comparison">
      <h1>Recipe Comparison</h1>
      
      {/* Recipe Cards */}
      <Grid container spacing={2}>
        {comparison.recipes.map(recipe => (
          <Grid item xs={3} key={recipe.id}>
            <Card className="recipe-card">
              <h3>{recipe.name}</h3>
              <p>Version: {recipe.version}</p>
              <p>Created: {new Date(recipe.created_at).toLocaleDateString()}</p>
              <Badge color="primary">Active</Badge>
            </Card>
          </Grid>
        ))}
      </Grid>
      
      {/* Parameter Comparison Table */}
      <Card className="parameter-table">
        <h2>Parameter Comparison</h2>
        <table>
          <thead>
            <tr>
              <th>Parameter</th>
              {comparison.recipes.map(r => <th key={r.id}>{r.name}</th>)}
              <th>Mean</th>
            </tr>
          </thead>
          <tbody>
            {comparison.parameters.map(row => (
              <tr key={row.parameter}>
                <td>{row.parameter}</td>
                {comparison.recipes.map(r => (
                  <td key={r.id} style={{
                    backgroundColor: row[r.name]?.color === 'red' ? '#ffebee' :
                                    row[r.name]?.color === 'orange' ? '#fff3e0' :
                                    row[r.name]?.color === 'yellow' ? '#fffde7' : 'transparent'
                  }}>
                    {row[r.name]?.value}
                  </td>
                ))}
                <td>{row.mean.toFixed(1)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
```

## Killer Differentiator: What Makes This Special?
- **No other tool** offers side-by-side recipe comparison with OES overlay.
- **Color-coded differences** make it instantly obvious what changed.
- **Statistical significance** tells you if differences matter.
- **Similarity matrix** helps group recipes by behavior, not just parameters.
