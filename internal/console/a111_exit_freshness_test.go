package console

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type a111AdvancingQuarantineCommander struct {
	fakePositionPolicyCommander
	clock            *fakeClock
	advance          time.Duration
	runtimeCalls     int
	listCalls        int
	quarantineCalls  int
	afterQuarantines func()
}

func (c *a111AdvancingQuarantineCommander) Runtime(ctx context.Context) (positionpolicy.ManagementRuntime, error) {
	c.runtimeCalls++
	return c.fakePositionPolicyCommander.Runtime(ctx)
}

func (c *a111AdvancingQuarantineCommander) List(ctx context.Context) ([]positionpolicy.State, error) {
	c.listCalls++
	return c.fakePositionPolicyCommander.List(ctx)
}

func (c *a111AdvancingQuarantineCommander) Quarantines(context.Context) ([]exitquarantine.Row, error) {
	c.quarantineCalls++
	if c.afterQuarantines != nil {
		c.afterQuarantines()
		return nil, nil
	}
	c.clock.advance(c.advance)
	return nil, nil
}

func (*a111AdvancingQuarantineCommander) PreviewQuarantineRelease(context.Context,
	exitquarantine.Request) (exitquarantine.Preview, error) {
	panic("unexpected quarantine preview")
}

func (*a111AdvancingQuarantineCommander) ReleaseQuarantine(context.Context,
	exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	panic("unexpected quarantine release")
}

// a111MarkerReadClock advances only on the second direct response-time sample
// taken by handlePositionManagement after Quarantines returns. Calls made later
// by policyAction for token timestamps cannot stand in for that sample.
type a111MarkerReadClock struct {
	mu             sync.Mutex
	now            time.Time
	advance        time.Duration
	armed          bool
	handlerSamples int
}

func newA111MarkerReadClock() *a111MarkerReadClock {
	return &a111MarkerReadClock{now: time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC)}
}

func (c *a111MarkerReadClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.armed {
		return c.now
	}
	pc, _, _, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if !ok || fn == nil || !strings.Contains(fn.Name(), ".handlePositionManagement") {
		return c.now
	}
	c.handlerSamples++
	if c.handlerSamples == 2 {
		c.now = c.now.Add(c.advance)
	}
	return c.now
}

func (c *a111MarkerReadClock) arm(advance time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advance, c.armed, c.handlerSamples = advance, true, 0
}

func (c *a111MarkerReadClock) samples() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.handlerSamples
}

// a111DecorateReadClock moves only between the marker-read and response-time
// samples owned by decoratePositionRows. The overview handler takes another
// clock sample for its broker strip after decoration; that later sample cannot
// stand in for the safety verdict already attached to the row.
type a111DecorateReadClock struct {
	mu              sync.Mutex
	now             time.Time
	advance         time.Duration
	armed           bool
	decorateSamples int
}

func newA111DecorateReadClock() *a111DecorateReadClock {
	return &a111DecorateReadClock{now: time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC)}
}

func (c *a111DecorateReadClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.armed {
		return c.now
	}
	pc, _, _, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if !ok || fn == nil || !strings.Contains(fn.Name(), ".decoratePositionRows") {
		return c.now
	}
	c.decorateSamples++
	if c.decorateSamples == 2 {
		c.now = c.now.Add(c.advance)
	}
	return c.now
}

func (c *a111DecorateReadClock) arm(advance time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advance, c.armed, c.decorateSamples = advance, true, 0
}

func (c *a111DecorateReadClock) samples() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.decorateSamples
}

func a111ConsoleSnapshot(observedAt string) journal.ExitSnapshotView {
	return journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
		Line: exitpolicy.ExitLineSnapshot{
			SnapshotID: "a111-console-snapshot", DecisionID: "a111-console-decision",
			ObservationID: "a111-console-observation", PositionID: "a111-console-position",
			PositionGeneration: 1, EntryPrice: "200", InitialStop: "190", ObservedPrice: "205",
			CurrentProtection: "195", HighWater: "210", RatchetLevel: exitpolicy.LevelNone,
			ActiveRung: exitpolicy.NoRung, NextTarget: "220", NextProtection: "200",
			Action: exitpolicy.ActionNone, ProjectedQuantity: "0",
		},
		ObservationSource: "quote_fetched_at", ObservedAt: observedAt,
	}}
}

func TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"
	commander := &fakePositionPolicyCommander{state: state}
	h := newLivePositionHarness(t, func(options *Options) { options.PositionPolicies = commander })
	line := seedUnchangedJudgement(t, h, "2026-07-27T00:59:30Z") // exact 30-second boundary
	holdEngineMarker(t, h.marker, h.clock.Now())
	h.authenticate(t)

	positions := positionHTMLRow(t, h.page(t, "/positions"), state.Symbol)
	management := h.page(t, "/position-management")
	for _, want := range []string{
		`class="safety-line"`,
		`role="group" aria-label="현재 exit 보호선과 다음 단계"`,
		`.safety-line { display: block; margin-top: 0.2rem; font-size: 1rem;`,
	} {
		if !strings.Contains(management, want) {
			t.Errorf("/position-management is missing safety-line presentation %q:\n%s", want, management)
		}
	}
	if strings.Contains(management, `class="submetric">현재 보호선`) {
		t.Errorf("/position-management renders safety values as provenance submetrics:\n%s", management)
	}
	for path, rendered := range map[string]string{
		"/positions":           positions,
		"/position-management": management,
	} {
		if !strings.Contains(rendered, "평가 완료") {
			t.Errorf("%s does not identify the exact-boundary line as evaluated:\n%s", path, rendered)
		}
		for _, value := range exitLineValues(line) {
			if strings.TrimSpace(value) != "" && !strings.Contains(rendered, value) {
				t.Errorf("%s is missing actionable line value %q:\n%s", path, value, rendered)
			}
		}
	}
}

func TestA111PositionManagementRechecksFreshnessAfterQuarantineRead(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"

	assertPostQuarantineStale := func(t *testing.T, h *dashboardHarness,
		commander *a111AdvancingQuarantineCommander, line exitpolicy.ExitLineSnapshot,
		wantStatus, wantReason string) {
		t.Helper()
		h.authenticate(t)
		commander.runtimeCalls, commander.listCalls, commander.quarantineCalls = 0, 0, 0
		page := h.page(t, "/position-management")
		if commander.runtimeCalls != 1 || commander.listCalls != 1 || commander.quarantineCalls != 1 {
			t.Fatalf("position-management reads = runtime %d, list %d, quarantines %d; want one each",
				commander.runtimeCalls, commander.listCalls, commander.quarantineCalls)
		}
		marker := "<code>" + state.AccountRef + "</code>"
		at := strings.Index(page, marker)
		if at < 0 {
			t.Fatalf("position-management row is missing %q", marker)
		}
		start := strings.LastIndex(page[:at], "<tr")
		end := strings.Index(page[at:], "</tr>")
		if start < 0 || end < 0 {
			t.Fatalf("position-management row is not bounded around %q", marker)
		}
		row := page[start : at+end+len("</tr>")]
		for _, want := range []string{
			wantStatus,
			wantReason,
			`현재 보호선 <strong>—</strong> · 다음 익절 <strong>—</strong> · 다음 보호선 <strong>—</strong>`,
			`관측가 — · 고점 — · 평가 2026-07-27T01:00:00Z`,
		} {
			if !strings.Contains(row, want) {
				t.Errorf("post-quarantine row is missing %q:\n%s", want, row)
			}
		}
		assertLineClosed(t, row, line)
		for _, falseJournalState := range []string{
			"exit 평가 근거를 확인할 수 없어",
			"저장된 exit 평가를 찾을 수 없다",
			"원장을 읽을 수 없다",
		} {
			if strings.Contains(page, falseJournalState) {
				t.Errorf("readable journal was reported as %q:\n%s", falseJournalState, page)
			}
		}
	}

	t.Run("unwired marker crosses observation bound", func(t *testing.T) {
		commander := &a111AdvancingQuarantineCommander{
			fakePositionPolicyCommander: fakePositionPolicyCommander{state: state},
			advance:                     30*time.Second + time.Nanosecond,
		}
		h := newDashboardHarness(t, func(options *Options) { options.PositionPolicies = commander })
		commander.clock = h.clock
		seedJournal(t, h.journal)
		line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a111-rpc-age",
			"70500", "70000", "68000", "0", exitpolicy.LevelNone)
		writeViewSnapshot(t, h.journal, line, recovery, h.clock.Now().Format(time.RFC3339Nano))
		assertPostQuarantineStale(t, h, commander, line,
			"오래된 평가", "평가 시각이 표시 허용 범위를 지났다")
	})

	t.Run("running marker crosses engine stale bound", func(t *testing.T) {
		commander := &a111AdvancingQuarantineCommander{
			fakePositionPolicyCommander: fakePositionPolicyCommander{state: state},
			advance:                     enginelock.StaleAfter + time.Second,
		}
		h := newLivePositionHarness(t, func(options *Options) { options.PositionPolicies = commander })
		commander.clock = h.clock
		line := seedUnchangedJudgement(t, h, h.clock.Now().Format(time.RFC3339Nano))
		holdEngineMarker(t, h.marker, h.clock.Now())
		assertPostQuarantineStale(t, h.dashboardHarness, commander, line,
			"엔진 정지", "엔진이 실행 중이 아니어서 보호선이 갱신되지 않는다")
	})
}

