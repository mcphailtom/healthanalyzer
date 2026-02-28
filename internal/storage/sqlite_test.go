package storage

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SQLiteStoreSuite struct {
	suite.Suite
	store *SQLiteStore
	ctx   context.Context
}

func TestSQLiteStoreSuite(t *testing.T) {
	suite.Run(t, new(SQLiteStoreSuite))
}

func (s *SQLiteStoreSuite) SetupTest() {
	store, err := NewSQLiteStore(":memory:")
	s.Require().NoError(err)
	s.store = store
	s.ctx = context.Background()
}

func (s *SQLiteStoreSuite) TearDownTest() {
	if s.store != nil {
		s.store.Close()
	}
}

// --- Entry Tests ---

func (s *SQLiteStoreSuite) TestSaveAndGetEntryByID() {
	entry := Entry{
		ID:           "entry-1",
		Date:         date(2026, 2, 28),
		Category:     "sleep",
		InputText:    "Slept 7 hours, woke up once",
		AnalysisText: "Adequate sleep duration with one interruption.",
	}

	s.Require().NoError(s.store.SaveEntry(s.ctx, entry))

	got, err := s.store.GetEntryByID(s.ctx, "entry-1")
	s.Require().NoError(err)

	s.Equal(entry.ID, got.ID)
	s.Equal(entry.Date.Format(time.DateOnly), got.Date.Format(time.DateOnly))
	s.Equal(entry.Category, got.Category)
	s.Equal(entry.InputText, got.InputText)
	s.Equal(entry.AnalysisText, got.AnalysisText)
	s.False(got.CreatedAt.IsZero())
}

func (s *SQLiteStoreSuite) TestSaveEntryGeneratesID() {
	entry := Entry{
		Date:      date(2026, 2, 28),
		Category:  "sleep",
		InputText: "Slept well",
	}

	s.Require().NoError(s.store.SaveEntry(s.ctx, entry))

	entries, err := s.store.GetEntriesByDateRange(s.ctx, "sleep", date(2026, 2, 28), date(2026, 2, 28))
	s.Require().NoError(err)
	s.Require().Len(entries, 1)
	s.NotEmpty(entries[0].ID)
}

func (s *SQLiteStoreSuite) TestGetEntriesByDateRange() {
	entries := []Entry{
		{ID: "e1", Date: date(2026, 2, 25), Category: "sleep", InputText: "day 1"},
		{ID: "e2", Date: date(2026, 2, 26), Category: "sleep", InputText: "day 2"},
		{ID: "e3", Date: date(2026, 2, 27), Category: "sleep", InputText: "day 3"},
		{ID: "e4", Date: date(2026, 2, 27), Category: "food", InputText: "food day 3"},
		{ID: "e5", Date: date(2026, 2, 28), Category: "sleep", InputText: "day 4"},
	}
	for _, e := range entries {
		s.Require().NoError(s.store.SaveEntry(s.ctx, e))
	}

	got, err := s.store.GetEntriesByDateRange(s.ctx, "sleep", date(2026, 2, 26), date(2026, 2, 27))
	s.Require().NoError(err)
	s.Require().Len(got, 2)
	s.Equal("e3", got[0].ID) // DESC order
	s.Equal("e2", got[1].ID)
}

func (s *SQLiteStoreSuite) TestGetEntriesByDateRangeExcludesOtherCategories() {
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{ID: "e1", Date: date(2026, 2, 27), Category: "sleep", InputText: "sleep"}))
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{ID: "e2", Date: date(2026, 2, 27), Category: "food", InputText: "food"}))

	got, err := s.store.GetEntriesByDateRange(s.ctx, "food", date(2026, 2, 27), date(2026, 2, 27))
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("e2", got[0].ID)
}

func (s *SQLiteStoreSuite) TestGetEntryByIDNotFound() {
	_, err := s.store.GetEntryByID(s.ctx, "nonexistent")
	s.Error(err)
}

