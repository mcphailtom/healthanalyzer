package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

//go:embed vec_schema.sql
var vecSchemaSQL string

func init() {
	sqlite_vec.Auto()
}

// EmbeddingDimension is the number of floats in each embedding vector.
// Must match the embedding model output (e.g. 1536 for OpenAI text-embedding-3-small).
const EmbeddingDimension = 1536

// SQLiteStore implements Store using SQLite and sqlite-vec.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at the given path,
// applies the schema, and returns a ready-to-use store.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := applySchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func applySchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("core schema: %w", err)
	}
	if _, err := db.Exec(vecSchemaSQL); err != nil {
		return fmt.Errorf("vec schema: %w", err)
	}
	return nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- Entries ---

func (s *SQLiteStore) SaveEntry(ctx context.Context, entry Entry) error {
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO entries (id, date, category, input_text, analysis_text, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		entry.ID,
		entry.Date.Format(time.DateOnly),
		entry.Category,
		entry.InputText,
		entry.AnalysisText,
		entry.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetEntriesByDateRange(ctx context.Context, category string, from, to time.Time) ([]Entry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, date, category, input_text, analysis_text, created_at
		 FROM entries
		 WHERE category = ? AND date >= ? AND date <= ?
		 ORDER BY date DESC`,
		category,
		from.Format(time.DateOnly),
		to.Format(time.DateOnly),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEntries(rows)
}

func (s *SQLiteStore) GetEntryByID(ctx context.Context, id string) (Entry, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, date, category, input_text, analysis_text, created_at
		 FROM entries WHERE id = ?`, id)

	return scanEntry(row)
}

// --- Findings ---

func (s *SQLiteStore) SaveFinding(ctx context.Context, finding Finding) error {
	if finding.ID == "" {
		finding.ID = uuid.NewString()
	}
	if finding.CreatedAt.IsZero() {
		finding.CreatedAt = time.Now().UTC()
	}

	categoriesJSON, err := json.Marshal(finding.Categories)
	if err != nil {
		return fmt.Errorf("marshal categories: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO findings (id, date, finding_text, categories, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		finding.ID,
		finding.Date.Format(time.DateOnly),
		finding.FindingText,
		string(categoriesJSON),
		finding.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetRecentFindings(ctx context.Context, limit int) ([]Finding, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, date, finding_text, categories, created_at
		 FROM findings
		 ORDER BY date DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFindings(rows)
}

// --- Summaries ---

func (s *SQLiteStore) SaveSummary(ctx context.Context, summary Summary) error {
	if summary.ID == "" {
		summary.ID = uuid.NewString()
	}
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO summaries (id, period_start, period_end, category, summary_text, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		summary.ID,
		summary.PeriodStart.Format(time.DateOnly),
		summary.PeriodEnd.Format(time.DateOnly),
		summary.Category,
		summary.SummaryText,
		summary.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetSummaries(ctx context.Context, category string, from, to time.Time) ([]Summary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, period_start, period_end, category, summary_text, created_at
		 FROM summaries
		 WHERE category = ? AND period_start >= ? AND period_end <= ?
		 ORDER BY period_start DESC`,
		category,
		from.Format(time.DateOnly),
		to.Format(time.DateOnly),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSummaries(rows)
}

// --- Metrics ---

func (s *SQLiteStore) SaveMetrics(ctx context.Context, entryID string, metrics []Metric) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO metrics (id, entry_id, key, value, created_at)
		 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, m := range metrics {
		id := m.ID
		if id == "" {
			id = uuid.NewString()
		}
		if _, err := stmt.ExecContext(ctx, id, entryID, m.Key, m.Value, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetMetrics(ctx context.Context, category string, key string, from, to time.Time) ([]Metric, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT m.id, m.entry_id, m.key, m.value, m.created_at
		 FROM metrics m
		 JOIN entries e ON e.id = m.entry_id
		 WHERE e.category = ? AND m.key = ? AND e.date >= ? AND e.date <= ?
		 ORDER BY e.date DESC`,
		category,
		key,
		from.Format(time.DateOnly),
		to.Format(time.DateOnly),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMetrics(rows)
}

// --- Embeddings ---

func (s *SQLiteStore) SaveEmbedding(ctx context.Context, sourceID string, sourceType string, embedding []float32) error {
	blob, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("serialize embedding: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO vec_embeddings (source_id, source_type, embedding)
		 VALUES (?, ?, ?)`,
		sourceID, sourceType, blob,
	)
	return err
}

func (s *SQLiteStore) SearchSimilar(ctx context.Context, embedding []float32, sourceType string, limit int) ([]SimilarResult, error) {
	blob, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return nil, fmt.Errorf("serialize query embedding: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT v.source_id, v.distance,
		        e.id, e.date, e.category, e.input_text, e.analysis_text, e.created_at
		 FROM vec_embeddings v
		 JOIN entries e ON e.id = v.source_id
		 WHERE v.embedding MATCH ?
		   AND v.source_type = ?
		   AND k = ?`,
		blob, sourceType, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SimilarResult
	for rows.Next() {
		var (
			sr        SimilarResult
			sourceID  string
			dateStr   string
			createdAt string
		)
		if err := rows.Scan(
			&sourceID, &sr.Distance,
			&sr.Entry.ID, &dateStr, &sr.Entry.Category,
			&sr.Entry.InputText, &sr.Entry.AnalysisText, &createdAt,
		); err != nil {
			return nil, err
		}
		sr.Entry.Date, _ = time.Parse(time.DateOnly, dateStr)
		sr.Entry.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		results = append(results, sr)
	}
	return results, rows.Err()
}

// --- scan helpers ---

func scanEntries(rows *sql.Rows) ([]Entry, error) {
	var entries []Entry
	for rows.Next() {
		e, err := scanEntryFromRows(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func scanEntryFromRows(rows *sql.Rows) (Entry, error) {
	var e Entry
	var dateStr, createdAt string
	if err := rows.Scan(&e.ID, &dateStr, &e.Category, &e.InputText, &e.AnalysisText, &createdAt); err != nil {
		return Entry{}, err
	}
	e.Date, _ = time.Parse(time.DateOnly, dateStr)
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return e, nil
}

func scanEntry(row *sql.Row) (Entry, error) {
	var e Entry
	var dateStr, createdAt string
	if err := row.Scan(&e.ID, &dateStr, &e.Category, &e.InputText, &e.AnalysisText, &createdAt); err != nil {
		return Entry{}, err
	}
	e.Date, _ = time.Parse(time.DateOnly, dateStr)
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return e, nil
}

func scanFindings(rows *sql.Rows) ([]Finding, error) {
	var findings []Finding
	for rows.Next() {
		var f Finding
		var dateStr, createdAt, categoriesJSON string
		if err := rows.Scan(&f.ID, &dateStr, &f.FindingText, &categoriesJSON, &createdAt); err != nil {
			return nil, err
		}
		f.Date, _ = time.Parse(time.DateOnly, dateStr)
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		_ = json.Unmarshal([]byte(categoriesJSON), &f.Categories)
		findings = append(findings, f)
	}
	return findings, rows.Err()
}

func scanSummaries(rows *sql.Rows) ([]Summary, error) {
	var summaries []Summary
	for rows.Next() {
		var s Summary
		var periodStart, periodEnd, createdAt string
		if err := rows.Scan(&s.ID, &periodStart, &periodEnd, &s.Category, &s.SummaryText, &createdAt); err != nil {
			return nil, err
		}
		s.PeriodStart, _ = time.Parse(time.DateOnly, periodStart)
		s.PeriodEnd, _ = time.Parse(time.DateOnly, periodEnd)
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

func scanMetrics(rows *sql.Rows) ([]Metric, error) {
	var metrics []Metric
	for rows.Next() {
		var m Metric
		var createdAt string
		if err := rows.Scan(&m.ID, &m.EntryID, &m.Key, &m.Value, &createdAt); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// Ensure SQLiteStore implements Store at compile time.
var _ Store = (*SQLiteStore)(nil)
