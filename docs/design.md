# HealthAnalyzer -- Technical Design

## Architecture

HealthAnalyzer uses a multi-agent architecture to track and analyse daily health data across categories. Each category has a dedicated sub-agent responsible for analysis, and a top-level orchestrator coordinates cross-category insights.

### Agent Hierarchy

```
User Input
    |
    v
Top-Level Agent (Orchestrator)
    |-- classifies input into categories
    |-- dispatches to relevant sub-agents
    |-- collects sub-agent outputs
    |-- synthesises cross-category analysis
    |-- persists findings
    |
    +-- Sleep Sub-Agent
    +-- Macro Nutrients Sub-Agent
    +-- Cycle Tracker Sub-Agent
    +-- Digestion Sub-Agent
    +-- Weight Sub-Agent
    +-- Feelings Sub-Agent
    +-- Energy Sub-Agent
    +-- Mind Sub-Agent
    +-- Activity Sub-Agent
    +-- ... (extensible via registry)
```

### Categories

| Category | Input Style | Example Metrics |
|----------|-------------|-----------------|
| **Sleep** | Manual entry or device import | total_hours, rem_minutes, deep_minutes, awake_count |
| **Macro Nutrients** | Free text (photo support planned) | calories, protein_g, carbs_g, fat_g, fibre_g |
| **Cycle Tracker** | Day number, symptoms | cycle_day, phase |
| **Digestion** | Symptom selection or free text, tied to meals | symptom_type, severity |
| **Weight** | Single number | weight_kg |
| **Feelings** | Free text or selection from: ok, bloated, gassy, heartburn, nauseous, fine, happy, sad, sensitive, angry, confident, excited, irritable, anxious, insecure, grateful, indifferent | mood_score |
| **Energy** | Free text or selection from: exhausted, tired, ok, energetic, fully energised | energy_level |
| **Mind** | Free text or selection from: forgetful, brain fog, calm, stressed, focused, distracted, motivated, unmotivated, creative, productive, unproductive | mental_state_score |
| **Activity** | Free text | activity_type, duration_minutes, intensity |

Digestion entries are correlated with meals via the agent's analysis text rather than explicit foreign keys -- the agent references the relevant meal naturally.

Cycle tracking is a high-value cross-category signal. The orchestrator correlates cycle phase with patterns in mood, energy, mind, digestion, and weight.

### Adding New Categories

New categories are added by implementing the `SubAgent` interface and registering via the `category.Register()` function. No schema changes or configuration is required -- categories are identified by a string label. Each sub-agent provides its own system prompt and metric extraction rules.

## Agent Tools

Agents are equipped with tools that allow them to dynamically query stored data during analysis. This enables agents to make retrieval decisions based on what they observe in the current input rather than relying solely on pre-assembled context.

### Tool Design

Each sub-agent and the orchestrator receive a set of tools they can call during an LLM completion. The tools are thin wrappers around the storage interface:

```go
// AgentTool defines a function an agent can call during analysis.
type AgentTool struct {
    Name        string
    Description string
    Parameters  map[string]ToolParameter
    Execute     func(ctx context.Context, args map[string]any) (string, error)
}
```

### Available Tools

**Sub-agent tools** (available to all sub-agents):

| Tool | Purpose |
|------|---------|
| `get_recent_entries` | Fetch entries for this category within a date range |
| `search_similar` | Semantic search for entries similar to a given text |
| `get_metrics` | Fetch historical metric values for this category (e.g., sleep_hours over the last 30 days) |
| `get_summaries` | Retrieve periodic summaries for this category |

**Orchestrator tools** (available to the top-level agent):

| Tool | Purpose |
|------|---------|
| `get_category_entries` | Fetch entries for any category within a date range |
| `get_category_metrics` | Fetch metrics for any category (enables cross-category correlation) |
| `get_findings` | Retrieve past cross-category findings |
| `search_all` | Semantic search across all categories |

### Analysis Flow

With tools, the analysis flow becomes a conversation rather than a single prompt-response:

1. Sub-agent receives today's input and its system prompt
2. An initial context window is provided: last 7 days of entries + top 5 semantically similar entries + relevant summaries
3. The agent analyses the input and may call tools for additional data ("let me check what happened on the same cycle day last month" or "let me look at sleep data from the last time this energy pattern appeared")
4. The agent produces its analysis text and extracted metrics
5. Entry, metrics, and embedding are persisted

This approach lets the agent decide what historical context is relevant based on what it sees, rather than the application pre-guessing.

