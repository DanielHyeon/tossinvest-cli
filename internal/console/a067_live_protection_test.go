package console

// a067: the protection line is a state, not a measurement.
//
// The engine writes exit_states.last_observed_at only from record(), and
// judgeRatchet/judgeLadder return before record() whenever the judgement moved
// nothing. So that column is the last time the line *changed*, and treating it
// as an observation heartbeat under the broker cache TTL emptied the whole 라인
// column on a perfectly healthy engine. These tests pin the replacement: the
// line is drawn while something is maintaining it, and closed when nothing is —
// or when this position has been taken out of judgement entirely.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// livePositionHarness is the dashboard harness with an engine marker wired, which
// is what production always has (cmd/tossctl/console.go always resolves one) and
// what no other console test has ever had.
type livePositionHarness struct {
	*dashboardHarness
	marker string
}

func newLivePositionHarness(t *testing.T, tweak ...func(*Options)) *livePositionHarness {
	t.Helper()
	marker := filepath.Join(t.TempDir(), enginelock.MarkerFileName)
	opts := append([]func(*Options){func(o *Options) { o.EngineMarker = marker }}, tweak...)
	return &livePositionHarness{dashboardHarness: newDashboardHarness(t, opts...), marker: marker}
}

// seedUnchangedJudgement writes a canonical snapshot whose last change is
// observedAt — the 042660/IONQ/TSLA shape from the 2026-08-03 ledger, where the
// price sits below the recorded high-water so every judgement since has moved
// nothing and therefore persisted nothing.
func seedUnchangedJudgement(t *testing.T, h *livePositionHarness, observedAt string) exitpolicy.ExitLineSnapshot {
	t.Helper()
	seedJournal(t, h.journal)
	line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a067",
		"70500", "70000", "68000", "0", exitpolicy.LevelNone)
	writeViewSnapshot(t, h.journal, line, recovery, observedAt)
	return line
}

// exitLineValues are the five the 라인 column shows, in the spec's order.
func exitLineValues(line exitpolicy.ExitLineSnapshot) []string {
	return []string{line.NextTarget, line.InitialStop, line.NextProtection,
		line.CurrentProtection, line.HighWater}
}

func assertLineShown(t *testing.T, row string, line exitpolicy.ExitLineSnapshot) {
	t.Helper()
	for _, want := range exitLineValues(line) {
		if strings.TrimSpace(want) == "" {
			continue
		}
		if !strings.Contains(row, want) {
			t.Errorf("the 라인 column is missing %q:\n%s", want, row)
		}
	}
}

func assertLineClosed(t *testing.T, row string, line exitpolicy.ExitLineSnapshot) {
	t.Helper()
	for _, unwanted := range exitLineValues(line) {
		if strings.TrimSpace(unwanted) == "" {
			continue
		}
		if strings.Contains(row, unwanted) {
			t.Errorf("the 라인 column exposes %q when it should be closed:\n%s", unwanted, row)
		}
	}
}

// TestTheProtectionLineSurvivesAJudgementThatChangedNothing is a067 tasks 2.1 and
// 2.2: the ledger's own values, hours after the last change, on both screens that
// draw a protection line.
func TestTheProtectionLineSurvivesAJudgementThatChangedNothing(t *testing.T) {
	h := newLivePositionHarness(t)
	line := seedUnchangedJudgement(t, h, "2026-07-26T19:00:00Z") // six hours before the clock
	holdEngineMarker(t, h.marker, h.clock.Now())
	h.authenticate(t)

	for _, screen := range []string{"/positions", "/dashboard"} {
		row := positionHTMLRow(t, h.page(t, screen), "005930")
		if !strings.Contains(row, "평가 완료") {
			t.Errorf("%s calls a live engine's protection line stale:\n%s", screen, row)
		}
		assertLineShown(t, row, line)
	}
}

// TestAStoppedEngineClosesTheProtectionLine is task 2.3. Nothing is maintaining
// the line, and the screen says which of the two things is wrong.
func TestAStoppedEngineClosesTheProtectionLine(t *testing.T) {
	h := newLivePositionHarness(t)
	line := seedUnchangedJudgement(t, h, "2026-07-27T00:59:45Z") // fifteen seconds old
	holdEngineMarker(t, h.marker, h.clock.Now().Add(-2*enginelock.StaleAfter))
	h.authenticate(t)

	page := h.page(t, "/positions")
	row := positionHTMLRow(t, page, "005930")
	assertLineClosed(t, row, line)
	if !strings.Contains(row, "엔진 정지") {
		t.Errorf("a stopped engine is not named on the row:\n%s", row)
	}
	if !strings.Contains(page, "엔진이 실행 중이 아니어서 보호선이 갱신되지 않는다") {
		t.Error("the page does not say why the protection line is closed")
	}
}

