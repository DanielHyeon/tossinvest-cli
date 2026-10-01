package engine_test

// a091 — 한 주도 못 판 손절은 critical 이다(tasks 3.x · 4.x).
//
// 하네스 둘(tasks 3 머리):
//   (가) newExitHarness 그대로 — 가짜 알림 수집기(종류 · 호출 수 · 필드).
//   (나) a091Harness — 실제 obs.RecordOnly + 실제 원장 + 실제 게이트 + 로그 캡처(JSON). 행 · 등급 · 게이트 · 모드 · 로그 줄을 잰다.
//
// 이 파일의 시험이 지키는 경계: 제출 수량 · 제출 수 · 발의 해제/재발의(§0.3 · §0.9)는 한 글자도 바뀌지 않고, 바뀌는 것은 보고(종류 ·
// 등급 · 문구 · 로그)뿐이다.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// --- 하네스 (나) ------------------------------------------------------------------------------------------------------

// a091Buf 는 로그 캡처 버퍼 — 관측 goroutine 과 시험 goroutine 이 같이 읽고 씀.
type a091Buf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *a091Buf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *a091Buf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

// lines 는 JSON 로그 줄을 사전으로 돌려줌.
func (s *a091Buf) lines(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(s.String()), "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		m := map[string]any{}
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("log line is not JSON: %q", l)
		}
		out = append(out, m)
	}
	return out
}

// ofEvent 는 그 종류의 로그 줄들.
func (s *a091Buf) ofEvent(t *testing.T, e obs.EventType) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, l := range s.lines(t) {
		if l["event"] == string(e) {
			out = append(out, l)
		}
	}
	return out
}

// a091Floor 는 호출마다 함수를 부르는 하한 공급자 — ctx 를 받아 취소 · 오류를 흉내 냄.
type a091Floor struct {
	fn func(ctx context.Context) (riskcalc.ConfirmedFloor, bool, error)
}

func (f *a091Floor) ConfirmedFloor(ctx context.Context, _, _ string) (riskcalc.ConfirmedFloor, bool, error) {
	return f.fn(ctx)
}

func a091Zero(bound string) *a091Floor {
	return &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
		return riskcalc.ConfirmedFloor{Quantity: "0", Bound: bound}, true, nil
	}}
}

func a091Failing(err error) *a091Floor {
	return &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
		return riskcalc.ConfirmedFloor{}, true, err
	}}
}

type a091Rig struct {
	*exitHarness
	log      *a091Buf
	notifier *obs.Notifier
}

// a091Harness 는 (나) — 알림 켜짐 여부와 하한 공급자를 받음. 알림기는 하네스의 원장 · 게이트를 씀(생산과 같은 기록 전용 입구).
func a091Harness(t *testing.T, enabled bool, floor engine.FloorSource, mutate func(*engine.ExitObserverOptions)) *a091Rig {
	t.Helper()
	buf := &a091Buf{}
	logger := obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clock.NewFake(exitNow)})
	r := &a091Rig{log: buf}
	r.exitHarness = newExitHarness(t, func(o *engine.ExitObserverOptions) {
		r.notifier = &obs.Notifier{Journal: o.Journal, Gate: o.Retrier.Gate, AccountRef: o.AccountRef,
			Clock: clock.System(), Log: logger}
		o.Alerts = obs.RecordOnly{N: r.notifier}
		o.ZeroFloorLog = logger // 생산 모양: 관측자 전체 Log 는 nil, a091 줄은 전용 싱크로
		o.NotificationsEnabled = enabled
		if floor != nil {
			o.Floor = floor
		}
		if mutate != nil {
			mutate(o)
		}
	})
	return r
}

// alertsCause 는 이관 버퍼로 넘어간(normal) 마지막 0주 알림의 payload cause 를 돌려줌 — 하네스 (나) 의 알림기에는 이관 버퍼가 없어
// 일반 등급은 버려지므로, 같은 사건을 가짜 수집기로 한 번 더 관측해 잼.
func (r *a091Rig) alertsCause(t *testing.T) string {
	t.Helper()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		o.Floor = a091Zero(riskcalc.FloorBoundHoldings)
		o.NotificationsEnabled = true
	})
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 67900)
	h.observe()
	a, ok := h.alerts.first(obs.EventExitProposalCapped)
	if !ok {
		t.Fatal("no capped alert")
	}
	return fmt.Sprint(a.Fields["cause"])
}

func (r *a091Rig) rows(typ obs.EventType) []journal.Alert {
	return a092PendingOfType(r.t, r.journal, string(typ))
}

// breach 는 보호 청산(기준선 이탈 전량)을 만든다 — 10주 · 손절 68000 · 관측 67900.
func (r *a091Rig) breach() journal.Position {
	p := r.entry("005930", "10", "70000", "68000", "70000")
	r.quote("005930", 67900)
	return p
}

// --- 3.1 · 4.1 — 알림 켜짐 · 보호 · 하한 0 ---------------------------------------------------------------------------------

