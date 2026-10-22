package engine_test

// a163 tasks 2.6~2.9 · 2.11(b) · 2.12 · 2.14 · 2.15 — 체결 나이 보고(design D4 · D8).
//
// 보고 토글 ON 관측자에서: 판정이 끝난 뒤 구간이 바뀐 포지션마다 SourceAgeLog 한 줄, 브로커 시각이 없거나 해독 불가인 포지션은
// source_missing 일반 등급 알림 한 번(래치 · 주기당 8건). OFF 관측자에서는 줄 · 알림 · 래치 0 이고 판정 · 제출은 ON 과 같음.
// 하네스의 fakePrices 는 FetchedAt 을 비워 두므로 받은 시각은 관측자의 사이클 대체(validatedAt = 가짜 시계 지금)임 —
// design D4 「나이 = observedQuote.FetchedAt − 브로커 시각」 의 대체 경로(exitloop.go:867-870)를 그대로 지남.

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// --- 고정물 ---------------------------------------------------------------------

// a163StampedPrices 는 하네스의 fakePrices 응답에 종목별 브로커 데이터 시각 원문을 붙임(어댑터가 하는 일과 같은 모양).
type a163StampedPrices struct {
	inner  *fakePrices
	stamps map[string]string
}

func (p *a163StampedPrices) Prices(ctx context.Context, symbols []string) ([]domain.Quote, error) {
	quotes, err := p.inner.Prices(ctx, symbols)
	for i := range quotes {
		quotes[i].SourceTimestamp = p.stamps[quotes[i].Symbol]
	}
	return quotes, err
}

type a163Fixture struct {
	*exitHarness
	stamps    map[string]string
	logs      *a090Buffer
	lastOrder string
}

var a163KST = time.FixedZone("KST", 9*60*60)

// a163Setup 은 exit 하네스에 시각을 붙이는 가격 원천과 SourceAgeLog 를 꽂음. on 이 보고 토글.
func a163Setup(t *testing.T, on bool, mutate func(*engine.ExitObserverOptions)) *a163Fixture {
	t.Helper()
	f := &a163Fixture{stamps: map[string]string{}, logs: &a090Buffer{}}
	f.exitHarness = newExitHarness(t, func(o *engine.ExitObserverOptions) {
		o.Prices = &a163StampedPrices{inner: o.Prices.(*fakePrices), stamps: f.stamps}
		o.SourceAgeLog = obs.NewLogger(obs.LogOptions{Writer: f.logs, JSON: true, Clock: o.Clock})
		o.SourceAgeReport = on
		if mutate != nil {
			mutate(o)
		}
	})
	record := f.submit.record
	f.submit.record = func(req execgw.PlaceRequest) string {
		id := record(req)
		f.lastOrder = id
		return id
	}
	return f
}

// ago 는 받은 시각(가짜 시계 지금) 기준 d 전의 브로커 원문 시각임 — 실측 형식(KST 오프셋 · 밀리초).
func (f *a163Fixture) ago(d time.Duration) string {
	return f.clk.Now().Add(-d).In(a163KST).Format("2006-01-02T15:04:05.000Z07:00")
}

func (f *a163Fixture) cycle() engine.ExitCycle {
	f.t.Helper()
	f.clk.Advance(5 * time.Second)
	return f.observe()
}

func a163Lines(t *testing.T, raw string, event obs.EventType) []map[string]any {
	t.Helper()
	var out []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		var line map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("log line is not JSON: %q", scanner.Text())
		}
		if event == "" || line[obs.FieldEvent] == string(event) {
			out = append(out, line)
		}
	}
	return out
}

func (f *a163Fixture) ageLines() []map[string]any {
	return a163Lines(f.t, f.logs.String(), obs.EventExitSourceAge)
}

// bucketsOf 는 종목별 구간 줄의 bucket 열임 — 로거가 position_id 값을 가림(redact)으로 처리하므로 종목으로 고름.
func (f *a163Fixture) bucketsOf(symbol string) []string {
	var out []string
	for _, l := range f.ageLines() {
		if l[obs.FieldSymbol] == symbol {
			out = append(out, fmt.Sprint(l["bucket"]))
		}
	}
	return out
}