### Future: External Reference Tools

A future enhancement may add tools for external reference data (nutritional databases, hormone cycle phase calculators, exercise physiology references). These would be added as additional tools without changing the agent framework. This is not planned for initial implementation but the tool interface is designed to accommodate it.

## System Prompts

Each sub-agent's behaviour is governed by a system prompt that defines:

- What the category covers and how to interpret input
- Metric extraction rules (which metrics to extract, how to scale qualitative values to numeric)
- Analysis guidelines (what patterns to look for, what to flag)
- Output format expectations

System prompts are stored as embedded text files alongside each sub-agent implementation:

```
internal/category/sleep/prompt.txt
internal/category/nutrition/prompt.txt
internal/category/digestion/prompt.txt
internal/category/cycle/prompt.txt
internal/category/weight/prompt.txt
internal/category/feelings/prompt.txt
internal/category/energy/prompt.txt
internal/category/mind/prompt.txt
internal/category/activity/prompt.txt
```

This means prompt refinement is a text change, not a code change. Prompts are version-controlled and reviewable.

The orchestrator also has its own system prompt that directs cross-category correlation. Key correlation rules include:

- Cycle phase against mood, energy, digestion, and weight patterns
- Sleep quality against next-day energy and cognitive function
- Macro nutrient patterns against weight trends and digestion symptoms
- Activity levels against sleep quality and energy

## Storage

All persistence uses a single SQLite database file with the sqlite-vec extension for vector search.

### Schema

```sql
-- Raw entries and their analyses
CREATE TABLE entries (
    id            TEXT PRIMARY KEY,
    date          TEXT NOT NULL,          -- ISO 8601 date
    category      TEXT NOT NULL,
    input_text    TEXT NOT NULL,
    analysis_text TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL
);

-- Cross-category insights from the top-level agent
CREATE TABLE findings (
    id           TEXT PRIMARY KEY,
    date         TEXT NOT NULL,
    finding_text TEXT NOT NULL,
    categories   TEXT NOT NULL DEFAULT '[]',  -- JSON array of category names
    created_at   TEXT NOT NULL
);

-- Periodic rollup summaries
CREATE TABLE summaries (
    id           TEXT PRIMARY KEY,
    period_start TEXT NOT NULL,
    period_end   TEXT NOT NULL,
    category     TEXT NOT NULL,
    summary_text TEXT NOT NULL,
    created_at   TEXT NOT NULL
);

-- Extracted numeric metrics from analyses
CREATE TABLE metrics (
    id         TEXT PRIMARY KEY,
    entry_id   TEXT NOT NULL REFERENCES entries(id),
    key        TEXT NOT NULL,
    value      REAL NOT NULL,
    created_at TEXT NOT NULL
);

-- Vector embeddings for semantic search (sqlite-vec virtual table)
CREATE VIRTUAL TABLE vec_embeddings USING vec0(
    source_id   TEXT,
    source_type TEXT,
    embedding   float[1536]  -- dimension matches embedding model output
);
```

### Design Principles

- **No category-specific columns.** All categories share the same table structure. The LLM agents perform analysis; SQL handles organisation and retrieval only.
- **Hybrid retrieval.** Sub-agents use both temporal queries (recent N days) and semantic similarity (vector search) to assemble initial context, then can dynamically query for more via tools.
- **Metrics table** stores extracted key-value numeric data (e.g., `sleep_hours: 7.5`, `energy_level: 4`) for charting, trend visualisation, and agent tool queries.

### Vectors and Embeddings

An embedding is a fixed-length array of floats that represents the semantic meaning of a text. Similar texts produce similar embeddings. The sqlite-vec extension enables nearest-neighbour search over these embeddings directly within SQLite.

Retrieval flow:
1. Generate an embedding of the current input
2. Query sqlite-vec for the N most similar past entries
3. Return those entries as additional context for the sub-agent
4. The agent may perform additional semantic searches via tools during analysis

This allows retrieval by meaning ("find entries similar to this pattern") rather than only by date or category.

## LLM Providers

The application supports mixed provider configuration: one provider for text completion (agent reasoning) and another for embedding generation.

### Supported Providers

| Provider  | Completion | Embeddings | Notes |
|-----------|:----------:|:----------:|-------|
| OpenAI    | Yes        | Yes        | GPT-4o, text-embedding-3-small |
| Anthropic | Yes        | No         | Claude Sonnet/Haiku; pair with another provider for embeddings |
| Ollama    | Yes        | Yes        | Local models; OpenAI-compatible API on localhost:11434 |