func TestA091AProtectiveZeroIsACriticalRow(t *testing.T) {
	r := a091Harness(t, true, a091Zero(riskcalc.FloorBoundSellable), nil)
	p := r.breach()
	r.observe()

	if len(r.submit.places) != 0 {
		t.Fatalf("places = %d, want 0 — a floor of zero authorises no sale", len(r.submit.places))
	}
	rows := r.rows(obs.EventExitStopSoldNothing)
	if len(rows) != 1 {
		t.Fatalf("rows of %s = %d, want 1", obs.EventExitStopSoldNothing, len(rows))
	}
	row := rows[0]
	if want := string(obs.EventExitStopSoldNothing) + "|" + p.ID; row.EventKey != want {
		t.Errorf("key = %q, want %q", row.EventKey, want)
	}
	if row.Severity != string(obs.SeverityCritical) {
		t.Errorf("severity = %q, want critical", row.Severity)
	}
	if want := "005930 손절이 한 주도 나가지 않았다"; row.Title != want {
		t.Errorf("title = %q, want %q", row.Title, want)
	}
	wantBody := "005930 · " + r.clk.Now().UTC().Format(time.RFC3339) + " 관측: 매도가능 수량이 0이다 — 다른 미체결 매도가 주식을 잡고 있을 수 있다. " +
		"제안 10주 중 0주를 제출했다 — 손절이 나가지 않았다. 불일치가 해소되면 같은 단계를 다시 제안한다."
	if row.Body != wantBody {
		t.Errorf("body:\n got %s\nwant %s", row.Body, wantBody)
	}
	if strings.Contains(row.Title+row.Body, "일부") {
		t.Errorf("a zero-share report must not say part went out:\n%s\n%s", row.Title, row.Body)
	}
	if len(r.rows(obs.EventExitProposalCapped)) != 0 {
		t.Error("a capped row appeared — the protective zero is one kind")
	}
}

// --- 3.2 — 하한 계산 실패(B2)는 같은 종류 · 등급 · 키, 원문 오류는 제목 · 본문 · payload 에 없음 -----------------------------------

func TestA091AFloorThatCannotBeComputedIsTheSameReport(t *testing.T) {
	raw := "official: holdings read refused for acct-exit: HTTP 500 upstream"
	r := a091Harness(t, true, a091Failing(errors.New(raw)), nil)
	p := r.breach()
	r.observe()

	if len(r.submit.places) != 0 {
		t.Fatal("B2 must submit nothing")
	}
	rows := r.rows(obs.EventExitStopSoldNothing)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if want := string(obs.EventExitStopSoldNothing) + "|" + p.ID; rows[0].EventKey != want {
		t.Errorf("key = %q, want %q (the same key as the zero-floor path)", rows[0].EventKey, want)
	}
	if !strings.Contains(rows[0].Body, "확정 하한을 계산하지 못했다") {
		t.Errorf("body does not name the cause:\n%s", rows[0].Body)
	}
	for _, s := range []string{raw, "HTTP 500", "official:"} {
		if strings.Contains(rows[0].Title+rows[0].Body+rows[0].Payload, s) {
			t.Errorf("the raw error %q reached the alert row", s)
		}
	}
}

// --- 3.2a — 알림 꺼짐: B2 는 알림 0(오늘처럼), 끝은 옛 종류 normal 하나. outbox · 게이트 · 모드 무변화 -------------------------------

func TestA091WithAlertsOffNothingChangesButTheWording(t *testing.T) {
	t.Run("B2 raises no alert", func(t *testing.T) {
		h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
			o.Floor = a091Failing(errors.New("holdings read failed"))
			o.NotificationsEnabled = false
		})
		h.entry("005930", "10", "70000", "68000", "70000")
		h.quote("005930", 67900)
		h.observe()
		if n := len(h.alerts.events); n != 0 {
			t.Fatalf("alerts = %d (%+v), want 0 — today B2 only logs", n, h.alerts.events)
		}
	})
	t.Run("the zero floor stays a normal capped alert", func(t *testing.T) {
		h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
			o.Floor = a091Zero(riskcalc.FloorBoundSellable)
			o.NotificationsEnabled = false
		})
		h.entry("005930", "10", "70000", "68000", "70000")
		h.quote("005930", 67900)
		h.observe()
		if h.alerts.count(obs.EventExitStopSoldNothing) != 0 || h.alerts.count(obs.EventExitProposalCapped) != 1 {
			t.Fatalf("alerts = %+v, want exactly one %s", h.alerts.events, obs.EventExitProposalCapped)
		}
		a, _ := h.alerts.first(obs.EventExitProposalCapped)
		if want := "005930 청산이 확정 하한에 걸려 한 주도 나가지 않았다"; a.Title != want {
			t.Errorf("title = %q, want %q", a.Title, want)
		}
	})
	t.Run("B2 leaves exactly one account-free log line in the production shape", func(t *testing.T) {
		r := a091Harness(t, false, a091Failing(errors.New("holdings read failed for "+exitAccount)), nil)
		p := r.breach()
		r.observe()
		lines := r.log.ofEvent(t, obs.EventExitProposalCapped)
		if len(lines) != 1 {
			t.Fatalf("capped lines = %d, want exactly 1:\n%s", len(lines), r.log.String())
		}
		l := lines[0]
		if l["level"] != "ERROR" || l["cause"] != "floor_unknown" || l["symbol"] != "005930" || l["position_id"] != p.ID {
			t.Errorf("B2 line = %v, want level ERROR, cause floor_unknown, the symbol and the position", l)
		}
		if _, ok := l["account"]; ok || strings.Contains(fmt.Sprint(l), exitAccount) {
			t.Errorf("the B2 line carries the account: %v", l)
		}
		if len(r.log.ofEvent(t, obs.EventExitStopSoldNothing)) != 0 {
			t.Error("an alerts-off engine logged the new kind")
		}
	})
	t.Run("the take-profit and the alerts-off bodies say what happened", func(t *testing.T) {
		h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
			o.Floor = a091Zero(riskcalc.FloorBoundSellable)
		})
		h.entry("005930", "10", "70000", "68000", "70000")
		h.quote("005930", 67900)
		h.observe()
		a, _ := h.alerts.first(obs.EventExitProposalCapped)
		want := "005930 · " + h.clk.Now().UTC().Format(time.RFC3339) + " 관측: 매도가능 수량이 0이다 — 다른 미체결 매도가 주식을 잡고 있을 수 있다. " +
			"제안 10주 중 0주를 제출했다 — 손절이 나가지 않았다. 불일치가 해소되면 같은 단계를 다시 제안한다."
		if a.Body != want {
			t.Errorf("alerts-off protective body:\n got %s\nwant %s", a.Body, want)
		}
	})
	for _, cause := range []struct {
		name  string
		floor engine.FloorSource
	}{
		{"B2", a091Failing(errors.New("holdings read failed"))},
		{"zero floor", a091Zero(riskcalc.FloorBoundSellable)},
	} {
		t.Run("no row, no latch, no mode: "+cause.name, func(t *testing.T) {
			r := a091Harness(t, false, cause.floor, nil)
			r.breach()
			r.observe()
			if rows, _ := r.journal.PendingAlerts(context.Background(), 0); len(rows) != 0 {
				t.Errorf("outbox rows = %+v, want none with alerts off", rows)
			}
			if rej := r.gate.CheckEntry(); rej != nil {
				t.Errorf("entry gate = %v, want open", rej)
			}
			if got := r.mode(); got != journal.ModeNormal {
				t.Errorf("mode = %s, want NORMAL", got)
			}
		})
	}
}

