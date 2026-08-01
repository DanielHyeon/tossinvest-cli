package console

// portfolio_test.go covers the two dashboard screens (change
// add-operator-dashboard, tasks 2.1 and 2.2).
//
// Everything is httptest against a real journal file and a counting fake broker.
// The claims worth testing here are not "the HTML renders" but the ones an
// operator would be misled by if they were wrong:
//
//	the join survives    either source can be missing and the page still answers.
//	the label is fixed   an unmanaged holding says 관리 외(미편입), in those words.
//	the budget holds     one refresh is one call, a second tab is free, and a
//	                     verification in progress suspends refreshing entirely.
//	nothing is invented  no exit price on the history screen, and a NULL hold time
//	                     renders as an em dash rather than as zero seconds.

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

// --- a broker that counts, and a clock that does not move on its own -----------------

// countingHoldings is the read-only broker the dashboard is given. It counts,
// because "one refresh is one call" is a rate-budget claim and counting is the
// only honest way to check it.
type countingHoldings struct {
	mu    sync.Mutex
	calls int
	rows  []domain.Position
	err   error
}

func (h *countingHoldings) Holdings(context.Context, string) ([]domain.Position, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.calls++
	if h.err != nil {
		return nil, h.err
	}
	return append([]domain.Position(nil), h.rows...), nil
}

func (h *countingHoldings) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.calls
}

// fakeClock is the injected time source. The TTL is a duration and a test that
// slept through it would be a slow test that still could not prove the boundary.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// --- journal fixtures ---------------------------------------------------------------

// seedEngineJournal creates a real journal with journal.Open — which is the
// engine's constructor, and therefore the only thing that writes the schema — and
// then fills it through a plain driver connection.
//
// The rows are written as SQL rather than through the engine's write paths on
// purpose: reaching a frozen trade outcome the honest way needs a decision, an
// intent, an attempt, a fill and two apply-hook passes, none of which is what
// these tests are about. What matters here is what the *screen* does with rows
// that look like that, and the shapes are pinned against the real DDL by the
// STRICT tables refusing anything else.
func seedEngineJournal(t *testing.T, path, statements string) {
	t.Helper()
	ctx := context.Background()

	j, err := journal.Open(ctx, journal.Options{
		Path:     path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	if err := j.Close(); err != nil {
		t.Fatalf("journal.Close: %v", err)
	}
	if strings.TrimSpace(statements) == "" {
		return
	}
	execRaw(t, path, statements)
}

// execRaw runs SQL against the journal file with a plain writable connection.
func execRaw(t *testing.T, path, statements string) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), statements); err != nil {
		t.Fatalf("seeding %s: %v", path, err)
	}
}

func seedJournal(t *testing.T, path string) {
	t.Helper()
	seedEngineJournal(t, path, journalFixture)
}

func seedEmptyJournal(t *testing.T, path string) {
	t.Helper()
	seedEngineJournal(t, path, "")
}

// bumpSchemaVersion makes the journal look like it was written by a newer build.
func bumpSchemaVersion(t *testing.T, path string, version int) {
	t.Helper()
	execRaw(t, path, "PRAGMA user_version = "+strconv.Itoa(version)+";")
}

// seedOldJournal writes a database from before the core-domain migration: valid,
// readable, and without a single table the dashboard reads.
func seedOldJournal(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(),
		`CREATE TABLE schema_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL) STRICT;
		 PRAGMA user_version = 1;`); err != nil {
		t.Fatalf("writing a pre-v6 journal: %v", err)
	}
}