### Interface

```go
type CompletionClient interface {
    Complete(ctx context.Context, messages []Message, tools []AgentTool) (string, error)
}

type EmbeddingClient interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}
```

The `Complete` method accepts tools so that the provider handles tool-call loops (the LLM requests a tool call, the provider executes it, feeds the result back, and continues until the LLM produces a final response).

Provider implementations are selected at startup based on TOML configuration. The Ollama implementation reuses the OpenAI SDK pointed at the local endpoint.

### Configuration

```toml
[completion]
provider = "anthropic"
api_key  = "sk-..."
model    = "claude-sonnet-4-20250514"

[embedding]
provider = "openai"
api_key  = "sk-..."
model    = "text-embedding-3-small"
```

API keys can also be set via environment variables (`HEALTHANALYZER_COMPLETION_API_KEY`, `HEALTHANALYZER_EMBEDDING_API_KEY`).

## Interfaces

### Web (Primary)

HTMX-powered web interface served via `healthanalyzer serve`. Uses Go's `net/http` and `html/template` with HTMX for partial page updates.

Routes:
- `GET /` -- input form and recent analysis
- `POST /entries` -- submit health input, returns analysis via HTMX swap
- `GET /history` -- browse past entries
- `GET /history/:date` -- view a specific day

### TUI (Secondary)

Bubbletea terminal interface launched via `healthanalyzer`. Provides the same functionality as the web interface for terminal users.

Both interfaces share all business logic through the agent and storage layers.

## Context Window Management

Sub-agents receive an initial context window assembled before the first LLM call, then can expand their context dynamically via tools.

### Initial Context (Pre-assembled)

1. **Recent history**: Last 7 days of daily entries for the category (chronological)
2. **Similar past entries**: Top 5 semantically similar past entries (vector search)
3. **Periodic summaries**: Relevant weekly/monthly summaries

### Dynamic Context (Agent-driven via tools)

During analysis, the agent can request additional data:
- Entries from specific date ranges
- Metrics for trend analysis
- Semantically similar entries with different search terms
- Summaries from specific periods

This hybrid approach provides a solid baseline of context while allowing the agent to investigate further when patterns warrant it.

### Summary Generation

A summary command (`healthanalyzer summarise --period week`) condenses older daily analyses into weekly or monthly summaries. Summaries are stored with their own embeddings for semantic retrieval. This prevents unbounded context growth as history accumulates.

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/mattn/go-sqlite3` | SQLite driver (CGo, required for sqlite-vec) |
| `github.com/openai/openai-go` | OpenAI API client |
| `github.com/anthropics/anthropic-sdk-go` | Anthropic API client |
| `github.com/stretchr/testify` | Test assertions and suites |
| `github.com/vektra/mockery` | Mock generation |

## Build

Requires a C compiler (CGo) due to the SQLite driver and sqlite-vec extension.

```bash
go build -o healthanalyzer ./cmd/healthanalyzer
```

## Implementation Phases

### Phase 1: Project Foundation
Project scaffold, Go module, directory structure, configuration loading, basic CLI entrypoint.

### Phase 2: Storage Layer
SQLite schema, migrations, repository implementation, sqlite-vec integration, Store interface with full CRUD and vector search.

### Phase 3: LLM Client Layer
CompletionClient and EmbeddingClient implementations for OpenAI, Anthropic, and Ollama. Provider factory based on configuration. Tool-call loop handling within each provider implementation.

### Phase 4: Agent Framework
Agent tool definitions and execution. Sub-agent interface with tool-equipped analysis. Base sub-agent with common retrieval flow. System prompts for all 9 categories. Top-level orchestrator with classification, dispatch, and cross-category synthesis.

### Phase 5: Web Interface
HTTP server, HTMX templates, input submission, analysis display, history browsing.

### Phase 6: TUI Interface
Bubbletea models for input, analysis display, and history navigation.

### Phase 7: Summary Generation
Periodic rollup summaries, CLI command, integration with sub-agent context assembly.

### Phase 8: Multimodal Input
Image support for meal photo analysis. Extends the LLM interface to accept image attachments. Vision model support across providers (GPT-4o, Claude, LLaVA via Ollama). Image storage on disk with file path references in entries.

### Phase 9: Device and App Import
Import sleep and activity data from wearables and health apps (e.g., Apple Health, Fitbit). Import layer normalises device data into natural language descriptions before passing to the relevant sub-agent for analysis. This preserves the agent-driven analysis model while allowing structured data sources.