func (f *a163Fixture) missingAlerts() []obs.Event {
	var out []obs.Event
	for _, e := range f.alerts.events {
		if e.Type == obs.EventExitSourceMissing {
			out = append(out, e)
		}
	}
	return out
}

func (f *a163Fixture) missingAlertsFor(positionID string) int {
	n := 0
	for _, e := range f.missingAlerts() {
		if e.Fields["position_id"] == positionID {
			n++
		}
	}
	return n
}

// closePosition 은 방금 제출된 청산 매도를 전량 체결로 기록해 포지션을 닫음.
func (f *a163Fixture) closePosition(symbol, quantity, price string) {
	f.t.Helper()
	if f.lastOrder == "" {
		f.t.Fatal("no exit order was submitted to fill")
	}
	if _, err := f.journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: f.lastOrder, Symbol: symbol, Market: "kr", State: "CLOSED_FILLED", Terminal: true,
		Quantity: quantity, FilledQuantity: quantity, AveragePrice: price,
		AccountRef: exitAccount, TradingDay: "2026-03-30", Side: "SELL",
		ObservedAt: f.clk.Now().Format(time.RFC3339),
	}); err != nil {
		f.t.Fatalf("RecordFill: %v", err)
	}
}

// --- 2.6 알림 · 래치 · 재무장 · 정리 · 상한 ------------------------------------------

func TestA163AMissingOrUnparseableTimeIsAlertedOnceAndRearmsWhenATimeArrives(t *testing.T) {
	f := a163Setup(t, true, nil)
	a := f.entry("005930", "10", "70000", "68000", "70000")
	b := f.entry("000660", "4", "120000", "118000", "120000")
	c := f.entry("035420", "5", "200000", "190000", "200000")
	f.quote("005930", 70100)
	f.quote("000660", 120100)
	f.quote("035420", 200100)
	f.stamps["005930"] = ""                                  // missing
	f.stamps["000660"] = "garbage"                           // unparseable
	f.stamps["035420"] = f.ago(9*time.Hour + 30*time.Minute) // gt_1h — 로그만

	if c1 := f.cycle(); c1.Judged != 3 {
		t.Fatalf("judged = %d, want 3", c1.Judged)
	}
	if f.missingAlertsFor(a.ID) != 1 || f.missingAlertsFor(b.ID) != 1 || f.missingAlertsFor(c.ID) != 0 {
		t.Fatalf("alerts a=%d b=%d c=%d, want 1 · 1 · 0 (a stale time is logged, never alerted — H1)",
			f.missingAlertsFor(a.ID), f.missingAlertsFor(b.ID), f.missingAlertsFor(c.ID))
	}
	f.cycle()
	if got := len(f.missingAlerts()); got != 2 {
		t.Fatalf("alerts after a second cycle = %d, want still 2 — the latch holds a reason until it clears", got)
	}
	f.stamps["005930"] = f.ago(3 * time.Second)
	f.cycle()
	if got := len(f.missingAlerts()); got != 2 {
		t.Fatalf("alerts when a's time arrives = %d, want 2", got)
	}
	f.stamps["005930"] = ""
	f.cycle()
	if f.missingAlertsFor(a.ID) != 2 || f.missingAlertsFor(b.ID) != 1 {
		t.Fatalf("after a's time vanished again: a=%d b=%d, want a rearmed (2) and b still latched (1)",
			f.missingAlertsFor(a.ID), f.missingAlertsFor(b.ID))
	}
	for _, e := range f.missingAlerts() {
		if obs.SeverityOf(e.Type) != obs.SeverityNormal {
			t.Fatalf("source_missing is graded %s, want normal (H2)", obs.SeverityOf(e.Type))
		}
	}
}

