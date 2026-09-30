package engine_test

// a090 — 관측되지 않은 보유 포지션은 세어지고 알려짐(tasks §2 RED).
//
// 무엇을 재나: 보유·exit 대상 포지션 중 그 주기에 판정에 닿지 못한 것이 포지션 단위로 세어지고(주기 결과 · normal 로그),
// 마지막 판정 뒤 60초를 넘으면 critical 알림이 알림기의 기록 입구로 한 번 적재되고, 운영 모드가 ENTRY_BLOCKED 로 조여지며
// 그 공지도 적재만 되는지. 단언은 알리미 스파이가 아니라 **원장의 outbox 행과 운영 모드 행**에 둠(tasks 2.3).
//
// 알림기는 진짜 obs.Notifier(멈춘 publisher) 를 기록 전용(obs.RecordOnly)으로 꽂음 — 생산 조립(Context.ExitObserver)과 같은 모양.

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// --- 고정물 ---------------------------------------------------------------------

// a090Buffer 는 경합 없이 읽을 수 있는 로그 싱크임.
type a090Buffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *a090Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *a090Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// a090CountingAlerts 는 기록 전용 알림기를 감싸 관측 루프의 Notify 호출 수를 셈 — a090 경로는 Notify 가 아니라 기록 입구를 써야 함.
type a090CountingAlerts struct {
	ro       obs.RecordOnly
	mu       sync.Mutex
	notifies int
	records  map[obs.EventType]int
	events   []obs.Event // 입구에 넘어간 사건 전부(성공 · 실패 불문) — 실패한 기록은 원장 행이 없어 행 검사로는 안 보임
}

func (a *a090CountingAlerts) Notify(ctx context.Context, e obs.Event) error {
	a.mu.Lock()
	a.notifies++
	a.mu.Unlock()
	return a.ro.Notify(ctx, e)
}

func (a *a090CountingAlerts) RecordCritical(ctx context.Context, e obs.Event, remindAfter time.Duration) error {
	a.mu.Lock()
	if a.records == nil {
		a.records = map[obs.EventType]int{}
	}
	a.records[e.Type]++
	a.events = append(a.events, e)
	a.mu.Unlock()
	return a.ro.N.RecordCritical(ctx, e, remindAfter)
}

func (a *a090CountingAlerts) seen() []obs.Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]obs.Event(nil), a.events...)
}

func (a *a090CountingAlerts) recorded(t obs.EventType) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.records[t]
}

func (a *a090CountingAlerts) count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.notifies
}

type a090Fixture struct {
	*exitHarness
	pub  *a092StuckPublisher
	n    *obs.Notifier
	logs *a090Buffer
}

// a090Setup 은 exit 하네스에 진짜 알림기(기록 전용)와 a090 전용 로거를 꽂음.
func a090Setup(t *testing.T, mutate func(*engine.ExitObserverOptions)) *a090Fixture {
	t.Helper()
	f := &a090Fixture{logs: &a090Buffer{}}
	f.pub = &a092StuckPublisher{release: make(chan struct{})}
	t.Cleanup(func() { close(f.pub.release) })
	f.exitHarness = newExitHarness(t, func(o *engine.ExitObserverOptions) {
		f.n = &obs.Notifier{
			Publisher: f.pub, Journal: o.Journal, Gate: o.Retrier.Gate, AccountRef: o.AccountRef,
			Clock: clock.System(), Attempts: 1, RetryDelay: time.Millisecond,
		}
		ro := obs.RecordOnly{N: f.n}
		o.Alerts = ro
		o.Announcer = ro
		o.Critical = f.n
		o.UnobservedLog = obs.NewLogger(obs.LogOptions{Writer: f.logs, JSON: true, Clock: o.Clock})
		if mutate != nil {
			mutate(o)
		}
	})
	return f
}

// cycles 는 5초 간격 주기를 n 번 돌림(주기 전에 시계를 전진).
func (f *a090Fixture) cycles(n int) engine.ExitCycle {
	f.t.Helper()
	var c engine.ExitCycle
	for i := 0; i < n; i++ {
		f.clk.Advance(5 * time.Second)
		c = f.observe()
	}
	return c
}

func (f *a090Fixture) outageRows() []journal.Alert {
	f.t.Helper()
	return a092PendingOfType(f.t, f.journal, string(obs.EventExitObservationOutage))
}

// noticeRows 는 a090 강화 공지 행만 돌려줌(key operating_mode:<mode>:<전이 id> — 계좌 없음). 입구의 동기 경로 공지(key 에 계좌)는 뺌.
func (f *a090Fixture) noticeRows() []journal.Alert {
	f.t.Helper()
	var out []journal.Alert
	for _, row := range a092PendingOfType(f.t, f.journal, string(obs.EventOperatingMode)) {
		if strings.HasPrefix(row.EventKey, "operating_mode:"+journal.ModeEntryBlocked+":") {
			out = append(out, row)
		}
	}
	return out
}

// a090Tightenings 는 관측 두절 트리거로 커밋된 모드 행 수를 셈 — 입구의 전달 실패 승격(다른 트리거)은 뺌.
func (f *a090Fixture) a090Tightenings() []journal.OperatingModeRecord {
	f.t.Helper()
	var out []journal.OperatingModeRecord
	for _, m := range f.modeRows() {
		if m.Cause == journal.ModeTriggerExitObservationOutage {
			out = append(out, m)
		}
	}
	return out
}

func (f *a090Fixture) modeRows() []journal.OperatingModeRecord {
	f.t.Helper()
	rows, err := f.journal.OperatingModeHistory(context.Background(), exitAccount)
	if err != nil {
		f.t.Fatalf("OperatingModeHistory: %v", err)
	}
	return rows
}

// lines 는 a090 전용 로거가 낸 exit.position_unobserved 줄을 파싱해 돌려줌.
func (f *a090Fixture) lines() []map[string]any {
	f.t.Helper()
	var out []map[string]any
	for _, raw := range strings.Split(strings.TrimSpace(f.logs.String()), "\n") {
		if raw == "" {
			continue
		}
		var line map[string]any
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			f.t.Fatalf("log line is not JSON: %q", raw)
		}
		if line["event"] == string(obs.EventExitPositionUnobserved) {
			out = append(out, line)
		}
	}
	return out
}

// relax 는 운영자의 완화(사람 결정)를 원장에 씀.
func (f *a090Fixture) relax() {
	f.t.Helper()
	if _, _, err := f.journal.TransitionOperatingMode(context.Background(), journal.TransitionModeRequest{
		AccountRef: exitAccount, Mode: journal.ModeNormal, Cause: "operator checked", Actor: journal.ModeActorOperator,
		Approval: "OPS-a090", Auditor: a090NopAuditor{},
	}); err != nil {
		f.t.Fatalf("relaxing: %v", err)
	}
}

type a090NopAuditor struct{}

func (a090NopAuditor) RecordAction(string, string, string, string) error { return nil }

func a090Payload(t *testing.T, a journal.Alert) map[string]any {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal([]byte(a.Payload), &fields); err != nil {
		t.Fatalf("payload of %s is not JSON: %q", a.EventKey, a.Payload)
	}
	return fields
}

