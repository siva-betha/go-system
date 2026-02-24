# Feature: Golden Recipe Library

Create a searchable library of "golden" recipes with performance metrics, chamber requirements, and qualification history.

## User Story
As a process engineer, I want to find the best recipe for a given application, see which chambers it runs well on, and understand its historical performance across the entire fab.

## Key Capabilities
- **Advanced Search**: Filter by material (Oxide, Poly-Si, Nitride), target etch rate, selectivity targets, and chamber type.
- **Chamber Compatibility Benchmarking**: View statistical distribution of endpoint times and yields across all chambers running the "Golden" version.
- **Chamber Qualification (CQ)**: Track which specific chambers are qualified to run a recipe and when they last passed qualification.
- **Cloning Logic**: Easily clone a golden recipe to a new experiment with full traceability back to the parent.
- **Annotation & BKM Library**: Shared notes and "Best Known Methods" from engineers on optimizing specific recipes.

## Implementation: Golden Recipe Schema

```sql
-- Database Schema for Golden Library
CREATE TABLE golden_recipes (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    material_app VARCHAR(100), -- e.g., 'STI Etch'
    target_etch_rate FLOAT,
    target_selectivity FLOAT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE recipe_qualifications (
    id UUID PRIMARY KEY,
    recipe_id UUID REFERENCES golden_recipes(id),
    chamber_id VARCHAR(50),
    qualified_date TIMESTAMP,
    expiry_date TIMESTAMP,
    status VARCHAR(20) -- 'Qualified', 'Warning', 'Expired'
);
```

## Killer Differentiator: Institutional Memory
- **Prevents "Recipe Drift"**: Centralized source of truth for high-performance process configurations.
- **Cross-Chamber Transparency**: Identify why specific chambers outperform others even when running identical recipes.
- **Knowledge Transfer**: Accelerates onboarding by giving new engineers access to proven, qualified processes.
