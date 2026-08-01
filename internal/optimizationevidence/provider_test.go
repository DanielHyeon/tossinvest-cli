package optimizationevidence

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

type dashboardReaderStub struct {
	view  performance.DashboardView
	err   error
	query performance.Query
	calls int
}

func (s *dashboardReaderStub) Dashboard(_ context.Context, query performance.Query) (performance.DashboardView, error) {
	s.calls++
	s.query = query
	if s.view.Query.AsOf.IsZero() {
		s.view.Query = query
	}
	return s.view, s.err
}

func completeDashboard() performance.DashboardView {
	metrics := make([]performance.MetricSummary, 0, len(RequiredMetricKeys()))
	for _, key := range RequiredMetricKeys() {
		metrics = append(metrics, performance.MetricSummary{
			Key: key, Value: "1", Samples: performance.DefaultMinimumSample,
			Status: performance.StatusComplete, Provenance: "fixture@v1",
		})
	}
	return performance.DashboardView{
		NewestSourceAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		States:         performance.StateCounts{Complete: performance.DefaultMinimumSample},
		Aggregates: []performance.Aggregate{{
			Market: "kr", LaneID: "lane-a", LaneVersion: "lane/v1",
			PolicyID: "policy-a", PolicyVersion: "policy/v1",
			Samples: performance.DefaultMinimumSample, Status: performance.StatusComplete,
			SemanticsVersion: performance.SemanticsVersion, ObservationProvenance: "fixture@v1",
			Metrics: metrics,
		}},
	}
}

func TestProviderUsesPersistedSourceTimeAndMarksOldEvidenceStale(t *testing.T) {
	asOf := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	view := completeDashboard()
	view.NewestSourceAt = asOf.Add(-MaxEvidenceAge - time.Nanosecond)
	evidence, err := New(&dashboardReaderStub{view: view}, func() time.Time { return asOf }).ReadEvidence(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Status != strategyopt.EvidenceStale || !evidence.ObservedAt.Equal(view.NewestSourceAt) ||
		!slices.Contains(evidence.Missing, MissingStale) {
		t.Fatalf("evidence=%+v, want stale at authoritative persisted source time", evidence)
	}
}

func TestProviderRejectsDashboardThatDoesNotEchoTheExactFixedQuery(t *testing.T) {
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	view := completeDashboard()
	view.Query = performance.DefaultQuery(asOf)
	view.Query.Market = "client-invented"
	evidence, err := New(&dashboardReaderStub{view: view}, func() time.Time { return asOf }).ReadEvidence(context.Background())
	if err == nil || evidence.Status != strategyopt.EvidenceUnavailable || evidence.Digest != "" ||
		!slices.Contains(evidence.Missing, MissingQueryMismatch) {
		t.Fatalf("evidence=%+v err=%v, want unavailable query mismatch", evidence, err)
	}
}

func TestProviderRefusesMissingOrFuturePersistedSourceTime(t *testing.T) {
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	for _, sourceAt := range []time.Time{{}, asOf.Add(time.Nanosecond)} {
		view := completeDashboard()
		view.NewestSourceAt = sourceAt
		evidence, err := New(&dashboardReaderStub{view: view}, func() time.Time { return asOf }).ReadEvidence(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if evidence.Status != strategyopt.EvidenceStale || !slices.Contains(evidence.Missing, MissingSourceTime) {
			t.Fatalf("sourceAt=%s evidence=%+v, want stale fail-closed", sourceAt, evidence)
		}
	}
}

func TestProviderUsesOneServerFixedThirtyDayAllScopeQuery(t *testing.T) {
	asOf := time.Date(2026, 8, 1, 12, 34, 56, 789, time.FixedZone("fixture", 9*60*60))
	reader := &dashboardReaderStub{view: completeDashboard()}
	evidence, err := New(reader, func() time.Time { return asOf }).ReadEvidence(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if reader.calls != 1 {
		t.Fatalf("Dashboard calls=%d, want 1", reader.calls)
	}
	if reader.query.PeriodDays != 30 || reader.query.Market != performance.AllMarkets ||
		reader.query.Lane != performance.AllLanes || reader.query.CompleteOnly ||
		reader.query.MinimumSample != performance.DefaultMinimumSample {
		t.Fatalf("query=%+v, want fixed 30d/all/all including missing lineage for fail-closed assessment", reader.query)
	}
	if !reader.query.AsOf.Equal(asOf.UTC()) || !evidence.ObservedAt.Equal(reader.view.NewestSourceAt) {
		t.Fatalf("query as-of=%s evidence observed=%s want persisted %s", reader.query.AsOf, evidence.ObservedAt, reader.view.NewestSourceAt)
	}
	if evidence.Status != strategyopt.EvidenceComplete || len(evidence.Digest) != 64 || len(evidence.Missing) != 0 {
		t.Fatalf("evidence=%+v", evidence)
	}
}

func TestProviderDigestIsDeterministicAndInsensitiveToDashboardSliceOrder(t *testing.T) {
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	first := completeDashboard()
	second := completeDashboard()
	slices.Reverse(second.Aggregates[0].Metrics)
	one, err := New(&dashboardReaderStub{view: first}, func() time.Time { return asOf }).ReadEvidence(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	two, err := New(&dashboardReaderStub{view: second}, func() time.Time { return asOf }).ReadEvidence(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if one.Digest == "" || one.Digest != two.Digest {
		t.Fatalf("digests=%q/%q, want identical canonical SHA-256", one.Digest, two.Digest)
	}
}

func TestProviderFailsClosedForEveryMissingEvidenceClass(t *testing.T) {
	tests := []struct {
		name string
		edit func(*performance.DashboardView)
		want string
	}{
		{name: "no complete lineage", edit: func(v *performance.DashboardView) {
			v.States.Complete = 0
		}, want: MissingCompleteLineage},
		{name: "link missing", edit: func(v *performance.DashboardView) {
			v.States.LinkMissing = 1
		}, want: MissingLink},
		{name: "not measured", edit: func(v *performance.DashboardView) {
			v.States.NotMeasured = 1
		}, want: MissingNotMeasured},
		{name: "insufficient sample", edit: func(v *performance.DashboardView) {
			v.States.Complete = performance.DefaultMinimumSample - 1
			v.States.InsufficientSample = 1
			v.Aggregates[0].Samples = performance.DefaultMinimumSample - 1
			v.Aggregates[0].Status = performance.StatusInsufficientSample
		}, want: MissingMinimumSample},
		{name: "required metric status", edit: func(v *performance.DashboardView) {
			v.Aggregates[0].Metrics[0].Status = performance.StatusNotMeasured
			v.Aggregates[0].Metrics[0].Value = ""
		}, want: MissingRequiredMetricPrefix + RequiredMetricKeys()[0]},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			view := completeDashboard()
			tc.edit(&view)
			evidence, err := New(&dashboardReaderStub{view: view}, func() time.Time {
				return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
			}).ReadEvidence(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if evidence.Status != strategyopt.EvidenceInsufficient || !slices.Contains(evidence.Missing, tc.want) ||
				len(evidence.Digest) != 64 || evidence.ObservedAt.IsZero() {
				t.Fatalf("evidence=%+v, want insufficient with %q and deterministic observation", evidence, tc.want)
			}
		})
	}
}

func TestProviderDashboardErrorIsUnavailableAndNeverLaunderedAsEvidence(t *testing.T) {
	readerErr := errors.New("fixture dashboard unavailable")
	evidence, err := New(&dashboardReaderStub{err: readerErr}, func() time.Time {
		return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	}).ReadEvidence(context.Background())
	if !errors.Is(err, readerErr) {
		t.Fatalf("error=%v, want wrapped dashboard error", err)
	}
	if evidence.Status != strategyopt.EvidenceUnavailable || evidence.Digest != "" ||
		!evidence.ObservedAt.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) ||
		!slices.Contains(evidence.Missing, MissingDashboardError) {
		t.Fatalf("evidence=%+v", evidence)
	}
}

func TestProviderRejectsNilDependenciesAndZeroClock(t *testing.T) {
	validReader := &dashboardReaderStub{view: completeDashboard()}
	validClock := func() time.Time { return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC) }
	tests := []struct {
		name     string
		provider *Provider
	}{
		{name: "nil provider"},
		{name: "nil reader", provider: New(nil, validClock)},
		{name: "nil clock", provider: New(validReader, nil)},
		{name: "zero clock", provider: New(validReader, func() time.Time { return time.Time{} })},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evidence, err := tc.provider.ReadEvidence(context.Background())
			if err == nil || evidence.Status != strategyopt.EvidenceUnavailable || evidence.Digest != "" ||
				!slices.Contains(evidence.Missing, MissingDashboardError) {
				t.Fatalf("evidence=%+v err=%v, want unavailable fail-closed", evidence, err)
			}
		})
	}
}