// a090Raw 는 원장 파일에 직접 붙어 실패 주입 트리거를 걸고 뗌.
func a090Raw(t *testing.T, h *exitHarness) *sql.DB {
	t.Helper()
	return openRaw(t, h)
}

func a090Exec(t *testing.T, db *sql.DB, stmt string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), stmt); err != nil {
		t.Fatalf("%s: %v", stmt, err)
	}
}

// a090Pair 는 판정되는 형제 A(000001)와 대상 B(005930) 두 포지션을 만들고 둘 다 한 번 판정시킴.
func (f *a090Fixture) a090Pair() (sibling, target journal.Position) {
	f.t.Helper()
	sibling = f.entry("000001", "10", "70000", "68000", "70000")
	target = f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("000001", 70100)
	f.quote("005930", 70100)
	if c := f.observe(); c.Judged != 2 || c.Err != nil {
		f.t.Fatalf("first cycle = %+v, want both judged", c)
	}
	return sibling, target
}

// --- R1 · R2 ------------------------------------------------------------------

func TestA090R1AZeroQuoteSiblingIsCountedAndLoggedAtNormalGrade(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	f.quote("005930", 0)
	f.clk.Advance(5 * time.Second)
	cycle := f.observe()
	if cycle.Err != nil || cycle.Judged != 1 {
		t.Fatalf("cycle = %+v, want the other symbol judged and no error", cycle)
	}
	if cycle.Unobserved != 1 {
		t.Fatalf("Unobserved = %d, want 1", cycle.Unobserved)
	}
	lines := f.lines()
	if len(lines) != 1 {
		t.Fatalf("unobserved log lines = %d (%v), want the streak's start line", len(lines), lines)
	}
	if lines[0]["position_id"] != target.ID || lines[0]["cause"] != "no_quote" ||
		lines[0]["severity"] != string(obs.SeverityNormal) {
		t.Fatalf("start line = %v, want position %s cause no_quote at normal grade", lines[0], target.ID)
	}
	if rows := f.outageRows(); len(rows) != 0 {
		t.Fatalf("below the threshold an outage row was recorded: %+v", rows)
	}
}

func TestA090R2AnAbsentSiblingIsCountedLikeAZeroQuote(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	delete(f.prices.last, "005930")
	f.clk.Advance(5 * time.Second)
	cycle := f.observe()
	if cycle.Err != nil || cycle.Judged != 1 || cycle.Unobserved != 1 {
		t.Fatalf("cycle = %+v, want judged 1 · unobserved 1", cycle)
	}
	lines := f.lines()
	if len(lines) != 1 || lines[0]["position_id"] != target.ID || lines[0]["cause"] != "no_quote" {
		t.Fatalf("start lines = %v", lines)
	}
	// 매 주기 반복하지 않음.
	f.cycles(2)
	if got := len(f.lines()); got != 1 {
		t.Fatalf("unobserved lines after three cycles = %d, want one start line", got)
	}
}

// --- R3 · R3a -------------------------------------------------------------------

func TestA090R3APositionUnjudgedPastTheThresholdIsRecordedOnceAndTightens(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	delete(f.prices.last, "005930")

	// 판정 뒤 55초: 임계 아래.
	f.cycles(11)
	if rows := f.outageRows(); len(rows) != 0 {
		t.Fatalf("at 55s an outage row exists: %+v", rows)
	}
	if f.mode() != journal.ModeNormal {
		t.Fatalf("mode = %s at 55s, want NORMAL", f.mode())
	}
	// 60초.
	cycle := f.cycles(1)
	rows := f.outageRows()
	if len(rows) != 1 {
		t.Fatalf("outage rows at 60s = %d, want 1", len(rows))
	}
	row := rows[0]
	parts := strings.Split(row.EventKey, "|")
	if len(parts) != 3 || parts[0] != string(obs.EventExitObservationOutage) || parts[1] != target.ID || parts[2] == "" {
		t.Fatalf("key = %q, want type|position|streak", row.EventKey)
	}
	fields := a090Payload(t, row)
	if fields["position_id"] != target.ID || fields["symbol"] != "005930" || fields["cause"] != "no_quote" {
		t.Fatalf("fields = %v", fields)
	}
	if secs, _ := fields["unobserved_seconds"].(float64); secs < 60 {
		t.Fatalf("unobserved_seconds = %v, want >= 60", fields["unobserved_seconds"])
	}
	if !strings.Contains(row.Body, target.ID) {
		t.Fatalf("body does not name the position: %q", row.Body)
	}
	if row.ClaimedBy != "" {
		t.Fatalf("the exit record holds a lease by %q", row.ClaimedBy)
	}
	// R3a: 같은 순회 뒤 모드 ENTRY_BLOCKED · 공지는 outbox 행(계좌 없는 key).
	if f.mode() != journal.ModeEntryBlocked || !cycle.Escalated {
		t.Fatalf("mode = %s escalated=%v, want ENTRY_BLOCKED in the same cycle", f.mode(), cycle.Escalated)
	}
	modes := f.modeRows()
	last := modes[len(modes)-1]
	if last.Cause != journal.ModeTriggerExitObservationOutage {
		t.Fatalf("mode cause = %s", last.Cause)
	}
	notices := f.noticeRows()
	if len(notices) != 1 || notices[0].EventKey != "operating_mode:"+journal.ModeEntryBlocked+":"+last.ID {
		t.Fatalf("notice rows = %+v, want key operating_mode:ENTRY_BLOCKED:<transition id>", notices)
	}
	if f.pub.count() != 0 {
		t.Fatalf("publisher calls = %d, want 0 — the exit goroutine records and does not send", f.pub.count())
	}

	// 같은 연속의 다음 주기: 반복 없음.
	f.cycles(2)
	if len(f.outageRows()) != 1 || len(f.noticeRows()) != 1 || len(f.modeRows()) != len(modes) {
		t.Fatalf("the same streak re-recorded: outage=%d notices=%d modes=%d",
			len(f.outageRows()), len(f.noticeRows()), len(f.modeRows()))
	}
}

// --- R3b 기점 = 마지막 판정 ------------------------------------------------------

func TestA090R3bTheThresholdRunsFromTheLastJudgementAcrossYieldsAndBlackouts(t *testing.T) {
	for _, variant := range []string{"yield", "blackout"} {
		t.Run(variant, func(t *testing.T) {
			f := a090Setup(t, nil)
			_, target := f.a090Pair()
			// 판정 뒤 55초 동안 포지션 단위 처리가 없는 주기들.
			switch variant {
			case "yield":
				f.slo.behind = true
				f.cycles(11)
				f.slo.behind = false
			case "blackout":
				f.prices.err = errors.New("down")
				f.cycles(11)
				f.prices.err = nil
			}
			if rows := f.outageRows(); len(rows) != 0 {
				t.Fatalf("a position outage row before any partial cycle: %+v", rows)
			}
			// 첫 부분 응답 주기(60초): 첫 미스부터 60초를 더 기다리지 않음.
			delete(f.prices.last, "005930")
			f.cycles(1)
			rows := f.outageRows()
			if len(rows) != 1 || !strings.Contains(rows[0].EventKey, target.ID) {
				t.Fatalf("outage rows = %+v, want the position alerted at 60s from its last judgement", rows)
			}
		})
	}
}