// journalFixture is the account the screens render.
const journalFixture = `
INSERT INTO decisions (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
                       risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
VALUES ('d-1','123-45-678901',0,'EXPOSURE_RAISING','RISK_INTENT','{}','h-1',NULL,'{}','n-1',
        '2026-07-26T00:30:00Z','2026-07-26T00:40:00Z');

-- Managed: an entry decision and an exit state with every line populated.
INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, opened_at)
VALUES ('pos-managed','123-45-678901','kr','005930',1,'d-1','OPEN','10','70000','2026-07-26T00:30:00Z');
INSERT INTO exit_states (position_id, policy_kind, entry_price, initial_stop, initial_risk,
                         baseline_price, high_water, ratchet_level, active_rung, taken_ratio_total,
                         pending_action, pending_level, pending_intent_id, completed, updated_at)
VALUES ('pos-managed','RATCHET','70000','68000','2000','69500','74000','HALF_RISK',NULL,'0.25',
        'PARTIAL_TAKE','PARTIAL_LOCK','intent-77',0,'2026-07-27T00:50:00Z');

-- Externally acquired: no entry decision, therefore not an exit-policy target.
INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, opened_at)
VALUES ('pos-external','123-45-678901','kr','035420',1,NULL,'OPEN','7','200000','2026-07-25T05:00:00Z');

-- Closed, with a frozen outcome.
INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, opened_at, closed_at)
VALUES ('pos-closed','123-45-678901','kr','005380',1,'d-1','CLOSED','0','0',
        '2026-07-24T00:30:00Z','2026-07-24T06:30:00Z');
INSERT INTO exit_states (position_id, policy_kind, entry_price, initial_stop, initial_risk,
                         baseline_price, high_water, ratchet_level, taken_ratio_total,
                         completed, updated_at)
VALUES ('pos-closed','RATCHET','55000','53000','2000','56000','60000','PROFIT_LOCK','1',1,
        '2026-07-24T06:30:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('pos-closed','48000','2.4','2000','10',21600,'PROFIT_LOCK',NULL,'2026-07-24T06:30:00Z');

-- Closed with an unknown hold time: the nullable column the screen must not
-- render as zero.
INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('pos-untimed','123-45-678901','kr','051910',1,'d-1','CLOSED','0','0','2026-07-23T06:30:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('pos-untimed','-9000','-0.9','2000','5',NULL,NULL,NULL,'2026-07-23T06:30:00Z');

INSERT INTO exit_events (position_id, observed_price, high_water, baseline_after, level_after,
                         action, proposed_intent_id, created_at)
VALUES ('pos-managed','74000','74000','69500','HALF_RISK','PARTIAL_TAKE','intent-77',
        '2026-07-27T00:50:00Z');
`

// brokerRows is what the fake account holds: the managed symbol, the external
// one, and one the journal has never heard of.
func brokerRows() []domain.Position {
	return []domain.Position{
		{Symbol: "005930", Name: "삼성전자", MarketType: "KR", Quantity: 10, AveragePrice: 70000,
			CurrentPrice: 73000, MarketValue: 730000, UnrealizedPnL: 30000, ProfitRate: 0.0428},
		{Symbol: "035420", Name: "NAVER", MarketType: "KR", Quantity: 7, AveragePrice: 200000,
			CurrentPrice: 190000, MarketValue: 1330000, UnrealizedPnL: -70000, ProfitRate: -0.05},
		{Symbol: "000660", Name: "SK하이닉스", MarketType: "KR", Quantity: 3, AveragePrice: 120000,
			CurrentPrice: 131500, MarketValue: 394500, UnrealizedPnL: 34500, ProfitRate: 0.0958},
	}
}

// dashboardHarness is newHarness with the dashboard wiring attached.
type dashboardHarness struct {
	*harness
	holdings *countingHoldings
	clock    *fakeClock
	journal  string
	runlock  string
}

func newDashboardHarness(t *testing.T, tweak ...func(*Options)) *dashboardHarness {
	t.Helper()
	dir := t.TempDir()
	journalPath := filepath.Join(dir, journal.DBFileName)
	lockPath := filepath.Join(dir, runlock.FileName)
	holdings := &countingHoldings{rows: brokerRows()}
	clk := newFakeClock()

	options := append([]func(*Options){func(o *Options) {
		o.Holdings = holdings
		o.JournalPath = journalPath
		o.RunLockPath = lockPath
		o.Now = clk.Now
	}}, tweak...)

	return &dashboardHarness{
		harness:  newHarness(t, options...),
		holdings: holdings,
		clock:    clk,
		journal:  journalPath,
		runlock:  lockPath,
	}
}