// TestAQuarantinedPositionIsNotDrawnAsProtected is tasks 2.4 and 2.5.
//
// This is the safety precondition of the whole change. Before a067 the console
// had never read exit_snapshot_quarantines, and the only reason a quarantined
// row was not drawn with an actionable line was that the age bound happened to
// catch it. Take the age bound away without this and the screen starts asserting
// protection for a position the engine refuses to judge at all.
func TestAQuarantinedPositionIsNotDrawnAsProtected(t *testing.T) {
	h := newLivePositionHarness(t)
	line := seedUnchangedJudgement(t, h, "2026-07-27T00:59:45Z")
	// pos-seed has an exit state but no canonical snapshot — the 042660 shape,
	// adopted and never yet moved by a judgement. Its reason would otherwise be
	// "아직 exit 평가가 기록되지 않았다", which is true and useless: the fact that
	// matters is that no judgement is coming.
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,
		 quantity,avg_price,opened_at)
		VALUES ('pos-seed','123-45-678901','kr','111111',1,'d-1','OPEN','2','83500','2026-07-27T00:30:00Z');
		INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,
		 baseline_price,high_water,ratchet_level,taken_ratio_total,completed,updated_at)
		VALUES ('pos-seed','RATCHET','83500','81000','2500','81000','83500','NONE','0',0,'2026-07-27T00:30:00Z');
		INSERT INTO exit_snapshot_quarantines
		 (position_id,position_generation,quarantine_version,reason,evidence,quarantined_at)
		VALUES ('pos-managed',1,1,'ambiguous_recovery',
		        'exitpolicy: recovery candidate identity mismatch','2026-07-27T00:30:00Z'),
		       ('pos-seed',1,1,'ambiguous_recovery','no usable protection state','2026-07-27T00:31:00Z');`)
	holdEngineMarker(t, h.marker, h.clock.Now())
	h.authenticate(t)

	// Both screens, because the spec forbids two screens from giving one holding
	// two different protection answers — and both read the same ledger seam.
	for _, screen := range []string{"/positions", "/dashboard"} {
		page := h.page(t, screen)
		row := positionHTMLRow(t, page, "005930")
		assertLineClosed(t, row, line)
		if !strings.Contains(row, "판정 격리") {
			t.Errorf("%s does not mark a quarantined position as unjudged:\n%s", screen, row)
		}
		if !strings.Contains(page, "이 포지션은 지금 exit 판정 대상이 아니다") {
			t.Errorf("%s does not say the position is out of judgement", screen)
		}
		seed := positionHTMLRow(t, page, "111111")
		if !strings.Contains(seed, "이 포지션은 지금 exit 판정 대상이 아니다") {
			t.Errorf("%s loses the quarantine when there is no canonical snapshot:\n%s", screen, seed)
		}
	}
}

// TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne is task 2.5. The
// active-quarantine index is per (position, generation) and the engine only ever
// asks about the current one; a re-adopted position must not stay closed because
// its previous instance was quarantined.
func TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne(t *testing.T) {
	h := newLivePositionHarness(t)
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		UPDATE positions SET instance_seq=2 WHERE id='pos-managed';
		INSERT INTO exit_snapshot_quarantines
		 (position_id,position_generation,quarantine_version,reason,evidence,quarantined_at)
		VALUES ('pos-managed',1,1,'ambiguous_recovery','stale generation','2026-07-27T00:30:00Z');`)
	line, recovery := ratchetViewSnapshot(t, "pos-managed", 2, "10", "obs-a067-gen2",
		"70500", "70000", "68000", "0", exitpolicy.LevelNone)
	writeViewSnapshot(t, h.journal, line, recovery, "2026-07-26T19:00:00Z")
	holdEngineMarker(t, h.marker, h.clock.Now())
	h.authenticate(t)

	row := positionHTMLRow(t, h.page(t, "/positions"), "005930")
	if strings.Contains(row, "판정 격리") {
		t.Errorf("a previous generation's quarantine closed the current one:\n%s", row)
	}
	assertLineShown(t, row, line)
}

