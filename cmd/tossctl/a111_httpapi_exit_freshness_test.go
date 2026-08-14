package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type a111HTTPRouteSession struct {
	router   http.Handler
	reader   *httpAPIReader
	holdings *httpAPIHoldingsFixture
	now      *time.Time
}

// a111SlowHoldingsFixture advances the same deterministic clock that the
// production reader uses while its one cache-miss broker read is in flight.
// It keeps the temporal regression test free of wall-clock scheduling.
type a111SlowHoldingsFixture struct {
	calls   int
	rows    []domain.Position
	advance func()
}

func (f *a111SlowHoldingsFixture) Holdings(context.Context, string) ([]domain.Position, error) {
	f.calls++
	if f.advance != nil {
		f.advance()
	}
	return append([]domain.Position(nil), f.rows...), nil
}

// a111SlowManagementRuntimeFixture advances time only after Positions has
// completed its cache lookup and entered the local projection reads.
type a111SlowManagementRuntimeFixture struct {
	calls   int
	advance func()
}

func (f *a111SlowManagementRuntimeFixture) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	f.calls++
	if f.advance != nil {
		f.advance()
	}
	return positionpolicy.ManagementRuntime{}, nil
}

// a111ClockSequence models time crossing a boundary while the marker file is
// being read. The route must consume cache, pre-marker and post-marker samples
// in that order; extra reads repeat the final value and are caught by calls.
type a111ClockSequence struct {
	calls  int
	values []time.Time
}

func (c *a111ClockSequence) Now() time.Time {
	index := c.calls
	c.calls++
	if index >= len(c.values) {
		index = len(c.values) - 1
	}
	return c.values[index]
}