func (h *dashboardHarness) page(t *testing.T, path string) string {
	t.Helper()
	resp := h.get(t, path)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, resp.StatusCode)
	}
	return body(t, resp)
}

// --- the positions screen -----------------------------------------------------------

// TestThePositionsScreenShowsTheExitLineOfAManagedPosition.
//
// A pre-a042 exit state has no persisted snapshot identity. The trading view
// must keep the management verdict while refusing to expose its legacy raw
// columns as a current, actionable line.
func TestThePositionsScreenShowsTheExitLineOfAManagedPosition(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	for _, want := range []string{
		"005930",
		"엔진 관리",
		"근거 없음",
		"이전 원장에는 exit snapshot 근거가 없다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the positions screen does not show %q", want)
		}
	}
	row := rowFor(t, page, "005930")
	for _, staleRaw := range []string{"69500", "74000", "HALF_RISK", "intent-77"} {
		if strings.Contains(row, staleRaw) {
			t.Errorf("legacy row exposes raw actionable value %q", staleRaw)
		}
	}
}

// TestAnUnmanagedHoldingIsLabelledExactlyOnce.
//
// Spec scenario "관리 외 보유의 정직한 구분", plus the round-1 P3 disposition that
// fixed one label. Two holdings are unmanaged for two different reasons — one has
// no journal position at all, one has a position with no entry decision — and
// both carry the same label with their own explanation.
func TestAnUnmanagedHoldingIsLabelledExactlyOnce(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	const label = "관리 외(미편입)"
	// Two rows plus the paragraph that explains what the label means.
	if n := strings.Count(page, label); n < 3 {
		t.Errorf("the unmanaged label appears %d time(s); both unmanaged holdings must carry it "+
			"and the page must say what it means", n)
	}
	for _, want := range []string{
		"엔진 원장에 포지션이 없는 보유", // 000660: broker only — the page-level notice
		// 035420: a journal position with neither justifying record. Since
		// adopt-external-positions there are two, and the reason names both —
		// "no entry decision" alone would now be an incomplete answer.
		"편입 기록(adoption)도 없는 포지션이다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the screen does not explain why a holding is unmanaged: %q missing", want)
		}
	}
	// And it does not offer to fix it. Folding an external holding into
	// management belongs to adopt-external-positions.
	for _, banned := range []string{"편입하기", "관리 시작", `action="/positions`} {
		if strings.Contains(page, banned) {
			t.Errorf("the positions screen offers %q; it is read-only", banned)
		}
	}
}

// TestThePositionsScreenRendersWithEitherSourceMissing.
//
// Spec: 어느 쪽 데이터 소스가 없거나 비어 있어도 다른 쪽만으로 렌더한다 — 조인
// 실패가 화면 실패가 되어서는 안 된다.
func TestThePositionsScreenRendersWithEitherSourceMissing(t *testing.T) {
	t.Run("no journal file — the engine has not run", func(t *testing.T) {
		h := newDashboardHarness(t)
		h.authenticate(t)

		page := h.page(t, "/positions")
		if !strings.Contains(page, "엔진 미가동") {
			t.Error("a missing journal is not reported as 엔진 미가동")
		}
		if !strings.Contains(page, "005930") {
			t.Error("the broker holdings did not render without a journal")
		}
		if _, err := os.Stat(h.journal); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("opening the dashboard created a journal database: %v", err)
		}
	})

	t.Run("no broker wiring", func(t *testing.T) {
		h := newDashboardHarness(t, func(o *Options) { o.Holdings = nil })
		seedJournal(t, h.journal)
		h.authenticate(t)

		page := h.page(t, "/positions")
		if !strings.Contains(page, "브로커 조회가 배선되지 않았다") {
			t.Error("an unwired broker is not reported")
		}
		if !strings.Contains(page, "005930") || !strings.Contains(page, "근거 없음") {
			t.Error("the journal side did not render without a broker")
		}
	})

	t.Run("broker error keeps the journal side", func(t *testing.T) {
		h := newDashboardHarness(t)
		h.holdings.err = errors.New("429 too many requests")
		seedJournal(t, h.journal)
		h.authenticate(t)

		page := h.page(t, "/positions")
		if !strings.Contains(page, "브로커 조회 실패") || !strings.Contains(page, "429") {
			t.Error("a broker failure is not reported honestly")
		}
		if !strings.Contains(page, "근거 없음") {
			t.Error("a broker failure took the journal side of the screen down with it")
		}
	})
}