// TestAConsoleWithNoEngineMarkerKeepsTheObservationAgeBound is task 2.6.
//
// The operator-console spec forbids drawing an unwired marker as a stopped
// engine, so a console that cannot see the engine keeps exactly the judgement it
// had before a067 — including its wording. This is the regression guard for
// every existing console test, none of which wires a marker.
func TestAConsoleWithNoEngineMarkerKeepsTheObservationAgeBound(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a067-unwired",
		"70500", "70000", "68000", "0", exitpolicy.LevelNone)
	writeViewSnapshot(t, h.journal, line, recovery, "2026-07-26T19:00:00Z")
	h.authenticate(t)

	page := h.page(t, "/positions")
	row := positionHTMLRow(t, page, "005930")
	assertLineClosed(t, row, line)
	for _, want := range []string{"오래된 평가", "평가 시각이 표시 허용 범위를 지났다"} {
		if !strings.Contains(page, want) {
			t.Errorf("an unwired console lost the pre-a067 wording %q", want)
		}
	}
	if strings.Contains(page, "엔진 정지") {
		t.Error("an unwired marker was reported as a stopped engine")
	}
}

// TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted is tasks 2.7 and 2.8: a
// running engine does not excuse evidence that fails its own integrity checks.
func TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted(t *testing.T) {
	for _, c := range []struct{ name, observedAt, reason string }{
		{"observed in the future", "2026-07-27T02:00:00Z", "평가 시각이 현재보다 미래다"},
		// An unparseable stamp never reaches the freshness test: the stored
		// snapshot fails to decode first, and that is the reason the operator
		// gets. The case is here so a067 cannot quietly make a corrupt snapshot
		// renderable by removing the age bound above it.
		{"unparseable observation", "not-a-timestamp", "저장된 exit snapshot의 무결성을 확인할 수 없다"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newLivePositionHarness(t)
			line := seedUnchangedJudgement(t, h, c.observedAt)
			holdEngineMarker(t, h.marker, h.clock.Now())
			h.authenticate(t)

			page := h.page(t, "/positions")
			assertLineClosed(t, positionHTMLRow(t, page, "005930"), line)
			if !strings.Contains(page, c.reason) {
				t.Errorf("the page does not give the reason %q", c.reason)
			}
		})
	}
}

// --- the stock name on /position-management ------------------------------------

func usPolicyState() positionpolicy.State {
	state := managedPolicyState()
	state.PositionID, state.Market, state.Symbol = "p-2", "us", "IONQ"
	return state
}

type twoPolicyCommander struct {
	fakePositionPolicyCommander
	states []positionpolicy.State
}

func (c *twoPolicyCommander) List(context.Context) ([]positionpolicy.State, error) {
	return append([]positionpolicy.State(nil), c.states...), nil
}

// TestPositionManagementNamesTheStock is tasks 3.1, 3.2 and 3.4.
//
// The 계좌·종목 cell had no name field at all. US only looked correct because a
// ticker reads like a word: 아이온큐 was as absent as 한화오션 was.
func TestPositionManagementNamesTheStock(t *testing.T) {
	commander := &twoPolicyCommander{states: []positionpolicy.State{managedPolicyState(), usPolicyState()}}
	h := newDashboardHarness(t, func(o *Options) { o.PositionPolicies = commander })
	h.holdings.rows = []domain.Position{
		{Symbol: "005930", Name: "삼성전자", MarketType: "KR", Quantity: 10},
		{Symbol: "IONQ", Name: "아이온큐", MarketType: "US", Quantity: 4},
		// Same ticker, other market: its name must not reach the KR row.
		{Symbol: "005930", Name: "WRONG MARKET", MarketType: "US", Quantity: 1},
	}
	h.authenticate(t)
	h.page(t, "/positions") // the one screen that fills the cache

	page := h.page(t, "/position-management")
	for _, want := range []string{"삼성전자", "아이온큐"} {
		if !strings.Contains(page, want) {
			t.Errorf("/position-management does not name the stock %q", want)
		}
	}
	if strings.Contains(page, "WRONG MARKET") {
		t.Error("a name was taken from a holding in another market")
	}
}

// TestPositionManagementStaysSilentAndCallsNobodyWhenTheNameIsUnknown is task
// 3.3. This screen makes no broker call today and gains none here.
func TestPositionManagementStaysSilentAndCallsNobodyWhenTheNameIsUnknown(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState()}
	h := policyHarness(t, commander)
	h.authenticate(t)

	before := h.holdings.count()
	page := h.page(t, "/position-management")
	if got := h.holdings.count(); got != before {
		t.Errorf("/position-management made %d broker call(s); it must make none", got-before)
	}
	if !strings.Contains(page, "005930") {
		t.Error("the row disappeared when its name was unknown")
	}
	if strings.Contains(page, "삼성전자") {
		t.Error("a name was rendered from an empty holdings cache")
	}
}
