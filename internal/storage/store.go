package storage

import (
	"context"
	"time"
)

// Store defines all persistence operations for the application.
type Store interface {
	// Entry operations
	SaveEntry(ctx context.Context, entry Entry) error
	GetEntriesByDateRange(ctx context.Context, category string, from, to time.Time) ([]Entry, error)
	GetEntryByID(ctx context.Context, id string) (Entry, error)

	// Finding operations
	SaveFinding(ctx context.Context, finding Finding) error
	GetRecentFindings(ctx context.Context, limit int) ([]Finding, error)

	// Summary operations
	SaveSummary(ctx context.Context, summary Summary) error
	GetSummaries(ctx context.Context, category string, from, to time.Time) ([]Summary, error)

	// Metric operations
	SaveMetrics(ctx context.Context, entryID string, metrics []Metric) error
	GetMetrics(ctx context.Context, category string, key string, from, to time.Time) ([]Metric, error)

	// Vector operations
	SaveEmbedding(ctx context.Context, sourceID string, sourceType string, embedding []float32) error
	SearchSimilar(ctx context.Context, embedding []float32, sourceType string, limit int) ([]SimilarResult, error)

	// Lifecycle
	Close() error
}