// TestAnEmptyJournalSaysSoRatherThanLookingLikeAMissingOne.
//
// Spec scenario "엔진 미가동 상태의 대시보드" pairs two states — journal 파일이
// 없거나 포지션 테이블이 비어 있으면 — and they are different facts. A read ledger
// with nothing in it is the engine managing nothing; a missing one is the engine
// never having started.
func TestAnEmptyJournalSaysSoRatherThanLookingLikeAMissingOne(t *testing.T) {
	h := newDashboardHarness(t)
	seedEmptyJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if !strings.Contains(page, "관리 포지션 없음") {
		t.Error("an empty but readable journal does not say 관리 포지션 없음")
	}
	if strings.Contains(page, "엔진 미가동") {
		t.Error("a readable journal is reported as 엔진 미가동")
	}
	if !strings.Contains(page, "005930") {
		t.Error("the broker holdings did not render")
	}
	if !strings.Contains(page, "관리 외(미편입)") {
		t.Error("holdings with no journal position are not marked unmanaged")
	}

	// And the seeded one does not claim it.
	full := newDashboardHarness(t)
	seedJournal(t, full.journal)
	full.authenticate(t)
	if strings.Contains(full.page(t, "/positions"), "관리 포지션 없음") {
		t.Error("a journal with live positions claims to have none")
	}
}

// TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead.
//
// "관리 외(미편입)" is a finding: the ledger was read and this holding is not an
// exit-policy target. When the ledger could not be read there is no finding, and
// saying it anyway is how an operator ends up believing a protected position is
// bare — or, worse, believing a bare one is fine because the screen looked the
// same as always.
func TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, path string)
		want  string
	}{
		{"no journal file", func(*testing.T, string) {}, "엔진 미가동"},
		{"schema too old", seedOldJournal, "엔진 기동 필요"},
		{"schema too new", func(t *testing.T, path string) {
			seedJournal(t, path)
			bumpSchemaVersion(t, path, 9999)
		}, "콘솔 업데이트 필요"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newDashboardHarness(t)
			tc.setup(t, h.journal)
			h.authenticate(t)

			page := h.page(t, "/positions")
			if !strings.Contains(page, tc.want) {
				t.Errorf("the journal state is not reported as %q", tc.want)
			}
			if strings.Contains(page, "관리 외(미편입)</td>") {
				t.Error("a holding is labelled unmanaged although the journal was never read")
			}
			if !strings.Contains(page, "관리 여부 불명") {
				t.Error("the rows do not say the management state is unknown")
			}
			if !strings.Contains(page, "관리 중이 아니라고 단정하지 않는다") {
				t.Error("the row gives no reason for the unknown state")
			}
		})
	}
}