// --- R3c 순서 ------------------------------------------------------------------

func TestA090R3cAStopIsSubmittedBeforeTheAlertAndTheNextCycleIsNotDelayed(t *testing.T) {
	var alerts *a090CountingAlerts
	var spy *a111SubmitSpy
	f := a090Setup(t, func(o *engine.ExitObserverOptions) {
		ro := o.Alerts.(obs.RecordOnly)
		alerts = &a090CountingAlerts{ro: ro}
		o.Alerts = alerts
		o.Critical = alerts
		spy = &a111SubmitSpy{delegate: o.Submit}
		o.Submit = spy
	})
	a := f.entry("000001", "10", "70000", "68000", "70000")
	b := f.entry("005930", "10", "70000", "68000", "70000")
	c := f.entry("035720", "10", "70000", "68000", "70000")
	for _, s := range []string{"000001", "005930", "035720"} {
		f.quote(s, 70100)
	}
	f.observe()
	delete(f.prices.last, "000001") // A 미관측
	f.cycles(11)

	var outageAtPlace, modeAtPlace = -1, ""
	spy.beforePlace = func(execgw.PlaceRequest) {
		if outageAtPlace < 0 {
			outageAtPlace = len(f.outageRows())
			modeAtPlace = f.mode()
		}
	}
	f.quote("005930", 67900) // B 손절 조건
	f.clk.Advance(5 * time.Second)
	a092ObserveWithin(t, f.exitHarness, 5*time.Second)
	if spy.places != 1 {
		t.Fatalf("places = %d, want B's liquidation", spy.places)
	}
	if outageAtPlace != 0 || modeAtPlace != journal.ModeNormal {
		t.Fatalf("at B's submission: outage rows=%d mode=%s — the alert/tightening ran before the stop", outageAtPlace, modeAtPlace)
	}
	if rows := f.outageRows(); len(rows) != 1 || !strings.Contains(rows[0].EventKey, a.ID) {
		t.Fatalf("A's outage row after the cycle = %+v", rows)
	}
	if f.mode() != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s after the cycle", f.mode())
	}
	// 강화 뒤 다음 주기: C 가 곧바로 판정·제출됨(전송이 멈춰 있어도).
	f.quote("035720", 67900)
	f.clk.Advance(5 * time.Second)
	a092ObserveWithin(t, f.exitHarness, 2*time.Second)
	if spy.places != 2 {
		t.Fatalf("places = %d after the next cycle, want C's liquidation too", spy.places)
	}
	if alerts.count() != 0 {
		t.Fatalf("the loop called Notify %d times — a090's alert and notice must use the record entrance", alerts.count())
	}
	if f.pub.count() != 0 {
		t.Fatalf("publisher calls = %d", f.pub.count())
	}
	_ = b
	_ = c
}

// --- R3d · R3g ① 적재 실패 -------------------------------------------------------

func TestA090R3dARecordFailureLatchesTheGateAtTheEntranceAndIsRetried(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `CREATE TRIGGER a090_fail_outage BEFORE INSERT ON alert_outbox
		WHEN NEW.event_type = 'exit.observation_outage' BEGIN SELECT RAISE(ABORT, 'disk full acct-exit'); END`)
	delete(f.prices.last, "005930")
	cycle := f.cycles(12)
	if cycle.Err != nil || cycle.Judged != 1 {
		t.Fatalf("the loop stopped on a record failure: %+v", cycle)
	}
	if rej := f.gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonAlertUndelivered {
		t.Fatalf("entry check = %v, want the entrance's producer latch (%s)", rej, execgw.ReasonAlertUndelivered)
	}
	if len(f.outageRows()) != 0 {
		t.Fatal("a row exists although the insert failed")
	}
	// 알림이 적재되지 않은 연속은 a090 이 모드를 조이지 않음(알림 없는 조임 금지). 입구의 승격은 별도 트리거.
	for _, m := range f.modeRows() {
		if m.Cause == journal.ModeTriggerExitObservationOutage {
			t.Fatalf("a090 tightened without a recorded alert: %+v", m)
		}
	}
	// 저장소 복구 → 다음 주기 재시도 성공(같은 연속, 한 행).
	a090Exec(t, db, `DROP TRIGGER a090_fail_outage`)
	f.cycles(1)
	rows := f.outageRows()
	if len(rows) != 1 || !strings.Contains(rows[0].EventKey, target.ID) {
		t.Fatalf("outage rows after recovery = %+v, want one", rows)
	}
	f.cycles(1)
	if len(f.outageRows()) != 1 {
		t.Fatal("the recovered streak was recorded twice")
	}
}

// --- R3e 에피소드 ----------------------------------------------------------------

func TestA090R3eANewStreakIsANewRowAndTheSameStreakIsOne(t *testing.T) {
	var alerts *a090CountingAlerts
	f := a090Setup(t, func(o *engine.ExitObserverOptions) {
		alerts = &a090CountingAlerts{ro: o.Alerts.(obs.RecordOnly)}
		o.Alerts = alerts
		o.Critical = alerts
	})
	_, target := f.a090Pair()
	delete(f.prices.last, "005930")
	f.cycles(12)
	first := f.outageRows()
	if len(first) != 1 {
		t.Fatalf("first streak rows = %d", len(first))
	}
	// 같은 연속은 입구에 한 번만 감 — 원장의 key 중복 제거에 기대지 않음.
	f.cycles(3)
	if got := alerts.recorded(obs.EventExitObservationOutage); got != 1 {
		t.Fatalf("the same streak reached the record entrance %d times, want 1", got)
	}
	f.relax()
	f.cycles(2) // R3g ④: 같은 연속에서 재강화 없음
	if f.mode() != journal.ModeNormal {
		t.Fatalf("the same streak re-tightened after the operator relaxed: mode=%s", f.mode())
	}
	// 판정 → 연속 해제.
	f.quote("005930", 70100)
	f.cycles(1)
	// 새 연속.
	delete(f.prices.last, "005930")
	f.cycles(12)
	rows := f.outageRows()
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want a new row for the new streak", len(rows))
	}
	if rows[0].EventKey == rows[1].EventKey || !strings.Contains(rows[1].EventKey, target.ID) {
		t.Fatalf("keys = %q · %q", rows[0].EventKey, rows[1].EventKey)
	}
	if got := alerts.recorded(obs.EventExitObservationOutage); got != 2 {
		t.Fatalf("record entrance calls = %d over two streaks, want 2", got)
	}
	if f.mode() != journal.ModeEntryBlocked {
		t.Fatalf("the second streak did not tighten: mode=%s", f.mode())
	}
	notices := f.noticeRows()
	if len(notices) != 2 || notices[0].EventKey == notices[1].EventKey {
		t.Fatalf("notice rows = %+v, want one per transition", notices)
	}
}

// --- R3f 단조 경과 ---------------------------------------------------------------

