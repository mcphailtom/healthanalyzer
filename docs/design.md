# HealthAnalyzer -- Technical Design

## Architecture

HealthAnalyzer uses a multi-agent architecture to track and analyse daily health data across categories (food, sleep, activity, and user-defined categories added over time).

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
    +-- Food Sub-Agent
    +-- Sleep Sub-Agent
    +-- Activity Sub-Agent
    +-- ... (extensible via registry)
```

The **orchestrator** receives raw user input, determines which categories are relevant, and dispatches to the appropriate sub-agents. Each sub-agent:

1. Retrieves recent chronological history from storage
2. Retrieves semantically similar past entries via vector search
3. Retrieves periodic summaries for longer-term context
4. Assembles a prompt with this context and the day's input
5. Calls the LLM for analysis
6. Extracts metrics from the analysis
7. Persists the entry, metrics, and embedding

The orchestrator then collects all sub-agent outputs, retrieves recent cross-category findings, and produces an overall analysis identifying trends and correlations.

### Adding New Categories

New categories are added by implementing the `SubAgent` interface and registering via the `category.Register()` function. No schema changes or configuration is required -- categories are identified by a string label.

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
- **Hybrid retrieval.** Sub-agents use both temporal queries (recent N days) and semantic similarity (vector search) to assemble context.
- **Metrics table** stores extracted key-value numeric data (e.g., `sleep_hours: 7.5`) for potential future charting and trend visualisation without requiring re-analysis.

### Vectors and Embeddings

An embedding is a fixed-length array of floats that represents the semantic meaning of a text. Similar texts produce similar embeddings. The sqlite-vec extension enables nearest-neighbour search over these embeddings directly within SQLite.

Retrieval flow:
1. Generate an embedding of the current input
2. Query sqlite-vec for the N most similar past entries
3. Return those entries as additional context for the sub-agent

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
    Complete(ctx context.Context, messages []Message) (string, error)
}

type EmbeddingClient interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}
```

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

Sub-agents assemble context from three sources:

1. **Recent history**: Last 7 days of daily entries for the category (chronological)
2. **Similar past entries**: Top 5 semantically similar past entries (vector search)
3. **Periodic summaries**: Relevant weekly/monthly summaries

This hybrid approach balances recency with relevance while keeping token usage predictable.

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
CompletionClient and EmbeddingClient implementations for OpenAI, Anthropic, and Ollama. Provider factory based on configuration.

### Phase 4: Agent Framework
Sub-agent interface, base sub-agent with common retrieval/analysis flow, food/sleep/activity sub-agents, top-level orchestrator with classification, dispatch, and cross-category synthesis.

### Phase 5: Web Interface
HTTP server, HTMX templates, input submission, analysis display, history browsing.

### Phase 6: TUI Interface
Bubbletea models for input, analysis display, and history navigation.

### Phase 7: Summary Generation
Periodic rollup summaries, CLI command, integration with sub-agent context assembly.
