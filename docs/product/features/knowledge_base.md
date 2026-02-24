# Feature: Process Knowledge Base

Build a searchable knowledge base of process issues and solutions.

## User Story
As a process engineer, I want to search for similar issues others have faced and see how they resolved them.

## Implementation Details

### Semantic Search Engine
- **Automated Indexing**: Every resolved "Scrap Alert" or "RCA Report" is automatically tagged and indexed.
- **Natural Language Retrieval**: Uses vector embeddings to find "conceptually similar" issues even if different keywords are used.

### Backend: Knowledge Service
```python
# services/knowledge/wiki.py

class ProcessWiki:
    def search_similar_cases(self, incident_context: dict):
        """
        Find historical solutions based on current process state
        """
        query_vector = self.embedder.encode(incident_context)
        
        # Search vector DB for top matches
        results = self.vector_db.search(query_vector, limit=5)
        
        return [
            {
                'title': r.title,
                'resolution': r.solution_text,
                'similarity': r.score,
                'validated_by': r.author
            } for r in results
        ]
```

## Killer Differentiator: Tribal Knowledge, Digitized
- **Endless Memory**: Prevents the same mistake from being made twice by different shifts.
- **Smart Onboarding**: New engineers can "learn" the history of a specific chamber just by browsing its knowledge base entries.
- **Closing the Loop**: Directly links historical fixes to current live alerts, suggesting a "Proven Solution" in the notification.
