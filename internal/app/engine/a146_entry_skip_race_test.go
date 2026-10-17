//go:build tossos_testseams

package engine

// a146 — 경합 검출기 아래서 도는 감독자 스킵 시험(Makefile RACE_ENGINE_FILES · RACE_ENGINE_TESTS, 구현 리뷰 P3-2): 스킵 기록은 감독자 잠금 아래에서
// 일어나고 알림은 잠금 밖에서 불리므로, 잠금을 지우는 변이는 -race 만 잡음.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// D3: 감독자가 주기 ctx 에 실은 기록자로 스킵을 세고(주기마다), 첫 원인을 지키며, 알림은 (계좌 일 키, 종목)마다 한 번.
func TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol(t *testing.T) {
	for _, effective := range []bool{true, false} {
		var calls atomic.Int32
		var mu sync.Mutex
		var notices, usNotices []string
		cycle := func(ctx context.Context) error {
			calls.Add(1)
			recordStrategyEntrySkip(ctx, a146Skip("005930"))
			recordStrategyEntrySkip(ctx, a146Skip("000660"))
			return nil
		}
		fake := clock.NewFake(time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)) // KST 10:00 — 일 키 10-17
		// US 주기도 자기 종목 스킵을 기록함 — 시장 귀속(기록자가 자기 worker 에 셈)을 재게(구현 리뷰 P3-1). calls 는 KR 주기만 셈.
		usCycle := func(ctx context.Context) error {
			recordStrategyEntrySkip(ctx, &strategyEntrySkip{market: strategyrouter.MarketUS, symbol: "AAPL", entry: "100.00", stop: "95.00", fx: "USD->KRW sha256:fx",
				cause: errors.New("execgw: q_final refused: EXISTING_GUARDIAN_CAP")})
			return nil
		}
		worker := func(market StrategyMarket, cycle StrategyCycle) StrategyMarketWorker {
			if !effective {
				return StrategyMarketWorker{Market: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: cycle}
			}
			return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
				AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
				EvidenceDigest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				LatchRevision:      1, PollInterval: DefaultStrategyCycleLimit}
		}
		supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: fake, CycleLimit: MaximumStrategyCycleLimit,
			Workers: []StrategyMarketWorker{worker(StrategyMarketKR, cycle), worker(StrategyMarketUS, usCycle)},
			EntrySkipNotice: func(market StrategyMarket, day, symbol, cause string) {
				mu.Lock()
				defer mu.Unlock()
				if market == StrategyMarketKR {
					notices = append(notices, day+" "+symbol+" "+cause)
				} else {
					usNotices = append(usNotices, string(market)+" "+symbol)
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
		advance(4)
		snapshot, _ := supervisor.Snapshot(StrategyMarketKR)
		mu.Lock()
		first := append([]string(nil), notices...)
		mu.Unlock()
		if snapshot.Latched || snapshot.SwallowedCycleErrors != 0 || snapshot.GuardianCapZeroSkips < 6 ||
			!strings.Contains(snapshot.FirstGuardianCapZeroSkip, "005930") || strings.Contains(snapshot.FirstGuardianCapZeroSkip, "000660") {
			t.Fatalf("effective=%v snapshot=%+v — want ≥6 skips counted, the first cause kept, no latch, nothing swallowed", effective, snapshot)
		}
		if len(first) != 2 || !strings.HasPrefix(first[0], "2026-10-17 005930 ") || !strings.HasPrefix(first[1], "2026-10-17 000660 ") {
			t.Fatalf("effective=%v notices=%v — want exactly one per symbol for account day 2026-10-17", effective, first)
		}
		usSnapshot, _ := supervisor.Snapshot(StrategyMarketUS)
		mu.Lock()
		us := append([]string(nil), usNotices...)
		mu.Unlock()
		if strings.Join(us, ",") != "US AAPL" || usSnapshot.GuardianCapZeroSkips == 0 || !strings.Contains(usSnapshot.FirstGuardianCapZeroSkip, "AAPL") ||
			strings.Contains(snapshot.FirstGuardianCapZeroSkip, "AAPL") {
			t.Fatalf("effective=%v us notices=%v us=%+v kr=%+v — want US skips counted and announced on the US market only", effective, us, usSnapshot, snapshot)
		}
		// KST 08:00 을 넘겨 다음 일 키 → 종목마다 한 번 더.
		fake.Set(time.Date(2026, 10, 17, 23, 0, 0, 0, time.UTC))
		advance(calls.Load() + 3)
		mu.Lock()
		second := append([]string(nil), notices...)
		mu.Unlock()
		cancel()
		<-done
		if len(second) != 4 || !strings.HasPrefix(second[2], "2026-10-18 005930 ") || !strings.HasPrefix(second[3], "2026-10-18 000660 ") {
			t.Fatalf("effective=%v notices=%v — want one more per symbol for account day 2026-10-18", effective, second)
		}
	}
}

// F5 뒤집힘(a146): 권한 갱신 전용(refreshOnly — 오늘 생산 감독자) 시장 — 삼킴 계수가 오르지 않고, 뒤 범위가 굶지 않음.
// (편집 전 HEAD 에서는 매 주기 삼킴 + precheck 1 회였음.)
func TestA146ACapZeroEntryIsNotASwallowedFaultOnARefreshOnlyMarket(t *testing.T) {
	fixture, counter := a146CapZeroFixture(t)
	snapshot, calls, _ := a146RunMarket(t, fixture, false, 3)
	if snapshot.Latched || calls < 3 || snapshot.SwallowedCycleErrors != 0 || snapshot.FirstSwallowedFailure != "" {
		t.Fatalf("cycles=%d snapshot=%+v — want nothing swallowed for cap-0 skips", calls, snapshot)
	}
	if got := counter.prechecks.Load(); got != 2*calls || snapshot.GuardianCapZeroSkips != uint64(2*calls) {
		t.Fatalf("prechecks=%d skips=%d cycles=%d — want the second scope evaluated every cycle", got, snapshot.GuardianCapZeroSkips, calls)
	}
}