func TestA163AHeldPositionThatWasNotJudgedKeepsItsLatch(t *testing.T) {
	f := a163Setup(t, true, nil)
	a := f.entry("005930", "10", "70000", "68000", "70000")
	d := f.entry("000660", "4", "120000", "118000", "120000")
	f.quote("005930", 70100)
	f.quote("000660", 120100)
	f.stamps["005930"], f.stamps["000660"] = "", ""
	f.cycle()
	if f.missingAlertsFor(d.ID) != 1 {
		t.Fatalf("d alerts = %d, want 1", f.missingAlertsFor(d.ID))
	}
	delete(f.prices.last, "000660") // d 는 보유 · 대상이지만 이번 주기 판정되지 않음(no_quote)
	f.cycle()
	if got := f.observer.SourceAgeLatchForTest()[d.ID]; got != "missing" {
		t.Fatalf("d's latch after an unjudged cycle = %q, want it unchanged (missing)", got)
	}
	f.quote("000660", 120100)
	f.cycle()
	if f.missingAlertsFor(d.ID) != 1 || f.missingAlertsFor(a.ID) != 1 {
		t.Fatalf("alerts a=%d d=%d, want 1 · 1 — an unjudged cycle must not rearm the latch",
			f.missingAlertsFor(a.ID), f.missingAlertsFor(d.ID))
	}
}

func TestA163AClosedPositionLeavesTheLatch(t *testing.T) {
	f := a163Setup(t, true, nil)
	a := f.entry("005930", "10", "70000", "68000", "70000")
	c := f.entry("000660", "4", "120000", "118000", "120000")
	f.quote("005930", 67000) // 손절선 아래 — 첫 주기에 청산 제출
	f.quote("000660", 120100)
	f.stamps["005930"], f.stamps["000660"] = "", ""
	if cycle := f.cycle(); cycle.Proposed != 1 {
		t.Fatalf("proposed = %d, want a's liquidation", cycle.Proposed)
	}
	if latch := f.observer.SourceAgeLatchForTest(); len(latch) != 2 {
		t.Fatalf("latch = %v, want both judged positions", latch)
	}
	f.closePosition("005930", "10", "67000")
	f.cycle()
	latch := f.observer.SourceAgeLatchForTest()
	if _, kept := latch[a.ID]; kept || len(latch) != 1 || latch[c.ID] != "missing" {
		t.Fatalf("latch after a closed = %v, want only c — an id outside this cycle's held set is removed", latch)
	}
}

func TestA163AtMostEightSourceAlertsLeaveInOneCycle(t *testing.T) {
	f := a163Setup(t, true, nil)
	for i := 0; i < 12; i++ {
		symbol := fmt.Sprintf("1000%02d", i)
		f.entry(symbol, "1", "10000", "9000", "10000")
		f.quote(symbol, 10010)
		f.stamps[symbol] = ""
	}
	counts := []int{}
	for i := 0; i < 3; i++ {
		before := len(f.missingAlerts())
		if cycle := f.cycle(); cycle.Judged != 12 {
			t.Fatalf("cycle %d judged %d, want all 12 whatever the alert cap", i+1, cycle.Judged)
		}
		counts = append(counts, len(f.missingAlerts())-before)
	}
	if fmt.Sprint(counts) != "[8 4 0]" {
		t.Fatalf("alerts per cycle = %v, want [8 4 0] — the cap carries the rest to the next cycle", counts)
	}
}

// --- 2.7 순서 -------------------------------------------------------------------

// a163OrderProbe 는 첫 source_missing 알림 순간의 제출 수를 적음.
type a163OrderProbe struct {
	inner  engine.ExitAlerter
	places func() int
	at     int
}

func (p *a163OrderProbe) Notify(ctx context.Context, e obs.Event) error {
	if e.Type == obs.EventExitSourceMissing && p.at < 0 {
		p.at = p.places()
	}
	return p.inner.Notify(ctx, e)
}

