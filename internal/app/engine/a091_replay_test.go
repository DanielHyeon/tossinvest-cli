package engine_test

// a091 tasks 5.1 · 5.1a · 5.3 · 5.4 — 8/2 재생(생산 배달 실행자 경유, 팔 넷) · 재알림 창 경계 · 몫 실측 · 뒤쪽 보호 포지션.
//
// 8/2 의 모양(design 「8/2 원장 재독」): 보유 5 · 매도가능 0(한정 항 Sellable) · 13 관측 / 3분. 여기서는 전량 이탈 손절이 매 관측
// 0주로 깎이는 것을 13번 되풀이하고, 그 사이사이 생산 조립(`Context.AlertDeliverer`)의 배달 실행자 사이클을 돌린다.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

type a091Transport struct {
	mu    sync.Mutex
	calls int
	fail  bool
}

func (p *a091Transport) Publish(context.Context, obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	if p.fail {
		return errors.New("transport is down")
	}
	return nil
}

func (p *a091Transport) count() int { p.mu.Lock(); defer p.mu.Unlock(); return p.calls }

type a091ReplayResult struct {
	pendingNew, allRows int
	sends               int
	latched             bool
	mode                string
	undeliveredLines    int
	noPublisherLines    int
}

// a091Replay 는 13 관측과 그 사이의 배달 사이클을 돌린다. pub 이 nil 이면 publisher 없음.
func a091Replay(t *testing.T, enabled bool, pub *a091Transport) a091ReplayResult {
	t.Helper()
	r := a091Harness(t, enabled, a091Zero(riskcalc.FloorBoundSellable), nil)
	if pub != nil {
		r.notifier.Publisher = pub
	}
	logger := r.notifier.Log
	aux, err := (&engine.Context{Journal: r.journal, Entry: r.gate, Notifier: r.notifier, AccountRef: exitAccount,
		Log: logger}).AlertDeliverer(r.clk)
	if err != nil {
		t.Fatalf("AlertDeliverer: %v", err)
	}
	r.breach()

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); _ = aux.Run(runCtx) }()
	t.Cleanup(func() { cancel(); <-done })

	for obsN := 0; obsN < 13; obsN++ {
		r.observe()
		if !r.clk.WaitForSleepers(1, 10*time.Second) {
			t.Fatalf("observation %d: the executor never slept", obsN)
		}
		r.clk.Advance(15 * time.Second) // 13 관측 / 3분
	}
	if !r.clk.WaitForSleepers(1, 10*time.Second) {
		t.Fatal("the executor did not settle after the last observation")
	}
	if len(r.submit.places) != 0 {
		t.Fatalf("places = %d, want 0 across the replay", len(r.submit.places))
	}

	res := a091ReplayResult{}
	res.pendingNew = len(r.rows(obs.EventExitStopSoldNothing))
	all, _ := r.journal.PendingAlerts(context.Background(), 0)
	res.allRows = len(all)
	if pub != nil {
		res.sends = pub.count()
	}
	_, res.latched = r.gate.Blocks()[execgw.ReasonAlertUndelivered]
	res.mode = r.mode()
	res.undeliveredLines = len(r.log.ofEvent(t, obs.EventAlertUndelivered))
	for _, l := range r.log.ofEvent(t, obs.EventAlertUndelivered) {
		if strings.Contains(fmt.Sprint(l["detail"]), "no publisher is configured") {
			res.noPublisherLines++
		}
	}
	return res
}

func TestA091TheAugustSecondReplay(t *testing.T) {
	t.Run("(i) alerts on, working transport", func(t *testing.T) {
		pub := &a091Transport{}
		res := a091Replay(t, true, pub)
		t.Logf("measured: %+v", res)
		if res.sends != 1 || res.pendingNew != 0 || res.latched || res.mode != journal.ModeNormal || res.undeliveredLines != 0 {
			t.Fatalf("got %+v, want one send, the row settled, no latch, NORMAL — 13 observations fold into one episode", res)
		}
	})
	t.Run("(ii) alerts on, failing transport", func(t *testing.T) {
		pub := &a091Transport{fail: true}
		res := a091Replay(t, true, pub)
		t.Logf("measured: %+v", res)
		if res.pendingNew != 1 || res.sends != 13 || res.undeliveredLines != 1 || !res.latched || res.mode != journal.ModeEntryBlocked {
			t.Fatalf("got %+v, want one PENDING row, 13 attempts (one per delivery cycle), one undelivered line, the latch and ENTRY_BLOCKED (intended a092)", res)
		}
	})
	t.Run("(iii) alerts on, no publisher", func(t *testing.T) {
		res := a091Replay(t, true, nil)
		t.Logf("measured: %+v", res)
		if res.pendingNew != 1 || !res.latched || res.mode != journal.ModeEntryBlocked || res.undeliveredLines != 14 || res.noPublisherLines != 13 {
			t.Fatalf("got %+v, want one PENDING row, the latch, ENTRY_BLOCKED and 14 undelivered lines (13 no-publisher, one per delivery cycle, + the latch)", res)
		}
	})
	t.Run("(iv) alerts off", func(t *testing.T) {
		res := a091Replay(t, false, nil)
		t.Logf("measured: %+v", res)
		if res.allRows != 0 || res.sends != 0 || res.latched || res.mode != journal.ModeNormal || res.undeliveredLines != 0 {
			t.Fatalf("got %+v, want no row, no latch, NORMAL — an alerts-off engine does not stop over a report it cannot send", res)
		}
	})
}