// a090AnchorClock 은 벽시계와 경과를 따로 움직이는 **앵커 인식** 시계임. Now 가 돌려준 값마다 그 순간의 경과 누적을 기억해,
// 그 값을 앵커로 넘기면 Since 가 「그 뒤 흐른 경과」를 냄(발급 직후 0). a111 분리 픽스처는 역행 뒤 만든 앵커를 틀리게 잰다
// (design D3 5판 R3-7).
type a090AnchorClock struct {
	mu     sync.Mutex
	wall   time.Time
	mono   time.Duration
	issued map[int64]time.Duration
}

func newA090AnchorClock(at time.Time) *a090AnchorClock {
	return &a090AnchorClock{wall: at, issued: map[int64]time.Duration{}}
}

func (c *a090AnchorClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.issued[c.wall.UnixNano()] = c.mono
	return c.wall
}

func (c *a090AnchorClock) Since(at time.Time) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	if m, ok := c.issued[at.UnixNano()]; ok {
		return c.mono - m
	}
	return c.wall.Sub(at)
}

func (c *a090AnchorClock) Sleep(ctx context.Context, _ time.Duration) error { return ctx.Err() }

func (c *a090AnchorClock) advance(d time.Duration) {
	c.mu.Lock()
	c.wall = c.wall.Add(d)
	c.mono += d
	c.mu.Unlock()
}

func (c *a090AnchorClock) rewindWall(d time.Duration) {
	c.mu.Lock()
	c.wall = c.wall.Add(-d)
	c.mu.Unlock()
}

func TestA090R3fElapsedIsMonotonicAcrossAWallClockRollback(t *testing.T) {
	var ck *a090AnchorClock
	f := a090Setup(t, func(o *engine.ExitObserverOptions) {
		ck = newA090AnchorClock(exitNow)
		o.Clock = ck
	})
	_, target := f.a090Pair()
	step := func(n int) {
		for i := 0; i < n; i++ {
			ck.advance(5 * time.Second)
			f.observe()
		}
	}
	delete(f.prices.last, "005930")
	// 연속 1: 기점 W1(경과 0). 30초 뒤 벽시계를 47초 되감고 30초 더 → 경과 60, 벽시계 차이는 13초.
	// 되감는 폭은 5초 격자에서 벗어나게(47초) — 이 픽스처는 벽시계 값으로 앵커를 찾으므로 역행 뒤 옛 값과 겹치면 안 됨.
	step(6)
	ck.rewindWall(47 * time.Second)
	step(5)
	if rows := f.outageRows(); len(rows) != 0 {
		t.Fatalf("at elapsed 55s: %+v", rows)
	}
	step(1)
	rows := f.outageRows()
	if len(rows) != 1 {
		t.Fatalf("a wall-clock rollback lengthened the window: rows=%d at elapsed 60s", len(rows))
	}
	// 연속 해제(판정) — 그 판정 순간의 벽시계를 W2 로 기억.
	f.quote("005930", 70100)
	step(1)
	delete(f.prices.last, "005930")
	// 역행 **뒤** 만든 앵커: 벽시계를 되감은 직후 판정 → 발급 직후 경과 0, 이후 정확한 진행.
	ck.rewindWall(22 * time.Second) // 격자 밖(앞 역행과도 다른 나머지)
	f.quote("005930", 70100)
	f.observe() // 판정(기점 = 역행 뒤 앵커)
	delete(f.prices.last, "005930")
	f.observe() // 경과 0 인 미관측 — 연속 시작만
	if got := len(f.outageRows()); got != 1 {
		t.Fatalf("an anchor made after the rollback measured elapsed > 0 at creation: rows=%d", got)
	}
	step(11)
	if got := len(f.outageRows()); got != 1 {
		t.Fatalf("at 55s after the post-rollback anchor: rows=%d", got)
	}
	step(1)
	rows = f.outageRows()
	if len(rows) != 2 || rows[0].EventKey == rows[1].EventKey || !strings.Contains(rows[1].EventKey, target.ID) {
		t.Fatalf("rows = %+v, want a second, distinct streak row at 60s", rows)
	}
}

// --- R3g 실패 전이 -----------------------------------------------------------------

func TestA090R3gATighteningCommitFailureIsRetriedNextCycle(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `CREATE TRIGGER a090_fail_mode BEFORE INSERT ON operating_modes
		BEGIN SELECT RAISE(ABORT, 'mode store acct-exit down'); END`)
	delete(f.prices.last, "005930")
	f.cycles(12)
	if len(f.outageRows()) != 1 {
		t.Fatal("the alert was not recorded")
	}
	if f.mode() != journal.ModeNormal {
		t.Fatalf("mode = %s although the commit failed", f.mode())
	}
	// ⓐ 모드 기록 실패에서는 공지가 입구에 가지 않음 — 공지 행 없음.
	if len(f.noticeRows()) != 0 {
		t.Fatalf("a notice row without a committed transition: %+v", f.noticeRows())
	}
	a090Exec(t, db, `DROP TRIGGER a090_fail_mode`)
	f.cycles(1)
	if f.mode() != journal.ModeEntryBlocked {
		t.Fatalf("the failed tightening was not retried: mode=%s", f.mode())
	}
	if len(f.noticeRows()) != 1 || len(f.outageRows()) != 1 {
		t.Fatalf("after the retry: notices=%d outage=%d", len(f.noticeRows()), len(f.outageRows()))
	}
}

func TestA090R3gACommittedTighteningWhoseNoticeFailedIsAnnouncedExactlyOnce(t *testing.T) {
	f := a090Setup(t, nil)
	sibling, _ := f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `CREATE TRIGGER a090_fail_notice BEFORE INSERT ON alert_outbox
		WHEN NEW.event_type = 'engine.operating_mode' BEGIN SELECT RAISE(ABORT, 'notice store acct-exit down'); END`)
	delete(f.prices.last, "005930")
	f.cycles(12)
	if f.mode() != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s — the transition must be committed", f.mode())
	}
	if len(f.a090Tightenings()) != 1 {
		t.Fatalf("a090 tightenings = %d, want 1", len(f.a090Tightenings()))
	}
	if len(f.noticeRows()) != 0 {
		t.Fatal("the notice was recorded although the insert failed")
	}
	// 운영자 완화 → 같은 연속에서 a090 재강화 없음(공지 재시도 실패로 입구가 자기 트리거로 승격하는 것은 별개).
	f.relax()
	f.cycles(2)
	if got := len(f.a090Tightenings()); got != 1 {
		t.Fatalf("the transition was retried: a090 tightenings = %d, want 1", got)
	}
	// 연속을 끝내고(판정), B1·B4 주기를 끼운 뒤 포지션이 사라짐(B3 포함) — 공지는 여전히 재시도 대상.
	f.quote("005930", 70100)
	f.cycles(1)
	a090Exec(t, db, `DROP TRIGGER a090_fail_notice`)
	f.slo.behind = true
	f.cycles(1)
	f.slo.behind = false
	f.prices.err = errors.New("down")
	f.cycles(1)
	f.prices.err = nil
	if got := len(f.noticeRows()); got != 0 {
		t.Fatalf("a yield or blackout cycle retried the notice: rows=%d", got)
	}
	f.cycles(1)
	notices := f.noticeRows()
	if len(notices) != 1 {
		t.Fatalf("notice rows = %d, want exactly one after recovery", len(notices))
	}
	blocked := f.a090Tightenings()[0]
	if notices[0].EventKey != "operating_mode:"+journal.ModeEntryBlocked+":"+blocked.ID {
		t.Fatalf("notice key = %q, want the committed transition %s", notices[0].EventKey, blocked.ID)
	}
	f.cycles(2)
	if len(f.noticeRows()) != 1 {
		t.Fatal("the notice was recorded twice")
	}
	_ = sibling
}