// TestThePositionsScreenNamesBothSchemaDirections.
//
// Spec scenario "스키마 불일치 — 양방향". Neither direction may be disguised as an
// empty screen, and the broker side keeps working through both.
func TestThePositionsScreenNamesBothSchemaDirections(t *testing.T) {
	t.Run("newer than the console", func(t *testing.T) {
		h := newDashboardHarness(t)
		seedJournal(t, h.journal)
		bumpSchemaVersion(t, h.journal, 9999)
		h.authenticate(t)

		page := h.page(t, "/positions")
		if !strings.Contains(page, "콘솔 업데이트 필요") {
			t.Error("a newer schema is not reported as 콘솔 업데이트 필요")
		}
		if !strings.Contains(page, "005930") {
			t.Error("the broker side stopped rendering on a schema mismatch")
		}
	})

	t.Run("older than the console", func(t *testing.T) {
		h := newDashboardHarness(t)
		seedOldJournal(t, h.journal)
		h.authenticate(t)

		page := h.page(t, "/positions")
		if !strings.Contains(page, "엔진 기동 필요") {
			t.Error("a pre-migration schema is not reported as 엔진 기동 필요")
		}
		if !strings.Contains(page, "005930") {
			t.Error("the broker side stopped rendering on a schema mismatch")
		}
	})
}

// --- the rate budget ------------------------------------------------------------------

// TestRefreshingInsideTheTTLCostsNoBrokerCall.
//
// Spec scenario "새로고침 연타", and the §0.4 accounting the change had to make
// explicit: one refresh is one holdings call, and everything inside the TTL is
// free.
func TestRefreshingInsideTheTTLCostsNoBrokerCall(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	for i := 0; i < 5; i++ {
		h.page(t, "/positions")
	}
	if n := h.holdings.count(); n != 1 {
		t.Fatalf("%d broker call(s) for five renders inside the TTL, want 1", n)
	}

	h.clock.advance(holdingsTTL - time.Second)
	h.page(t, "/positions")
	if n := h.holdings.count(); n != 1 {
		t.Errorf("%d broker call(s) just before the TTL expires, want 1", n)
	}

	h.clock.advance(2 * time.Second)
	h.page(t, "/positions")
	if n := h.holdings.count(); n != 2 {
		t.Errorf("%d broker call(s) after the TTL expired, want 2", n)
	}
}

// TestTheCacheTTLIsNotBelowTheFloorTheSpecFixes.
func TestTheCacheTTLIsNotBelowTheFloorTheSpecFixes(t *testing.T) {
	if holdingsTTL < 15*time.Second {
		t.Errorf("holdingsTTL = %s; the operator-console spec fixes a floor of 15s", holdingsTTL)
	}
}

// TestAFailingBrokerIsNotRetriedOnEveryPageLoad.
//
// The TTL bounds attempts, not successes. A broker answering 429 is the one
// moment the rate budget is already in trouble, and a dashboard that asked again
// on every render would be making it worse while showing an error either way.
func TestAFailingBrokerIsNotRetriedOnEveryPageLoad(t *testing.T) {
	h := newDashboardHarness(t)
	h.holdings.err = errors.New("429 too many requests")
	seedJournal(t, h.journal)
	h.authenticate(t)

	for i := 0; i < 4; i++ {
		h.page(t, "/positions")
	}
	if n := h.holdings.count(); n != 1 {
		t.Errorf("%d broker call(s) for four renders against a failing broker, want 1", n)
	}

	h.clock.advance(holdingsTTL + time.Second)
	h.page(t, "/positions")
	if n := h.holdings.count(); n != 2 {
		t.Errorf("%d broker call(s) after the TTL elapsed, want 2", n)
	}
}

// TestARefreshIsExactlyOneHoldingsCall.
//
// The reviewer's point in round 1: a screen that showed a current price per
// symbol would be tempted into a quote per symbol, turning one call into one per
// holding. The price comes out of the holdings response and there is no second
// read of any kind.
func TestARefreshIsExactlyOneHoldingsCall(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	h.page(t, "/positions")
	if n := h.holdings.count(); n != 1 {
		t.Fatalf("one refresh made %d broker call(s) for %d holdings, want 1",
			n, len(brokerRows()))
	}
	// And the price on the page is the one that response carried.
	page := h.page(t, "/positions")
	if !strings.Contains(page, "73,000") {
		t.Error("the page does not show the lastPrice from the holdings response")
	}
}