// 5.1a — 정착(운영자 승인) 행 뒤 같은 키: 재알림 창(1h) 안 재무장 0, 지나면 재무장 1(본문 교체).
func TestA091TheReminderWindowDecidesTheNextEpisode(t *testing.T) {
	r := a091Harness(t, true, a091Zero(riskcalc.FloorBoundSellable), nil)
	p := r.breach()
	r.observe()
	rows := r.rows(obs.EventExitStopSoldNothing)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	ctx := context.Background()
	if err := r.notifier.Acknowledge(ctx, "a091-operator", rows[0].ID); err != nil { // 정착 — 운영자 승인
		t.Fatalf("Acknowledge: %v", err)
	}
	r.clk.Advance(30 * time.Minute)
	r.observe()
	if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 0 {
		t.Fatalf("PENDING rows = %d inside the window, want 0 — a settled episode is not re-armed", n)
	}
	r.clk.Advance(31 * time.Minute) // 승인 뒤 61분
	r.observe()
	again := r.rows(obs.EventExitStopSoldNothing)
	if len(again) != 1 || again[0].EventKey != string(obs.EventExitStopSoldNothing)+"|"+p.ID {
		t.Fatalf("rows = %+v past the window, want the same key re-armed", again)
	}
	if again[0].Body == rows[0].Body {
		t.Errorf("the re-armed row kept the first episode's body — a new episode carries its own time")
	}
}

// --- 5.3 — 보고 몫, 칸별(design D5 7판) ----------------------------------------------------------------------------------
//
// 재는 양: **한 포지션의 관측 사이클 전체 경과**(ObserveOnce) — 보고 호출(로그 줄 · `n.mu` 대기 · 기록 · 실패 시 래치와 승격)을 전부
// 담는 상한이다(구현 리뷰 codex #1: Notify 만 재면 B2 오류 줄 · 실패 줄이 빠짐). 시세 · 판정 · 해제 트랜잭션도 들어가므로 보고 몫보다
// 크게 잰다(보수 방향). 소유 칸은 B2 · B7 두 경로를 다 잰다.

const a091ReportShare = 750 * time.Millisecond

// a091ObservationPeriod 는 exit 관측 주기 기본값(5s) — 정본 「몫은 주기보다 작아야 한다」의 판정 기준.
const a091ObservationPeriod = 5 * time.Second

// a091MeasuredAckWorst 는 승인 경합 칸의 실측 최악(2026-10-01: 1.06~1.28s, `-race` · 커버리지 실행 포함)을 올림한 값 — 5.4 의 실측 주입 변형이 씀.
const a091MeasuredAckWorst = 1300 * time.Millisecond

// a091FailOutboxWrites 는 원장의 알림 행 쓰기만 실패시킴(SQL 트리거) — 운영 모드 쓰기는 살아 있어 승격 트랜잭션이 실제로 돈다
// (구현 리뷰 codex #2: 닫힌 원장은 승격까지 실패시켜 그 비용을 재지 못함).
func a091FailOutboxWrites(t *testing.T, r *a091Rig) {
	t.Helper()
	db, err := sql.Open("sqlite", r.dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, stmt := range []string{
		`CREATE TRIGGER a091_fail_insert BEFORE INSERT ON alert_outbox BEGIN SELECT RAISE(ABORT, 'a091 injected outbox failure'); END`,
		`CREATE TRIGGER a091_fail_update BEFORE UPDATE ON alert_outbox BEGIN SELECT RAISE(ABORT, 'a091 injected outbox failure'); END`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("installing the outbox failure: %v", err)
		}
	}
}

