package storage

import "time"

// Entry represents a single health data submission and its analysis.
type Entry struct {
	ID           string
	Date         time.Time
	Category     string
	InputText    string
	AnalysisText string
	CreatedAt    time.Time
}

// Finding represents a cross-category insight from the top-level agent.
type Finding struct {
	ID          string
	Date        time.Time
	FindingText string
	Categories  []string
	CreatedAt   time.Time
}

// Summary represents a periodic rollup of analyses for a category.
type Summary struct {
	ID          string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Category    string
	SummaryText string
	CreatedAt   time.Time
}

// Metric represents a single extracted numeric measurement from an analysis.
type Metric struct {
	ID        string
	EntryID   string
	Key       string
	Value     float64
	CreatedAt time.Time
}

// SimilarResult pairs a stored entry with its similarity score from vector search.
type SimilarResult struct {
	Entry    Entry
	Distance float32
}
