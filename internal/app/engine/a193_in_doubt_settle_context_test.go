package engine_test

// a193 — production 조립(Context.InDoubtSettler)을 **거쳐** 만든 실행자를 실제로 돌리는 시험(Manager 검증 반려 1회차: 조립 경로 공백).
//
// 다른 a193 시험은 newInDoubtSettle(deps) 로 직접 조립하므로, Context 의 원장 · 해소기 · 알림기 · 계좌가 실행자에 실리는지는 여기서만 잼.
// 실 원장(exitHarness) + 실 execgw.Resolver(가짜 목록) + 실 obs.Notifier(원장 outbox) 로 조립함. 값은 합성(종목·계좌·주문번호 아님).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const a193CtxSymbol = "Z00002"

// a193CtxPager 는 주문 목록 가짜임 — open 이 있으면 OPEN 목록에 싣고, err 가 있으면 매번 실패.
type a193CtxPager struct {
	open []json.RawMessage
	err  error
}

func (p *a193CtxPager) OrdersPageRaw(_ context.Context, q execgw.OrderQuery, _ string) (execgw.OrderPage, error) {
	if p.err != nil {
		return execgw.OrderPage{}, p.err
	}
	if q.Status == "OPEN" {
		return execgw.OrderPage{Orders: p.open}, nil
	}
	return execgw.OrderPage{}, nil
}

// a193ContextSettler 는 production 조립 함수로 실행자를 만듦 — 시험이 조종석에 앉지 않게 Context 만 채움.
func a193ContextSettler(h *exitHarness, account string, pager execgw.OrderPager, cfg execgw.ResolveConfig) engine.AuxiliaryExecutor {
	ectx := &engine.Context{
		Journal:    h.journal,
		Resolver:   &execgw.Resolver{Journal: h.journal, Orders: pager, Order: nil, Clock: h.clk, Gate: h.gate, Config: cfg},
		Notifier:   &obs.Notifier{Journal: h.journal, Gate: h.gate, AccountRef: exitAccount},
		AccountRef: account,
	}
	return ectx.InDoubtSettler(h.clk)
}

// a193InDoubtSell 은 합성 종목의 매도 PLACE 를 IN_DOUBT 로 남기고 그 attempt id 를 돌려줌.
func a193InDoubtSell(h *exitHarness, intentID string) string {
	h.t.Helper()
	h.a094SellAttemptOn(a193CtxSymbol, intentID, journal.StateInDoubt)
	pending, err := h.journal.PendingAttempts(context.Background())
	if err != nil {
		h.t.Fatal(err)
	}
	for _, rec := range pending {
		if rec.IntentID == intentID {
			return rec.ID
		}
	}
	h.t.Fatalf("no attempt for %s", intentID)
	return ""
}

// a193RunSettler 는 실행자를 goroutine 에서 돌리고, 정리(취소 · 반환 확인)를 Cleanup 에 둠. Run 이 일찍 돌아오면(배선 누락) 그 오류를 돌려주는 채널.
func a193RunSettler(t *testing.T, exec engine.AuxiliaryExecutor) <-chan error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- exec.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("the settler did not return after cancellation")
		}
	})
	return done
}

func a193OutboxKeys(t *testing.T, h *exitHarness) []string {
	t.Helper()
	alerts, err := h.journal.PendingAlerts(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(alerts))
	for _, a := range alerts {
		keys = append(keys, a.EventKey)
	}
	return keys
}