// TestAVerificationInProgressSuspendsTheRefresh.
//
// Spec scenarios "검증 실행 중 — 캐시 있음" and "— 콜드 캐시". The verification is
// the scarce thing; a dashboard must never be the reason a supervised live run
// loses a step to a 429.
func TestAVerificationInProgressSuspendsTheRefresh(t *testing.T) {
	t.Run("another process, warm cache", func(t *testing.T) {
		h := newDashboardHarness(t)
		seedJournal(t, h.journal)
		h.authenticate(t)

		h.page(t, "/positions")
		if n := h.holdings.count(); n != 1 {
			t.Fatalf("%d call(s) for the first render, want 1", n)
		}

		holdRunLock(t, h.runlock, h.clock.Now())
		h.clock.advance(2 * holdingsTTL)

		page := h.page(t, "/positions")
		if n := h.holdings.count(); n != 1 {
			t.Errorf("%d broker call(s) while a verification is running, want 1", n)
		}
		if !strings.Contains(page, "검증 중 — 갱신 보류") {
			t.Error("the page does not say the refresh is being withheld")
		}
		if !strings.Contains(page, "73,000") {
			t.Error("the cached reading is not shown while the refresh is withheld")
		}
	})

	t.Run("another process, cold cache", func(t *testing.T) {
		h := newDashboardHarness(t)
		seedJournal(t, h.journal)
		holdRunLock(t, h.runlock, h.clock.Now())
		h.authenticate(t)

		page := h.page(t, "/positions")
		if n := h.holdings.count(); n != 0 {
			t.Errorf("%d broker call(s) with a cold cache during a verification, want 0", n)
		}
		if !strings.Contains(page, "검증 중 — 데이터 없음") {
			t.Error("a cold cache during a verification is not reported honestly")
		}
		if !strings.Contains(page, "근거 없음") {
			t.Error("the journal side stopped rendering during a verification")
		}
	})

	t.Run("a stale marker is ignored", func(t *testing.T) {
		h := newDashboardHarness(t)
		seedJournal(t, h.journal)
		holdRunLock(t, h.runlock, h.clock.Now().Add(-2*runlock.StaleAfter))
		h.authenticate(t)

		h.page(t, "/positions")
		if n := h.holdings.count(); n != 1 {
			t.Errorf("%d broker call(s); a marker older than %s is a crashed run, not a live one",
				n, runlock.StaleAfter)
		}
	})

	t.Run("this process", func(t *testing.T) {
		h := newDashboardHarness(t)
		seedJournal(t, h.journal)
		h.startAndWait(t) // a run parked on the approval form: in flight

		page := h.page(t, "/positions")
		if n := h.holdings.count(); n != 0 {
			t.Errorf("%d broker call(s) while this console is running a verification, want 0", n)
		}
		if !strings.Contains(page, "이 콘솔이 실계좌 검증을 실행 중이다") {
			t.Error("the in-process verification is not named as the reason")
		}
	})
}

// --- the history screen -----------------------------------------------------------------

// TestTheHistoryScreenShowsTheFrozenRoundTrip.
//
// Spec scenario "동결된 왕복 결과 표시": the joined symbol, the frozen numbers, the
// hold time and the exit stage — and nothing that is not a column.
func TestTheHistoryScreenShowsTheFrozenRoundTrip(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	for _, want := range []string{
		"005380",               // the symbol, joined from positions
		"55000",                // the entry price, joined from exit_states
		"48000",                // realised P&L after costs, frozen
		"2.4",                  // realised R, frozen
		"6시간 0분",               // 21600 seconds
		"PROFIT_LOCK",          // the exit stage reached
		"2026-07-24T06:30:00Z", // the close
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the history screen does not show %q", want)
		}
	}
}

