//go:build tossos_testseams

package engine

// a137 재리뷰 P1-N1: 측정 불가의 조용한 종료도 흔적을 남김 — 시장별 계수 · 첫 원인(계좌 값 · 원문 오류 없음)이 감독자 스냅숏에,
// 알림 줄은 계좌 일 키마다 한 번(주기마다가 아님). 잠금 · 삼킴 계수 · 운영 모드는 여전히 건드리지 않음.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
)

type a137FlatCampaigns struct{}

func (a137FlatCampaigns) CurrentPositionCampaignCAS(context.Context, string, string, string) (journal.PositionCampaignCASRead, error) {
	return journal.PositionCampaignCASRead{State: "FLAT"}, nil
}

type a137RefusingDispatcher struct{ err error }

func (d a137RefusingDispatcher) dispatch(context.Context, strategyhandoff.Delivered) (execgw.Outcome, error) {
	return execgw.Outcome{}, d.err
}

func TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay(t *testing.T) {
	unmeasurable := errors.Join(&execgw.IssueRefusal{Stage: execgw.StageChain, Reason: "INPUT_UNAVAILABLE"},
		&execgw.DailyLossUnmeasurable{Cause: errors.New("dailyloss: today's realised loss cannot be measured: KR 005930 pos-1a2b3c4d — closed today without a frozen outcome row")})
	for _, effective := range []bool{true, false} {
		var calls atomic.Int32
		var mu sync.Mutex
		var notices []string
		cycle := func(ctx context.Context) error {
			calls.Add(1)
			return dispatchStrategyMarketHandoffs(ctx, a137FlatCampaigns{}, a137RefusingDispatcher{err: unmeasurable},
				strategyhandoff.AdmitEachOwnerScope(true, a137Selected()))
		}
		fake := clock.NewFake(time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC)) // KST 10:00 — 일 키 10-14
		worker := func(market StrategyMarket) StrategyMarketWorker {
			if !effective {
				return StrategyMarketWorker{Market: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: cycle}
			}
			return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
				AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
				EvidenceDigest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				LatchRevision:      1, PollInterval: DefaultStrategyCycleLimit}
		}
		supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: fake, CycleLimit: MaximumStrategyCycleLimit,
			Workers: []StrategyMarketWorker{worker(StrategyMarketKR), worker(StrategyMarketUS)},
			DailyLossNotice: func(market StrategyMarket, day, cause string) {
				mu.Lock()
				defer mu.Unlock()
				if market == StrategyMarketKR {
					notices = append(notices, day+" "+cause)
				}
			}})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- supervisor.Run(ctx) }()
		<-supervisor.Ready()
		advance := func(until int32) {
			deadline := time.Now().Add(3 * time.Second)
			for calls.Load() < until && time.Now().Before(deadline) {
				fake.Advance(DefaultStrategyCycleLimit)
				time.Sleep(time.Millisecond)
			}
			time.Sleep(20 * time.Millisecond)
		}
		advance(6)
		snapshot, _ := supervisor.Snapshot(StrategyMarketKR)
		mu.Lock()
		first := append([]string(nil), notices...)
		mu.Unlock()
		if snapshot.Latched || snapshot.SwallowedCycleErrors != 0 || snapshot.DailyLossUnmeasurableCycles < 2 ||
			!strings.Contains(snapshot.FirstDailyLossUnmeasurable, "closed today without a frozen outcome row") ||
			strings.Contains(snapshot.FirstDailyLossUnmeasurable, "acct") {
			t.Fatalf("effective=%v snapshot=%+v — want counted, not latched, not swallowed, no account text", effective, snapshot)
		}
		if len(first) != 1 || !strings.HasPrefix(first[0], "2026-10-14 ") {
			t.Fatalf("effective=%v notices=%v — want exactly one for account day 2026-10-14", effective, first)
		}
		// KST 08:00 을 넘겨 다음 일 키로 → 한 번 더.
		fake.Set(time.Date(2026, 10, 14, 23, 0, 0, 0, time.UTC))
		advance(calls.Load() + 3)
		mu.Lock()
		second := append([]string(nil), notices...)
		mu.Unlock()
		cancel()
		<-done
		if len(second) != 2 || !strings.HasPrefix(second[1], "2026-10-15 ") {
			t.Fatalf("effective=%v notices=%v — want one more for account day 2026-10-15", effective, second)
		}
	}
}

// 생산 조립의 알림은 경고 줄 하나 — 이벤트 이름 · 시장 · 계좌 일 키 · 매도 무관 문구, 계좌 값 없음. 로그가 없으면 알림 자리도 없음.
func TestA137TheProductionNoticeIsOneWarningLine(t *testing.T) {
	if (&Context{}).strategyDailyLossNotice() != nil {
		t.Fatal("without a logger there must be no notice seam")
	}
	var buf strings.Builder
	c := &Context{Log: obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true})}
	c.strategyDailyLossNotice()(StrategyMarketKR, "2026-10-14", "dailyloss: today's realised loss cannot be measured: KR 005930 pos-1a2b3c4d — closed today without a frozen outcome row")
	line := buf.String()
	for _, want := range []string{string(obs.EventDailyLossUnmeasurable), `"market":"KR"`, `"account_day":"2026-10-14"`, "sells, stops and exits are unaffected"} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q lacks %q", line, want)
		}
	}
	if strings.Count(line, "\n") != 1 {
		t.Fatalf("want exactly one line, got %q", line)
	}
}

// 생산 조립이 알림 자리를 실제로 감독자에 넘김(배선 한 줄을 못 박음).
func TestA137TheProductionSupervisorCarriesTheNotice(t *testing.T) {
	var buf strings.Builder
	c := &Context{Log: obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true}),
		Entry: execgw.NewEntryGate(clock.NewFake(time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC)), nil)}
	supervisor, err := c.NewRefreshingPairedStrategyEntrySupervisor(clock.NewFake(time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	if supervisor.dailyLossNotice == nil {
		t.Fatal("the production supervisor has no daily-loss notice seam")
	}
}