// --- 3.3 · 4.2 — 부분 캡은 종류 · 등급 · 문구 무변화 -------------------------------------------------------------------------

func TestA091APartialCapIsUnchanged(t *testing.T) {
	r := a091Harness(t, true, &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
		return riskcalc.ConfirmedFloor{Quantity: "3", Bound: riskcalc.FloorBoundHoldings}, true, nil
	}}, nil)
	r.breach()
	r.observe()
	if len(r.submit.places) != 1 || r.submit.places[0].Intent.Quantity != 3 {
		t.Fatalf("places = %+v, want the capped liquidation of 3", r.submit.places)
	}
	if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 0 {
		t.Errorf("rows of the new kind = %d, want 0 — part of the stop went out", n)
	}
	lines := r.log.ofEvent(t, obs.EventExitProposalCapped)
	if len(lines) != 1 {
		t.Fatalf("capped log lines = %d, want 1", len(lines))
	}
	if lines[0]["severity"] != string(obs.SeverityNormal) {
		t.Errorf("severity = %v, want normal", lines[0]["severity"])
	}
	want := "10 제안 · RECONCILE 확정 하한이 허용하는 수량 3 (broker holdings) · 잔여 7는 매도되지 않는다. " +
		"불일치가 해소되면 같은 단계를 다시 제안한다."
	if lines[0]["detail"] != want {
		t.Errorf("partial-cap body changed:\n got %v\nwant %s", lines[0]["detail"], want)
	}
}

// --- 3.3a (엔진) — 보유 0(Holdings 한정 0)은 옛 종류 normal · 본문이 원인을 말함 -----------------------------------------------

func TestA091AZeroHoldingIsNotAFailedStop(t *testing.T) {
	r := a091Harness(t, true, a091Zero(riskcalc.FloorBoundHoldings), nil)
	r.breach()
	r.observe()
	if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 0 {
		t.Fatalf("rows of the new kind = %d, want 0 — the account no longer holds the shares", n)
	}
	lines := r.log.ofEvent(t, obs.EventExitProposalCapped)
	if len(lines) != 1 || !strings.Contains(fmt.Sprint(lines[0]["detail"]), "계좌에 보유가 없다") {
		t.Fatalf("capped lines = %+v, want one that says the account holds none", lines)
	}
	a := r.alertsCause(t)
	if a != "no_holding" {
		t.Errorf("cause = %q, want no_holding", a)
	}
}

// --- 3.3b — 종료 취소는 출처로 판정 -------------------------------------------------------------------------------------