func TestProviderFailsClosedForLineageSemanticsAndRequiredMetricShape(t *testing.T) {
	tests := []struct {
		name string
		edit func(*performance.DashboardView)
		want string
	}{
		{name: "blank aggregate lineage", edit: func(v *performance.DashboardView) {
			v.Aggregates[0].LaneVersion = " "
		}, want: MissingCompleteLineage},
		{name: "semantics drift", edit: func(v *performance.DashboardView) {
			v.Aggregates[0].SemanticsVersion = "lane-performance/future"
		}, want: missingSemanticsVersion},
		{name: "missing required metric", edit: func(v *performance.DashboardView) {
			v.Aggregates[0].Metrics = v.Aggregates[0].Metrics[1:]
		}, want: MissingRequiredMetricPrefix + RequiredMetricKeys()[0]},
		{name: "duplicate required metric", edit: func(v *performance.DashboardView) {
			v.Aggregates[0].Metrics = append(v.Aggregates[0].Metrics, v.Aggregates[0].Metrics[0])
		}, want: MissingRequiredMetricPrefix + RequiredMetricKeys()[0]},
		{name: "undersampled required metric", edit: func(v *performance.DashboardView) {
			v.Aggregates[0].Metrics[0].Samples = performance.DefaultMinimumSample - 1
		}, want: MissingRequiredMetricPrefix + RequiredMetricKeys()[0]},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			view := completeDashboard()
			tc.edit(&view)
			evidence, err := New(&dashboardReaderStub{view: view}, func() time.Time {
				return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
			}).ReadEvidence(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if evidence.Status != strategyopt.EvidenceInsufficient || !slices.Contains(evidence.Missing, tc.want) {
				t.Fatalf("evidence=%+v, want %q", evidence, tc.want)
			}
		})
	}
}