func TestA163TheReportComesAfterEveryJudgementAndBeforeTheUnobservedSettlement(t *testing.T) {
	probe := &a163OrderProbe{at: -1}
	f := a163Setup(t, true, func(o *engine.ExitObserverOptions) {
		probe.inner = o.Alerts
		o.Alerts = probe
	})
	probe.places = func() int { return len(f.submit.places) }
	f.UnobservedLogTo(t)
	first := f.entry("005930", "10", "70000", "68000", "70000")
	second := f.entry("000660", "4", "120000", "118000", "120000")
	_ = f.entry("035420", "5", "200000", "190000", "200000") // 시세 없음 — a090 미관측 줄을 냄
	f.quote("005930", 70100)
	f.quote("000660", 117000) // 둘째가 손절선 아래
	f.stamps["005930"], f.stamps["000660"] = "", ""
	f.cycle()
	if probe.at != 1 {
		t.Fatalf("submissions at the first report = %d, want 1 — the second position's stop must already be out (safety 4)", probe.at)
	}
	if f.missingAlertsFor(first.ID) != 1 || f.missingAlertsFor(second.ID) != 1 {
		t.Fatalf("alerts first=%d second=%d, want 1 · 1", f.missingAlertsFor(first.ID), f.missingAlertsFor(second.ID))
	}
	lastAge, firstUnobserved := -1, -1
	for i, l := range a163Lines(t, f.logs.String(), "") {
		switch l[obs.FieldEvent] {
		case string(obs.EventExitSourceAge):
			lastAge = i
		case string(obs.EventExitPositionUnobserved):
			if firstUnobserved < 0 {
				firstUnobserved = i
			}
		}
	}
	if lastAge < 0 || firstUnobserved < 0 || lastAge > firstUnobserved {
		t.Fatalf("line order age=%d unobserved=%d, want every source-age line before the a090 settlement's first line", lastAge, firstUnobserved)
	}
}

// UnobservedLogTo 는 a090 줄을 같은 버퍼에 쓰게 해 두 보고의 순서를 한 열에서 읽음.
func (f *a163Fixture) UnobservedLogTo(t *testing.T) {
	t.Helper()
	opts := f.observer.OptionsForTest()
	opts.UnobservedLog = obs.NewLogger(obs.LogOptions{Writer: f.logs, JSON: true, Clock: opts.Clock})
	observer, err := engine.NewExitObserver(opts)
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	f.observer = observer
}

// --- 2.8 계좌 없음 · 필드 목록 --------------------------------------------------------

func TestA163NoReportLineOrAlertCarriesTheAccount(t *testing.T) {
	f := a163Setup(t, true, nil)
	f.entry("005930", "10", "70000", "68000", "70000")
	f.entry("000660", "4", "120000", "118000", "120000")
	f.entry("035420", "5", "200000", "190000", "200000")
	f.quote("005930", 70100)
	f.quote("000660", 120100)
	f.quote("035420", 200100)
	f.stamps["005930"], f.stamps["000660"], f.stamps["035420"] = "", strings.Repeat("x", 100), f.ago(2*time.Minute)
	f.cycle()
	allowed := map[string]bool{
		"position_id": true, obs.FieldMarket: true, obs.FieldSymbol: true, "bucket": true, "prev_bucket": true,
		"age_seconds": true, "source_timestamp": true,
		// 로거 자신의 열
		"time": true, "level": true, "msg": true, obs.FieldEvent: true, obs.FieldSubject: true, obs.FieldSeverity: true,
	}
	lines := f.ageLines()
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	for _, l := range lines {
		for key := range l {
			if !allowed[key] {
				t.Errorf("source-age line carries %q, outside design D4's field list: %v", key, l)
			}
		}
		if raw, ok := l["source_timestamp"].(string); ok && (l["bucket"] != "unparseable" || len(raw) > 64) {
			t.Errorf("source_timestamp = %q on bucket %v — only an unparseable time, first 64 bytes", raw, l["bucket"])
		}
	}
	if strings.Contains(f.logs.String(), exitAccount) {
		t.Fatalf("the source-age log carries the account ref:\n%s", f.logs.String())
	}
	alerts, _ := json.Marshal(f.missingAlerts())
	if len(f.missingAlerts()) != 2 || strings.Contains(string(alerts), exitAccount) {
		t.Fatalf("alerts = %s, want two and no account ref", alerts)
	}
	for _, e := range f.missingAlerts() {
		for key := range e.Fields {
			if !allowed[key] {
				t.Errorf("source_missing alert carries %q: %v", key, e.Fields)
			}
		}
	}
}