func TestA091OnlyACancellationOnlyFailureIsSuppressed(t *testing.T) {
	type arm struct {
		name    string
		err     func() error
		cancel  bool // 하한 조회 안에서 호출자 ctx 를 취소하는가
		wantRow bool
	}
	auth := fmt.Errorf("engine: reading the holding of 005930 for the confirmed floor: %w", official.ErrAuth)
	for _, a := range []arm{
		{"(i) cancellation only, ctx done", func() error {
			return fmt.Errorf("engine: reading the holding: %w", &url.Error{Op: "Get", URL: "x", Err: context.Canceled})
		}, true, false},
		{"(ii) a real failure, then shutdown", func() error { return errors.New("broker said no") }, true, true},
		{"(iv) an HTTP deadline with the loop alive", func() error {
			return fmt.Errorf("engine: reading the holding: %w", context.DeadlineExceeded)
		}, false, true},
		{"(v) a joined auth refusal and cancellation", func() error {
			return errors.Join(auth, fmt.Errorf("escalating: %w", context.Canceled))
		}, true, true},
		{"(viii) a cancellation-only error while the loop is alive", func() error {
			return fmt.Errorf("engine: reading the holding: %w", context.Canceled) // 루프가 아닌 하위 ctx 의 취소 — 종료가 아님
		}, false, true},
		{"(vi) shutdown during the retry backoff reports the last transient failure", nil, true, true},
		{"(vii) the production client erases the cause", func() error {
			return fmt.Errorf("%w: %s", official.ErrTransport, context.Canceled) // official/client.go doRequest 모양
		}, true, true},
	} {
		t.Run(a.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var retrier *execgw.Retrier
			floor := &a091Floor{fn: func(fctx context.Context) (riskcalc.ConfirmedFloor, bool, error) {
				if a.err == nil { // (vi): 실제 Retrier — 첫 시도 일시 실패, 대기 중 종료
					go func() {
						clk := retrier.Clock.(*clock.Fake)
						if clk.WaitForSleepers(1, 5*time.Second) {
							cancel()
						}
					}()
					err := retrier.Query(fctx, execgw.QueryHoldings, func(context.Context) error { return official.ErrServer })
					return riskcalc.ConfirmedFloor{}, true, fmt.Errorf("engine: reading the holding: %w", err)
				}
				if a.cancel {
					cancel()
				}
				return riskcalc.ConfirmedFloor{}, true, a.err()
			}}
			r := a091Harness(t, true, floor, func(o *engine.ExitObserverOptions) {
				o.Retrier.Policy = execgw.RetryPolicy{MaxAttempts: 3, Budget: time.Minute, BaseBackoff: time.Second, MaxBackoff: time.Second}
				retrier = o.Retrier
			})
			r.breach()
			r.observer.ObserveOnce(ctx)
			got := len(r.rows(obs.EventExitStopSoldNothing))
			if a.wantRow && got != 1 {
				t.Fatalf("rows = %d, want 1 — this failure is not a shutdown", got)
			}
			if !a.wantRow && got != 0 {
				t.Fatalf("rows = %d, want 0 — the only failure was the shutdown itself", got)
			}
			if rej := r.gate.CheckEntry(); rej != nil && rej.Reason == execgw.ReasonAlertUndelivered {
				t.Errorf("a false undelivered latch: %v", rej)
			}
		})
	}
}

// (iii) — 판정 뒤 · 기록 전에 종료가 끼어든다: 기록은 끝난 ctx 를 보지 않으므로 행 1 · 가짜 래치 0.
type a091CancelBeforeRecord struct {
	inner  engine.ExitAlerter
	cancel func()
}

func (c *a091CancelBeforeRecord) Notify(ctx context.Context, e obs.Event) error {
	if e.Type == obs.EventExitStopSoldNothing {
		c.cancel()
	}
	return c.inner.Notify(ctx, e)
}

func TestA091AShutdownBetweenJudgementAndRecordStillRecords(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wrap := &a091CancelBeforeRecord{cancel: cancel}
	r := a091Harness(t, true, a091Failing(errors.New("broker said no")), func(o *engine.ExitObserverOptions) {
		o.Alerts = wrap
	})
	wrap.inner = obs.RecordOnly{N: r.notifier}
	r.breach()
	r.observer.ObserveOnce(ctx)
	if ctx.Err() == nil {
		t.Fatal("arrangement: the shutdown never came")
	}
	if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 1 {
		t.Fatalf("rows = %d, want 1", n)
	}
	if rej := r.gate.CheckEntry(); rej != nil {
		t.Errorf("entry gate = %v, want open — a cancelled ctx must not fail the record", rej)
	}
}

func TestA091TheCancellationPredicate(t *testing.T) {
	for _, c := range []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"bare", context.Canceled, true},
		{"wrapped", fmt.Errorf("a: %w", context.Canceled), true},
		{"url error", &url.Error{Op: "Get", URL: "x", Err: context.Canceled}, true},
		{"joined cancellations", errors.Join(context.Canceled, fmt.Errorf("b: %w", context.Canceled)), true},
		{"joined with a real failure", errors.Join(errors.New("auth"), context.Canceled), false},
		{"deadline", context.DeadlineExceeded, false},
		{"erased by the official client", fmt.Errorf("%w: %s", official.ErrTransport, context.Canceled), false},
		{"a multi-error of nothing", a091Multi{nil, nil}, false},
		{"a wrapper that unwraps to nothing", a091NilUnwrap{}, false},
	} {
		if got := engine.CancellationOnlyForTest(c.err); got != c.want {
			t.Errorf("%s: cancellationOnly = %v, want %v", c.name, got, c.want)
		}
	}
}

// a091NilUnwrap 은 Unwrap 이 nil 을 돌려주는 감싼 오류 — 잎이 없으므로 취소뿐이 아님.
type a091NilUnwrap struct{}

func (a091NilUnwrap) Error() string { return "wrapped nothing" }
func (a091NilUnwrap) Unwrap() error { return nil }