func TestA090R3gAFailureAfterTheOperatorClearedLatchesAgain(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `CREATE TRIGGER a090_fail_outage BEFORE INSERT ON alert_outbox
		WHEN NEW.event_type = 'exit.observation_outage' BEGIN SELECT RAISE(ABORT, 'disk full'); END`)
	delete(f.prices.last, "005930")
	f.cycles(12)
	if rej := f.gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonAlertUndelivered {
		t.Fatalf("first failure did not latch: %v", rej)
	}
	// ⑤ 입구는 기록 오류가 돌아온 뒤 잠금 — 그 잠금을 운영자가 풀면(해제) 다음 실패는 새 증거로 다시 잠금(⑥).
	f.gate.Clear(execgw.ReasonAlertUndelivered)
	if rej := f.gate.CheckEntry(); rej != nil && rej.Reason == execgw.ReasonAlertUndelivered {
		t.Fatal("clear did not clear")
	}
	f.cycles(1)
	if rej := f.gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonAlertUndelivered {
		t.Fatalf("a new failure after the clear did not latch again: %v", rej)
	}
}

// --- R4 · R5 · R6 ----------------------------------------------------------------

func TestA090R4AJudgementEndsTheStreakWithOneReleaseLine(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	delete(f.prices.last, "005930")
	f.cycles(3)
	f.quote("005930", 70100)
	cycle := f.cycles(1)
	if cycle.Unobserved != 0 || cycle.Judged != 2 {
		t.Fatalf("cycle = %+v", cycle)
	}
	lines := f.lines()
	if len(lines) != 2 {
		t.Fatalf("lines = %v, want start + release", lines)
	}
	release := lines[1]
	if release["position_id"] != target.ID {
		t.Fatalf("release = %v", release)
	}
	if secs, ok := release["unobserved_seconds"].(float64); !ok || secs < 15 {
		t.Fatalf("release unobserved_seconds = %v", release["unobserved_seconds"])
	}
	// 래치 해제: 새 연속은 새 시작 줄.
	delete(f.prices.last, "005930")
	f.cycles(1)
	if got := len(f.lines()); got != 3 {
		t.Fatalf("lines = %d, want a new start line", got)
	}
}

func TestA090R5WhenNoSymbolAnswersOnlyTheAccountLadderMoves(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	f.prices.err = errors.New("down")
	var last engine.ExitCycle
	for i := 0; i < 13; i++ {
		last = f.cycles(1)
		if last.Unobserved != 0 {
			t.Fatalf("a blackout cycle counted positions: %+v", last)
		}
	}
	for _, row := range f.outageRows() {
		if strings.Count(row.EventKey, "|") == 2 {
			t.Fatalf("a position row in a blackout: %q", row.EventKey)
		}
	}
	if len(f.outageRows()) != 1 {
		t.Fatalf("account outage rows = %d, want the account ladder's one", len(f.outageRows()))
	}
	if len(f.lines()) != 0 {
		t.Fatalf("a blackout wrote position lines: %v", f.lines())
	}
}

func TestA090R6AYieldIsUnchanged(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	f.slo.behind = true
	cycle := f.cycles(1)
	if !cycle.Deferred || cycle.Unobserved != 0 || cycle.Judged != 0 {
		t.Fatalf("yield cycle = %+v", cycle)
	}
	if f.prices.calls != 1 {
		t.Fatalf("the yield read prices: calls=%d", f.prices.calls)
	}
}

// --- R7 임대 만료 --------------------------------------------------------------------

func TestA090R7AQuoteThatExpiredMidCycleIsCountedAsExpired(t *testing.T) {
	var spy *a111SubmitSpy
	f := a090Setup(t, func(o *engine.ExitObserverOptions) {
		spy = &a111SubmitSpy{delegate: o.Submit}
		o.Submit = spy
	})
	first := f.entry("000001", "10", "70000", "68000", "70000")
	later := f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("000001", 70100)
	f.quote("005930", 70100)
	f.observe()
	// 첫 포지션 제출이 16초 걸리면 뒤 포지션 시세의 사용 임대(15초)가 끝남.
	spy.beforePlace = func(execgw.PlaceRequest) { f.clk.Advance(16 * time.Second) }
	f.quote("000001", 67900)
	f.clk.Advance(5 * time.Second)
	cycle := f.observe()
	if cycle.Judged != 1 || cycle.Unobserved != 1 {
		t.Fatalf("cycle = %+v, want first judged and later unobserved", cycle)
	}
	lines := f.lines()
	if len(lines) != 1 || lines[0]["position_id"] != later.ID || lines[0]["cause"] != "quote_expired" {
		t.Fatalf("lines = %v, want later's start with cause quote_expired", lines)
	}
	if len(f.outageRows()) != 0 {
		t.Fatal("below the threshold an expired quote is not an alert")
	}
	_ = first
}

// --- R8 보유 종료 ------------------------------------------------------------------

func TestA090R8APositionThatStopsBeingHeldIsForgottenWithoutAnAlert(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	delete(f.prices.last, "005930")
	f.cycles(6)
	// 계좌에서 사라짐(외부 조정으로 0).
	ctx := context.Background()
	watermark, err := f.journal.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatalf("FillWatermark: %v", err)
	}
	if _, err := f.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: exitAccount, Market: "kr", Symbol: "005930", Kind: "EXTERNAL",
		ExpectedPrevQuantity: "10", ExpectedFillWatermark: watermark, NewQuantity: "0", NewAvgPrice: "0",
		BrokerAsOf: f.clk.Now().Format(time.RFC3339), Evidence: "sold by hand",
	}); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	cycle := f.cycles(10)
	if cycle.Unobserved != 0 || len(f.outageRows()) != 0 {
		t.Fatalf("a position that left was alerted: cycle=%+v rows=%+v", cycle, f.outageRows())
	}
	if got := f.observer.UnobservedRecordsForTest(); got != 1 {
		t.Fatalf("records held = %d, want only the sibling's — the departed position's record must be dropped", got)
	}
	_ = target
}

// --- R10 · R11 -------------------------------------------------------------------

func TestA090R10OnePriceReadPerCycle(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	delete(f.prices.last, "005930")
	before := f.prices.calls
	f.cycles(13)
	if got := f.prices.calls - before; got != 13 {
		t.Fatalf("price reads = %d over 13 cycles, want 13", got)
	}
}

