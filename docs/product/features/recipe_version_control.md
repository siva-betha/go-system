# Feature: Recipe Version Control & History

Create a Git-like version control system for recipes with full history, branching, and rollback capabilities.

## User Story
As a process engineer, I want to track every change to recipes, understand who changed what and when, and easily roll back to previous versions when experiments fail.

## Key Capabilities
- **Every recipe change is committed** with a mandatory message explaining the rationale.
- **Branching** for experimental variants (e.g., "Branch_LowPressure_Test") without affecting the master production recipe.
- **Tagging** for releases (e.g., "Production", "Qualified", "Golden").
- **Visual Diff view** between any two versions with highlighted parameter changes.
- **Impact analysis** before deploying changes to a physical tool.
- **Approval workflow** requiring peer or manager sign-off for production recipes.

## Implementation

### Backend: Versioning Schema
Recipes are stored in a relational database with a linked `recipe_history` table. Each entry in `recipe_history` represents a snapshot of the recipe parameters.

```python
# services/recipes/versioning.py

class RecipeVersionManager:
    async def commit_change(self, recipe_id: str, changes: dict, 
                            message: str, user_id: str):
        """
        Creates a new version of the recipe and updates the pointer
        """
        async with self.db.transaction():
            # Get current version
            current = await self.db.recipes.find_one({"id": recipe_id})
            
            # Create new history entry
            new_version_num = current['version'] + 1
            history_entry = {
                "recipe_id": recipe_id,
                "version": new_version_num,
                "parameters": {**current['parameters'], **changes},
                "commit_message": message,
                "created_by": user_id,
                "created_at": datetime.utcnow()
            }
            await self.db.recipe_history.insert_one(history_entry)
            
            # Update main recipe pointer
            await self.db.recipes.update_one(
                {"id": recipe_id},
                {"$set": {"version": new_version_num, "parameters": history_entry['parameters']}}
            )
            
            return new_version_num

    async def create_branch(self, source_recipe_id: str, branch_name: str):
        """
        Clone a recipe into a new 'branch' recipe
        """
        source = await self.db.recipes.find_one({"id": source_recipe_id})
        branch = source.copy()
        branch['id'] = str(uuid.uuid4())
        branch['name'] = f"{source['name']} ({branch_name})"
        branch['parent_id'] = source_recipe_id
        branch['is_branch'] = True
        
        await self.db.recipes.insert_one(branch)
        return branch['id']
```

## Killer Differentiator: Software Rigor for Hardware
- **Eliminates "Silent" Changes**: No parameter can change without an entry in the audit log.
- **Experimental Safety**: Engineers can experiment on branches without risking production downtime.
- **Compliance Ready**: Automated reports for CQ (Chamber Qualification) showing exactly which recipe version was used for every wafer.
- **Collaborative**: Multiple engineers can suggest changes via "Proposal" branches that get reviewed before merging.