func TestA111PositionManagementSamplesResponseTimeAfterMarkerRead(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"

	assertPostMarkerStale := func(t *testing.T, h *dashboardHarness,
		clock *a111MarkerReadClock, commander *a111AdvancingQuarantineCommander,
		line exitpolicy.ExitLineSnapshot, wantStatus, wantReason string) {
		t.Helper()
		h.authenticate(t)
		commander.runtimeCalls, commander.listCalls, commander.quarantineCalls = 0, 0, 0
		page := h.page(t, "/position-management")
		if commander.runtimeCalls != 1 || commander.listCalls != 1 || commander.quarantineCalls != 1 {
			t.Fatalf("position-management reads = runtime %d, list %d, quarantines %d; want one each",
				commander.runtimeCalls, commander.listCalls, commander.quarantineCalls)
		}
		if got := clock.samples(); got < 2 {
			t.Errorf("handlePositionManagement response-time samples after quarantine = %d, want pre/post marker samples", got)
		}
		marker := "<code>" + state.AccountRef + "</code>"
		at := strings.Index(page, marker)
		if at < 0 {
			t.Fatalf("position-management row is missing %q", marker)
		}
		start := strings.LastIndex(page[:at], "<tr")
		end := strings.Index(page[at:], "</tr>")
		if start < 0 || end < 0 {
			t.Fatalf("position-management row is not bounded around %q", marker)
		}
		row := page[start : at+end+len("</tr>")]
		for _, want := range []string{
			wantStatus,
			wantReason,
			`현재 보호선 <strong>—</strong> · 다음 익절 <strong>—</strong> · 다음 보호선 <strong>—</strong>`,
			`관측가 — · 고점 — · 평가 2026-07-27T01:00:00Z`,
		} {
			if !strings.Contains(row, want) {
				t.Errorf("post-marker row is missing %q:\n%s", want, row)
			}
		}
		assertLineClosed(t, row, line)
		for _, falseJournalState := range []string{
			"exit 평가 근거를 확인할 수 없어",
			"저장된 exit 평가를 찾을 수 없다",
			"원장을 읽을 수 없다",
		} {
			if strings.Contains(page, falseJournalState) {
				t.Errorf("readable journal was reported as %q:\n%s", falseJournalState, page)
			}
		}
	}

	newHarness := func(t *testing.T, clock *a111MarkerReadClock,
		commander *a111AdvancingQuarantineCommander, engineMarker string) *dashboardHarness {
		t.Helper()
		return newDashboardHarness(t, func(options *Options) {
			options.PositionPolicies = commander
			options.EngineMarker = engineMarker
			options.Now = clock.Now
		})
	}

	t.Run("unwired marker crosses observation bound during marker read", func(t *testing.T) {
		clock := newA111MarkerReadClock()
		commander := &a111AdvancingQuarantineCommander{
			fakePositionPolicyCommander: fakePositionPolicyCommander{state: state},
		}
		commander.afterQuarantines = func() { clock.arm(30*time.Second + time.Nanosecond) }
		h := newHarness(t, clock, commander, "")
		seedJournal(t, h.journal)
		line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a111-marker-age",
			"70500", "70000", "68000", "0", exitpolicy.LevelNone)
		writeViewSnapshot(t, h.journal, line, recovery, clock.Now().Format(time.RFC3339Nano))
		assertPostMarkerStale(t, h, clock, commander, line,
			"오래된 평가", "평가 시각이 표시 허용 범위를 지났다")
	})

	t.Run("running marker crosses engine stale bound during marker read", func(t *testing.T) {
		clock := newA111MarkerReadClock()
		commander := &a111AdvancingQuarantineCommander{
			fakePositionPolicyCommander: fakePositionPolicyCommander{state: state},
		}
		commander.afterQuarantines = func() { clock.arm(2 * time.Nanosecond) }
		marker := filepath.Join(t.TempDir(), enginelock.MarkerFileName)
		h := newHarness(t, clock, commander, marker)
		seedJournal(t, h.journal)
		line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a111-marker-stop",
			"70500", "70000", "68000", "0", exitpolicy.LevelNone)
		writeViewSnapshot(t, h.journal, line, recovery, clock.Now().Format(time.RFC3339Nano))
		holdEngineMarker(t, marker, clock.Now().Add(-enginelock.StaleAfter+time.Nanosecond))
		assertPostMarkerStale(t, h, clock, commander, line,
			"엔진 정지", "엔진이 실행 중이 아니어서 보호선이 갱신되지 않는다")
	})
}

