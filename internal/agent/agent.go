package agent

import (
	"context"

	"github.com/mcphailtom/healthanalyzer/internal/storage"
)

// SubAgentInput provides all context a sub-agent needs to analyse a day's input.
type SubAgentInput struct {
	Date          string
	RawInput      string
	RecentHistory []storage.Entry
	SimilarPast   []storage.Entry
	Summaries     []storage.Summary
}

// SubAgentOutput contains the results of a sub-agent's analysis.
type SubAgentOutput struct {
	AnalysisText string
	Metrics      []storage.Metric
}

// SubAgent analyses health data for a specific category.
type SubAgent interface {
	Category() string
	Analyse(ctx context.Context, input SubAgentInput) (SubAgentOutput, error)
}

// AnalysisResult holds the top-level agent's combined analysis.
type AnalysisResult struct {
	CategoryAnalyses map[string]SubAgentOutput
	OverallAnalysis  string
	Finding          *storage.Finding
}

// Orchestrator coordinates input classification, sub-agent dispatch,
// and cross-category analysis.
type Orchestrator interface {
	Analyse(ctx context.Context, date string, input string) (AnalysisResult, error)
}