// a193WaitFor 는 cond 가 참이 될 때까지 기다림. drive 가 참이면 해소 · 실행자의 가짜 시계 대기를 5초씩 밀어 줌. Run 이 일찍 반환하면 실패.
func a193WaitFor(t *testing.T, h *exitHarness, done <-chan error, drive bool, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for !cond() {
		select {
		case err := <-done:
			t.Fatalf("the settler stopped before %s: %v", what, err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		if drive && h.clk.WaitForSleepers(1, 20*time.Millisecond) {
			h.clk.Advance(5 * time.Second)
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func a193Contains(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

func a193AttemptState(t *testing.T, h *exitHarness, id string) journal.AttemptState {
	t.Helper()
	rec, err := h.journal.LookupAttempt(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return rec.State
}

func TestA193TheContextSettlerRecordsFirstSightInTheOutbox(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	id := a193InDoubtSell(h, "exit-a193-ctx-1")
	done := a193RunSettler(t, a193ContextSettler(h, exitAccount, &a193CtxPager{err: errors.New("synthetic 5xx")}, execgw.ResolveConfig{}))
	key := string(obs.EventOrderInDoubt) + "|attempt:" + id
	a193WaitFor(t, h, done, false, "the first-sight outbox row", func() bool { return a193Contains(a193OutboxKeys(t, h), key) })
}

func TestA193TheContextSettlerResolvesTheAttempt(t *testing.T) {
	h, _, _ := a094Harness(t, nil)
	id := a193InDoubtSell(h, "exit-a193-ctx-2")
	order := json.RawMessage(fmt.Sprintf(`{"orderId":"ord-synthetic-ctx","symbol":%q,"side":"SELL","status":"PENDING","quantity":"10","price":"67000","orderedAt":%q}`,
		a193CtxSymbol, h.clk.Now().UTC().Format(time.RFC3339)))
	done := a193RunSettler(t, a193ContextSettler(h, exitAccount, &a193CtxPager{open: []json.RawMessage{order}}, execgw.ResolveConfig{}))
	a193WaitFor(t, h, done, true, "the attempt to settle", func() bool { return a193AttemptState(t, h, id) == journal.StateConfirmed })
}

func TestA193TheContextSettlerFindsTheArmedProposalOfItsAccount(t *testing.T) {
	cases := []struct {
		name      string
		account   string
		wantAlert bool
	}{
		// 계좌가 실려야 무장 발의를 찾음 → exit 발의의 park 는 기존 recordParkCause 몫이라 실행자는 알리지 않음.
		{"the engine account finds the proposal", exitAccount, false},
		// 계좌가 비면 발의를 못 찾음 → 실행자가 park 알림을 냄(= 같은 park 의 중복 알림). 두 경우가 구별됨을 보이는 대조군.
		{"an empty account cannot find it", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _, _ := a094Harness(t, nil)
			p := h.entry(a193CtxSymbol, "10", "70000", "68000", "70000")
			h.quote(a193CtxSymbol, 70100)
			h.observe()
			intent := "exit-a193-ctx-armed"
			h.a094ArmOn(p, a193CtxSymbol, intent, string(exitpolicy.ActionBaselineBreach))
			id := a193InDoubtSell(h, intent)
			done := a193RunSettler(t, a193ContextSettler(h, tc.account, &a193CtxPager{}, execgw.ResolveConfig{}))
			a193WaitFor(t, h, done, true, "the park", func() bool { return a193AttemptState(t, h, id) == journal.StateUnresolvedInDoubt })
			// park 뒤 실행자가 그 결정(알림 여부)을 마치고 다음 대기에 들 때까지 — 그 전에는 알림이 아직 안 났을 수 있음.
			if !h.clk.WaitForSleepers(1, 10*time.Second) {
				t.Fatal("the settler never went back to sleep after the park")
			}
			parkKey := string(obs.EventOrderUnresolved) + "|attempt:" + id
			keys := a193OutboxKeys(t, h)
			if got := a193Contains(keys, parkKey); got != tc.wantAlert {
				t.Fatalf("settler park alert present = %v, want %v (outbox %v)", got, tc.wantAlert, keys)
			}
			if !a193Contains(keys, string(obs.EventOrderInDoubt)+"|attempt:"+id) {
				t.Fatalf("first-sight row missing (outbox %v)", keys)
			}
		})
	}
}

func TestA193TheContextSettlerDerivesTheBoundFromTheResolverBudget(t *testing.T) {
	cases := []struct {
		name      string
		cfg       execgw.ResolveConfig
		wantStale bool
	}{
		// 해소 예산 1분 → 10분 상한이 아니라 그 두 배 2분.
		{"a one-minute budget makes the bound two minutes", execgw.ResolveConfig{MaxDuration: time.Minute}, true},
		// 예산을 안 주면 기본 5분 → 10분 — 2분에는 알리지 않음.
		{"the default budget keeps ten minutes", execgw.ResolveConfig{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _, _ := a094Harness(t, nil)
			id := a193InDoubtSell(h, "exit-a193-ctx-bound")
			h.clk.Advance(2 * time.Minute)
			done := a193RunSettler(t, a193ContextSettler(h, exitAccount, &a193CtxPager{err: errors.New("synthetic 5xx")}, tc.cfg))
			first := string(obs.EventOrderInDoubt) + "|attempt:" + id
			// 1단(처음 봄 → 10분 검사)이 끝난 뒤를 처음 봄 행으로 판단함 — 같은 스윕에서 처음 봄 바로 뒤에 10분 검사가 돎.
			a193WaitFor(t, h, done, false, "the first sweep", func() bool { return a193Contains(a193OutboxKeys(t, h), first) })
			if !h.clk.WaitForSleepers(1, 10*time.Second) {
				t.Fatal("the first sweep never reached the resolution's sleep")
			}
			stale := a193Contains(a193OutboxKeys(t, h), string(obs.EventOrderInDoubt)+"|stale:"+id)
			if stale != tc.wantStale {
				t.Fatalf("stale alert at two minutes = %v, want %v", stale, tc.wantStale)
			}
		})
	}
}

// 조립이 원장 · 해소기를 실어야 실행자가 돈다 — 둘 중 하나가 빠지면 Run 이 바로 반환함(위 시험들이 잡음). 여기서는 이름으로 그 반환을 고정.
func TestA193TheContextSettlerStopsLoudlyWhenTheContextIsUnwired(t *testing.T) {
	exec := (&engine.Context{}).InDoubtSettler(nil)
	err := exec.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no journal or resolver") {
		t.Fatalf("an unwired context settler returned %v", err)
	}
}