// --- 2.9 진입 래치 · 모드 · outbox 0 ---------------------------------------------------

func TestA163TheReportNeverBlocksEntryOrEscalates(t *testing.T) {
	f := a090Setup(t, func(o *engine.ExitObserverOptions) {
		o.Prices = &a163StampedPrices{inner: o.Prices.(*fakePrices), stamps: map[string]string{"005930": "", "000660": "garbage"}}
		o.SourceAgeReport = true
		o.SourceAgeLog = obs.NewLogger(obs.LogOptions{Writer: &a090Buffer{}, JSON: true, Clock: o.Clock})
	})
	f.entry("005930", "10", "70000", "68000", "70000")
	f.entry("000660", "4", "120000", "118000", "120000")
	f.quote("005930", 70100)
	f.quote("000660", 120100)
	f.cycles(3)
	if rows := a092PendingOfType(t, f.journal, string(obs.EventExitSourceMissing)); len(rows) != 0 {
		t.Fatalf("outbox rows for source_missing = %d, want 0 — a normal-grade report is never durable (H2)", len(rows))
	}
	if mode := f.mode(); mode != journal.ModeNormal {
		t.Fatalf("operating mode = %s, want NORMAL — a report must not tighten the account", mode)
	}
	if latch := f.gate.EntryLatch(); latch.Blocked {
		t.Fatalf("the entry gate is blocked (%s %s) — a report must not latch entries", latch.Reason, latch.Detail)
	}
}

// --- 2.11(b) 생산 조립 ------------------------------------------------------------------

func TestA163TheAssembledObserverWritesItsLinesToTheDedicatedLogger(t *testing.T) {
	broker, srv := newA163Broker(t)
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	eng.Config.Engine.ExitSourceAgeReport = true
	sink := &a090Buffer{}
	logger := obs.NewLogger(obs.LogOptions{Writer: sink, JSON: true})
	observer, err := eng.ExitObserver(engine.ExitObserverOptions{
		Costs: costs.DefaultModel(), Clock: clock.System(), SourceAgeLog: logger,
	})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	opts := observer.OptionsForTest()
	if opts.SourceAgeLog != logger || opts.Log != nil {
		t.Fatalf("SourceAgeLog passed=%v Log=%v, want the dedicated logger through and no whole-observer log", opts.SourceAgeLog == logger, opts.Log)
	}
	s := &e2eStack{t: t, broker: broker.e2eBroker, engine: eng, observer: observer, clk: clock.NewFake(time.Now().UTC())}
	s.entry("005930", "10", "70000", "68000")
	// 실제 브로커 원문 모양의 묵은 시각 — 어댑터 · observe · 보고를 지나 gt_1h 로 분류돼야 함(중간에서 버려지면 missing 이 됨).
	broker.set("005930", "70100", a163StaleStamp)
	if cycle := observer.ObserveOnce(context.Background()); cycle.Err != nil || cycle.Judged != 1 {
		t.Fatalf("assembled cycle = %+v, want one judged position", cycle)
	}
	if lines := a163Lines(t, sink.String(), obs.EventExitSourceAge); len(lines) != 1 || lines[0]["bucket"] != "gt_1h" {
		t.Fatalf("dedicated logger lines = %v, want one gt_1h line — the broker time must travel adapter → observation → report", lines)
	}
}

// --- 2.12 빈도 · 분류 -------------------------------------------------------------------

func TestA163AnAgeLineIsWrittenOnlyWhenTheBucketChanges(t *testing.T) {
	f := a163Setup(t, true, nil)
	f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("005930", 70100)
	for i := 0; i < 3; i++ {
		f.stamps["005930"] = f.ago(3 * time.Second) // 매 주기 새 체결 — 같은 구간
		f.cycle()
	}
	if got := f.bucketsOf("005930"); fmt.Sprint(got) != "[le_15s]" {
		t.Fatalf("lines over three le_15s cycles = %v, want one", got)
	}
	f.stamps["005930"] = f.ago(2 * time.Minute)
	f.cycle()
	got := f.ageLines()
	if fmt.Sprint(f.bucketsOf("005930")) != "[le_15s le_5m]" || got[len(got)-1]["prev_bucket"] != "le_15s" {
		t.Fatalf("lines after a bucket change = %v", got)
	}
}

