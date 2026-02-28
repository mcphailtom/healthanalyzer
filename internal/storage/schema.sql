CREATE TABLE IF NOT EXISTS entries (
    id            TEXT PRIMARY KEY,
    date          TEXT NOT NULL,
    category      TEXT NOT NULL,
    input_text    TEXT NOT NULL,
    analysis_text TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_entries_date_category ON entries(date, category);
CREATE INDEX IF NOT EXISTS idx_entries_category ON entries(category);

CREATE TABLE IF NOT EXISTS findings (
    id           TEXT PRIMARY KEY,
    date         TEXT NOT NULL,
    finding_text TEXT NOT NULL,
    categories   TEXT NOT NULL DEFAULT '[]',
    created_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_findings_date ON findings(date);

CREATE TABLE IF NOT EXISTS summaries (
    id           TEXT PRIMARY KEY,
    period_start TEXT NOT NULL,
    period_end   TEXT NOT NULL,
    category     TEXT NOT NULL,
    summary_text TEXT NOT NULL,
    created_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_summaries_category_period ON summaries(category, period_start, period_end);

CREATE TABLE IF NOT EXISTS metrics (
    id         TEXT PRIMARY KEY,
    entry_id   TEXT NOT NULL REFERENCES entries(id),
    key        TEXT NOT NULL,
    value      REAL NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metrics_entry_id ON metrics(entry_id);
CREATE INDEX IF NOT EXISTS idx_metrics_key ON metrics(key);