// a091Multi 는 잎이 비었을 수 있는 다중 오류 — errors.Join 은 nil 잎을 거르지만 다른 구현은 그렇지 않을 수 있음.
type a091Multi []error

func (m a091Multi) Error() string   { return "multi" }
func (m a091Multi) Unwrap() []error { return m }

// --- 3.3c — 종료 중 보고 대기: 기록이 막혀 있으면 루프는 기다렸다 행을 쓰고 돌아온다(기한 없음 — 이름 붙은 대가) ---------------

type a091BlockingAlerts struct {
	inner   engine.ExitAlerter
	release chan struct{}
	entered chan struct{}
	ctxErr  error
}

func (b *a091BlockingAlerts) Notify(ctx context.Context, e obs.Event) error {
	if e.Type == obs.EventExitStopSoldNothing {
		close(b.entered)
		<-b.release
		b.ctxErr = ctx.Err()
	}
	return b.inner.Notify(ctx, e)
}

func TestA091ShutdownWaitsForTheReportWithoutADeadline(t *testing.T) {
	blocking := &a091BlockingAlerts{release: make(chan struct{}), entered: make(chan struct{})}
	r := a091Harness(t, true, a091Failing(errors.New("broker said no")), func(o *engine.ExitObserverOptions) {
		o.Alerts = blocking
	})
	blocking.inner = obs.RecordOnly{N: r.notifier} // 실제 기록 — 막힘이 풀린 뒤 원장에 행이 생기는지 잼
	h := r.exitHarness
	r.breach()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { h.observer.ObserveOnce(ctx); close(done) }()
	select {
	case <-blocking.entered:
	case <-done:
		t.Fatal("the cycle returned without recording the report")
	case <-time.After(5 * time.Second):
		t.Fatal("the report was never recorded")
	}
	cancel() // 종료가 보고 도중에 옴
	select {
	case <-done:
		t.Fatal("the loop returned while the report was still recording")
	case <-time.After(50 * time.Millisecond):
	}
	close(blocking.release)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the loop did not return after the report finished — deadlock")
	}
	if blocking.ctxErr != nil {
		t.Errorf("the report saw ctx.Err() = %v, want nil — it records with context.WithoutCancel", blocking.ctxErr)
	}
	if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 1 {
		t.Errorf("rows = %d after the shutdown, want 1 — the report survived the cancellation", n)
	}
	if rej := r.gate.CheckEntry(); rej != nil {
		t.Errorf("entry gate = %v, want open — no false undelivered latch", rej)
	}
}

// --- 3.4 — 주문 액션 전수 표 · 익절은 옛 종류 ------------------------------------------------------------------------------

// a091OrderableActions 는 exitpolicy 의 Action 상수를 AST 로 열거해 Orderable 인 것을 돌려줌 — 새 액션이 생기면 이 표가 깨짐.
func a091OrderableActions(t *testing.T) []exitpolicy.Action {
	t.Helper()
	// go test 는 패키지 디렉터리에서 돈다 — 상대 경로라 -trimpath 빌드에서도 같은 파일을 읽음(runtime.Caller 경로는 모듈 상대가 됨).
	dir := filepath.Join("..", "..", "exitpolicy")
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []exitpolicy.Action
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, e.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.CONST {
				continue
			}
			for _, s := range g.Specs {
				vs := s.(*ast.ValueSpec)
				id, ok := vs.Type.(*ast.Ident)
				if !ok || id.Name != "Action" {
					continue
				}
				for i := range vs.Names {
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok {
						t.Fatalf("Action constant %s is not a literal", vs.Names[i].Name)
					}
					a := exitpolicy.Action(strings.Trim(lit.Value, `"`))
					if a.Orderable() {
						out = append(out, a)
					}
				}
			}
		}
	}
	return out
}

func TestA091TheProtectiveSplitCoversEveryOrderableAction(t *testing.T) {
	actions := a091OrderableActions(t)
	if len(actions) != 5 {
		t.Fatalf("orderable actions = %v, want 5 — a new action needs a decision here (protective or take-profit)", actions)
	}
	protective := map[exitpolicy.Action]bool{exitpolicy.ActionBaselineBreach: true, exitpolicy.ActionLadderStop: true}
	for _, a := range actions {
		if got := engine.IsProtectiveForTest(a); got != protective[a] {
			t.Errorf("isProtective(%s) = %v, want %v", a, got, protective[a])
		}
	}
}

func TestA091ATakeProfitThatSoldNothingKeepsItsGrade(t *testing.T) {
	t.Run("zero floor", func(t *testing.T) {
		r := a091Harness(t, true, a091Zero(riskcalc.FloorBoundSellable), nil)
		r.entry("005930", "10", "70000", "68000", "70000")
		r.quote("005930", 72000) // +1.0R → 40% 익절
		r.observe()
		if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 0 {
			t.Fatalf("rows of the new kind = %d for a take-profit, want 0", n)
		}
		lines := r.log.ofEvent(t, obs.EventExitProposalCapped)
		if len(lines) != 1 {
			t.Fatalf("capped lines = %d, want 1", len(lines))
		}
		want := "005930 · " + r.clk.Now().UTC().Format(time.RFC3339) + " 관측: 매도가능 수량이 0이다 — 다른 미체결 매도가 주식을 잡고 있을 수 있다. " +
			"제안 4주 중 0주를 제출했다 — 청산이 나가지 않았다. 불일치가 해소되면 같은 단계를 다시 제안한다."
		if lines[0]["detail"] != want {
			t.Errorf("take-profit body:\n got %v\nwant %s — a take-profit is not a stop", lines[0]["detail"], want)
		}
	})
	t.Run("floor cannot be computed", func(t *testing.T) {
		h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
			o.Floor = a091Failing(errors.New("holdings read failed"))
			o.NotificationsEnabled = true
		})
		h.entry("005930", "10", "70000", "68000", "70000")
		h.quote("005930", 72000)
		h.observe()
		if n := len(h.alerts.events); n != 0 {
			t.Fatalf("alerts = %+v, want none — a take-profit B2 only logs", h.alerts.events)
		}
	})
}