func TestA111PositionManagementNeverResurrectsAStoppedMarkerAfterClockRollback(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"
	clock := newA111MarkerReadClock()
	commander := &a111AdvancingQuarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: state},
	}
	// The first handler sample reads the marker while it is already stale. The
	// second sample rolls wall time backward across that boundary; it may make
	// RefreshedAt look recent, but it cannot overturn the stopped read verdict.
	commander.afterQuarantines = func() { clock.arm(-2 * time.Nanosecond) }
	markerPath := filepath.Join(t.TempDir(), enginelock.MarkerFileName)
	h := newDashboardHarness(t, func(options *Options) {
		options.PositionPolicies = commander
		options.EngineMarker = markerPath
		options.Now = clock.Now
	})
	seedJournal(t, h.journal)
	line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a111-marker-rollback",
		"70500", "70000", "68000", "0", exitpolicy.LevelNone)
	observedAt := clock.Now().Add(-time.Second).Format(time.RFC3339Nano)
	writeViewSnapshot(t, h.journal, line, recovery, observedAt)
	holdEngineMarker(t, markerPath, clock.Now().Add(-enginelock.StaleAfter-time.Nanosecond))

	h.authenticate(t)
	commander.runtimeCalls, commander.listCalls, commander.quarantineCalls = 0, 0, 0
	page := h.page(t, "/position-management")
	if commander.runtimeCalls != 1 || commander.listCalls != 1 || commander.quarantineCalls != 1 {
		t.Fatalf("position-management reads = runtime %d, list %d, quarantines %d; want one each",
			commander.runtimeCalls, commander.listCalls, commander.quarantineCalls)
	}
	if got := clock.samples(); got != 2 {
		t.Fatalf("handlePositionManagement marker-bound samples = %d, want exactly pre/post", got)
	}
	marker := "<code>" + state.AccountRef + "</code>"
	at := strings.Index(page, marker)
	if at < 0 {
		t.Fatalf("position-management row is missing %q", marker)
	}
	start := strings.LastIndex(page[:at], "<tr")
	end := strings.Index(page[at:], "</tr>")
	if start < 0 || end < 0 {
		t.Fatalf("position-management row is not bounded around %q", marker)
	}
	row := page[start : at+end+len("</tr>")]
	for _, want := range []string{
		"엔진 정지",
		"엔진이 실행 중이 아니어서 보호선이 갱신되지 않는다",
		`현재 보호선 <strong>—</strong> · 다음 익절 <strong>—</strong> · 다음 보호선 <strong>—</strong>`,
		`관측가 — · 고점 — · 평가 ` + observedAt,
	} {
		if !strings.Contains(row, want) {
			t.Errorf("rolled-back marker row is missing %q:\n%s", want, row)
		}
	}
	assertLineClosed(t, row, line)
	for _, falseJournalState := range []string{
		"exit 평가 근거를 확인할 수 없어",
		"저장된 exit 평가를 찾을 수 없다",
		"원장을 읽을 수 없다",
	} {
		if strings.Contains(page, falseJournalState) {
			t.Errorf("readable journal was reported as %q:\n%s", falseJournalState, page)
		}
	}
}

func TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"

	for _, route := range []string{"/positions", "/dashboard"} {
		route := route
		t.Run(strings.TrimPrefix(route, "/")+"/observation age", func(t *testing.T) {
			commander := &countingPolicyCommander{states: []positionpolicy.State{state}}
			h := newDashboardHarness(t, func(options *Options) { options.PositionPolicies = commander })
			seedJournal(t, h.journal)
			line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a111-policy-age",
				"70500", "70000", "68000", "0", exitpolicy.LevelNone)
			writeViewSnapshot(t, h.journal, line, recovery, h.clock.Now().Format(time.RFC3339Nano))

			h.authenticate(t)
			h.page(t, "/positions") // fill broker and policy caches before isolating the policy miss
			h.Console.enginePolicy.invalidate()
			runtimeBefore, listBefore := commander.reads()
			brokerBefore := h.holdings.count()
			var advanceOnce sync.Once
			commander.mu.Lock()
			commander.duringRead = func() {
				advanceOnce.Do(func() { h.clock.advance(30*time.Second + time.Nanosecond) })
			}
			commander.mu.Unlock()

			row := positionHTMLRow(t, h.page(t, route), state.Symbol)
			runtimeAfter, listAfter := commander.reads()
			if runtimeAfter-runtimeBefore != 1 || listAfter-listBefore != 1 {
				t.Fatalf("%s policy cache miss reads = runtime %d, list %d; want one each",
					route, runtimeAfter-runtimeBefore, listAfter-listBefore)
			}
			if got := h.holdings.count() - brokerBefore; got != 0 {
				t.Fatalf("%s made %d broker reads while rechecking a policy-delayed line; want none", route, got)
			}
			assertA111HoldingsRouteClosed(t, route, row, line,
				"오래된 평가", "평가 시각이 표시 허용 범위를 지났다")
		})

		t.Run(strings.TrimPrefix(route, "/")+"/engine marker", func(t *testing.T) {
			commander := &countingPolicyCommander{states: []positionpolicy.State{state}}
			h := newLivePositionHarness(t, func(options *Options) { options.PositionPolicies = commander })
			line := seedUnchangedJudgement(t, h, h.clock.Now().Format(time.RFC3339Nano))
			holdEngineMarker(t, h.marker, h.clock.Now())

			h.authenticate(t)
			h.page(t, "/positions")
			h.Console.enginePolicy.invalidate()
			runtimeBefore, listBefore := commander.reads()
			brokerBefore := h.holdings.count()
			var advanceOnce sync.Once
			commander.mu.Lock()
			commander.duringRead = func() {
				advanceOnce.Do(func() { h.clock.advance(enginelock.StaleAfter + time.Nanosecond) })
			}
			commander.mu.Unlock()

			row := positionHTMLRow(t, h.page(t, route), state.Symbol)
			runtimeAfter, listAfter := commander.reads()
			if runtimeAfter-runtimeBefore != 1 || listAfter-listBefore != 1 {
				t.Fatalf("%s policy cache miss reads = runtime %d, list %d; want one each",
					route, runtimeAfter-runtimeBefore, listAfter-listBefore)
			}
			if got := h.holdings.count() - brokerBefore; got != 0 {
				t.Fatalf("%s made %d broker reads while rechecking a policy-delayed line; want none", route, got)
			}
			assertA111HoldingsRouteClosed(t, route, row, line,
				"엔진 정지", "엔진이 실행 중이 아니어서 보호선이 갱신되지 않는다")
		})
	}
}

func TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"

	for _, route := range []string{"/positions", "/dashboard"} {
		route := route
		t.Run(strings.TrimPrefix(route, "/"), func(t *testing.T) {
			clock := newA111DecorateReadClock()
			commander := &countingPolicyCommander{states: []positionpolicy.State{state}}
			markerPath := filepath.Join(t.TempDir(), enginelock.MarkerFileName)
			h := newDashboardHarness(t, func(options *Options) {
				options.PositionPolicies = commander
				options.EngineMarker = markerPath
				options.Now = clock.Now
			})
			seedJournal(t, h.journal)
			line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a111-policy-rollback",
				"70500", "70000", "68000", "0", exitpolicy.LevelNone)
			observedAt := clock.Now().Add(-time.Second).Format(time.RFC3339Nano)
			writeViewSnapshot(t, h.journal, line, recovery, observedAt)
			holdEngineMarker(t, markerPath, clock.Now().Add(-enginelock.StaleAfter-time.Nanosecond))

			h.authenticate(t)
			h.page(t, "/positions")
			h.Console.enginePolicy.invalidate()
			runtimeBefore, listBefore := commander.reads()
			brokerBefore := h.holdings.count()
			var armOnce sync.Once
			commander.mu.Lock()
			commander.duringRead = func() { armOnce.Do(func() { clock.arm(-2 * time.Nanosecond) }) }
			commander.mu.Unlock()

			row := positionHTMLRow(t, h.page(t, route), state.Symbol)
			runtimeAfter, listAfter := commander.reads()
			if runtimeAfter-runtimeBefore != 1 || listAfter-listBefore != 1 {
				t.Fatalf("%s policy cache miss reads = runtime %d, list %d; want one each",
					route, runtimeAfter-runtimeBefore, listAfter-listBefore)
			}
			if got := h.holdings.count() - brokerBefore; got != 0 {
				t.Fatalf("%s made %d broker reads while rechecking a rolled-back marker; want none", route, got)
			}
			if got := clock.samples(); got != 2 {
				t.Fatalf("%s decorate marker-bound samples = %d, want exactly pre/post", route, got)
			}
			assertA111HoldingsRouteClosed(t, route, row, line,
				"엔진 정지", "엔진이 실행 중이 아니어서 보호선이 갱신되지 않는다")
		})
	}
}

func assertA111HoldingsRouteClosed(t *testing.T, route, row string, line exitpolicy.ExitLineSnapshot,
	wantStatus, wantReason string) {
	t.Helper()
	for _, want := range []string{
		wantStatus,
		wantReason,
		`<span>익절</span><strong>—</strong>`,
		`<span>손절</span><strong>—</strong>`,
		`<span>추적 회수</span><strong>—</strong>`,
		`<span>기준</span><strong>—</strong>`,
		`<span>고점</span><strong>—</strong>`,
		`진입가 <strong>—</strong> · 관측가 <strong>—</strong>`,
		`현재 보호선 <strong>—</strong> · 다음 익절 <strong>—</strong>`,
		`최초 손절 <strong>—</strong> · 워터마크 <strong>—</strong>`,
		`다음 보호선 <strong>—</strong> · 예상 수량 <strong>—</strong>`,
	} {
		if !strings.Contains(row, want) {
			t.Errorf("%s delayed row is missing %q:\n%s", route, want, row)
		}
	}
	assertLineClosed(t, row, line)
}

func TestA111PositionManagementKeepsJournalFailureTruth(t *testing.T) {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.AccountRef = "123-45-678901"

	assertFailedClosed := func(t *testing.T, h *livePositionHarness, wantState, wantDetail string) {
		t.Helper()
		h.authenticate(t)
		page := h.page(t, "/position-management")
		for _, want := range []string{
			wantState,
			"exit 평가 근거를 확인할 수 없어 현재 보호선과 다음 exit 단계를 표시하지 않는다.",
			`현재 보호선 <strong>—</strong> · 다음 익절 <strong>—</strong> · 다음 보호선 <strong>—</strong>`,
		} {
			if !strings.Contains(page, want) {
				t.Errorf("journal failure page is missing %q:\n%s", want, page)
			}
		}
		if strings.Contains(page, "저장된 exit 평가를 찾을 수 없다") {
			t.Errorf("journal failure was misreported as no saved evaluation:\n%s", page)
		}
		if wantDetail != "" && !strings.Contains(page, wantDetail) {
			t.Errorf("journal failure page discarded detail %q:\n%s", wantDetail, page)
		}
	}

	t.Run("unwired", func(t *testing.T) {
		h := newLivePositionHarness(t, func(options *Options) {
			options.PositionPolicies = &fakePositionPolicyCommander{state: state}
			options.JournalPath = ""
		})
		assertFailedClosed(t, h, "원장 경로가 배선되지 않았다", "")
	})

	t.Run("open failure", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(blocker, []byte("blocked"), 0o600); err != nil {
			t.Fatalf("write journal path blocker: %v", err)
		}
		h := newLivePositionHarness(t, func(options *Options) {
			options.PositionPolicies = &fakePositionPolicyCommander{state: state}
			options.JournalPath = filepath.Join(blocker, journal.DBFileName)
		})
		assertFailedClosed(t, h, "원장을 읽을 수 없다", filepath.Join(blocker, journal.DBFileName))
	})

	t.Run("read failure", func(t *testing.T) {
		h := newLivePositionHarness(t, func(options *Options) {
			options.PositionPolicies = &fakePositionPolicyCommander{state: state}
		})
		seedJournal(t, h.journal)
		execRaw(t, h.journal, "DROP TABLE exit_snapshot_quarantines")
		assertFailedClosed(t, h, "원장을 읽을 수 없다", "no such table: exit_snapshot_quarantines")
	})
}

