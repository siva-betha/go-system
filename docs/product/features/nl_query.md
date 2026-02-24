# Feature: Natural Language Process Query

Query your process data using natural language.

## User Story
As a process engineer, I want to ask questions like "Show me all wafers with endpoint time > 60s last week" and get instant visualizations.

## Implementation Details

### NLP-to-Query Engine
- **LLM Integration**: Uses a fine-tuned Large Language Model (like GPT-4 or a custom Llama variant) that understands semiconductor domain terminology.
- **Schema Mapping**: The model translates natural language into InfluxQL (for time-series data) and SQL (for metadata).

### Backend: Query Translator
```python
# services/ai/nlp_query.py

class ProcessQueryAssistant:
    def process_request(self, user_question: str):
        """
        Translate English question to executable code and viz
        """
        # Step 1: Extract intent and filters
        query_plan = self.llm.generate_plan(user_question)
        
        # Step 2: Execute data fetch
        data = self.executor.run(query_plan)
        
        # Step 3: Suggest visualization type
        return {
            'data': data,
            'viz_type': 'ScatterPlot' if 'correlation' in user_question else 'Histogram',
            'explanation': "I found 42 wafers where the CF4 flow was above 110sccm."
        }
```

## Killer Differentiator: Instant Engineering
- **Democratizes Data**: No need to learn SQL or complex BI tools. Any technician can "ask" the data a question.
- **Voice-Enabled**: Can be integrated with AR headsets or mobile apps for hands-free queries on the fab floor.
- **Contextual Learning**: The system learns from engineer corrections (e.g., "Actually, I wanted to see Chamber 4, not 3").