// --- 3.5 — 관측 결과 무변화(§0.3 · §0.9) — 0주 뒤 레벨 해제 · 하한이 풀리면 같은 레벨 전량 재발의 ------------------------------

func TestA091TheOutcomeIsUnchanged(t *testing.T) {
	for _, c := range []struct {
		name  string
		floor func(lifted *bool) engine.FloorSource
	}{
		{"zero floor", func(lifted *bool) engine.FloorSource {
			return &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
				return riskcalc.ConfirmedFloor{Quantity: "0", Bound: riskcalc.FloorBoundSellable}, !*lifted, nil
			}}
		}},
		{"floor cannot be computed", func(lifted *bool) engine.FloorSource {
			return &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
				if *lifted {
					return riskcalc.ConfirmedFloor{}, false, nil
				}
				return riskcalc.ConfirmedFloor{}, true, errors.New("holdings read failed")
			}}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			lifted := false
			r := a091Harness(t, true, c.floor(&lifted), nil)
			p := r.breach()
			r.observe()
			if len(r.submit.places) != 0 {
				t.Fatal("a zero floor authorises no sale")
			}
			if r.state(p.ID).Pending() {
				t.Fatal("the level must stay proposable, so the armed proposal is released")
			}
			lifted = true
			r.observe()
			if len(r.submit.places) != 1 || r.submit.places[0].Intent.Quantity != 10 {
				t.Fatalf("places = %+v after the floor lifted, want the whole 10 re-proposed", r.submit.places)
			}
		})
	}
}

// --- 3.7 — H2: 한 사건의 로그 줄과 알림은 같은 종류 ---------------------------------------------------------------------------

func TestA091TheLogAndTheAlertAreOneKind(t *testing.T) {
	for _, c := range []struct {
		name    string
		enabled bool
		want    obs.EventType
		other   obs.EventType
	}{
		{"alerts on", true, obs.EventExitStopSoldNothing, obs.EventExitProposalCapped},
		{"alerts off", false, obs.EventExitProposalCapped, obs.EventExitStopSoldNothing},
	} {
		t.Run(c.name, func(t *testing.T) {
			for _, floor := range []engine.FloorSource{a091Failing(errors.New("holdings read failed")), a091Zero(riskcalc.FloorBoundSellable)} {
				r := a091Harness(t, c.enabled, floor, nil)
				r.breach()
				r.observe()
				if n := len(r.log.ofEvent(t, c.want)); n == 0 {
					t.Fatalf("no %s log line:\n%s", c.want, r.log.String())
				}
				if n := len(r.log.ofEvent(t, c.other)); n != 0 {
					t.Errorf("%d %s log lines — the same event under two kinds:\n%s", n, c.other, r.log.String())
				}
			}
		})
	}
}

// --- 3.8 — 계좌 카나리: 보고 · 기록 실패 · 승격 줄 · 행 어디에도 계좌 원문이 없음 --------------------------------------------------

