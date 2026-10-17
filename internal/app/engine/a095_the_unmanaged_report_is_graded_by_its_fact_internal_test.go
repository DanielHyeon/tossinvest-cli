package engine

// a095 tasks 2.1~2.17 — 무관리 보유 보고의 등급은 사실이 정함(engine-safety 델타 10판).
//
// 전부 **생산 배선**으로 잼: 실 원장 · 실 obs.Notifier(outbox · 진입 게이트 · 승격) · 실 배달 실행자(alertDeliverer).
// 가짜는 브로커 표면(보유 · 주문 · 예수금 · 시세)과 전송 수단(a098RecordingPublisher)뿐임. 발신 가짜만으로 재면
// 「critical 로 매겼다」와 「outbox 에 행이 섰다」를 구별하지 못함(r3 N1 codex 제안).
//
// 편입 시도 실패의 두 범주(design D1 「시도 실패」의 경계):
//   ① 편입 전 거절 — adoptOne B1(SyntheticStop 실패). 픽스처: DefaultStopPct 0(ValidateStopPct 가 거절).
//   ② 영속 실패   — adoptOne B2(AdoptPosition 트랜잭션 실패). 픽스처: position_adoptions INSERT 를 막는 시험 전용 트리거.

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const a095Account = "acct-a095"

var a095Now = time.Date(2026, 9, 30, 1, 0, 0, 0, time.UTC)

// --- 브로커 표면 가짜 ------------------------------------------------------------

type a095Orders struct{}

func (a095Orders) OrdersPageRaw(context.Context, execgw.OrderQuery, string) (execgw.OrderPage, error) {
	return execgw.OrderPage{}, nil
}

// a095Holdings 는 원시 경로로 보유를 답함. onCall 은 수집 도중 시계를 움직이는 픽스처(묵은 스냅샷)용.
type a095Holdings struct {
	mu     sync.Mutex
	items  []reconcile.RawHolding
	onCall func()
}

func (h *a095Holdings) Positions(ctx context.Context) ([]domain.Position, error) {
	raw, _ := h.PositionsRaw(ctx)
	out := make([]domain.Position, 0, len(raw))
	for _, r := range raw {
		out = append(out, domain.Position{Symbol: r.Symbol, MarketType: r.Market})
	}
	return out, nil
}

func (h *a095Holdings) PositionsRaw(context.Context) ([]reconcile.RawHolding, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.onCall != nil {
		h.onCall()
	}
	return append([]reconcile.RawHolding(nil), h.items...), nil
}

func (h *a095Holdings) set(symbol, quantity string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := range h.items {
		if h.items[i].Symbol == symbol {
			h.items[i].Quantity = quantity
			return
		}
	}
	h.items = append(h.items, reconcile.RawHolding{Symbol: symbol, Market: "kr", Quantity: quantity,
		AveragePrice: "55000"})
}

type a095Balance struct{}

func (a095Balance) BuyingPower(context.Context, string) (float64, error) { return 1_000_000, nil }

type a095Prices struct {
	mu   sync.Mutex
	last map[string]float64
	err  error
}

func (p *a095Prices) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return nil, p.err
	}
	var out []domain.Quote
	for _, s := range symbols {
		if last, ok := p.last[s]; ok {
			out = append(out, domain.Quote{Symbol: s, Last: last, Currency: "KRW"})
		}
	}
	return out, nil
}

// a095Capture 는 발신된 사건을 기록하고 실 알림기로 넘김 — key · 종류를 재는 창이며 판정은 알림기가 함.
type a095Capture struct {
	inner  ExitAlerter
	events []obs.Event
}

func (c *a095Capture) Notify(ctx context.Context, e obs.Event) error {
	c.events = append(c.events, e)
	if c.inner == nil {
		return nil
	}
	return c.inner.Notify(ctx, e)
}

func (c *a095Capture) keysContaining(part string) []string {
	var out []string
	for _, e := range c.events {
		if strings.Contains(e.Key, part) {
			out = append(out, e.Key)
		}
	}
	return out
}

// --- 픽스처 -----------------------------------------------------------------------

type a095Fixture struct {
	t        *testing.T
	path     string
	j        *journal.Journal
	clk      *clock.Fake
	gate     *execgw.EntryGate
	pub      *a098RecordingPublisher
	n        *obs.Notifier
	capture  *a095Capture
	holdings *a095Holdings
	prices   *a095Prices
	opts     ReconcileDriverOptions
	d        *ReconcileDriver
	del      *alertDeliverer
	side     *sql.DB
}

type a095Option func(*a095Fixture)

func a095NotificationsOff(f *a095Fixture) { f.opts.NotificationsEnabled = false }
func a095NoTopic(f *a095Fixture)          { f.n.Publisher = nil }
func a095StopPctRefused(f *a095Fixture)   { f.opts.Adoption.DefaultStopPct = 0 } // 범주 ①
func a095Adoption(a config.Adoption) a095Option {
	return func(f *a095Fixture) { f.opts.Adoption = a }
}