func TestA090R11ARestartForgetsTheRecordAndStartsANewEpisode(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	delete(f.prices.last, "005930")
	f.cycles(12)
	if len(f.outageRows()) != 1 {
		t.Fatal("first episode not recorded")
	}
	f.relax()
	// 재시작: 같은 원장 위 새 관측자.
	opts := f.observer.OptionsForTest()
	restarted, err := engine.NewExitObserver(opts)
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	f.observer = restarted
	// 첫 처리 주기(+5초)가 새 기점 — 그로부터 60초(+65초)에 새 에피소드.
	f.cycles(12)
	if len(f.outageRows()) != 1 {
		t.Fatal("the restarted observer kept the old base — it cannot know it")
	}
	f.cycles(1)
	rows := f.outageRows()
	if len(rows) != 2 || rows[0].EventKey == rows[1].EventKey || !strings.Contains(rows[1].EventKey, target.ID) {
		t.Fatalf("rows = %+v, want a new episode 60s after the restart", rows)
	}
}

// --- R12 · R13 workingSet 탈락 ----------------------------------------------------------

func TestA090R12APositionWhoseStateCannotOpenIsCountedAndAlerted(t *testing.T) {
	for _, alone := range []bool{false, true} {
		name := "with_sibling"
		if alone {
			name = "alone_B3"
		}
		t.Run(name, func(t *testing.T) {
			f := a090Setup(t, nil)
			if !alone {
				f.entry("000001", "10", "70000", "68000", "70000")
				f.quote("000001", 70100)
			}
			db := a090Raw(t, f.exitHarness)
			a090Exec(t, db, `CREATE TRIGGER a090_no_state BEFORE INSERT ON exit_states
				WHEN (SELECT symbol FROM positions WHERE id = NEW.position_id) = '005930'
				BEGIN SELECT RAISE(ABORT, 'no stop'); END`)
			target := f.entry("005930", "10", "70000", "68000", "70000")
			f.quote("005930", 70100)
			first := f.observe()
			if first.Unobserved != 1 {
				t.Fatalf("first cycle = %+v, want the unopenable position counted", first)
			}
			f.cycles(11)
			if len(f.outageRows()) != 0 {
				t.Fatal("alerted before 60s from first sighting")
			}
			f.cycles(1)
			rows := f.outageRows()
			if len(rows) != 1 || !strings.Contains(rows[0].EventKey, target.ID) {
				t.Fatalf("rows = %+v, want the unopenable position alerted", rows)
			}
			if a090Payload(t, rows[0])["cause"] != "not_in_working_set" {
				t.Fatalf("cause = %v", a090Payload(t, rows[0])["cause"])
			}
		})
	}
}

func TestA090R13QuarantineReadAndWriteFailuresAreCounted(t *testing.T) {
	t.Run("active_quarantine_read_B14", func(t *testing.T) {
		f := a090Setup(t, nil)
		_, target := f.a090Pair()
		db := a090Raw(t, f.exitHarness)
		a090Exec(t, db, `ALTER TABLE exit_snapshot_quarantines RENAME TO a090_moved`)
		cycle := f.cycles(1)
		if cycle.Unobserved != 2 {
			t.Fatalf("cycle = %+v, want both positions counted (the read fails for every position)", cycle)
		}
		f.cycles(11)
		rows := f.outageRows()
		if len(rows) != 2 {
			t.Fatalf("rows = %d, want both positions alerted at 60s", len(rows))
		}
		_ = target
	})
	t.Run("quarantine_write_B12", func(t *testing.T) {
		f := a090Setup(t, nil)
		_, target := f.a090Pair()
		db := a090Raw(t, f.exitHarness)
		a090Exec(t, db, `CREATE TRIGGER a090_no_quarantine BEFORE INSERT ON exit_snapshot_quarantines
			BEGIN SELECT RAISE(ABORT, 'quarantine store down'); END`)
		if _, err := db.Exec(`UPDATE exit_states SET snapshot_status=NULL WHERE position_id=?`, target.ID); err != nil {
			t.Fatal(err)
		}
		cycle := f.cycles(1)
		if cycle.Unobserved != 1 || cycle.Judged != 1 {
			t.Fatalf("cycle = %+v, want the corrupt position counted", cycle)
		}
		f.cycles(11)
		if rows := f.outageRows(); len(rows) != 1 || !strings.Contains(rows[0].EventKey, target.ID) {
			t.Fatalf("rows = %+v", rows)
		}
	})
}

// --- R14 · R16 범위와 「관측됨」 ------------------------------------------------------------

func TestA090R14UnmanagedAndCompletedPositionsAreNotCounted(t *testing.T) {
	f := a090Setup(t, nil)
	sibling := f.entry("000001", "10", "70000", "68000", "70000")
	completed := f.entry("005930", "10", "70000", "68000", "70000")
	f.quote("000001", 70100)
	f.quote("005930", 70100)
	f.observe()
	db := a090Raw(t, f.exitHarness)
	if _, err := db.Exec(`UPDATE exit_states SET completed=1 WHERE position_id=?`, completed.ID); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	watermark, err := f.journal.FillWatermark(ctx, "000660")
	if err != nil {
		t.Fatalf("FillWatermark: %v", err)
	}
	if _, err := f.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: exitAccount, Market: "kr", Symbol: "000660", Kind: "EXTERNAL",
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark, NewQuantity: "7", NewAvgPrice: "120000",
		BrokerAsOf: f.clk.Now().Format(time.RFC3339), Evidence: "bought by hand",
	}); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	cycle := f.cycles(13)
	if cycle.Unobserved != 0 || len(f.outageRows()) != 0 || len(f.lines()) != 0 {
		t.Fatalf("out-of-scope positions counted: cycle=%+v rows=%+v lines=%v", cycle, f.outageRows(), f.lines())
	}
	_ = sibling
}

func TestA090R16AJudgementThatEndsAtOnceStillCountsAsObserved(t *testing.T) {
	t.Run("quarantined_refused", func(t *testing.T) {
		f := a090Setup(t, nil)
		_, target := f.a090Pair()
		if _, err := f.journal.QuarantineExitSnapshot(context.Background(), target.ID, target.InstanceSeq,
			"stored_snapshot_corrupt", "test"); err != nil {
			t.Fatalf("QuarantineExitSnapshot: %v", err)
		}
		cycle := f.cycles(13)
		if cycle.Unobserved != 0 || len(f.outageRows()) != 0 {
			t.Fatalf("a refused judgement counted as unobserved: cycle=%+v rows=%+v", cycle, f.outageRows())
		}
	})
	t.Run("selector_stamp_failure", func(t *testing.T) {
		f := a090Setup(t, nil)
		_, target := f.a090Pair()
		supersededQuarantine(t, f.exitHarness, target)
		db := a090Raw(t, f.exitHarness)
		a090Exec(t, db, `CREATE TRIGGER a090_no_stamp BEFORE UPDATE ON exit_snapshot_quarantines
			BEGIN SELECT RAISE(ABORT, 'stamp store down'); END`)
		cycle := f.cycles(13)
		if cycle.Err == nil {
			t.Fatal("the stamp failure fixture did not reach the stamp")
		}
		if cycle.Unobserved != 0 || len(f.outageRows()) != 0 {
			t.Fatalf("a judgement that failed at the stamp counted as unobserved: cycle=%+v", cycle)
		}
	})
}

// --- R17 계좌 카나리 ------------------------------------------------------------------