func TestA163EveryBucketIsClassifiedFromTheReceiptTime(t *testing.T) {
	f := a163Setup(t, true, nil)
	cases := []struct {
		symbol, stamp, want string
	}{
		{"100001", "", "missing"},
		{"100002", "2026-03-30T09:59:58.000+0900", "unparseable"}, // 계약 형식 변화(+0900)
		{"100003", "", "ahead"},
		{"100004", "", "le_15s"},
		{"100005", "", "le_5m"},
		{"100006", "", "le_1h"},
		{"100007", "", "gt_1h"},
		{"100008", "", "le_15s"}, // 경계 15 s 포함
		{"100009", "", "le_5m"},  // 경계 5 분 포함
		{"100010", "", "le_1h"},  // 경계 1 시간 포함
	}
	for _, c := range cases {
		f.entry(c.symbol, "1", "10000", "9000", "10000")
		f.quote(c.symbol, 10010)
	}
	f.clk.Advance(5 * time.Second)
	f.stamps["100001"] = ""
	f.stamps["100002"] = cases[1].stamp
	f.stamps["100003"] = f.ago(-5 * time.Minute)
	f.stamps["100004"] = f.ago(3 * time.Second)
	f.stamps["100005"] = f.ago(2 * time.Minute)
	f.stamps["100006"] = f.ago(30 * time.Minute)
	f.stamps["100007"] = f.ago(9*time.Hour + 30*time.Minute)
	f.stamps["100008"] = f.ago(15 * time.Second)
	f.stamps["100009"] = f.ago(5 * time.Minute)
	f.stamps["100010"] = f.ago(time.Hour)
	if cycle := f.observe(); cycle.Judged != len(cases) {
		t.Fatalf("judged = %d, want %d — no bucket may skip a judgement", cycle.Judged, len(cases))
	}
	for _, c := range cases {
		if got := f.bucketsOf(c.symbol); fmt.Sprint(got) != "["+c.want+"]" {
			t.Errorf("%s (%q) buckets = %v, want [%s]", c.symbol, f.stamps[c.symbol], got, c.want)
		}
	}
	// 받은 시각이 비어 사이클 대체로 잰 나이 — 0 시각으로 재면 거대 음수가 되어 ahead 로 떨어짐(P3-4).
	for _, l := range f.ageLines() {
		if l["bucket"] == "le_15s" {
			if age, _ := l["age_seconds"].(float64); age < 0 || age > 15 {
				t.Errorf("le_15s age_seconds = %v, want 0..15 measured from the observation's receipt", l["age_seconds"])
			}
		}
	}
}

// --- 2.14 OFF = 오늘 동작 ------------------------------------------------------------------

type a163Snapshot struct {
	Cycles  []string
	Places  []string
	Ledger  []string
	Settled int
}