func a091OutboxCount(t *testing.T, r *a091Rig, typ obs.EventType) int {
	t.Helper()
	db, err := sql.Open("sqlite", r.dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM alert_outbox WHERE event_type = ?`, string(typ)).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestA091TheReportFitsItsShare(t *testing.T) {
	const reps = 20
	var ackWindows [][2]time.Time // 승인 칸의 승인 호출 구간(겹침 증명)
	type cell struct {
		name   string
		floor  func() engine.FloorSource
		fail   bool
		before func(t *testing.T, r *a091Rig) (stop func() int)
		owned  bool
	}
	b7 := func() engine.FloorSource { return a091Zero(riskcalc.FloorBoundSellable) }
	b2 := func() engine.FloorSource { return a091Failing(errors.New("holdings read failed")) }
	ack := func(t *testing.T, r *a091Rig) func() int {
		ctx := context.Background()
		fill := func() {
			for i := 0; i < 100; i++ {
				if _, err := r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-backlog-" + time.Now().Format(time.RFC3339Nano) + fmt.Sprint(i),
					Type: "execgw.order_unresolved", Severity: "critical", Title: "t"}); err != nil {
					t.Errorf("backlog insert: %v", err)
					return
				}
			}
		}
		fill() // 측정 전에 밀린 행 100 이 이미 있다(준비 장벽 — codex #3)
		stop := make(chan struct{})
		acks := 0
		var mu sync.Mutex
		ackWindows = ackWindows[:0]
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				start := time.Now()
				if err := r.notifier.Acknowledge(ctx, "a091-operator"); err != nil {
					t.Errorf("Acknowledge: %v", err)
					return
				}
				end := time.Now()
				mu.Lock()
				acks++
				ackWindows = append(ackWindows, [2]time.Time{start, end})
				mu.Unlock()
				fill()
			}
		}()
		return func() int { close(stop); wg.Wait(); mu.Lock(); defer mu.Unlock(); return acks }
	}
	pool := func(t *testing.T, r *a091Rig) func() int {
		ctx := context.Background()
		stop := make(chan struct{})
		var wg sync.WaitGroup
		writes := 0
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-pool-" + time.Now().Format(time.RFC3339Nano),
					Type: "execgw.order_unresolved", Severity: "normal", Title: "t"}); err != nil {
					t.Errorf("pool write: %v", err)
					return
				}
				writes++
			}
		}()
		return func() int { close(stop); wg.Wait(); return writes }
	}
	for _, c := range []cell{
		{"uncontended ledger, B7", b7, false, nil, true},
		{"uncontended ledger, B2", b2, false, nil, true},
		{"the record fails and escalates, B7", b7, true, nil, true},
		{"the record fails and escalates, B2", b2, true, nil, true},
		{"a busy connection pool, B7", b7, false, pool, true},
		{"a busy connection pool, B2", b2, false, pool, true},
		{"an operator acknowledging a 100-row backlog, B7", b7, false, ack, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := a091Harness(t, true, c.floor(), nil)
			if c.fail {
				a091FailOutboxWrites(t, r)
			}
			var stop func() int
			if c.before != nil {
				stop = c.before(t, r)
			}
			r.breach()
			var worst time.Duration
			var cycles [][2]time.Time
			for i := 0; i < reps; i++ {
				start := time.Now()
				r.observe()
				end := time.Now()
				if d := end.Sub(start); d > worst {
					worst = d
				}
				cycles = append(cycles, [2]time.Time{start, end})
				r.clk.Advance(5 * time.Second)
			}
			load := -1
			if stop != nil {
				load = stop()
				if load == 0 {
					t.Fatalf("the contention never ran during the measurement — this cell measured nothing")
				}
			}
			t.Logf("cell %q: %d cycles, observed worst %s (share %s), contention ops %d", c.name, reps, worst, a091ReportShare, load)
			if c.fail {
				if got := r.mode(); got != journal.ModeEntryBlocked {
					t.Fatalf("mode = %s, want ENTRY_BLOCKED — the escalation transaction did not run", got)
				}
			} else if n := a091OutboxCount(t, r, obs.EventExitStopSoldNothing); n != 1 {
				// 승인 칸에서는 운영자 승인이 그 행도 정착시키므로 PENDING 이 아니라 행 수로 잰다.
				t.Fatalf("outbox rows of the new kind = %d, want 1 — the report never recorded", n)
			}
			if !c.owned {
				// 겹침 증명(i2 codex): 승인 호출 구간과 시간이 겹친 측정 사이클 수 — 0 이면 이 칸은 경합을 재지 않았다.
				overlap := 0
				for _, cy := range cycles {
					for _, w := range ackWindows {
						if cy[0].Before(w[1]) && w[0].Before(cy[1]) {
							overlap++
							break
						}
					}
				}
				t.Logf("cycles overlapping an acknowledgement: %d of %d (acknowledgements %d)", overlap, reps, len(ackWindows))
				if overlap == 0 {
					t.Fatalf("no measured cycle overlapped an acknowledgement — this cell measured nothing")
				}
				// 승인 경합 칸은 측정 · 기록(design D5 (i) 7판): a092 정본이 이름 붙인 항이고 a091 은 빈도만 더함. 판정은 주기 유계 하나.
				if worst >= a091ObservationPeriod {
					t.Fatalf("observed worst %s reaches the %s observation period — the canonical bound fails", worst, a091ObservationPeriod)
				}
				return
			}
			if worst > a091ReportShare {
				t.Fatalf("observed worst %s exceeds the %s share — STOP and report (design D5 acceptance (i))", worst, a091ReportShare)
			}
		})
	}
}

// --- 5.4 — 뒤쪽 보호 포지션: 앞 포지션의 보고에 배정만큼 지연을 주입해도 그 사이클에 판정 · 제출된다 ------------------------------

type a091SymbolFloor map[string]riskcalc.ConfirmedFloor

func (f a091SymbolFloor) ConfirmedFloor(_ context.Context, _, symbol string) (riskcalc.ConfirmedFloor, bool, error) {
	c, ok := f[symbol]
	return c, ok, nil
}

type a091DelayedAlerts struct {
	inner   engine.ExitAlerter
	advance func()
	order   *[]string
}

func (a *a091DelayedAlerts) Notify(ctx context.Context, e obs.Event) error {
	if e.Type == obs.EventExitStopSoldNothing {
		*a.order = append(*a.order, "report")
		a.advance() // 보고가 배정만큼 걸렸다 — 관측 시계가 그만큼 흐른다
	}
	return a.inner.Notify(ctx, e)
}

func TestA091ALaterStopStillGoesOutInTheSameCycle(t *testing.T) {
	for _, delay := range []time.Duration{a091ReportShare, a091MeasuredAckWorst} {
		t.Run(delay.String(), func(t *testing.T) { a091LaterStopUnder(t, delay, true) })
	}
	// 양성 대조군: 시세 수명(15s)을 넘기는 지연이면 뒤 포지션은 그 사이클에 판정되지 않는다 — 이 시험이 실제로 시세 수명을 잰다는 증거.
	t.Run("positive control 20s", func(t *testing.T) { a091LaterStopUnder(t, 20*time.Second, false) })
}

// a091LaterStopUnder 는 앞 포지션의 보고가 delay 만큼 걸려도 뒤 보호 포지션이 그 사이클에 제출되는지 잰다 — 배정값과 실측
// 최악(승인 경합 셀) 둘 다(Manager 판정 보강 1: 증명은 측정값으로).
func a091LaterStopUnder(t *testing.T, delay time.Duration, wantPlaced bool) {
	var order []string
	delayed := &a091DelayedAlerts{order: &order}
	r := a091Harness(t, true, a091SymbolFloor{"000660": {Quantity: "0", Bound: riskcalc.FloorBoundSellable}},
		func(o *engine.ExitObserverOptions) { o.Alerts = delayed })
	delayed.inner = obs.RecordOnly{N: r.notifier}
	delayed.advance = func() { r.clk.Advance(delay) }
	record := r.submit.record
	r.submit.record = func(req execgw.PlaceRequest) string {
		order = append(order, "place:"+req.Intent.Symbol)
		return record(req)
	}

	// 관측 순서는 종목 순이다 — 0주 포지션(000660)이 앞, 보호 포지션(005930)이 뒤(순서 단언이 아래에서 확인).
	r.entry("000660", "10", "70000", "68000", "70000")
	r.entry("005930", "10", "70000", "68000", "70000")
	r.quote("005930", 67900)
	r.quote("000660", 67900)
	r.observe()

	if !wantPlaced {
		if len(r.submit.places) != 0 || len(order) != 1 || order[0] != "report" {
			t.Fatalf("places = %+v, order = %v — a %s delay must outlive the quote, or this test cannot see the quote lifetime", r.submit.places, order, delay)
		}
		return
	}
	if len(r.submit.places) != 1 || r.submit.places[0].Intent.Symbol != "005930" {
		t.Fatalf("places = %+v, want the 005930 stop in the same cycle", r.submit.places)
	}
	if len(order) != 2 || order[0] != "report" || order[1] != "place:005930" {
		t.Fatalf("order = %v, want the delayed report first and the later stop after it — otherwise this measured nothing", order)
	}
}