// 계좌 ref(exitAccount)가 a090 이 만든 로그 · 알림 행 · 공지 행(key · payload) 어디에도 없음. 실패 주입 셋에서도 a090 소유 로그에 없음.
func TestA090R17NoAccountReachesTheNewLinesAlertsOrNotices(t *testing.T) {
	check := func(t *testing.T, f *a090Fixture, where string) {
		t.Helper()
		for _, line := range f.lines() {
			encoded, _ := json.Marshal(line)
			if strings.Contains(string(encoded), exitAccount) {
				t.Fatalf("%s: the account reached an a090 log line: %s", where, encoded)
			}
		}
		for _, row := range append(f.outageRows(), f.noticeRows()...) {
			if strings.Count(row.EventKey, "|") != 2 && !strings.HasPrefix(row.EventKey, "operating_mode:"+journal.ModeEntryBlocked+":") {
				continue // a090 행만
			}
			if strings.Contains(row.EventKey, exitAccount) || strings.Contains(row.Payload, exitAccount) ||
				strings.Contains(row.Title, exitAccount) || strings.Contains(row.Body, exitAccount) {
				t.Fatalf("%s: the account reached an a090 row: %+v", where, row)
			}
		}
	}
	t.Run("success", func(t *testing.T) {
		f := a090Setup(t, nil)
		f.a090Pair()
		delete(f.prices.last, "005930")
		f.cycles(12)
		if len(f.outageRows()) != 1 || len(f.noticeRows()) != 1 || len(f.lines()) == 0 {
			t.Fatalf("the canary measured nothing: outage=%d notices=%d lines=%d",
				len(f.outageRows()), len(f.noticeRows()), len(f.lines()))
		}
		check(t, f, "success")
	})
	for _, inj := range []struct{ name, trigger string }{
		{"alert_record_failure", `CREATE TRIGGER a090_inj BEFORE INSERT ON alert_outbox
			WHEN NEW.event_type = 'exit.observation_outage' BEGIN SELECT RAISE(ABORT, 'disk full acct-exit'); END`},
		{"mode_record_failure", `CREATE TRIGGER a090_inj BEFORE INSERT ON operating_modes
			BEGIN SELECT RAISE(ABORT, 'mode store acct-exit down'); END`},
		{"notice_record_failure", `CREATE TRIGGER a090_inj BEFORE INSERT ON alert_outbox
			WHEN NEW.event_type = 'engine.operating_mode' BEGIN SELECT RAISE(ABORT, 'notice store acct-exit down'); END`},
	} {
		t.Run(inj.name, func(t *testing.T) {
			var spy *a090CountingAlerts
			f := a090Setup(t, func(o *engine.ExitObserverOptions) {
				spy = &a090CountingAlerts{ro: o.Alerts.(obs.RecordOnly)}
				o.Critical = spy
			})
			f.a090Pair()
			db := a090Raw(t, f.exitHarness)
			a090Exec(t, db, inj.trigger)
			delete(f.prices.last, "005930")
			f.cycles(12)
			// ⓐ 주입 도달 단언.
			switch inj.name {
			case "alert_record_failure":
				if len(f.outageRows()) != 0 {
					t.Fatal("the alert failure was not injected")
				}
			case "mode_record_failure":
				if f.mode() != journal.ModeNormal || len(f.outageRows()) != 1 || len(f.noticeRows()) != 0 {
					t.Fatalf("the mode failure was not injected: mode=%s outage=%d notices=%d",
						f.mode(), len(f.outageRows()), len(f.noticeRows()))
				}
			case "notice_record_failure":
				if f.mode() != journal.ModeEntryBlocked || len(f.noticeRows()) != 0 {
					t.Fatalf("the notice failure was not injected: mode=%s notices=%d", f.mode(), len(f.noticeRows()))
				}
			}
			// ⓑ ⓒ a090 소유 로그: 실패 줄이 있고, 주입 오류 문구(계좌 포함)를 옮기지 않음.
			var failure bool
			for _, line := range f.lines() {
				if _, ok := line["failure"]; ok {
					failure = true
				}
			}
			if !failure {
				t.Fatalf("no a090 failure line was written: %v", f.lines())
			}
			if strings.Contains(f.logs.String(), "down") || strings.Contains(f.logs.String(), "disk full") {
				t.Fatalf("an a090 line carried the raw error text: %s", f.logs.String())
			}
			check(t, f, inj.name)
			// 입구 인자 카나리(codex A090-I1): 실패한 기록은 행이 없으므로 입구에 넘어간 사건 자체를 봄 — 첫 시도와 재시도 모두
			// (한 주기 더 돌려 재시도 경로를 태움).
			f.cycles(1)
			events := spy.seen()
			if len(events) == 0 {
				t.Fatal("nothing reached the record entrance — the canary measured nothing")
			}
			for _, e := range events {
				encoded, _ := json.Marshal(e)
				if strings.Contains(string(encoded), exitAccount) {
					t.Fatalf("an event a090 handed to the entrance carries the account: %s", encoded)
				}
			}
			switch inj.name {
			case "alert_record_failure":
				if spy.recorded(obs.EventExitObservationOutage) < 2 {
					t.Fatalf("the failed alert was not retried: calls=%d", spy.recorded(obs.EventExitObservationOutage))
				}
			case "mode_record_failure":
				// 모드 커밋 실패 → 공지는 입구에 가지 않고 대기열도 비어 있음.
				if spy.recorded(obs.EventOperatingMode) != 0 || f.observer.PendingModeNoticesForTest() != 0 {
					t.Fatalf("a notice without a committed transition: calls=%d pending=%d",
						spy.recorded(obs.EventOperatingMode), f.observer.PendingModeNoticesForTest())
				}
			case "notice_record_failure":
				if spy.recorded(obs.EventOperatingMode) < 2 || f.observer.PendingModeNoticesForTest() != 1 {
					t.Fatalf("the failed notice was not queued and retried: calls=%d pending=%d",
						spy.recorded(obs.EventOperatingMode), f.observer.PendingModeNoticesForTest())
				}
			}
		})
	}
}

// R17 배선(엔진 쪽): Context.ExitObserver 는 UnobservedLog 를 통과시키고 Log 는 nil 로 둠.
func TestA090R17TheEngineAssemblyPassesTheDedicatedLoggerThrough(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	logger := obs.NewLogger(obs.LogOptions{Writer: &a090Buffer{}})
	observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel(), UnobservedLog: logger})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	opts := observer.OptionsForTest()
	if opts.UnobservedLog != logger {
		t.Fatal("the engine assembly dropped the dedicated unobserved logger")
	}
	if opts.Log != nil {
		t.Fatal("the observer's own logger is wired — its lines carry the account")
	}
}