func newA095(t *testing.T, options ...a095Option) *a095Fixture {
	t.Helper()
	clk := clock.NewFake(a095Now)
	path := filepath.Join(t.TempDir(), journal.DBFileName)
	j, err := journal.Open(context.Background(), journal.Options{
		Path: path, Clock: clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Exit: journal.ApplyExitFill}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	f := &a095Fixture{t: t, path: path, j: j, clk: clk,
		gate: execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
		pub:  &a098RecordingPublisher{}, holdings: &a095Holdings{},
		prices: &a095Prices{last: map[string]float64{}},
	}
	// 전송 시도 1회 — 재시도 대기(가짜 시계 Sleep)에 사이클이 멈추지 않게 함. 재알림 창은 10분.
	f.n = &obs.Notifier{Journal: j, Gate: f.gate, Publisher: f.pub, AccountRef: a095Account, Clock: clk,
		Attempts: 1, RemindAfter: 10 * time.Minute}
	f.capture = &a095Capture{inner: f.n}
	tracker := &reconcile.Tracker{Clock: clk, Gate: f.gate, Journal: j, AccountRef: a095Account}
	f.opts = ReconcileDriverOptions{
		Journal: j,
		Collector: &reconcile.Collector{Orders: a095Orders{}, Positions: f.holdings, Balance: a095Balance{},
			Currencies: []string{"KRW"}, AccountRef: a095Account, Clock: clk},
		Tracker:  tracker,
		Ingest:   &reconcile.Ingestor{Journal: j, AccountRef: a095Account, DefaultMarket: "kr"},
		Converge: &reconcile.Converger{Journal: j, Credit: tracker, AccountRef: a095Account},
		Prices:   f.prices, Alerts: f.capture, Names: &InstrumentNames{},
		AccountRef: a095Account, Clock: clk, DefaultMarket: "kr",
		Adoption:             config.Adoption{Enabled: true, DefaultStopPct: 0.05},
		NotificationsEnabled: true,
	}
	for _, o := range options {
		o(f)
	}
	f.rebuild()
	return f
}

// rebuild 는 같은 원장 위에 드라이버와 배달 실행자를 다시 세움 — 설정 전환(재조립)을 흉내 냄.
func (f *a095Fixture) rebuild() {
	f.t.Helper()
	d, err := NewReconcileDriver(f.opts)
	if err != nil {
		f.t.Fatalf("NewReconcileDriver: %v", err)
	}
	f.d = d
	f.del = &alertDeliverer{Journal: f.j, Publisher: f.n.Publisher, Clock: f.clk,
		Interval: alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a095-deliverer",
		Gate: f.gate, AccountRef: a095Account}
}

// restart 는 새 진입 게이트로 다시 기동하고 기동 복원(restoreAlertEntryLatch)을 태움.
func (f *a095Fixture) restart() {
	f.t.Helper()
	f.gate = execgw.NewEntryGate(f.clk, map[execgw.RequiredQuery]time.Duration{})
	if err := restoreAlertEntryLatch(context.Background(), f.j, f.gate); err != nil {
		f.t.Fatalf("restoreAlertEntryLatch: %v", err)
	}
	f.n = &obs.Notifier{Journal: f.j, Gate: f.gate, Publisher: f.n.Publisher, AccountRef: a095Account, Clock: f.clk,
		Attempts: 1, RemindAfter: 10 * time.Minute}
	f.capture = &a095Capture{inner: f.n}
	f.opts.Alerts = f.capture
	f.opts.Tracker = &reconcile.Tracker{Clock: f.clk, Gate: f.gate, Journal: f.j, AccountRef: a095Account}
	f.opts.Converge = &reconcile.Converger{Journal: f.j, Credit: f.opts.Tracker, AccountRef: a095Account}
	f.rebuild()
}

func (f *a095Fixture) holds(symbol, quantity string, last float64) {
	f.holdings.set(symbol, quantity)
	f.prices.mu.Lock()
	f.prices.last[symbol] = last
	f.prices.mu.Unlock()
}

func (f *a095Fixture) cycle() ReconcileCycle {
	f.t.Helper()
	done := make(chan ReconcileCycle, 1)
	go func() { done <- f.d.RunOnce(context.Background()) }()
	if !f.clk.WaitForSleepers(1, 5*time.Second) {
		f.t.Fatal("the driver never entered its stabilisation wait")
	}
	f.clk.Advance(reconcile.DefaultStabilisationInterval)
	return <-done
}

func (f *a095Fixture) position(symbol string) journal.Position {
	f.t.Helper()
	p, err := f.j.CurrentPosition(context.Background(), a095Account, "kr", symbol)
	if err != nil {
		f.t.Fatalf("CurrentPosition(%s): %v", symbol, err)
	}
	return p
}

func (f *a095Fixture) sideDB() *sql.DB {
	f.t.Helper()
	if f.side == nil {
		db, err := sql.Open("sqlite", "file:"+f.path+"?_pragma=busy_timeout(10000)")
		if err != nil {
			f.t.Fatalf("side handle: %v", err)
		}
		f.t.Cleanup(func() { _ = db.Close() })
		f.side = db
	}
	return f.side
}

func (f *a095Fixture) exec(q string) {
	f.t.Helper()
	if _, err := f.sideDB().Exec(q); err != nil {
		f.t.Fatalf("side exec %q: %v", q, err)
	}
}

// refuseAdoptions 는 범주 ② — 편입 트랜잭션을 원장 오류로 실패시킴. message 가 오류 문구(진단 원인)임.
func (f *a095Fixture) refuseAdoptions(message string) {
	f.exec(`DROP TRIGGER IF EXISTS a095_refuse_adoption`)
	f.exec(fmt.Sprintf(`CREATE TRIGGER a095_refuse_adoption BEFORE INSERT ON position_adoptions
		BEGIN SELECT RAISE(ABORT, '%s'); END`, message))
}

func (f *a095Fixture) allowAdoptions() { f.exec(`DROP TRIGGER IF EXISTS a095_refuse_adoption`) }

type a095Row struct {
	key, typ, severity, state, title, body string
	attempts                               int
}

func (f *a095Fixture) rows() []a095Row {
	f.t.Helper()
	rs, err := f.sideDB().Query(`SELECT event_key, event_type, severity, state, attempts, title, body
		FROM alert_outbox ORDER BY id`)
	if err != nil {
		f.t.Fatalf("reading alert_outbox: %v", err)
	}
	defer rs.Close()
	var out []a095Row
	for rs.Next() {
		var r a095Row
		if err := rs.Scan(&r.key, &r.typ, &r.severity, &r.state, &r.attempts, &r.title, &r.body); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

// criticalSends 는 편입 실패 critical 의 실제 전송 수임 — normal 보고도 같은 전송기로 best-effort 발행되므로 종류로 가름.
func (f *a095Fixture) criticalSends() int {
	f.pub.mu.Lock()
	defer f.pub.mu.Unlock()
	n := 0
	for _, sent := range f.pub.sent {
		if sent.Type == obs.EventExitPositionAdoptionFailed {
			n++
		}
	}
	return n
}

func (f *a095Fixture) latched() bool {
	_, ok := f.gate.Blocks()[execgw.ReasonAlertUndelivered]
	return ok
}

func (f *a095Fixture) entryBlocked() bool {
	f.t.Helper()
	snap, err := f.j.CurrentOperatingMode(context.Background(), a095Account)
	if err != nil {
		return false // 행이 없으면 승격도 없음
	}
	return snap.Mode == journal.ModeEntryBlocked
}

// assertNoCriticalConsequence 는 「normal 로 남았다」의 결과 셋 — outbox 행 · 게이트 래치 · 운영 모드 승격 — 을 함께 봄.
func (f *a095Fixture) assertNoCriticalConsequence(t *testing.T) {
	t.Helper()
	if rows := f.rows(); len(rows) != 0 {
		t.Errorf("outbox rows = %+v, want none: this report must not be critical", rows)
	}
	if f.latched() {
		t.Error("the entry gate latched on a report that must not be critical")
	}
	if f.entryBlocked() {
		t.Error("the operating mode escalated to ENTRY_BLOCKED on a report that must not be critical")
	}
}

func (f *a095Fixture) onlyRow(t *testing.T) a095Row {
	t.Helper()
	rows := f.rows()
	if len(rows) != 1 {
		t.Fatalf("outbox rows = %+v, want exactly one", rows)
	}
	return rows[0]
}

func a095CriticalFailureKey(posID string) string {
	return string(obs.EventExitPositionAdoptionFailed) + "|reconcile|enabled_failed|" + posID
}

// 2.6 (Q2(b) 정정 — Manager 판정 (가) 2026-09-30): 편입 꺼짐 + include 지정 종목의 시도 실패도 critical. include 지정은 운영자가
// 그 종목의 보호를 고른 것이고, 정본 exit-policy 「종목별 편입」이 include 경유 편입을 알림 규칙 전부에서 enabled 경유와 같게 둠.
func TestA095ADesignatedSymbolsFailureIsCriticalToo(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options []a095Option
		arrange func(*a095Fixture)
	}{
		{"category ① refused before adopting",
			[]a095Option{a095Adoption(config.Adoption{IncludeSymbols: []string{"005930"}})}, nil}, // pct 0
		{"category ② the adoption transaction failed",
			[]a095Option{a095Adoption(config.Adoption{IncludeSymbols: []string{"005930"}, DefaultStopPct: 0.05})},
			func(f *a095Fixture) { f.refuseAdoptions("a095 refusal") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA095(t, tc.options...)
			f.holds("005930", "10", 70000)
			if tc.arrange != nil {
				tc.arrange(f)
			}
			f.cycle()
			row := f.onlyRow(t)
			want := string(obs.EventExitPositionAdoptionFailed) + "|reconcile|include_failed|" + f.position("005930").ID
			if row.severity != string(obs.SeverityCritical) || row.key != want {
				t.Errorf("row = %+v, want a critical row keyed %q", row, want)
			}
		})
	}
	t.Run("notifications off keeps it normal", func(t *testing.T) {
		f := newA095(t, a095NotificationsOff, a095Adoption(config.Adoption{IncludeSymbols: []string{"005930"}}))
		f.holds("005930", "10", 70000)
		f.cycle()
		f.assertNoCriticalConsequence(t)
		if len(f.capture.keysContaining("|reconcile|include_failed|")) != 1 {
			t.Errorf("events = %+v, want one normal include_failed report", f.capture.events)
		}
	})
}

// --- 2.1 · 2.7 — exit 관측 자리는 normal, 두 자리의 key 는 다름 --------------------------

func TestA095TheExitObserverReportStaysNormalAndKeyedApart(t *testing.T) {
	// 같은 **종류**(exit.position_unmanaged)끼리 비교함 — critical 종류와 비교하면 종류만으로 늘 달라 key 분리를 재지 못함
	// (독립 리뷰 P3-3). 알림 꺼짐이라 대사 쪽 보고도 normal 종류로 남음.
	f := newA095(t, a095NotificationsOff, a095StopPctRefused)
	f.holds("005930", "10", 70000)
	f.cycle()
	p := f.position("005930")

	exitCapture := &a095Capture{inner: f.n}
	o := &ExitObserver{opts: ExitObserverOptions{Alerts: exitCapture, Names: &InstrumentNames{}},
		unmanaged: map[string]bool{}}
	before := len(f.rows())
	o.alertUnmanaged(context.Background(), p)

	if len(exitCapture.events) != 1 {
		t.Fatalf("exit observer events = %d, want 1", len(exitCapture.events))
	}
	exitEvent := exitCapture.events[0]
	if obs.SeverityOf(exitEvent.Type) != obs.SeverityNormal {
		t.Errorf("the exit observer's unmanaged report is %s; it stands before every stop judgement of "+
			"the cycle and must stay normal (결정 (1))", obs.SeverityOf(exitEvent.Type))
	}
	if got := len(f.rows()); got != before {
		t.Errorf("the exit observer's report wrote %d outbox row(s); normal reports have no durable row",
			got-before)
	}
	// 2.7 — 같은 포지션 · 같은 종류의 두 발신 자리는 다른 key 를 씀.
	reconcile := f.capture.keysContaining("|reconcile|")
	if len(reconcile) != 1 || f.capture.events[0].Type != exitEvent.Type {
		t.Fatalf("arrangement: want one reconcile report of the same kind, got %+v", f.capture.events)
	}
	if reconcile[0] == exitEvent.Key {
		t.Errorf("the reconcile report and the exit observer report share the key %q", exitEvent.Key)
	}
}

// --- 2.2 — 알림 켜짐 · 편입 켜짐 · 시도 실패 → critical, 생산 배선 --------------------

func TestA095AFailedAdoptionIsCriticalAndDurable(t *testing.T) {
	for _, tc := range []struct {
		name    string
		arrange func(*a095Fixture)
	}{
		{"category ① refused before adopting", func(f *a095Fixture) {}},
		{"category ② the adoption transaction failed", func(f *a095Fixture) { f.refuseAdoptions("a095 refusal") }},
		// 편입이 켜진 엔진의 include 지정 종목은 「편입 켜짐」 칸(9판 r8 N8 — alertUnmanaged 는 Enabled 를 Included 보다 먼저 봄).
		{"category ① on a symbol that is also designated", func(f *a095Fixture) {
			f.opts.Adoption.IncludeSymbols = []string{"005930"}
			f.rebuild()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := []a095Option{}
			if strings.HasPrefix(tc.name, "category ①") {
				options = append(options, a095StopPctRefused)
			}
			f := newA095(t, options...)
			f.holds("005930", "10", 70000)
			tc.arrange(f)

			cycle := f.cycle()
			if cycle.Adopted != 0 || cycle.Unmanaged != 1 {
				t.Fatalf("cycle = %+v, want the candidate unadopted and reported", cycle)
			}
			row := f.onlyRow(t)
			p := f.position("005930")
			if row.severity != string(obs.SeverityCritical) || row.typ != string(obs.EventExitPositionAdoptionFailed) {
				t.Errorf("row = %+v, want a critical %s row", row, obs.EventExitPositionAdoptionFailed)
			}
			if row.key != a095CriticalFailureKey(p.ID) {
				t.Errorf("key = %q, want %q", row.key, a095CriticalFailureKey(p.ID))
			}
			if row.state != journal.AlertDelivered || f.criticalSends() != 1 {
				t.Errorf("state = %s, sends = %d; a working transport delivers the critical report",
					row.state, f.criticalSends())
			}
			if obs.SeverityOf(obs.EventExitPositionAdoptionFailed) != obs.SeverityCritical {
				t.Error("the adoption-failure kind is not in the grading table")
			}
		})
	}
}

// --- 2.3 · 2.4 — 운영자가 고른 상태는 normal -------------------------------------------

func TestA095OperatorChosenStatesStayNormal(t *testing.T) {
	for _, tc := range []struct {
		name     string
		adoption config.Adoption
		cell     string
	}{
		{"excluded", config.Adoption{Enabled: true, DefaultStopPct: 0.05, ExcludeSymbols: []string{"005930"}}, "excluded"},
		{"adoption off and not designated", config.Adoption{}, "off_undesignated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA095(t, a095Adoption(tc.adoption))
			f.holds("005930", "10", 70000)
			if cycle := f.cycle(); cycle.Unmanaged != 1 {
				t.Fatalf("unmanaged = %d, want 1", cycle.Unmanaged)
			}
			f.assertNoCriticalConsequence(t)
			if len(f.capture.events) != 1 || f.capture.events[0].Type != obs.EventExitPositionUnmanaged {
				t.Fatalf("events = %+v, want one normal unmanaged report", f.capture.events)
			}
			if !strings.Contains(f.capture.events[0].Key, "|reconcile|"+tc.cell+"|") {
				t.Errorf("key = %q, want the fact cell %q", f.capture.events[0].Key, tc.cell)
			}
		})
	}
}

// --- 2.5 · 2.5a — 알림 꺼짐(topic 유지 포함)은 critical 아님, 켜짐 + topic 없음은 critical ------

func TestA095NotificationsOffNeverRecordsACritical(t *testing.T) {
	f := newA095(t, a095NotificationsOff, a095StopPctRefused) // 전송기는 남아 있음 — 꺼짐 + topic 유지
	// 이미 원장에 남은 critical 행은 정본대로 남음 — a095 는 그 행을 건드리지 않음.
	if _, err := f.j.EnqueueAlert(context.Background(), journal.Alert{EventKey: "a095-older", Type: "execgw.order_unresolved",
		Severity: "critical", Title: "older", Body: "older"}); err != nil {
		t.Fatal(err)
	}
	f.holds("005930", "10", 70000)
	f.cycle()

	rows := f.rows()
	if len(rows) != 1 || rows[0].key != "a095-older" || rows[0].state != journal.AlertPending {
		t.Fatalf("rows = %+v, want only the older PENDING row, untouched", rows)
	}
	if f.latched() || f.entryBlocked() {
		t.Error("a report on a notifications-off engine reached the gate or the operating mode")
	}
	if len(f.capture.events) != 1 || f.capture.events[0].Type != obs.EventExitPositionUnmanaged {
		t.Errorf("events = %+v, want the normal unmanaged report", f.capture.events)
	}
	// 재시작이 그 옛 행의 차단을 푸는 우회로가 아님(정본) — 새 게이트에도 래치가 섬.
	f.restart()
	if !f.latched() {
		t.Error("a restart with an undelivered critical row came up unlatched")
	}
}

func TestA095OnWithoutATopicIsStillCritical(t *testing.T) {
	f := newA095(t, a095NoTopic, a095StopPctRefused)
	f.holds("005930", "10", 70000)
	f.cycle()
	row := f.onlyRow(t)
	if row.severity != string(obs.SeverityCritical) || row.state != journal.AlertPending {
		t.Fatalf("row = %+v, want a PENDING critical row: on + no topic is on (판정은 설정 enabled)", row)
	}
	before := row.attempts
	if err := f.del.cycle(context.Background()); err != nil {
		t.Fatalf("deliverer cycle: %v", err)
	}
	if after := f.onlyRow(t); after.attempts <= before {
		t.Errorf("attempts %d → %d: the deliverer must count a missing transport as a failed attempt", before, after.attempts)
	}
}

// --- 2.5b + 조립 — 생산 조립은 로드된 설정의 enabled 를 읽음 --------------------------------

func a095LoadConfig(t *testing.T, notifications string) config.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	body := `{"schema_version":4,"trading":{},"engine":{"notifications":` + notifications + `}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewService(path).Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

func TestA095TheProductionAssemblyReadsTheLoadedSwitch(t *testing.T) {
	for _, tc := range []struct {
		name         string
		block        string
		wantCritical bool
	}{
		{"refused block written as enabled", `{"enabled":true,"base_url":"ftp://elsewhere","topic":"a095"}`, false},
		{"accepted block, enabled", `{"enabled":true,"base_url":"https://ntfy.example.internal","topic":"a095"}`, true},
		{"accepted block, disabled with a topic kept", `{"enabled":false,"topic":"a095"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA095(t)
			cfg := a095LoadConfig(t, tc.block)
			cfg.Engine.Adoption = config.Adoption{Enabled: true, DefaultStopPct: 0} // 범주 ①
			c := &Context{Config: cfg, Automation: AutomationStatus{Verified: true}, Journal: f.j,
				Reconcile: f.opts.Tracker, Ingest: f.opts.Ingest, Converge: f.opts.Converge,
				Notifier: f.n, AccountRef: a095Account, Names: &InstrumentNames{}}
			opts := ReconcileDriverOptions{Collector: f.opts.Collector, Prices: f.prices, Clock: f.clk,
				DefaultMarket:        "kr",
				NotificationsEnabled: !tc.wantCritical} // 호출자가 준 값은 조립이 덮어야 함
			d, err := c.ReconcileDriver(opts)
			if err != nil {
				t.Fatalf("Context.ReconcileDriver: %v", err)
			}
			f.d = d
			f.holds("005930", "10", 70000)
			f.cycle()
			rows := f.rows()
			if tc.wantCritical {
				if len(rows) != 1 || rows[0].severity != string(obs.SeverityCritical) {
					t.Fatalf("rows = %+v, want one critical row from the loaded enabled switch", rows)
				}
				return
			}
			if len(rows) != 0 {
				t.Fatalf("rows = %+v, want none: the loaded switch is off", rows)
			}
		})
	}
}