func TestA091NoLineOrRowCarriesTheAccount(t *testing.T) {
	for _, arm := range []struct {
		name       string
		enabled    bool
		takeProfit bool
	}{{"alerts off, protective", false, false}, {"alerts on, take-profit", true, true}, {"alerts off, take-profit", false, true}} {
		t.Run("B2 line: "+arm.name, func(t *testing.T) {
			r := a091Harness(t, arm.enabled, a091Failing(errors.New("holdings read failed for "+exitAccount)), nil)
			r.entry("005930", "10", "70000", "68000", "70000")
			if arm.takeProfit {
				r.quote("005930", 72000)
			} else {
				r.quote("005930", 67900)
			}
			r.observe()
			if len(r.log.ofEvent(t, obs.EventExitProposalCapped)) == 0 {
				t.Fatalf("arrangement: no B2 line:\n%s", r.log.String())
			}
			if strings.Contains(r.log.String(), exitAccount) {
				t.Errorf("the log carries the account %q:\n%s", exitAccount, r.log.String())
			}
		})
	}
	t.Run("recorded", func(t *testing.T) {
		r := a091Harness(t, true, a091Failing(errors.New("holdings read failed for "+exitAccount)), nil)
		r.breach()
		r.observe()
		if errs := r.log.ofEvent(t, obs.EventExitStopSoldNothing); len(errs) == 0 || func() bool {
			for _, l := range errs {
				if l["level"] == "ERROR" && strings.Contains(fmt.Sprint(l["error"]), "[account]") {
					return false
				}
			}
			return true
		}() {
			t.Errorf("no masked B2 error line of the new kind:\n%s", r.log.String())
		}
		rows, _ := r.journal.PendingAlerts(context.Background(), 0)
		if len(rows) == 0 {
			t.Fatal("arrangement: no row was recorded")
		}
		for _, row := range rows {
			if strings.Contains(row.Title+row.Body+row.Payload, exitAccount) {
				t.Errorf("row %s carries the account", row.EventKey)
			}
		}
		if strings.Contains(r.log.String(), exitAccount) {
			t.Errorf("the log carries the account %q:\n%s", exitAccount, r.log.String())
		}
	})
	t.Run("the record fails", func(t *testing.T) {
		closed, err := journal.Open(context.Background(), journal.Options{
			Path: filepath.Join(t.TempDir(), "closed.db"), Clock: clock.NewFake(exitNow),
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
		})
		if err != nil {
			t.Fatal(err)
		}
		_ = closed.Close()
		r := a091Harness(t, true, a091Failing(errors.New("holdings read failed for "+exitAccount)), nil)
		r.notifier.Journal = closed // 기록이 반드시 실패
		r.breach()
		r.observe()
		if rej := r.gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonAlertUndelivered {
			t.Fatalf("arrangement: the record did not fail (gate = %v)", rej)
		}
		if !strings.Contains(r.log.String(), "did not reach the operating mode") {
			t.Fatalf("arrangement: the escalation line is missing:\n%s", r.log.String())
		}
		if !strings.Contains(r.log.String(), "the report that the stop sold nothing could not be made durable") {
			t.Errorf("the report's own failure line is missing:\n%s", r.log.String())
		}
		if strings.Contains(r.log.String(), exitAccount) {
			t.Errorf("the failure path logs the account %q:\n%s", exitAccount, r.log.String())
		}
	})
}

// --- 3.9 — 겹침: 하한 조회의 401 은 모드 통지 기록 하나 + 새 종류 기록 하나, 그 밖은 0 ------------------------------------------

func TestA091AnAuthRefusalOnTheFloorReadRecordsTwo(t *testing.T) {
	var retrier *execgw.Retrier
	floor := &a091Floor{fn: func(ctx context.Context) (riskcalc.ConfirmedFloor, bool, error) {
		err := retrier.Query(ctx, execgw.QueryHoldings, func(context.Context) error { return official.ErrAuth })
		return riskcalc.ConfirmedFloor{}, true, fmt.Errorf("engine: reading the holding of 005930 for the confirmed floor: %w", err)
	}}
	r := a091Harness(t, true, floor, func(o *engine.ExitObserverOptions) {
		retrier = o.Retrier
	})
	ro := obs.RecordOnly{N: r.notifier}
	retrier.Escalate, retrier.AccountRef, retrier.Announcer = r.journal, exitAccount, ro
	r.breach()
	r.observe()
	all, _ := r.journal.PendingAlerts(context.Background(), 0)
	counts := map[string]int{}
	for _, a := range all {
		counts[a.Type]++
	}
	if counts[string(obs.EventOperatingMode)] != 1 || counts[string(obs.EventExitStopSoldNothing)] != 1 || len(all) != 2 {
		t.Fatalf("rows by type = %v, want one mode notice and one %s", counts, obs.EventExitStopSoldNothing)
	}
}

// --- 3.10 — 원인 계약: 행은 첫 원인, 로그는 관측마다 ------------------------------------------------------------------------

func TestA091TheRowKeepsTheFirstCauseAndTheLogKeepsEach(t *testing.T) {
	for _, c := range []struct {
		name          string
		first, second string // "b2" | "sellable"
	}{
		{"B2 then zero floor", "b2", "sellable"},
		{"zero floor then B2", "sellable", "b2"},
	} {
		t.Run(c.name, func(t *testing.T) {
			step := c.first
			floor := &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
				if step == "b2" {
					return riskcalc.ConfirmedFloor{}, true, errors.New("holdings read failed")
				}
				return riskcalc.ConfirmedFloor{Quantity: "0", Bound: riskcalc.FloorBoundSellable}, true, nil
			}}
			r := a091Harness(t, true, floor, nil)
			r.breach()
			r.observe()
			step = c.second
			r.clk.Advance(5 * time.Second)
			r.observe()

			cause := map[string]string{"b2": "확정 하한을 계산하지 못했다", "sellable": "매도가능 수량이 0"}
			rows := r.rows(obs.EventExitStopSoldNothing)
			if len(rows) != 1 {
				t.Fatalf("rows = %d, want 1 — one episode", len(rows))
			}
			if !strings.Contains(rows[0].Body, cause[c.first]) || strings.Contains(rows[0].Body, cause[c.second]) {
				t.Errorf("row body should keep the first cause %q:\n%s", cause[c.first], rows[0].Body)
			}
			var details []string
			for _, l := range r.log.ofEvent(t, obs.EventExitStopSoldNothing) {
				if d, ok := l["detail"].(string); ok && strings.Contains(d, "관측") {
					details = append(details, d)
				}
			}
			if len(details) != 2 || !strings.Contains(details[0], cause[c.first]) || !strings.Contains(details[1], cause[c.second]) {
				t.Errorf("per-observation log details = %q, want the %s cause then the %s cause", details, c.first, c.second)
			}
		})
	}
}