// --- Finding Tests ---

func (s *SQLiteStoreSuite) TestSaveAndGetFindings() {
	f := Finding{
		ID:          "f1",
		Date:        date(2026, 2, 28),
		FindingText: "Sleep quality correlates with caffeine intake.",
		Categories:  []string{"sleep", "food"},
	}

	s.Require().NoError(s.store.SaveFinding(s.ctx, f))

	got, err := s.store.GetRecentFindings(s.ctx, 10)
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(f.ID, got[0].ID)
	s.Equal(f.FindingText, got[0].FindingText)
	s.Equal(f.Categories, got[0].Categories)
}

func (s *SQLiteStoreSuite) TestGetRecentFindingsRespectsLimit() {
	for i := range 5 {
		f := Finding{
			ID:          fmtID("f", i),
			Date:        date(2026, 2, 24+i),
			FindingText: "finding",
			Categories:  []string{"sleep"},
		}
		s.Require().NoError(s.store.SaveFinding(s.ctx, f))
	}

	got, err := s.store.GetRecentFindings(s.ctx, 3)
	s.Require().NoError(err)
	s.Len(got, 3)
	// Most recent first
	s.Equal("f4", got[0].ID)
}

// --- Summary Tests ---

func (s *SQLiteStoreSuite) TestSaveAndGetSummaries() {
	sum := Summary{
		ID:          "s1",
		PeriodStart: date(2026, 2, 21),
		PeriodEnd:   date(2026, 2, 27),
		Category:    "sleep",
		SummaryText: "Average 7.2 hours, REM improving.",
	}

	s.Require().NoError(s.store.SaveSummary(s.ctx, sum))

	got, err := s.store.GetSummaries(s.ctx, "sleep", date(2026, 2, 1), date(2026, 2, 28))
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(sum.SummaryText, got[0].SummaryText)
	s.Equal(sum.PeriodStart.Format(time.DateOnly), got[0].PeriodStart.Format(time.DateOnly))
	s.Equal(sum.PeriodEnd.Format(time.DateOnly), got[0].PeriodEnd.Format(time.DateOnly))
}

func (s *SQLiteStoreSuite) TestGetSummariesFiltersByCategory() {
	s.Require().NoError(s.store.SaveSummary(s.ctx, Summary{
		ID: "s1", PeriodStart: date(2026, 2, 21), PeriodEnd: date(2026, 2, 27),
		Category: "sleep", SummaryText: "sleep summary",
	}))
	s.Require().NoError(s.store.SaveSummary(s.ctx, Summary{
		ID: "s2", PeriodStart: date(2026, 2, 21), PeriodEnd: date(2026, 2, 27),
		Category: "food", SummaryText: "food summary",
	}))

	got, err := s.store.GetSummaries(s.ctx, "food", date(2026, 2, 1), date(2026, 2, 28))
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("food summary", got[0].SummaryText)
}

// --- Metric Tests ---

func (s *SQLiteStoreSuite) TestSaveAndGetMetrics() {
	entry := Entry{ID: "e1", Date: date(2026, 2, 28), Category: "sleep", InputText: "slept"}
	s.Require().NoError(s.store.SaveEntry(s.ctx, entry))

	metrics := []Metric{
		{Key: "total_hours", Value: 7.5},
		{Key: "rem_minutes", Value: 95},
	}
	s.Require().NoError(s.store.SaveMetrics(s.ctx, "e1", metrics))

	got, err := s.store.GetMetrics(s.ctx, "sleep", "total_hours", date(2026, 2, 1), date(2026, 2, 28))
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("total_hours", got[0].Key)
	s.Equal(7.5, got[0].Value)
	s.Equal("e1", got[0].EntryID)
}

