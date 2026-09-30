package engine_test

// a091 tasks 5.1 · 5.1a · 5.3 · 5.4 — 8/2 재생(생산 배달 실행자 경유, 팔 넷) · 재알림 창 경계 · 몫 실측 · 뒤쪽 보호 포지션.
//
// 8/2 의 모양(design 「8/2 원장 재독」): 보유 5 · 매도가능 0(한정 항 Sellable) · 13 관측 / 3분. 여기서는 전량 이탈 손절이 매 관측
// 0주로 깎이는 것을 13번 되풀이하고, 그 사이사이 생산 조립(`Context.AlertDeliverer`)의 배달 실행자 사이클을 돌린다.

import (
	"context"
	"errors"
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
	return res
}

func TestA091TheAugustSecondReplay(t *testing.T) {
	t.Run("(i) alerts on, working transport", func(t *testing.T) {
		pub := &a091Transport{}
		res := a091Replay(t, true, pub)
		t.Logf("measured: %+v", res)
		if res.sends != 1 || res.pendingNew != 0 || res.latched || res.mode != journal.ModeNormal {
			t.Fatalf("got %+v, want one send, the row settled, no latch, NORMAL — 13 observations fold into one episode", res)
		}
	})
	t.Run("(ii) alerts on, failing transport", func(t *testing.T) {
		pub := &a091Transport{fail: true}
		res := a091Replay(t, true, pub)
		t.Logf("measured: %+v", res)
		if res.pendingNew != 1 || res.sends < obs.DefaultCriticalAttempts || !res.latched || res.mode != journal.ModeEntryBlocked {
			t.Fatalf("got %+v, want one PENDING row, at least %d attempts, the latch and ENTRY_BLOCKED (intended a092)",
				res, obs.DefaultCriticalAttempts)
		}
	})
	t.Run("(iii) alerts on, no publisher", func(t *testing.T) {
		res := a091Replay(t, true, nil)
		t.Logf("measured: %+v", res)
		if res.pendingNew != 1 || !res.latched || res.mode != journal.ModeEntryBlocked || res.undeliveredLines == 0 {
			t.Fatalf("got %+v, want one PENDING row, the latch, ENTRY_BLOCKED and the executor's no-publisher lines", res)
		}
	})
	t.Run("(iv) alerts off", func(t *testing.T) {
		res := a091Replay(t, false, nil)
		t.Logf("measured: %+v", res)
		if res.allRows != 0 || res.latched || res.mode != journal.ModeNormal {
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

// --- 5.3 — 보고 호출 하나의 전체 경과, 세 칸, 관측 최악 ≤ 750ms(a091 배정 「0주 보고 몫」) -------------------------------------

type a091TimedAlerts struct {
	inner engine.ExitAlerter
	mu    sync.Mutex
	worst time.Duration
	n     int
}

func (a *a091TimedAlerts) Notify(ctx context.Context, e obs.Event) error {
	if e.Type != obs.EventExitStopSoldNothing {
		return a.inner.Notify(ctx, e)
	}
	start := time.Now()
	err := a.inner.Notify(ctx, e)
	d := time.Since(start)
	a.mu.Lock()
	a.n++
	if d > a.worst {
		a.worst = d
	}
	a.mu.Unlock()
	return err
}

const a091ReportShare = 750 * time.Millisecond

// a091ObservationPeriod 는 exit 관측 주기 기본값(5s) — 정본 「몫은 주기보다 작아야 한다」의 판정 기준.
const a091ObservationPeriod = 5 * time.Second

// a091MeasuredAckWorst 는 승인 경합 셀의 실측 최악(2026-10-01: 1.17~1.28s)을 올림한 값 — 5.4 의 실측 주입 변형이 씀.
const a091MeasuredAckWorst = 1300 * time.Millisecond

func TestA091TheReportFitsItsShare(t *testing.T) {
	const reps = 20
	for _, cell := range []struct {
		name   string
		fail   bool
		before func(t *testing.T, r *a091Rig) (stop func())
		// owned 는 a091 이 소유한 셀인가 — 소유 셀만 750ms 통과/실패, 승인 경합 셀은 측정 · 기록(Manager 판정 2026-10-01).
		owned bool
	}{
		{"uncontended ledger", false, nil, true},
		{"uncontended ledger, the record fails", true, nil, true},
		{"an operator acknowledging a 100-row backlog", false, func(t *testing.T, r *a091Rig) func() {
			ctx := context.Background()
			stop := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for k := 0; ; k++ {
					select {
					case <-stop:
						return
					default:
					}
					for i := 0; i < 100; i++ {
						_, _ = r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-backlog-" + time.Now().Format(time.RFC3339Nano),
							Type: "execgw.order_unresolved", Severity: "critical", Title: "t"})
					}
					_ = r.notifier.Acknowledge(ctx, "a091-operator")
				}
			}()
			return func() { close(stop); wg.Wait() }
		}, false},
		{"a busy connection pool", false, func(t *testing.T, r *a091Rig) func() {
			ctx := context.Background()
			stop := make(chan struct{})
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
					_, _ = r.journal.EnqueueAlert(ctx, journal.Alert{EventKey: "a091-pool-" + time.Now().Format(time.RFC3339Nano),
						Type: "execgw.order_unresolved", Severity: "normal", Title: "t"})
				}
			}()
			return func() { close(stop); wg.Wait() }
		}, true},
	} {
		t.Run(cell.name, func(t *testing.T) {
			timed := &a091TimedAlerts{}
			r := a091Harness(t, true, a091Zero(riskcalc.FloorBoundSellable), nil)
			timed.inner = obs.RecordOnly{N: r.notifier}
			// 옵션은 조립 뒤라 알림 부품을 다시 끼울 수 없으니, 시간 재는 부품을 끼운 새 하네스를 만든다.
			r = a091Harness(t, true, a091Zero(riskcalc.FloorBoundSellable), func(o *engine.ExitObserverOptions) {
				o.Alerts = timed
			})
			timed.inner = obs.RecordOnly{N: r.notifier}
			if cell.fail {
				closed, err := journal.Open(context.Background(), journal.Options{
					Path: t.TempDir() + "/closed.db", Clock: r.clk,
					FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
				})
				if err != nil {
					t.Fatal(err)
				}
				_ = closed.Close()
				r.notifier.Journal = closed
			}
			if cell.before != nil {
				stop := cell.before(t, r)
				defer stop()
			}
			r.breach()
			for i := 0; i < reps; i++ {
				r.observe()
				r.clk.Advance(5 * time.Second)
			}
			if timed.n != reps {
				t.Fatalf("reports timed = %d, want %d", timed.n, reps)
			}
			t.Logf("cell %q: %d reports, observed worst %s (share %s)", cell.name, timed.n, timed.worst, a091ReportShare)
			if !cell.owned {
				// 승인 경합 셀은 측정 · 기록(design D5 (i) 7판): 이 대기는 a092 정본이 이름 붙인 항(승인은 잠금 아래 밀린 행을
				// 하나씩 — 행 수 비례)이고 a091 은 빈도만 더함. 판정은 관측 주기 유계 하나.
				if timed.worst >= a091ObservationPeriod {
					t.Fatalf("observed worst %s reaches the %s observation period — the canonical bound fails", timed.worst, a091ObservationPeriod)
				}
				return
			}
			if timed.worst > a091ReportShare {
				t.Fatalf("observed worst %s exceeds the %s share — STOP and report (design D5 acceptance (i))",
					timed.worst, a091ReportShare)
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
		t.Run(delay.String(), func(t *testing.T) { a091LaterStopUnder(t, delay) })
	}
}

// a091LaterStopUnder 는 앞 포지션의 보고가 delay 만큼 걸려도 뒤 보호 포지션이 그 사이클에 제출되는지 잰다 — 배정값과 실측
// 최악(승인 경합 셀) 둘 다(Manager 판정 보강 1: 증명은 측정값으로).
func a091LaterStopUnder(t *testing.T, delay time.Duration) {
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

	if len(r.submit.places) != 1 || r.submit.places[0].Intent.Symbol != "005930" {
		t.Fatalf("places = %+v, want the 005930 stop in the same cycle", r.submit.places)
	}
	if len(order) != 2 || order[0] != "report" || order[1] != "place:005930" {
		t.Fatalf("order = %v, want the delayed report first and the later stop after it — otherwise this measured nothing", order)
	}
}