func a163Run(t *testing.T, on bool) (a163Snapshot, *a163Fixture) {
	t.Helper()
	f := a163Setup(t, false, func(o *engine.ExitObserverOptions) {
		if on {
			o.SourceAgeReport = true
		}
	})
	var ids []string
	for _, e := range []struct{ symbol, price, stamp string }{
		{"005930", "70100", ""},                               // missing
		{"000660", "120100", "garbage"},                       // unparseable
		{"035420", "200100", "2026-03-30T00:30:00.000+09:00"}, // 9h30m 묵음
		{"051910", "9000", "2026-03-30T00:30:00.000+09:00"},   // 손절선(9500) 아래
	} {
		var p journal.Position
		if e.symbol == "051910" {
			p = f.entry(e.symbol, "3", "10000", "9500", "10000")
		} else if e.symbol == "000660" {
			p = f.entry(e.symbol, "4", "120000", "118000", "120000")
		} else if e.symbol == "035420" {
			p = f.entry(e.symbol, "5", "200000", "190000", "200000")
		} else {
			p = f.entry(e.symbol, "10", "70000", "68000", "70000")
		}
		ids = append(ids, p.ID)
		var price float64
		fmt.Sscan(e.price, &price)
		f.quote(e.symbol, price)
		f.stamps[e.symbol] = e.stamp
	}
	var snap a163Snapshot
	for i := 0; i < 3; i++ {
		c := f.cycle()
		snap.Cycles = append(snap.Cycles, fmt.Sprintf("%+v", c))
		snap.Settled += c.Unobserved
	}
	for _, p := range f.submit.places {
		snap.Places = append(snap.Places, fmt.Sprintf("%s/%s/%s", p.Intent.Symbol, decimalText(p.Intent.Quantity), decimalText(p.Intent.Price)))
	}
	sort.Strings(ids)
	for _, id := range ids {
		st := f.state(id)
		observedAt, source := "", ""
		if st.Snapshot.Snapshot != nil {
			observedAt, source = st.Snapshot.Snapshot.ObservedAt, st.Snapshot.Snapshot.ObservationSource
		}
		snap.Ledger = append(snap.Ledger, fmt.Sprintf("%s|%s|%s|%s|%s", st.Baseline, st.RatchetLevel, st.PendingAction, observedAt, source))
	}
	return snap, f
}

func TestA163TheToggleOffWritesNothingAndJudgesExactlyAsOn(t *testing.T) {
	off, offF := a163Run(t, false)
	on, onF := a163Run(t, true)
	if len(offF.ageLines()) != 0 || len(offF.missingAlerts()) != 0 || len(offF.observer.SourceAgeLatchForTest()) != 0 {
		t.Fatalf("toggle off: lines=%d alerts=%d latch=%v, want all zero (OFF = today's behaviour, safety 3)",
			len(offF.ageLines()), len(offF.missingAlerts()), offF.observer.SourceAgeLatchForTest())
	}
	if len(onF.ageLines()) == 0 || len(onF.missingAlerts()) != 2 {
		t.Fatalf("toggle on: lines=%d alerts=%d, want lines and the two source_missing alerts", len(onF.ageLines()), len(onF.missingAlerts()))
	}
	if fmt.Sprint(off) != fmt.Sprint(on) {
		t.Fatalf("judgement differs by toggle:\n off=%+v\n  on=%+v", off, on)
	}
	if len(off.Places) != 1 || !strings.HasPrefix(off.Places[0], "051910/3/") {
		t.Fatalf("places = %v, want the one stop below its line", off.Places)
	}
	if (engine.ExitObserverOptions{}).SourceAgeReport {
		t.Fatal("the zero ExitObserverOptions must be report-off")
	}
}

// --- 2.15 배선: 로드된 설정이 호출자 값을 이김, 기동 1회 읽기 -------------------------------------

func TestA163TheProductionAssemblyReadsTheLoadedToggle(t *testing.T) {
	for _, loaded := range []bool{true, false} {
		t.Run(fmt.Sprintf("loaded=%v", loaded), func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, smallLiveGate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, nil)
			srv, _ := interlockServer(t, "123-45")
			eng, err := openProtectedGateEngine(t, dir, srv, nil)
			if err != nil {
				t.Fatalf("production assembly: %v", err)
			}
			eng.Config.Engine.ExitSourceAgeReport = loaded
			observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel(), SourceAgeReport: !loaded})
			if err != nil {
				t.Fatalf("ExitObserver: %v", err)
			}
			if got := observer.OptionsForTest().SourceAgeReport; got != loaded {
				t.Fatalf("SourceAgeReport = %v, want the loaded %v — the caller's value must not win", got, loaded)
			}
			// 기동 뒤 설정이 바뀌어도(파일 · 메모리) 만들어진 관측자에는 닿지 않음 — 다음 기동부터.
			eng.Config.Engine.ExitSourceAgeReport = !loaded
			writeGateConfig(t, dir, smallLiveGate())
			if got := observer.OptionsForTest().SourceAgeReport; got != loaded {
				t.Fatalf("SourceAgeReport moved to %v after startup, want %v until the next start", got, loaded)
			}
		})
	}
}