func (s *a111HTTPRouteSession) positions(t *testing.T) httpapi.PositionsResource {
	t.Helper()
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/positions", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/positions = %d: %s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Resource string                    `json:"resource"`
		Data     httpapi.PositionsResource `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode positions envelope: %v", err)
	}
	if envelope.Resource != "positions" {
		t.Fatalf("route resource = %q, want positions", envelope.Resource)
	}
	return envelope.Data
}

func a111NewHTTPRouteSession(t *testing.T, observedAt, asOf time.Time,
	engineMarker string,
) *a111HTTPRouteSession {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), journal.DBFileName)
	writer, err := journal.Open(ctx, journal.Options{
		Path:     path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open journal: %v", err)
	}
	watermark, err := writer.FillWatermark(ctx, "AAPL")
	if err != nil {
		t.Fatalf("FillWatermark: %v", err)
	}
	fold, err := writer.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: "account-a111", Market: "us", Symbol: "AAPL", Kind: journal.AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: "3", NewAvgPrice: "200", BrokerAsOf: observedAt.Add(-time.Minute).Format(time.RFC3339Nano),
		Evidence: "A111 route fixture folds the authoritative broker holding",
	})
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	if _, err := writer.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: fold.Position.ID, Symbol: "AAPL", Market: "us", Quantity: "3",
		CostBasis: "200", ObservedPrice: "200", SyntheticStop: "190",
		ObservedAt: observedAt.Add(-time.Minute).Format(time.RFC3339Nano),
	}); err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	seed, err := writer.OpenAdoptedExitState(ctx, fold.Position.ID)
	if err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	evaluation := exitpolicy.RatchetSnapshotInput{
		Context: exitpolicy.SnapshotContext{
			PositionID: seed.PositionID, PositionGeneration: seed.PositionGeneration,
			ObservationID: "a111-http-route-observation", RemainingQuantity: "3",
		},
		Input: exitpolicy.RatchetInput{
			Entry: "200", InitialStop: "190", ObservedPrice: "205", HighWater: "200",
			Baseline: "190", RealBreakeven: "201", TakenRatioTotal: seed.TakenRatioTotal,
			Level: exitpolicy.Level(seed.RatchetLevel),
		},
	}
	line, err := exitpolicy.EvaluateRatchetSnapshot(evaluation)
	if err != nil {
		t.Fatalf("EvaluateRatchetSnapshot: %v", err)
	}
	line = line.ChangedFromState(evaluation.Input.HighWater, evaluation.Input.Baseline,
		evaluation.Input.Level, exitpolicy.NoRung)
	recovery := exitpolicy.NewRatchetRecoveryPolicy(evaluation)
	judgement := journal.ExitJudgement{
		PositionID: seed.PositionID, LifecycleGeneration: seed.LifecycleGeneration,
		Snapshot: line, RecoveryPolicy: recovery, ObservationSource: "quote_fetched_at", ObservedAt: observedAt,
		Provenance: journal.ExitDecisionProvenance{
			ObservationID: line.ObservationID, SnapshotID: line.SnapshotID,
			DecisionID: line.DecisionID, Policy: line.Policy,
		},
		ObservedPrice: line.ObservedPrice, HighWater: line.HighWater,
		Baseline: line.CurrentProtection, RatchetLevel: string(line.RatchetLevel), ActiveRung: line.ActiveRung,
	}
	if err := writer.RecordExitJudgement(ctx, judgement); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	readOnly, err := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	t.Cleanup(func() { _ = readOnly.Close() })
	holdings := &httpAPIHoldingsFixture{rows: []domain.Position{{
		Symbol: "AAPL", Name: "Apple", MarketCode: "US", Quantity: 3,
		AveragePrice: 200, CurrentPrice: 205, MarketValue: 615,
	}}}
	now := asOf
	reader := &httpAPIReader{
		now: func() time.Time { return now }, engineMarker: engineMarker, holdings: holdings,
		accountRef: func() (string, error) { return "account-a111", nil }, journal: readOnly,
	}
	router, err := httpapi.NewRouter(httpapi.Options{Reader: reader, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return &a111HTTPRouteSession{router: router, reader: reader, holdings: holdings, now: &now}
}

func a111HTTPRouteFixture(t *testing.T, observedAt, asOf time.Time) (httpapi.PositionsResource, int) {
	t.Helper()
	session := a111NewHTTPRouteSession(t, observedAt, asOf, "")
	resource := session.positions(t)
	return resource, session.holdings.calls
}

func a111HTTPStoredPosition(observedAt string) journal.PositionExit {
	return journal.PositionExit{
		Position: journal.Position{
			ID: "a111-http-position", AccountRef: "account-a111", Market: "us", Symbol: "AAPL",
			Quantity: "3", AdoptionID: "adoption-a111", State: "OPEN",
		},
		HasExit: true,
		Exit: journal.ExitState{
			PositionID: "a111-http-position", LifecycleGeneration: 1, PositionGeneration: 1,
			SnapshotStatus: journal.SnapshotStatusEvaluated,
			EntryPrice:     "200", InitialStop: "190", Baseline: "195", HighWater: "210",
			Snapshot: journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
				Line: exitpolicy.ExitLineSnapshot{
					SnapshotID: "a111-http-snapshot", DecisionID: "a111-http-decision",
					ObservationID: "a111-http-observation", PositionID: "a111-http-position",
					PositionGeneration: 1, EntryPrice: "200", InitialStop: "190", ObservedPrice: "205",
					CurrentProtection: "195", HighWater: "210", RatchetLevel: exitpolicy.LevelNone,
					ActiveRung: exitpolicy.NoRung, NextTarget: "220", NextProtection: "200",
					Action: exitpolicy.ActionNone, ProjectedQuantity: "0",
				},
				ObservationSource: "quote_fetched_at", ObservedAt: observedAt,
			}},
		},
	}
}

func TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	for _, live := range []operatorview.ExitLiveness{
		operatorview.ExitLivenessRunning,
		operatorview.ExitLivenessUnavailable,
		operatorview.ExitLivenessUnwired,
	} {
		for _, tc := range []struct {
			name   string
			age    time.Duration
			status string
		}{
			{"inside", 29999 * time.Millisecond, "fresh"},
			{"exact", 30 * time.Second, "fresh"},
			{"outside", 30*time.Second + time.Nanosecond, "stale"},
		} {
			t.Run(fmt.Sprint(live)+"/"+tc.name, func(t *testing.T) {
				stored := a111HTTPStoredPosition(observed.Format(time.RFC3339Nano))
				var got httpapi.Position
				applyStoredPositionWithLiveness(&got, stored, observed.Add(tc.age), live)
				if got.ExitLine.Status != tc.status {
					t.Fatalf("status = %q, want %q: %+v", got.ExitLine.Status, tc.status, got.ExitLine)
				}
				if tc.status == "stale" && (got.ExitLine.Reason == "" ||
					got.ExitLine.CurrentProtection != "—" || got.ExitLine.NextTarget != "—") {
					t.Fatalf("stale API line remained actionable: %+v", got.ExitLine)
				}
			})
		}
	}
}

func TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	resource, brokerCalls := a111HTTPRouteFixture(t, observed, observed.Add(30*time.Second))
	if brokerCalls != 1 {
		t.Fatalf("real route broker reads = %d, want one", brokerCalls)
	}
	if resource.Source != "official+journal-read-only" || len(resource.Items) != 1 {
		t.Fatalf("real route positions = %+v", resource)
	}
	got := resource.Items[0]
	if got.Symbol != "AAPL" || got.PositionID == "" || !got.InJournal || !got.InBroker {
		t.Fatalf("route did not join the production broker/journal position: %+v", got)
	}
	if got.ExitLine.Status != "fresh" || got.ExitLine.CurrentProtection == "—" ||
		got.ExitLine.NextTarget == "—" || got.ExitLine.EvaluatedAt != observed.Format(time.RFC3339Nano) {
		t.Fatalf("real /api/v1/positions hid exact-boundary evidence: %+v", got.ExitLine)
	}
}

func TestA111RealPositionsRouteUsesPostReadClockForFreshnessAndMarkerLiveness(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	advance := 30*time.Second + time.Nanosecond
	for _, tc := range []struct {
		name       string
		marker     func(*testing.T) (string, func())
		wantReason string
	}{
		{
			name:       "unavailable-marker-ages-the-exit-line-after-the-cache-miss",
			marker:     func(*testing.T) (string, func()) { return "", func() {} },
			wantReason: "observation_older_than_limit",
		},
		{
			name: "marker-expiring-during-the-cache-miss-is-stopped-after-the-read",
			marker: func(t *testing.T) (string, func()) {
				t.Helper()
				path := enginelock.MarkerPath(t.TempDir())
				held, err := enginelock.Hold(context.Background(), path,
					observed.Add(-enginelock.StaleAfter+30*time.Second))
				if err != nil {
					t.Fatalf("Hold engine marker: %v", err)
				}
				return path, held.Release
			},
			wantReason: "engine_not_running",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marker, release := tc.marker(t)
			t.Cleanup(release)
			session := a111NewHTTPRouteSession(t, observed, observed, marker)
			slowHoldings := &a111SlowHoldingsFixture{rows: session.holdings.rows, advance: func() {
				*session.now = observed.Add(advance)
			}}
			session.reader.holdings = slowHoldings

			resource := session.positions(t)
			if slowHoldings.calls != 1 {
				t.Fatalf("slow holdings calls = %d, want exactly one cache-miss read", slowHoldings.calls)
			}
			if len(resource.Items) != 1 || resource.Items[0].ExitLine.Status != "stale" ||
				resource.Items[0].ExitLine.Reason != tc.wantReason {
				t.Fatalf("post-read positions verdict = %+v, want stale/%s", resource.Items, tc.wantReason)
			}
			line := resource.Items[0].ExitLine
			for field, value := range map[string]string{
				"entry_price": line.EntryPrice, "initial_stop": line.InitialStop,
				"current_protection": line.CurrentProtection, "next_target": line.NextTarget,
				"next_protection": line.NextProtection, "projected_quantity": line.ProjectedQuantity,
				"observed_price": line.ObservedPrice, "high_water": line.HighWater,
			} {
				if value != "—" {
					t.Errorf("post-read stale line %s = %q, want closed dash", field, value)
				}
			}
		})
	}
}

func TestA111RealPositionsRouteUsesPostProjectionReadClockForFreshnessAndMarkerLiveness(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	advance := 30*time.Second + time.Nanosecond
	for _, tc := range []struct {
		name       string
		marker     func(*testing.T) (string, func())
		wantReason string
	}{
		{
			name:       "runtime-read-ages-the-exit-line-after-a-cache-hit",
			marker:     func(*testing.T) (string, func()) { return "", func() {} },
			wantReason: "observation_older_than_limit",
		},
		{
			name: "runtime-read-expires-the-marker-after-a-cache-hit",
			marker: func(t *testing.T) (string, func()) {
				t.Helper()
				path := enginelock.MarkerPath(t.TempDir())
				held, err := enginelock.Hold(context.Background(), path,
					observed.Add(-enginelock.StaleAfter+30*time.Second))
				if err != nil {
					t.Fatalf("Hold engine marker: %v", err)
				}
				return path, held.Release
			},
			wantReason: "engine_not_running",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marker, release := tc.marker(t)
			t.Cleanup(release)
			session := a111NewHTTPRouteSession(t, observed, observed, marker)

			warm := session.positions(t)
			if session.holdings.calls != 1 || len(warm.Items) != 1 ||
				warm.Items[0].ExitLine.Status != "fresh" {
				t.Fatalf("cache warm-up = calls %d, items %+v; want one fresh broker read",
					session.holdings.calls, warm.Items)
			}

			slowRuntime := &a111SlowManagementRuntimeFixture{advance: func() {
				*session.now = observed.Add(advance)
			}}
			session.reader.managementRuntime = slowRuntime
			resource := session.positions(t)
			if session.holdings.calls != 1 {
				t.Fatalf("projection-local read bypassed broker cache: holdings calls = %d, want one",
					session.holdings.calls)
			}
			if slowRuntime.calls != 1 {
				t.Fatalf("management runtime calls = %d, want exactly one projection-local read", slowRuntime.calls)
			}
			if len(resource.Items) != 1 || resource.Items[0].ExitLine.Status != "stale" ||
				resource.Items[0].ExitLine.Reason != tc.wantReason {
				t.Fatalf("post-projection-read positions verdict = %+v, want stale/%s",
					resource.Items, tc.wantReason)
			}
			line := resource.Items[0].ExitLine
			for field, value := range map[string]string{
				"entry_price": line.EntryPrice, "initial_stop": line.InitialStop,
				"current_protection": line.CurrentProtection, "next_target": line.NextTarget,
				"next_protection": line.NextProtection, "projected_quantity": line.ProjectedQuantity,
				"observed_price": line.ObservedPrice, "high_water": line.HighWater,
			} {
				if value != "—" {
					t.Errorf("post-projection-read stale line %s = %q, want closed dash", field, value)
				}
			}
		})
	}
}

func TestA111RealPositionsRouteUsesPostMarkerResponseClock(t *testing.T) {
	evidenceAt := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		markerAt   time.Time
		preMarker  time.Time
		postMarker time.Time
		wantReason string
	}{
		{
			name:       "evidence-crosses-thirty-seconds-during-marker-read",
			markerAt:   evidenceAt,
			preMarker:  evidenceAt.Add(30 * time.Second),
			postMarker: evidenceAt.Add(30*time.Second + time.Nanosecond),
			wantReason: "observation_older_than_limit",
		},
		{
			name:       "marker-alone-crosses-stale-after-during-marker-read",
			markerAt:   evidenceAt.Add(-enginelock.StaleAfter + time.Nanosecond),
			preMarker:  evidenceAt,
			postMarker: evidenceAt.Add(2 * time.Nanosecond),
			wantReason: "engine_not_running",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marker := enginelock.MarkerPath(t.TempDir())
			held, err := enginelock.Hold(context.Background(), marker, tc.markerAt)
			if err != nil {
				t.Fatalf("Hold engine marker: %v", err)
			}
			t.Cleanup(held.Release)
			session := a111NewHTTPRouteSession(t, evidenceAt, evidenceAt, marker)

			warm := session.positions(t)
			if session.holdings.calls != 1 || len(warm.Items) != 1 ||
				warm.Items[0].ExitLine.Status != "fresh" {
				t.Fatalf("cache warm-up = calls %d, items %+v; want one fresh broker read",
					session.holdings.calls, warm.Items)
			}

			clock := &a111ClockSequence{values: []time.Time{evidenceAt, tc.preMarker, tc.postMarker}}
			session.reader.now = clock.Now
			resource := session.positions(t)
			if session.holdings.calls != 1 {
				t.Errorf("marker-time transition bypassed broker cache: holdings calls = %d, want one",
					session.holdings.calls)
			}
			if clock.calls != 3 {
				t.Errorf("reader clock calls = %d, want cache, pre-marker and post-marker samples", clock.calls)
			}
			if len(resource.Items) != 1 || resource.Items[0].ExitLine.Status != "stale" ||
				resource.Items[0].ExitLine.Reason != tc.wantReason {
				t.Fatalf("post-marker positions verdict = %+v, want stale/%s",
					resource.Items, tc.wantReason)
			}
			line := resource.Items[0].ExitLine
			for field, value := range map[string]string{
				"entry_price": line.EntryPrice, "initial_stop": line.InitialStop,
				"current_protection": line.CurrentProtection, "next_target": line.NextTarget,
				"next_protection": line.NextProtection, "projected_quantity": line.ProjectedQuantity,
				"observed_price": line.ObservedPrice, "high_water": line.HighWater,
			} {
				if value != "—" {
					t.Errorf("post-marker stale line %s = %q, want closed dash", field, value)
				}
			}
		})
	}
}

func TestA111PostMarkerClockRollbackCannotResurrectAStoppedEngine(t *testing.T) {
	responseAt := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	markerAt := responseAt.Add(-enginelock.StaleAfter + time.Second)
	preMarker := markerAt.Add(enginelock.StaleAfter + time.Nanosecond)
	marker := enginelock.MarkerPath(t.TempDir())
	held, err := enginelock.Hold(context.Background(), marker, markerAt)
	if err != nil {
		t.Fatalf("Hold engine marker: %v", err)
	}
	t.Cleanup(held.Release)
	session := a111NewHTTPRouteSession(t, responseAt, responseAt, marker)

	warm := session.positions(t)
	if session.holdings.calls != 1 || len(warm.Items) != 1 ||
		warm.Items[0].ExitLine.Status != "fresh" {
		t.Fatalf("cache warm-up = calls %d, items %+v; want one fresh broker read",
			session.holdings.calls, warm.Items)
	}

	clock := &a111ClockSequence{values: []time.Time{responseAt, preMarker, responseAt}}
	session.reader.now = clock.Now
	resource := session.positions(t)
	if session.holdings.calls != 1 {
		t.Errorf("clock rollback bypassed broker cache: holdings calls = %d, want one",
			session.holdings.calls)
	}
	if clock.calls != 3 {
		t.Errorf("reader clock calls = %d, want cache, pre-marker and post-marker samples", clock.calls)
	}
	if len(resource.Items) != 1 || resource.Items[0].ExitLine.Status != "stale" ||
		resource.Items[0].ExitLine.Reason != "engine_not_running" {
		t.Fatalf("post-marker rollback verdict = %+v, want stopped stale/engine_not_running",
			resource.Items)
	}
	line := resource.Items[0].ExitLine
	for field, value := range map[string]string{
		"entry_price": line.EntryPrice, "initial_stop": line.InitialStop,
		"current_protection": line.CurrentProtection, "next_target": line.NextTarget,
		"next_protection": line.NextProtection, "projected_quantity": line.ProjectedQuantity,
		"observed_price": line.ObservedPrice, "high_water": line.HighWater,
	} {
		if value != "—" {
			t.Errorf("stopped rollback line %s = %q, want closed dash", field, value)
		}
	}
}

func TestA111CachedPositionsObserveRealEngineStopAndResumeWithoutContamination(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	firstAt := observed.Add(29 * time.Second)
	marker := enginelock.MarkerPath(t.TempDir())
	held, err := enginelock.Hold(context.Background(), marker, firstAt)
	if err != nil {
		t.Fatalf("Hold engine marker: %v", err)
	}
	t.Cleanup(held.Release)
	session := a111NewHTTPRouteSession(t, observed, firstAt, marker)

	warm := session.positions(t)
	if len(warm.Items) != 1 || warm.Items[0].ExitLine.Status != "fresh" ||
		warm.Items[0].ExitLine.CurrentProtection == "—" || warm.Items[0].ExitLine.NextTarget == "—" {
		t.Fatalf("running warm-cache response is not actionable: %+v", warm.Items)
	}
	if session.holdings.calls != 1 {
		t.Fatalf("warm-cache broker reads = %d, want one", session.holdings.calls)
	}

	held.Release()
	*session.now = firstAt.Add(time.Second) // same cached broker interval; evidence is exactly 30s old
	stopped := session.positions(t)
	if session.holdings.calls != 1 {
		t.Fatalf("engine transition bypassed the positions cache: broker reads=%d, want one", session.holdings.calls)
	}
	if len(stopped.Items) != 1 || stopped.Items[0].ExitLine.Status != "stale" ||
		stopped.Items[0].ExitLine.Reason != "engine_not_running" {
		t.Fatalf("cached response hid immediate engine stop: %+v", stopped.Items)
	}
	line := stopped.Items[0].ExitLine
	for field, value := range map[string]string{
		"entry_price": line.EntryPrice, "initial_stop": line.InitialStop,
		"current_protection": line.CurrentProtection, "next_target": line.NextTarget,
		"next_protection": line.NextProtection, "projected_quantity": line.ProjectedQuantity,
		"observed_price": line.ObservedPrice, "high_water": line.HighWater,
	} {
		if value != "—" {
			t.Errorf("stopped cached line %s = %q, want closed dash", field, value)
		}
	}

	resumedAt := firstAt.Add(500 * time.Millisecond) // still inside the per-position 30-second bound
	resumedMarker, err := enginelock.Hold(context.Background(), marker, resumedAt)
	if err != nil {
		t.Fatalf("resume engine marker: %v", err)
	}
	t.Cleanup(resumedMarker.Release)
	*session.now = resumedAt
	resumed := session.positions(t)
	if session.holdings.calls != 1 {
		t.Fatalf("engine resume bypassed the positions cache: broker reads=%d, want one", session.holdings.calls)
	}
	if len(resumed.Items) != 1 || resumed.Items[0].ExitLine.Status != "fresh" ||
		resumed.Items[0].ExitLine.Reason != "" {
		t.Fatalf("stopped projection contaminated cached running truth: %+v", resumed.Items)
	}
	resumedLine := resumed.Items[0].ExitLine
	for field, value := range map[string]string{
		"entry_price": resumedLine.EntryPrice, "initial_stop": resumedLine.InitialStop,
		"current_protection": resumedLine.CurrentProtection, "next_target": resumedLine.NextTarget,
		"next_protection": resumedLine.NextProtection, "projected_quantity": resumedLine.ProjectedQuantity,
		"observed_price": resumedLine.ObservedPrice, "high_water": resumedLine.HighWater,
	} {
		if value == "—" || value == "" {
			t.Errorf("resumed cached line %s was not restored: %q", field, value)
		}
	}
}

func TestA111RunningCachedPositionsRecheckTheExactPerPositionAgeBoundary(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	firstAt := observed.Add(29 * time.Second)
	marker := enginelock.MarkerPath(t.TempDir())
	held, err := enginelock.Hold(context.Background(), marker, firstAt)
	if err != nil {
		t.Fatalf("Hold engine marker: %v", err)
	}
	t.Cleanup(held.Release)
	session := a111NewHTTPRouteSession(t, observed, firstAt, marker)

	warm := session.positions(t)
	if len(warm.Items) != 1 || warm.Items[0].ExitLine.Status != "fresh" ||
		warm.Items[0].ExitLine.CurrentProtection == "—" || warm.Items[0].ExitLine.NextTarget == "—" {
		t.Fatalf("29-second warm-cache response is not actionable: %+v", warm.Items)
	}
	if session.holdings.calls != 1 {
		t.Fatalf("warm-cache broker reads = %d, want one", session.holdings.calls)
	}

	*session.now = observed.Add(30 * time.Second)
	exact := session.positions(t)
	if session.holdings.calls != 1 || len(exact.Items) != 1 || exact.Items[0].ExitLine.Status != "fresh" ||
		exact.Items[0].ExitLine.CurrentProtection == "—" || exact.Items[0].ExitLine.NextTarget == "—" {
		t.Fatalf("exact 30-second cached boundary is not fresh/actionable: calls=%d items=%+v",
			session.holdings.calls, exact.Items)
	}

	*session.now = observed.Add(30*time.Second + time.Nanosecond)
	aged := session.positions(t)
	if session.holdings.calls != 1 {
		t.Fatalf("age transition bypassed the positions cache: broker reads=%d, want one", session.holdings.calls)
	}
	if len(aged.Items) != 1 || aged.Items[0].ExitLine.Status != "stale" ||
		aged.Items[0].ExitLine.Reason != "observation_older_than_limit" {
		t.Fatalf("cached running response hid >30s age transition: %+v", aged.Items)
	}
	line := aged.Items[0].ExitLine
	for field, value := range map[string]string{
		"entry_price": line.EntryPrice, "initial_stop": line.InitialStop,
		"current_protection": line.CurrentProtection, "next_target": line.NextTarget,
		"next_protection": line.NextProtection, "projected_quantity": line.ProjectedQuantity,
		"observed_price": line.ObservedPrice, "high_water": line.HighWater,
	} {
		if value != "—" {
			t.Errorf("aged cached line %s = %q, want closed dash", field, value)
		}
	}
}

func TestA111HTTPAPIStoppedEngineIsImmediatelyStale(t *testing.T) {
	now := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	stored := a111HTTPStoredPosition(now.Format(time.RFC3339Nano))
	var got httpapi.Position
	applyStoredPositionWithLiveness(&got, stored, now, operatorview.ExitLivenessStopped)
	if got.ExitLine.Status != "stale" || got.ExitLine.Reason == "" ||
		got.ExitLine.CurrentProtection != "—" || got.ExitLine.NextTarget != "—" {
		t.Fatalf("stopped API projection = %+v", got.ExitLine)
	}
}

func TestA111HTTPAPIProjectionExactlyMatchesTheSharedOperatorVerdict(t *testing.T) {
	observed := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	stored := a111HTTPStoredPosition(observed.Format(time.RFC3339Nano))
	asOf := observed.Add(30*time.Second + time.Nanosecond)
	shared := operatorview.ApplyExitFreshness(stored.Exit.Snapshot, asOf, operatorview.ExitLivenessRunning)
	source := operatorview.Source{
		UnknownReason: shared.UnknownReason, StaleReason: shared.StaleReason,
		RemainingQuantity: stored.Position.Quantity, EffectiveSource: "persisted effective snapshot",
	}
	if shared.Snapshot != nil {
		source.Snapshot = &shared.Snapshot.Line
		source.ObservationSource = shared.Snapshot.ObservationSource
		source.ObservedAt = shared.Snapshot.ObservedAt
	}
	want := httpapi.ExitLineFrom(operatorview.BuildExitLine(source))
	var got httpapi.Position
	applyStoredPositionWithLiveness(&got, stored, asOf, operatorview.ExitLivenessRunning)
	if !reflect.DeepEqual(got.ExitLine, want) {
		t.Fatalf("HTTP exit line = %+v, shared = %+v", got.ExitLine, want)
	}
}

func TestA111HTTPAPIInvalidSiblingCannotBorrowRunningLiveness(t *testing.T) {
	now := time.Date(2026, 8, 14, 7, 0, 31, 0, time.UTC)
	freshStored := a111HTTPStoredPosition(now.Format(time.RFC3339Nano))
	staleStored := a111HTTPStoredPosition(now.Add(-31 * time.Second).Format(time.RFC3339Nano))
	staleStored.Position.ID = "a111-http-invalid-sibling"
	staleStored.Position.Symbol = "MSFT"
	staleStored.Exit.PositionID = staleStored.Position.ID

	var fresh, stale httpapi.Position
	applyStoredPositionWithLiveness(&fresh, freshStored, now, operatorview.ExitLivenessRunning)
	applyStoredPositionWithLiveness(&stale, staleStored, now, operatorview.ExitLivenessRunning)
	if fresh.ExitLine.Status != "fresh" {
		t.Fatalf("valid sibling = %+v", fresh.ExitLine)
	}
	if stale.ExitLine.Status != "stale" || stale.ExitLine.CurrentProtection != "—" ||
		stale.ExitLine.NextTarget != "—" {
		t.Fatalf("invalid sibling borrowed freshness: %+v", stale.ExitLine)
	}
}

func TestA111HTTPAPISeedCorruptAndPartialEvidenceStayHidden(t *testing.T) {
	now := time.Date(2026, 8, 14, 7, 0, 0, 0, time.UTC)
	for _, reason := range []string{
		"not_evaluated_yet",
		"invalid_effective_snapshot",
		"partial_evaluated_tuple",
		"flattened_snapshot_mismatch",
	} {
		t.Run(reason, func(t *testing.T) {
			stored := a111HTTPStoredPosition(now.Format(time.RFC3339Nano))
			stored.Exit.Snapshot = journal.ExitSnapshotView{UnknownReason: reason}
			var got httpapi.Position
			applyStoredPositionWithLiveness(&got, stored, now, operatorview.ExitLivenessRunning)
			if got.ExitLine.Status != "unknown" || got.ExitLine.CurrentProtection != "—" ||
				got.ExitLine.NextTarget != "—" || got.ExitLine.NextProtection != "—" {
				t.Fatalf("%s became actionable: %+v", reason, got.ExitLine)
			}
		})
	}
}