func TestA111ConsoleConsumesTheSharedFreshnessVerdict(t *testing.T) {
	observed := time.Date(2026, 8, 14, 6, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		local  protectionLiveness
		shared operatorview.ExitLiveness
	}{
		{"running", protectionLiveness{Wired: true, Running: true}, operatorview.ExitLivenessRunning},
		{"stopped", protectionLiveness{Wired: true, Running: false}, operatorview.ExitLivenessStopped},
		{"unwired", protectionLiveness{}, operatorview.ExitLivenessUnwired},
	} {
		for _, age := range []time.Duration{29999 * time.Millisecond, 30 * time.Second, 30*time.Second + time.Nanosecond} {
			t.Run(tc.name+"/"+age.String(), func(t *testing.T) {
				view := a111ConsoleSnapshot(observed.Format(time.RFC3339Nano))
				got := exitFreshness(view, observed.Add(age), tc.local)
				want := operatorview.ApplyExitFreshness(view, observed.Add(age), tc.shared)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("console freshness = %+v, shared = %+v", got, want)
				}
			})
		}
	}
}

func TestA111RunningConsoleDoesNotLetAValidSiblingKeepInvalidEvidenceFresh(t *testing.T) {
	now := time.Date(2026, 8, 14, 6, 0, 31, 0, time.UTC)
	live := protectionLiveness{Wired: true, Running: true}
	fresh := exitFreshness(a111ConsoleSnapshot(now.Format(time.RFC3339Nano)), now, live)
	aged := exitFreshness(a111ConsoleSnapshot(now.Add(-31*time.Second).Format(time.RFC3339Nano)), now, live)
	if fresh.Stale {
		t.Fatalf("valid sibling unexpectedly stale: %+v", fresh)
	}
	if !aged.Stale || aged.StaleReason != "observation_older_than_limit" {
		t.Fatalf("invalid sibling borrowed running liveness: %+v", aged)
	}
}

func TestA111ConsoleHidesStoppedFutureSeedAndCorruptEvidence(t *testing.T) {
	now := time.Date(2026, 8, 14, 6, 0, 0, 0, time.UTC)
	stopped := exitFreshness(a111ConsoleSnapshot(now.Format(time.RFC3339Nano)), now,
		protectionLiveness{Wired: true})
	stoppedLine := operatorview.BuildExitLine(operatorview.Source{
		Snapshot: &stopped.Snapshot.Line, StaleReason: stopped.StaleReason,
	})
	if !stoppedLine.Stale() || stoppedLine.Reason == "" ||
		stoppedLine.CurrentProtection != "—" || stoppedLine.NextTarget != "—" {
		t.Fatalf("stopped line became actionable: %+v", stoppedLine)
	}

	future := exitFreshness(a111ConsoleSnapshot(now.Add(time.Nanosecond).Format(time.RFC3339Nano)), now,
		protectionLiveness{Wired: true, Running: true})
	if !future.Stale || future.StaleReason != "observation_in_future" {
		t.Fatalf("future evidence = %+v", future)
	}
	for _, reason := range []string{"not_evaluated_yet", "invalid_effective_snapshot", "partial_evaluated_tuple"} {
		line := operatorview.BuildExitLine(operatorview.Source{UnknownReason: reason})
		if !line.Unknown() || line.CurrentProtection != "—" || line.NextTarget != "—" {
			t.Fatalf("%s became actionable: %+v", reason, line)
		}
	}
}