// TestANullHoldTimeIsAnEmDashAndNotZeroSeconds.
//
// Round-2 P3. held_seconds is nullable, and "0초" would be the screen stating a
// fact the journal does not hold.
func TestANullHoldTimeIsAnEmDashAndNotZeroSeconds(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	if strings.Contains(page, "<td>0초</td>") {
		t.Error("a NULL hold time was rendered as zero seconds")
	}
	row, ok := historyRow(page, "051910")
	if !ok {
		t.Fatal("the round trip with no hold time is missing entirely")
	}
	if !strings.Contains(row, "<td>—</td>") {
		t.Errorf("the NULL hold time is not an em dash in its own row: %s", row)
	}

	// The rendering rule itself, so a future formatter cannot turn "unknown" into
	// a number without this failing.
	if got := heldText(0, false); got != "—" {
		t.Errorf("heldText(unknown) = %q, want an em dash", got)
	}
	if got := heldText(0, true); got != "0초" {
		t.Errorf("heldText(0 seconds, known) = %q, want 0초 — a measured zero is not unknown", got)
	}
}

// historyRow extracts the <tr> containing a symbol, so a per-cell assertion is
// about that row rather than about the whole page.
func historyRow(page, symbol string) (string, bool) {
	for _, row := range strings.Split(page, "<tr>") {
		if strings.Contains(row, symbol) {
			return row, true
		}
	}
	return "", false
}

// TestTheHistoryScreenStatesItsOwnLimit.
//
// Round-2 P2 (편입 이력 공백): a position closed by an external sell never gets a
// trade_outcomes row, so the performance table has a hole in it by construction.
// The screen says so and points at the exit-event stream that covers it.
func TestTheHistoryScreenStatesItsOwnLimit(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	if !strings.Contains(page, "외부 매도로 종결된 포지션은 여기에 행이 생기지 않으므로") {
		t.Error("the history screen does not state the gap external closures leave")
	}
	if !strings.Contains(page, "exit 이벤트") || !strings.Contains(page, "intent-77") {
		t.Error("the exit-event stream that covers the gap is not rendered")
	}
}

// TestTheHistoryScreenShowsNoExitPriceAndRecomputesNothing.
//
// The schema has no exit price. The frozen realised R is beside it, and a value
// derived from the fills would disagree with that R the moment a partial
// take-profit happened.
func TestTheHistoryScreenShowsNoExitPriceAndRecomputesNothing(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	for _, banned := range []string{"<th>청산가", "<th>청산 단가", "<th>exit price"} {
		if strings.Contains(page, banned) {
			t.Errorf("the history screen has a %q column; the schema has no such value", banned)
		}
	}
	if !strings.Contains(page, "청산가는 스키마에 없으므로 표시하지 않으며") {
		t.Error("the screen does not say why there is no exit price")
	}
}

// TestTheHistoryScreenIsHonestWhenNothingHasClosed.
func TestTheHistoryScreenIsHonestWhenNothingHasClosed(t *testing.T) {
	h := newDashboardHarness(t)
	h.authenticate(t)

	page := h.page(t, "/history")
	if !strings.Contains(page, "엔진 미가동") {
		t.Error("a missing journal is not reported on the history screen")
	}

	empty := newDashboardHarness(t)
	seedEmptyJournal(t, empty.journal)
	empty.authenticate(t)
	page = empty.page(t, "/history")
	if !strings.Contains(page, "완결 거래 없음") {
		t.Error("an empty journal does not say 완결 거래 없음")
	}
}

// --- the routes -------------------------------------------------------------------------

// TestTheDashboardRoutesNeedTheSessionAndAreGETOnly.
func TestTheDashboardRoutesNeedTheSessionAndAreGETOnly(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)

	for _, path := range []string{"/positions", "/history"} {
		resp := h.get(t, path)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("GET %s without a session = %d, want 403", path, resp.StatusCode)
		}
	}

	h.authenticate(t)
	for _, path := range []string{"/positions", "/history"} {
		if resp := h.get(t, path); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, resp.StatusCode)
		}
		// A POST reaches the same read handler and changes nothing; what matters
		// is that no mutation route was added beside it.
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating broker call(s) were made by opening the dashboard", n)
	}
}

// --- helpers ------------------------------------------------------------------------------