// 전 종목 미응답(B4) 주기의 표시는 다음 주기로 새지 않음 — 그 사이 보유가 끝난 포지션이 미관측으로 세어지면 안 됨.
func TestA090ABlackoutCyclesMarksDoNotLeakIntoTheNextCycle(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	f.prices.err = errors.New("down")
	f.cycles(1) // B4: 두 포지션 모두 표시됐으나 처리 없음
	f.prices.err = nil
	// 두 포지션 다 사라짐 → 다음 주기는 표시 0 인 B3. 그 주기가 옛 주기의 표시를 쓰면 둘이 미관측으로 세어짐.
	ctx := context.Background()
	for _, symbol := range []string{"000001", "005930"} {
		watermark, err := f.journal.FillWatermark(ctx, symbol)
		if err != nil {
			t.Fatalf("FillWatermark: %v", err)
		}
		if _, err := f.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
			AccountRef: exitAccount, Market: "kr", Symbol: symbol, Kind: "EXTERNAL",
			ExpectedPrevQuantity: "10", ExpectedFillWatermark: watermark, NewQuantity: "0", NewAvgPrice: "0",
			BrokerAsOf: f.clk.Now().Format(time.RFC3339), Evidence: "sold by hand",
		}); err != nil {
			t.Fatalf("ApplyPositionAdjustment: %v", err)
		}
	}
	if cycle := f.cycles(1); cycle.Unobserved != 0 || cycle.Judged != 0 {
		t.Fatalf("cycle = %+v — the blackout cycle's marks leaked", cycle)
	}
	if got := f.observer.UnobservedRecordsForTest(); got != 0 {
		t.Fatalf("records held = %d after the account emptied", got)
	}
	if len(f.lines()) != 0 {
		t.Fatalf("lines = %v", f.lines())
	}
}

// 리뷰 P2-2 · P2-3: 강화는 커밋되고 공지만 실패한 뒤 저장소가 복구되고 **다음 처리 주기 전에** 운영자가 완화하면, 같은 연속은
// 다시 조이지 않음(ErrModeAnnouncementFailed 는 커밋 실패가 아님). 적재된 공지는 대기열에서 빠지고 입구를 다시 부르지 않음.
func TestA090R3gARecoveredNoticeDoesNotRetightenAfterTheOperatorRelaxes(t *testing.T) {
	var alerts *a090CountingAlerts
	f := a090Setup(t, func(o *engine.ExitObserverOptions) {
		alerts = &a090CountingAlerts{ro: o.Alerts.(obs.RecordOnly)}
		o.Alerts = alerts
		o.Critical = alerts
	})
	f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `CREATE TRIGGER a090_fail_notice BEFORE INSERT ON alert_outbox
		WHEN NEW.event_type = 'engine.operating_mode' BEGIN SELECT RAISE(ABORT, 'notice store down'); END`)
	delete(f.prices.last, "005930")
	f.cycles(12)
	if len(f.a090Tightenings()) != 1 || f.observer.PendingModeNoticesForTest() != 1 {
		t.Fatalf("setup: tightenings=%d pending=%d", len(f.a090Tightenings()), f.observer.PendingModeNoticesForTest())
	}
	a090Exec(t, db, `DROP TRIGGER a090_fail_notice`)
	f.relax()
	f.cycles(2)
	if got := len(f.a090Tightenings()); got != 1 || f.mode() != journal.ModeNormal {
		t.Fatalf("re-tightened the same streak after the operator relaxed: tightenings=%d mode=%s", got, f.mode())
	}
	if len(f.noticeRows()) != 1 || f.observer.PendingModeNoticesForTest() != 0 {
		t.Fatalf("notice rows=%d pending=%d, want the recovered notice once and an empty queue",
			len(f.noticeRows()), f.observer.PendingModeNoticesForTest())
	}
	calls := alerts.recorded(obs.EventOperatingMode)
	f.cycles(3)
	if got := alerts.recorded(obs.EventOperatingMode); got != calls {
		t.Fatalf("a settled notice kept reaching the entrance: %d → %d", calls, got)
	}
}

// 리뷰 P2-4: 작업 집합 오류(B2) 주기는 기점을 지우지 않음 — 그 시간도 마지막 판정부터의 경과에 듦(design D11).
func TestA090R3bAWorkingSetErrorKeepsTheBase(t *testing.T) {
	f := a090Setup(t, nil)
	_, target := f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `ALTER TABLE exit_states RENAME TO a090_moved_states`)
	for i := 0; i < 11; i++ {
		if c := f.cycles(1); c.Err == nil {
			t.Fatalf("the working-set fixture did not fail the cycle: %+v", c)
		}
	}
	a090Exec(t, db, `ALTER TABLE a090_moved_states RENAME TO exit_states`)
	delete(f.prices.last, "005930")
	f.cycles(1)
	rows := f.outageRows()
	if len(rows) != 1 || !strings.Contains(rows[0].EventKey, target.ID) {
		t.Fatalf("rows = %+v, want the position alerted 60s after its last judgement across B2 cycles", rows)
	}
}

// codex A090-I2 · tasks 2.3g ⑧: 공지 적재가 실패한 뒤 보유 포지션이 전부 사라져도, 저장소가 복구된 다음 처리 주기(표시 0 인 B3)에서
// 같은 전이의 공지가 정확히 한 번 적재되고 대기열이 비며 전이는 다시 일어나지 않음.
func TestA090R3gANoticeQueuedBeforeThePositionsLeftIsRecordedInTheEmptyCycle(t *testing.T) {
	f := a090Setup(t, nil)
	f.a090Pair()
	db := a090Raw(t, f.exitHarness)
	a090Exec(t, db, `CREATE TRIGGER a090_fail_notice BEFORE INSERT ON alert_outbox
		WHEN NEW.event_type = 'engine.operating_mode' BEGIN SELECT RAISE(ABORT, 'notice store down'); END`)
	delete(f.prices.last, "005930")
	f.cycles(12)
	if len(f.a090Tightenings()) != 1 || f.observer.PendingModeNoticesForTest() != 1 {
		t.Fatalf("setup: tightenings=%d pending=%d", len(f.a090Tightenings()), f.observer.PendingModeNoticesForTest())
	}
	transition := f.a090Tightenings()[0]
	// 두 포지션 모두 보유 종료.
	ctx := context.Background()
	for _, symbol := range []string{"000001", "005930"} {
		watermark, err := f.journal.FillWatermark(ctx, symbol)
		if err != nil {
			t.Fatalf("FillWatermark: %v", err)
		}
		if _, err := f.journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
			AccountRef: exitAccount, Market: "kr", Symbol: symbol, Kind: "EXTERNAL",
			ExpectedPrevQuantity: "10", ExpectedFillWatermark: watermark, NewQuantity: "0", NewAvgPrice: "0",
			BrokerAsOf: f.clk.Now().Format(time.RFC3339), Evidence: "sold by hand",
		}); err != nil {
			t.Fatalf("ApplyPositionAdjustment: %v", err)
		}
	}
	a090Exec(t, db, `DROP TRIGGER a090_fail_notice`)
	if c := f.cycles(1); c.Judged != 0 || c.Unobserved != 0 {
		t.Fatalf("the empty cycle = %+v, want B3 with nothing held", c)
	}
	notices := f.noticeRows()
	if len(notices) != 1 || notices[0].EventKey != "operating_mode:"+journal.ModeEntryBlocked+":"+transition.ID {
		t.Fatalf("notice rows = %+v, want exactly the queued transition %s", notices, transition.ID)
	}
	if f.observer.PendingModeNoticesForTest() != 0 || len(f.a090Tightenings()) != 1 {
		t.Fatalf("pending=%d tightenings=%d after the empty cycle", f.observer.PendingModeNoticesForTest(), len(f.a090Tightenings()))
	}
}