func (s *SQLiteStoreSuite) TestGetMetricsAcrossDateRange() {
	for i := range 3 {
		d := date(2026, 2, 26+i)
		id := fmtID("e", i)
		s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{
			ID: id, Date: d, Category: "sleep", InputText: "slept",
		}))
		s.Require().NoError(s.store.SaveMetrics(s.ctx, id, []Metric{
			{Key: "total_hours", Value: 6.5 + float64(i)},
		}))
	}

	got, err := s.store.GetMetrics(s.ctx, "sleep", "total_hours", date(2026, 2, 26), date(2026, 2, 28))
	s.Require().NoError(err)
	s.Len(got, 3)
}

func (s *SQLiteStoreSuite) TestGetMetricsFiltersByKey() {
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{
		ID: "e1", Date: date(2026, 2, 28), Category: "sleep", InputText: "slept",
	}))
	s.Require().NoError(s.store.SaveMetrics(s.ctx, "e1", []Metric{
		{Key: "total_hours", Value: 7},
		{Key: "rem_minutes", Value: 90},
	}))

	got, err := s.store.GetMetrics(s.ctx, "sleep", "rem_minutes", date(2026, 2, 1), date(2026, 2, 28))
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("rem_minutes", got[0].Key)
}

// --- Embedding Tests ---

func (s *SQLiteStoreSuite) TestSaveAndSearchEmbeddings() {
	// Create entries
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{
		ID: "e1", Date: date(2026, 2, 28), Category: "sleep",
		InputText: "slept badly", AnalysisText: "Poor sleep",
	}))
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{
		ID: "e2", Date: date(2026, 2, 27), Category: "sleep",
		InputText: "slept well", AnalysisText: "Good sleep",
	}))
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{
		ID: "e3", Date: date(2026, 2, 26), Category: "food",
		InputText: "ate pasta", AnalysisText: "Carb heavy",
	}))

	// Save embeddings -- use simple distinguishable vectors
	vec1 := makeVec(1536, 0.1)
	vec2 := makeVec(1536, 0.2)
	vec3 := makeVec(1536, 0.9)

	s.Require().NoError(s.store.SaveEmbedding(s.ctx, "e1", "entry", vec1))
	s.Require().NoError(s.store.SaveEmbedding(s.ctx, "e2", "entry", vec2))
	s.Require().NoError(s.store.SaveEmbedding(s.ctx, "e3", "entry", vec3))

	// Search with a query close to vec1
	query := makeVec(1536, 0.11)
	results, err := s.store.SearchSimilar(s.ctx, query, "entry", 2)
	s.Require().NoError(err)
	s.Require().Len(results, 2)

	// Closest should be e1 (vec 0.1 is closest to 0.11)
	s.Equal("e1", results[0].Entry.ID)
	s.Equal("slept badly", results[0].Entry.InputText)
	s.True(results[0].Distance < results[1].Distance)
}

func (s *SQLiteStoreSuite) TestSearchSimilarFiltersBySourceType() {
	s.Require().NoError(s.store.SaveEntry(s.ctx, Entry{
		ID: "e1", Date: date(2026, 2, 28), Category: "sleep", InputText: "slept",
	}))

	vec := makeVec(1536, 0.5)
	s.Require().NoError(s.store.SaveEmbedding(s.ctx, "e1", "entry", vec))

	// Search with a different source type should return nothing
	results, err := s.store.SearchSimilar(s.ctx, vec, "summary", 5)
	s.Require().NoError(err)
	s.Empty(results)
}

// --- Schema Tests ---

func (s *SQLiteStoreSuite) TestSchemaCreatesVecTable() {
	var version string
	err := s.store.db.QueryRow("SELECT vec_version()").Scan(&version)
	s.Require().NoError(err)
	s.NotEmpty(version)
}

// --- helpers ---

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func fmtID(prefix string, i int) string {
	return prefix + string(rune('0'+i))
}

func makeVec(dim int, val float32) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = val
	}
	// Add small variation based on position so vectors aren't identical across dims
	for i := range v {
		v[i] += float32(math.Sin(float64(i)*0.01)) * 0.001
	}
	return v
}