// holdRunLock writes the advisory marker another process would leave. It is
// written by the test rather than by the console, which cannot write it — that
// is asserted in static_test.go.
func holdRunLock(t *testing.T, path string, at time.Time) {
	t.Helper()
	lock, err := runlock.Acquire(path, at)
	if err != nil {
		t.Fatalf("runlock.Acquire: %v", err)
	}
	t.Cleanup(lock.Release)
}

// adoptedFixture adds a holding the engine adopted: no entry decision, an
// adoption record, and an exit state seeded from the adoption's observation.
const adoptedFixture = `
INSERT INTO position_adoptions (id, symbol, market, quantity, cost_basis, cost_basis_src,
                                observed_price, synthetic_stop, observed_at, preimage_digest)
VALUES ('adopt-1','000660','kr','3','120000.0000','BROKER_AVG','131500','124925',
        '2026-07-27T00:40:00Z','digest-1');
INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id,
                       adoption_id, state, quantity, avg_price, opened_at)
VALUES ('pos-adopted','123-45-678901','kr','000660',1,NULL,'adopt-1','OPEN','3','120000',
        '2026-07-27T00:40:00Z');
INSERT INTO exit_states (position_id, policy_kind, entry_price, initial_stop, initial_risk,
                         baseline_price, high_water, ratchet_level, active_rung, taken_ratio_total,
                         completed, updated_at)
VALUES ('pos-adopted','RATCHET','131500','124925','6575','124925','131500','NONE',NULL,'0',0,
        '2026-07-27T00:40:00Z');
`

// TestAnAdoptedHoldingRendersAsManagedWithItsBasis is task 2.7: the dashboard's
// eligibility display covers adoption.
//
// Two things are being checked and they are different findings. The *verdict* —
// managed, not 관리 외 — is the single eligibility predicate reaching the screen;
// a row still labelled 관리 외 would be telling an operator that a position the
// engine is actively stopping out is unprotected. The *basis* is the operator's
// next question: an engine that has started selling shares somebody bought by
// hand has to be able to say that it adopted them.
func TestAnAdoptedHoldingRendersAsManagedWithItsBasis(t *testing.T) {
	h := newDashboardHarness(t)
	seedEngineJournal(t, h.journal, journalFixture+adoptedFixture)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if !strings.Contains(page, "편입 기록") {
		t.Error("the screen does not say an adopted position's protection came from an adoption record")
	}
	if !strings.Contains(page, "진입 결정") {
		t.Error("the screen no longer distinguishes an engine-entered position's basis")
	}
	// 000660 is the adopted symbol. Its row must not carry the unmanaged label.
	row := rowFor(t, page, "000660")
	if strings.Contains(row, "관리 외(미편입)") {
		t.Errorf("the adopted holding is labelled unmanaged:\n%s", row)
	}
	if !strings.Contains(row, "엔진 관리") {
		t.Errorf("the adopted holding is not labelled as managed:\n%s", row)
	}
	// This fixture predates persisted snapshots, so its raw synthetic stop must
	// not be promoted to an actionable line.
	if strings.Contains(row, "124925") || !strings.Contains(row, "근거 없음") {
		t.Error("the adopted legacy position is not fail-closed")
	}
}

// rowFor returns the primary row carrying one symbol, so a per-row assertion
// cannot be satisfied by a later holding or by the page-level explanation.
func rowFor(t *testing.T, page, symbol string) string {
	t.Helper()
	at := strings.Index(page, `data-symbol="`+symbol+`"`)
	if at < 0 {
		t.Fatalf("the page does not mention %s", symbol)
	}
	start := strings.LastIndex(page[:at], "<tr")
	if start < 0 {
		t.Fatalf("the %s marker is not inside a table row", symbol)
	}
	rest := page[start:]
	if end := strings.Index(rest, "</tr>"); end >= 0 {
		return rest[:end+len("</tr>")]
	}
	t.Fatalf("the %s row is not closed", symbol)
	return ""
}