// --- 2.6 · 2.6a — Q2 답: 설정 거부 · include(꺼짐) 시도 실패 · 연기는 normal, 조건 칸이 갈림 ---------

func a095LoadAdoption(t *testing.T, block string) config.Adoption {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	body := `{"schema_version":4,"trading":{},"engine":{"adoption":` + block + `}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewService(path).Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg.Engine.Adoption
}

func TestA095Q2FactsStayNormalInTheirOwnCells(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options func(t *testing.T) []a095Option
		arrange func(*a095Fixture)
		cell    string
	}{
		{"refused block that asked for protection", func(t *testing.T) []a095Option {
			return []a095Option{a095Adoption(a095LoadAdoption(t, `{"enabled":true,"default_stop_pct":5}`))}
		}, nil, "rejected"},
		{"refused block that was switched off", func(t *testing.T) []a095Option {
			return []a095Option{a095Adoption(a095LoadAdoption(t, `{"enabled":false,"default_stop_pct":5}`))}
		}, nil, "rejected"},
		{"designated with adoption off, deferred", func(*testing.T) []a095Option {
			return []a095Option{a095Adoption(config.Adoption{IncludeSymbols: []string{"005930"}, DefaultStopPct: 0.05})}
		}, func(f *a095Fixture) { f.prices.err = fmt.Errorf("a095 quote outage") }, "include_deferred"},
		{"adoption on, the price read failed", func(*testing.T) []a095Option { return nil },
			func(f *a095Fixture) { f.prices.err = fmt.Errorf("a095 quote outage") }, "enabled_deferred"},
		{"adoption on, no quote for the symbol", func(*testing.T) []a095Option { return nil },
			func(f *a095Fixture) { delete(f.prices.last, "005930") }, "enabled_deferred"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA095(t, tc.options(t)...)
			f.holds("005930", "10", 70000)
			if tc.arrange != nil {
				tc.arrange(f)
			}
			if cycle := f.cycle(); cycle.Unmanaged != 1 {
				t.Fatalf("unmanaged = %d, want 1", cycle.Unmanaged)
			}
			f.assertNoCriticalConsequence(t)
			if len(f.capture.events) != 1 {
				t.Fatalf("events = %+v, want one report", f.capture.events)
			}
			e := f.capture.events[0]
			if e.Type != obs.EventExitPositionUnmanaged || !strings.Contains(e.Key, "|reconcile|"+tc.cell+"|") {
				t.Errorf("event = %s %q, want the normal kind in the cell %q", e.Type, e.Key, tc.cell)
			}
		})
	}
}

// --- 2.8 — 전이 상태(묵은 스냅샷)는 무알림 ----------------------------------------------------

func TestA095AStaleSnapshotStaysSilent(t *testing.T) {
	f := newA095(t, a095StopPctRefused)
	f.opts.SnapshotStaleness = time.Minute
	f.rebuild()
	calls := 0
	f.holdings.onCall = func() {
		calls++
		if calls%2 == 0 {
			f.clk.Advance(2 * time.Minute) // 둘째 수집 도중 시계가 움직여 스냅샷이 판정 순간 묵음
		}
	}
	f.holds("005930", "10", 70000)
	cycle := f.cycle()
	if cycle.Unmanaged != 0 || len(f.capture.events) != 0 {
		t.Errorf("cycle = %+v, events = %+v; a stale snapshot is a transition state and stays silent",
			cycle, f.capture.events)
	}
}

// --- 2.9 — fold 알림 어댑터는 생산에서 닿지 않음 ------------------------------------------------

func TestA095TheFoldAlertStaysUnwired(t *testing.T) {
	f := newA095(t)
	f.opts.Ingest = &reconcile.Ingestor{Journal: f.j, AccountRef: a095Account, DefaultMarket: "kr",
		Alert: notifierAlerter{notifier: f.n}}
	f.rebuild()
	if f.d.ingest.Alert != nil {
		t.Error("the driver's ingest copy carries the fold alert; IngestExternalPositions B12 would stop guarding it")
	}
	if f.opts.Ingest.Alert == nil {
		t.Error("the caller's Ingestor was modified; the driver must copy it")
	}
	if obs.SeverityOf(obs.EventExitPositionUnmanaged) != obs.SeverityNormal {
		t.Error("the unmanaged kind became critical; the fold and exit sites would go with it")
	}
}

// --- 2.12 — 연기(normal) → 같은 프로세스의 시도 실패(critical), 한 사이클 안의 묶음 ---------------------

func TestA095ADeferralThenAFailureIsRecorded(t *testing.T) {
	f := newA095(t, a095StopPctRefused)
	f.holds("005930", "10", 70000)
	f.prices.err = fmt.Errorf("a095 quote outage")
	f.cycle()
	if len(f.rows()) != 0 || len(f.capture.keysContaining("|enabled_deferred|")) != 1 {
		t.Fatalf("after the deferral: rows = %+v, events = %+v", f.rows(), f.capture.events)
	}
	f.prices.err = nil
	f.cycle()
	row := f.onlyRow(t)
	if row.severity != string(obs.SeverityCritical) {
		t.Fatalf("row = %+v; the earlier normal report latched the later critical failure away", row)
	}
}

// a095AdvancingWriter 는 로그 줄을 쓸 때마다 가짜 시계를 움직임 — adoptOne 의 거절 로그 뒤 다음 후보의 관측이 묵게 함.
type a095AdvancingWriter struct {
	clk *clock.Fake
	by  time.Duration
	buf bytes.Buffer
}

func (w *a095AdvancingWriter) Write(p []byte) (int, error) {
	w.clk.Advance(w.by)
	return w.buf.Write(p)
}

func TestA095OneCycleCanHoldAFailureAndADeferral(t *testing.T) {
	f := newA095(t, a095StopPctRefused)
	f.opts.Log = obs.NewLogger(obs.LogOptions{Writer: &a095AdvancingWriter{clk: f.clk,
		by: DefaultAdoptionPriceStaleness + time.Second}, JSON: true, Clock: f.clk})
	f.rebuild()
	f.holds("000660", "5", 150000) // 먼저 시도되고 실패함(범주 ①), 그 거절 로그가 시계를 움직임
	f.holds("005930", "10", 70000) // 관측이 묵어 adopt B7 로 남음 — 연기
	f.cycle()

	failed := f.capture.keysContaining("|enabled_failed|")
	deferred := f.capture.keysContaining("|enabled_deferred|")
	if len(failed) != 1 || len(deferred) != 1 {
		t.Fatalf("failed = %v, deferred = %v, events = %+v; want one of each", failed, deferred, f.capture.events)
	}
	row := f.onlyRow(t)
	if row.key != failed[0] {
		t.Errorf("the critical row is %q, want the attempted-and-failed one %q", row.key, failed[0])
	}
}

// --- 2.13 — 범주 ③(커밋 뒤 보호 미개설)은 critical 요구 밖 — 이름 붙은 경계 -----------------------

func TestA095AnAdoptionWithoutItsExitStateIsTheNamedBoundary(t *testing.T) {
	f := newA095(t)
	f.holds("005930", "10", 70000)
	f.exec(`CREATE TRIGGER a095_refuse_exit_state BEFORE INSERT ON exit_states
		BEGIN SELECT RAISE(ABORT, 'a095 exit state refusal'); END`)
	cycle := f.cycle()
	if cycle.Adopted != 1 {
		t.Fatalf("adopted = %d; adoptOne reports a committed adoption as adopted (issues I7)", cycle.Adopted)
	}
	for _, r := range f.rows() {
		if r.typ == string(obs.EventExitPositionAdoptionFailed) {
			t.Errorf("row %+v: category ③ is outside the critical requirement (후속 후보 I7)", r)
		}
	}
}

// --- 2.14 — 전이 행렬: 로드된 설정 전환 × 재시작 ------------------------------------------------

func TestA095TheSwitchTransitionMatrix(t *testing.T) {
	t.Run("on, pending row, then off across a restart", func(t *testing.T) {
		f := newA095(t, a095StopPctRefused)
		f.pub.fail = fmt.Errorf("a095 transport down")
		f.holds("005930", "10", 70000)
		f.cycle()
		if row := f.onlyRow(t); row.state != journal.AlertPending || row.severity != string(obs.SeverityCritical) {
			t.Fatalf("row = %+v, want a PENDING critical row", row)
		}
		f.opts.NotificationsEnabled = false
		f.restart()
		if !f.latched() {
			t.Error("the restart reopened entries over an undelivered critical row")
		}
		f.cycle()
		if len(f.capture.events) != 1 || f.capture.events[0].Type != obs.EventExitPositionUnmanaged {
			t.Errorf("events = %+v, want the new fact normal once off", f.capture.events)
		}
		if row := f.onlyRow(t); row.state != journal.AlertPending {
			t.Errorf("the older row is %s, want it left PENDING for the operator", row.state)
		}
	})
	t.Run("off, then on across a restart", func(t *testing.T) {
		f := newA095(t, a095NotificationsOff, a095StopPctRefused)
		f.holds("005930", "10", 70000)
		f.cycle()
		if len(f.rows()) != 0 {
			t.Fatalf("rows = %+v while off", f.rows())
		}
		f.opts.NotificationsEnabled = true
		f.restart()
		if f.latched() {
			t.Error("a restart with nothing undelivered latched the gate")
		}
		f.cycle()
		if row := f.onlyRow(t); row.severity != string(obs.SeverityCritical) {
			t.Errorf("row = %+v, want the failure critical once on", row)
		}
	})
}

// --- 2.15 — 배달됨 → 재알림 창 경과 → 여전히 실패면 다시 전송 ----------------------------------------

func TestA095ADeliveredFailureIsRemindedAfterTheWindow(t *testing.T) {
	f := newA095(t, a095StopPctRefused)
	f.holds("005930", "10", 70000)
	f.cycle()
	f.cycle()
	if got := f.criticalSends(); got != 1 {
		t.Fatalf("sends inside the window = %d, want 1", got)
	}
	first := f.onlyRow(t).body
	f.clk.Advance(11 * time.Minute)
	again := journal.RFC3339(f.clk.Now().Add(reconcile.DefaultStabilisationInterval))
	f.cycle()
	if got := f.criticalSends(); got != 2 {
		t.Errorf("sends after the window = %d, want 2: a memory latch must not hold a continuing critical", got)
	}
	// 재무장된 행은 통째로 이번 에피소드를 말함(정본) — 본문이 새 실패의 시각을 담아야 함(독립 리뷰 P3).
	if body := f.onlyRow(t).body; !strings.Contains(body, again) || body == first {
		t.Errorf("the re-armed row says %q; it must speak of the new failure at %s", body, again)
	}
}

// --- 2.16 — outbox 기록 실패 → 저장소 회복 → 같은 실패가 기록됨 · 사람 래치는 그대로 ------------------

func TestA095ARecordingFailureIsRetriedWhenTheStoreRecovers(t *testing.T) {
	f := newA095(t, a095StopPctRefused)
	f.holds("005930", "10", 70000)
	f.exec(`CREATE TRIGGER a095_refuse_outbox BEFORE INSERT ON alert_outbox
		BEGIN SELECT RAISE(ABORT, 'a095 outbox refusal'); END`)
	f.cycle()
	if len(f.rows()) != 0 {
		t.Fatalf("rows = %+v under a refusing outbox", f.rows())
	}
	if !f.latched() {
		t.Fatal("a critical alert that could not be recorded left the gate open (정본 「durable 기록의 실패」)")
	}
	f.exec(`DROP TRIGGER a095_refuse_outbox`)
	f.cycle()
	if row := f.onlyRow(t); row.severity != string(obs.SeverityCritical) {
		t.Errorf("row = %+v; the recovered store must receive the same failure on the next observation", row)
	}
	if !f.latched() {
		t.Error("the human-owned latch was released by the recovery")
	}
}

// --- 2.17 — 사실 식별자 ----------------------------------------------------------------------

func TestA095TheFactIdentity(t *testing.T) {
	t.Run("a deferral and a failure are different keys", func(t *testing.T) {
		f := newA095(t, a095StopPctRefused)
		f.holds("005930", "10", 70000)
		f.prices.err = fmt.Errorf("a095 quote outage")
		f.cycle()
		f.prices.err = nil
		f.cycle()
		d, fl := f.capture.keysContaining("|enabled_deferred|"), f.capture.keysContaining("|enabled_failed|")
		if len(d) != 1 || len(fl) != 1 || d[0] == fl[0] {
			t.Errorf("deferred %v · failed %v, want one each with different keys", d, fl)
		}
	})
	t.Run("an earlier normal fact does not latch a later different normal fact", func(t *testing.T) {
		f := newA095(t, a095NotificationsOff, a095StopPctRefused) // 꺼짐 — 둘 다 normal
		f.holds("005930", "10", 70000)
		f.prices.err = fmt.Errorf("a095 quote outage")
		f.cycle() // 연기
		f.prices.err = nil
		f.cycle() // 시도 실패 — 다른 사실
		f.cycle() // 같은 사실의 반복 — 억제
		d, fl := f.capture.keysContaining("|enabled_deferred|"), f.capture.keysContaining("|enabled_failed|")
		if len(d) != 1 || len(fl) != 1 {
			t.Errorf("deferred %v · failed %v, want each once: the latch is per fact, not per position", d, fl)
		}
	})
	t.Run("only the error text changed", func(t *testing.T) {
		f := newA095(t)
		f.holds("005930", "10", 70000)
		f.refuseAdoptions("first wording")
		f.cycle()
		f.refuseAdoptions("second wording")
		f.cycle()
		fl := f.capture.keysContaining("|enabled_failed|")
		if len(fl) != 2 || fl[0] != fl[1] {
			t.Errorf("keys = %v, want the same key twice", fl)
		}
		if rows := f.rows(); len(rows) != 1 {
			t.Errorf("rows = %+v, want one row for one fact", rows)
		}
	})
	t.Run("A B A with A settled is absorbed by the window", func(t *testing.T) {
		f := newA095(t, a095StopPctRefused)
		f.holds("005930", "10", 70000)
		f.cycle() // A 배달
		f.prices.err = fmt.Errorf("a095 quote outage")
		f.cycle() // B
		f.prices.err = nil
		f.cycle() // A — 창 안
		if got := f.criticalSends(); got != 1 {
			t.Errorf("sends = %d, want 1: a settled A inside the window is absorbed", got)
		}
	})
	t.Run("A B A with A pending is retried", func(t *testing.T) {
		f := newA095(t, a095StopPctRefused)
		f.pub.fail = fmt.Errorf("a095 transport down")
		f.holds("005930", "10", 70000)
		f.cycle()
		first := f.onlyRow(t).attempts
		f.prices.err = fmt.Errorf("a095 quote outage")
		f.cycle()
		f.prices.err = nil
		f.cycle()
		if again := f.onlyRow(t).attempts; again <= first {
			t.Errorf("attempts %d → %d: a PENDING A is unfinished, not a duplicate", first, again)
		}
	})
	t.Run("another sender holds A's lease", func(t *testing.T) {
		f := newA095(t, a095StopPctRefused)
		f.pub.fail = fmt.Errorf("a095 transport down")
		f.holds("005930", "10", 70000)
		f.cycle()
		row := f.onlyRow(t)
		claim, err := f.j.ClaimAlertForDelivery(context.Background(), journal.Alert{EventKey: row.key,
			Type: row.typ, Severity: row.severity, Title: row.title, Body: row.body}, 10*time.Minute, "a095-other")
		if err != nil || claim.Disposition != journal.ClaimAcquired {
			t.Fatalf("the other sender's claim = %+v, %v", claim, err)
		}
		f.pub.fail = nil
		f.cycle()
		if got := f.criticalSends(); got != 0 {
			t.Errorf("sends = %d, want 0 while another sender holds the lease", got)
		}
	})
}

// --- 2.11 — Q8 답: 해소 뒤에 배달된 critical 행은 그 시각의 사건을 말함 ---------------------------------

func TestA095AResolvedFailureStillSpeaksItsMoment(t *testing.T) {
	f := newA095(t)
	f.pub.fail = fmt.Errorf("a095 transport down")
	f.holds("005930", "10", 70000)
	f.refuseAdoptions("a095 refusal")
	f.cycle()
	moment := journal.RFC3339(f.clk.Now())
	row := f.onlyRow(t)
	if !strings.Contains(row.body, moment) {
		t.Fatalf("body = %q; the critical row must name the moment of the failure (%s)", row.body, moment)
	}
	if !f.latched() {
		t.Fatal("arrangement: the failed synchronous send should have latched the gate")
	}

	f.allowAdoptions()
	f.clk.Advance(time.Minute)
	if cycle := f.cycle(); cycle.Adopted != 1 {
		t.Fatalf("adopted = %d on the recovery cycle", cycle.Adopted)
	}
	if !f.latched() {
		t.Error("the adoption succeeding released the human-owned latch")
	}

	f.pub.fail = nil
	if err := f.del.cycle(context.Background()); err != nil {
		t.Fatalf("deliverer cycle: %v", err)
	}
	if got := f.onlyRow(t); got.state != journal.AlertDelivered {
		t.Fatalf("row state = %s, want DELIVERED", got.state)
	}
	f.pub.mu.Lock()
	sent := f.pub.sent[len(f.pub.sent)-1]
	f.pub.mu.Unlock()
	if !strings.Contains(sent.Body, moment) {
		t.Errorf("the late delivery says %q; it must still speak of %s", sent.Body, moment)
	}
	if !f.latched() {
		t.Error("a late delivery was counted as the operator's acknowledgement")
	}
}

// a095AdvancingPublisher 는 전송할 때마다 가짜 시계를 움직임 — 앞 후보의 동기 전송이 시간을 쓰는 동안 뒤 후보의 보고가 늦게
// 만들어지는 모양(codex 교차 리뷰 P2).
type a095AdvancingPublisher struct {
	inner *a098RecordingPublisher
	clk   *clock.Fake
	by    time.Duration
}

func (p *a095AdvancingPublisher) Publish(ctx context.Context, n obs.Notification) error {
	p.clk.Advance(p.by)
	return p.inner.Publish(ctx, n)
}

// 2.11 보강 — critical 문장의 시각은 보고를 만든 순간이 아니라 시도 실패를 관측한 순간임.
func TestA095EachFailureSpeaksTheMomentItFailed(t *testing.T) {
	f := newA095(t, a095StopPctRefused)
	f.n.Publisher = &a095AdvancingPublisher{inner: f.pub, clk: f.clk, by: 30 * time.Second}
	f.rebuild()
	f.holds("000660", "5", 150000)
	f.holds("005930", "10", 70000)
	failedAt := journal.RFC3339(f.clk.Now().Add(reconcile.DefaultStabilisationInterval))
	f.cycle()

	rows := f.rows()
	if len(rows) != 2 {
		t.Fatalf("rows = %+v, want two critical rows", rows)
	}
	for _, r := range rows {
		if !strings.Contains(r.body, failedAt) {
			t.Errorf("row %s says %q; both attempts failed at %s, before any send", r.key, r.body, failedAt)
		}
	}
}