// --- 3.2b — 생산 배선은 로드된 설정의 notifications.enabled 로 옵션을 덮는다(호출자 값 무시) ------------------------------------

func TestA091TheProductionAssemblyReadsTheLoadedSwitch(t *testing.T) {
	for _, loaded := range []bool{true, false} {
		t.Run(fmt.Sprintf("loaded=%v", loaded), func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, smallLiveGate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, nil)
			srv, _ := interlockServer(t, "123-45")
			// 조립이 Logger 를 받아 Context.Log 로 놓는 생산 경로 그대로 — 시험이 eng.Log 를 덮지 않는다(i2 보이스 A).
			eng, err := openProtectedGateEngineLogging(t, dir, srv, nil, &a091Buf{})
			if err != nil {
				t.Fatalf("production assembly: %v", err)
			}
			if eng.Log == nil {
				t.Fatal("the assembly dropped its logger — the a091 lines would have no sink")
			}
			eng.Config.Engine.Notifications.Enabled = loaded
			observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel(), NotificationsEnabled: !loaded})
			if err != nil {
				t.Fatalf("ExitObserver: %v", err)
			}
			if got := observer.OptionsForTest().ZeroFloorLog; got == nil || got != eng.Log {
				t.Errorf("ZeroFloorLog = %p, want the engine logger %p — the alerts-off B2 line must reach a sink", got, eng.Log)
			}
			if got := observer.OptionsForTest().NotificationsEnabled; got != loaded {
				t.Errorf("NotificationsEnabled = %v, want the loaded %v — the caller's value must not win", got, loaded)
			}
		})
	}
}

// --- 3.3a 보강 — 한정 항 다섯 전부: Holdings 만 옛 종류, 나머지 넷은 새 종류 · 원인 문구 · payload 고정 ------------------------------

func TestA091EveryFloorBoundHasItsGradeAndItsWords(t *testing.T) {
	for _, c := range []struct {
		bound, text string
		critical    bool
	}{
		{riskcalc.FloorBoundHoldings, "계좌에 보유가 없다 — 엔진 밖에서 종결되는 중일 수 있다", false},
		{riskcalc.FloorBoundSellable, "매도가능 수량이 0이다 — 다른 미체결 매도가 주식을 잡고 있을 수 있다", true},
		{riskcalc.FloorBoundLocalSells, "엔진의 미체결 매도가 남은 수량을 모두 쓰고 있다", true},
		{riskcalc.FloorBoundNoSnapshot, "계좌 스냅숏을 읽지 못했다", true},
		{riskcalc.FloorBoundStaleSnapshot, "계좌 스냅숏이 낡았다", true},
	} {
		t.Run(c.bound, func(t *testing.T) {
			r := a091Harness(t, true, a091Zero(c.bound), nil)
			p := r.breach()
			r.observe()
			rows := r.rows(obs.EventExitStopSoldNothing)
			if !c.critical {
				if len(rows) != 0 {
					t.Fatalf("rows = %d for %s, want 0 — the account holds none", len(rows), c.bound)
				}
				return
			}
			if len(rows) != 1 {
				t.Fatalf("rows = %d for %s, want 1 — the stop failed", len(rows), c.bound)
			}
			if !strings.Contains(rows[0].Body, " 관측: "+c.text+". ") {
				t.Errorf("body does not carry the %s cause %q:\n%s", c.bound, c.text, rows[0].Body)
			}
			var payload map[string]any
			if err := json.Unmarshal([]byte(rows[0].Payload), &payload); err != nil {
				t.Fatalf("payload: %v", err)
			}
			if payload["cause"] != "floor_zero" || payload["floor_bound"] != c.bound || payload["position_id"] != p.ID {
				t.Errorf("payload = %v, want cause floor_zero, floor_bound %s, the position", payload, c.bound)
			}
		})
	}
	t.Run("B2", func(t *testing.T) {
		r := a091Harness(t, true, a091Failing(errors.New("x")), nil)
		r.breach()
		r.observe()
		rows := r.rows(obs.EventExitStopSoldNothing)
		if len(rows) != 1 {
			t.Fatalf("rows = %d, want 1", len(rows))
		}
		var payload map[string]any
		_ = json.Unmarshal([]byte(rows[0].Payload), &payload)
		if _, has := payload["floor_bound"]; payload["cause"] != "floor_unknown" || has {
			t.Errorf("payload = %v, want cause floor_unknown and no floor_bound", payload)
		}
	})
}

// 종료 취소는 알림 없이 로그 한 줄 — cause shutdown(알림에 닿지 않는 유일한 원인 표지).
func TestA091AShutdownLeavesOnlyALine(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := a091Harness(t, true, &a091Floor{fn: func(context.Context) (riskcalc.ConfirmedFloor, bool, error) {
		cancel()
		return riskcalc.ConfirmedFloor{}, true, fmt.Errorf("x: %w", context.Canceled)
	}}, nil)
	r.breach()
	r.observer.ObserveOnce(ctx)
	lines := r.log.ofEvent(t, obs.EventExitProposalCapped)
	if len(lines) != 1 || lines[0]["cause"] != "shutdown" {
		t.Fatalf("lines = %v, want exactly one capped line with cause shutdown", lines)
	}
}
