package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 이 파일은 태스크 5.2 의 "합치기" 절을 **값으로** 확인한다. 옆 파일
// a112_refresh_wave_test.go 는 구조로 확인한다 — 잠금을 든 채 원격을 부르는
// 자리가 없다는 것. 구조만으로는 부족하다: 잠금을 안 들고도 두 시장이 각자
// 원격을 두 번 탈 수 있다.
//
// **왜 sleep 이 하나도 없는가.** `testing/synctest` 의 거품 안에서
// `synctest.Wait()` 는 다른 goroutine 이 전부 **확실히 멈춰 설 때까지**
// 기다린다. 그래서 "기다리는 시장이 아직 안 돌아왔다"를 시간이 아니라 상태로
// 확인할 수 있다.
//
// 그리고 이 도구가 여기서 특히 맞는 이유를 쟀다: **뮤텍스에 걸린 goroutine 은
// "확실히 멈춰 선" 것으로 치지 않는다**(Go 1.26 에서 측정 — `sync.Mutex.Lock`
// 에 걸린 goroutine 이 있으면 `synctest.Wait()` 가 영원히 돌아오지 않는다).
// 즉 아래 시험들은 "채널에서 기다린다"와 "잠금 뒤에 줄 서 있다"를 **구별한다**.
// 파도를 다시 잠금 안으로 넣는 편집은 이 파일을 초록으로 통과할 수 없다.

// refreshWaveSentinel 은 파도가 실어 나른 것을 알아볼 수 있게 만든 표식이다.
// 기다린 시장이 이 값을 받았다면 그것은 지도자의 파도를 탄 것이고, 자기 파도를
// 돌았다면 절대 이 값이 나오지 않는다(빈 Context 의 실제 수집은 달력·FX·저널을
// 하나도 못 읽어 다른 값을 만든다).
func refreshWaveSentinel() StrategyEntryProductionAssembly {
	return StrategyEntryProductionAssembly{
		Schedule: PairedStrategyScheduleSnapshot{
			KR: StrategyScheduleMarketSnapshot{Market: StrategyMarketKR, CalendarVersion: "WAVE-SENTINEL"},
		},
	}
}

func refreshWaveFixture() (*Context, *clock.Fake, time.Time) {
	fake := clock.NewFake(time.Date(2026, 9, 3, 2, 0, 0, 0, time.UTC))
	return &Context{}, fake, fake.Now().UTC()
}

// TestTwoMarketsRideOneAuthorityWaveInsteadOfTakingTurns 은 이 로트의 요점이다.
//
// 고치기 전에는 두 번째 시장이 mutex 뒤에 줄을 서고, 잠금을 물려받은 뒤에는
// 1초 창을 이미 벗어나(창은 파도의 **시작** 시각을 재므로) **자기 파도를 다시
// 돌았다.** 원격이 느릴수록 합치기가 정확히 꺼지는 모양이었다.
func TestTwoMarketsRideOneAuthorityWaveInsteadOfTakingTurns(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, fake, started := refreshWaveFixture()
		_, wave, leader := c.joinStrategyRefreshWave(started)
		if !leader || wave == nil {
			t.Fatal("첫 시장이 지도자가 되지 않았다")
		}

		type outcome struct {
			assembly StrategyEntryProductionAssembly
			err      error
		}
		arrived := make(chan outcome, 1)
		go func() {
			assembly, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), fake)
			arrived <- outcome{assembly: assembly, err: err}
		}()

		// 두 번째 시장이 확실히 멈춰 설 때까지 기다린다. 이 호출이 돌아온다는
		// 것 자체가 그 시장이 **채널에서** 기다린다는 뜻이다 — 잠금 뒤에 서
		// 있었다면 이 줄이 돌아오지 않는다.
		synctest.Wait()
		select {
		case got := <-arrived:
			t.Fatalf("파도가 발표되기 전에 두 번째 시장이 돌아왔다: err=%v", got.err)
		default:
		}

		// 파도가 도는 동안 공유 잠금은 비어 있어야 한다. 원격이 그 잠금 안에서
		// 돌면 이 줄이 실패한다.
		if !c.strategyRefreshMu.TryLock() {
			t.Fatal("파도가 도는 동안 공유 assembly mutex 가 잡혀 있다 — 원격이 잠금 안에서 돈다")
		}
		c.strategyRefreshMu.Unlock()

		c.publishStrategyRefreshWave(wave, started, refreshWaveSentinel(), nil)
		synctest.Wait()
		got := <-arrived
		if got.err != nil {
			t.Fatalf("기다린 시장이 오류를 받았다: %v", got.err)
		}
		if got.assembly.Schedule.KR.CalendarVersion != "WAVE-SENTINEL" {
			t.Fatalf("기다린 시장이 지도자의 파도를 타지 않고 자기 파도를 돌았다: %+v",
				got.assembly.Schedule.KR)
		}
	})
}

// TestOnlyOneMarketEverLeadsAWave 는 지도자가 하나뿐임을 본다. 둘이면 원격이
// 두 번 나가고, 그것이 이 로트가 없애려는 바로 그것이다.
func TestOnlyOneMarketEverLeadsAWave(t *testing.T) {
	c, _, started := refreshWaveFixture()
	_, first, firstLeads := c.joinStrategyRefreshWave(started)
	_, second, secondLeads := c.joinStrategyRefreshWave(started)
	if !firstLeads {
		t.Fatal("첫 호출자가 지도자가 되지 않았다")
	}
	if secondLeads {
		t.Fatal("두 번째 호출자도 지도자가 됐다 — 원격이 두 번 나간다")
	}
	if first != second {
		t.Fatal("두 호출자가 서로 다른 파도를 잡았다")
	}
}

// TestAFailedWaveReachesEveryMarketAndIsNeverCached 는 실패의 두 성질을 함께
// 본다. 기다린 시장이 **같은** 오류를 받고, 실패한 결과는 캐시에 들어가지
// 않는다. 캐시에 넣으면 다음 주기가 1초 동안 실패를 성공처럼 되돌려 준다.
func TestAFailedWaveReachesEveryMarketAndIsNeverCached(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, fake, started := refreshWaveFixture()
		_, wave, _ := c.joinStrategyRefreshWave(started)
		failure := errors.New("official calendar unavailable")

		arrived := make(chan error, 1)
		go func() {
			_, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), fake)
			arrived <- err
		}()
		synctest.Wait()

		c.publishStrategyRefreshWave(wave, started, StrategyEntryProductionAssembly{}, failure)
		synctest.Wait()
		if err := <-arrived; !errors.Is(err, failure) {
			t.Fatalf("기다린 시장이 받은 오류=%v — 지도자가 만난 것과 다르다", err)
		}
		if c.strategyRefresh != nil {
			t.Fatal("실패한 파도가 캐시에 들어갔다")
		}
		// 다음 호출자는 캐시를 만나지 않고 새 지도자가 되어야 한다.
		if _, _, leader := c.joinStrategyRefreshWave(started); !leader {
			t.Fatal("실패 뒤 다음 호출자가 지도자가 되지 못했다 — 파도가 자리에서 안 내려갔다")
		}
	})
}

// TestAMarketWaitingOnTheWaveLeavesWhenItsOwnCycleIsCancelled 은 오늘 없는
// 성질이다. `Lock()` 에 걸린 시장은 ctx 를 보지 못해, 자기 주기가 취소되어도
// 원격이 끝날 때까지 서 있는다.
func TestAMarketWaitingOnTheWaveLeavesWhenItsOwnCycleIsCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, fake, started := refreshWaveFixture()
		_, wave, _ := c.joinStrategyRefreshWave(started)
		ctx, cancel := context.WithCancel(context.Background())

		arrived := make(chan error, 1)
		go func() {
			_, err := c.refreshPairedStrategyEntryProductionAssembly(ctx, fake)
			arrived <- err
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()

		if err := <-arrived; !errors.Is(err, context.Canceled) {
			t.Fatalf("취소된 시장이 받은 값=%v — 취소가 아니다", err)
		}
		// 떠난 시장이 남은 시장의 파도를 끌고 가지 않는다.
		c.strategyRefreshMu.Lock()
		outstanding := c.strategyRefreshWave
		c.strategyRefreshMu.Unlock()
		if outstanding != wave {
			t.Fatal("기다리던 시장이 떠나면서 진행 중인 파도를 지웠다")
		}
	})
}

// TestTheCacheWindowStillMeasuresFromTheWaveStart 는 이 로트가 **바꾸지 않은**
// 것을 못 박는다. 1초 창은 여전히 파도의 시작 시각부터 잰다.
//
// 창을 완료 시각으로 옮기면 캐시의 수명이 파도 길이만큼 늘어난다. 그것은 이
// 태스크가 요구하지 않은 완화이고, 합치기는 창이 아니라 파도가 한다.
func TestTheCacheWindowStillMeasuresFromTheWaveStart(t *testing.T) {
	c, _, started := refreshWaveFixture()
	_, wave, _ := c.joinStrategyRefreshWave(started)
	c.publishStrategyRefreshWave(wave, started, refreshWaveSentinel(), nil)

	if cached, _, _ := c.joinStrategyRefreshWave(started.Add(999 * time.Millisecond)); cached == nil {
		t.Fatal("창 안인데 캐시를 못 받았다")
	}
	cached, _, leader := c.joinStrategyRefreshWave(started.Add(time.Second))
	if cached != nil {
		t.Fatal("창 밖인데 캐시를 받았다 — 창이 파도의 시작이 아니라 완료를 재고 있다")
	}
	if !leader {
		t.Fatal("창 밖 호출자가 새 지도자가 되지 못했다")
	}
}

// TestAPanickingWaveNeverStrandsTheMarketsWaitingOnIt 은 발표를 defer 로 거는
// 이유를 값으로 확인한다.
//
// 잠금은 defer 가 풀어 주지만 채널은 아무도 닫아 주지 않는다. 발표 없이
// 빠져나가면 기다리던 시장은 자기 주기의 30초 마감까지 서 있고, 감시견은 그것을
// abnormal 로 분류해 시장을 잠근다 — 원인은 다른 시장의 패닉인데.
//
// 그리고 패닉을 **성공**으로 발표하면 더 나쁘다: 기다린 시장이 빈 assembly 를
// 오류 없이 받고, `dispatch == nil` 이라 조용히 아무것도 안 하는 주기가 된다.
func TestAPanickingWaveNeverStrandsTheMarketsWaitingOnIt(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, fake, started := refreshWaveFixture()
		_, wave, _ := c.joinStrategyRefreshWave(started)

		arrived := make(chan error, 1)
		go func() {
			_, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), fake)
			arrived <- err
		}()
		synctest.Wait()

		rethrown := false
		func() {
			defer func() { rethrown = recover() != nil }()
			_, _ = c.collectStrategyRefreshWave(wave, started, func() (StrategyEntryProductionAssembly, error) {
				panic("official calendar client exploded")
			})
		}()
		if !rethrown {
			t.Fatal("패닉을 삼켰다 — invokeStrategyCycle 이 그것을 abnormal 주기로 분류할 수 없다")
		}

		synctest.Wait()
		err := <-arrived
		if err == nil {
			t.Fatal("패닉한 파도를 기다린 시장이 오류 없이 돌아왔다 — 빈 assembly 를 성공으로 받았다")
		}
		if !strings.Contains(err.Error(), "did not finish") {
			t.Fatalf("기다린 시장이 받은 오류=%v — 끝나지 않은 파도라고 말하지 않는다", err)
		}
		if c.strategyRefresh != nil {
			t.Fatal("패닉한 파도가 캐시에 들어갔다")
		}
	})
}

// TestTheMarketThatLeadsAWaveAlwaysPublishesIt 은 반증 배터리가 찾아낸 구멍을
// 막는다.
//
// 위 시험들은 전부 **시험이 지도자**다. 그래서 지도자가 파도를 만들어 놓고
// 발표하지 않는 변이(M7: 수집을 `collectStrategyRefreshWave` 를 건너뛰고 직접
// 부른다)가 살아남았다. 그 변이의 결과는 조용하지 않다: 파도가 자리에 남고
// 아무도 닫지 않으므로, **그 뒤의 모든 주기가** 그 죽은 파도에 합류해 자기
// ctx 가 죽을 때까지 기다린다. 두 시장의 진입이 함께 멈춘다.
//
// 이름을 얼리는 셈으로는 못 잡는다 — 그 변이는 수집을 원래 자리에 그대로 두기
// 때문이다. 잡으려면 **생산 경로가 지도자가 되는** 실행이 하나 있어야 한다.
func TestTheMarketThatLeadsAWaveAlwaysPublishesIt(t *testing.T) {
	c, fake, _ := refreshWaveFixture()
	first, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), fake)
	if err != nil {
		t.Fatalf("빈 Context 의 파도가 실패했다: %v", err)
	}
	if first.Supervisor == nil {
		t.Fatal("지도자가 빈 assembly 를 받았다")
	}

	c.strategyRefreshMu.Lock()
	outstanding, cached := c.strategyRefreshWave, c.strategyRefresh
	c.strategyRefreshMu.Unlock()
	if outstanding != nil {
		t.Fatal("지도자가 파도를 만들고 발표하지 않았다 — 뒤따르는 모든 주기가 그 죽은 파도에" +
			" 합류해 자기 ctx 가 죽을 때까지 기다린다. 두 시장의 진입이 함께 멈춘다")
	}
	if cached == nil {
		t.Fatal("성공한 파도가 캐시에 들어가지 않았다")
	}

	// 같은 창 안의 두 번째 시장은 **같은** 결과를 받는다. 포인터가 같다는 것이
	// 파도가 한 번이었다는 뜻이다.
	second, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), fake)
	if err != nil {
		t.Fatalf("두 번째 시장이 오류를 받았다: %v", err)
	}
	if second.Supervisor != first.Supervisor {
		t.Fatal("두 번째 시장이 자기 파도를 따로 돌았다 — 원격이 두 번 나간다")
	}
}

// TestARefreshWithoutAClockRefusesBeforeItCanMintAWave 는 nil 방어(B1)다.
// 시계가 없으면 창을 잴 수 없고, 창을 못 재면 캐시가 영원히 신선하거나 영원히
// 묵는다. 그 상태로 파도를 만드는 것보다 거절이 낫다.
func TestARefreshWithoutAClockRefusesBeforeItCanMintAWave(t *testing.T) {
	c, _, _ := refreshWaveFixture()
	if _, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), nil); err == nil {
		t.Fatal("시계 없이 새로 고침이 성공했다")
	}
	c.strategyRefreshMu.Lock()
	outstanding := c.strategyRefreshWave
	c.strategyRefreshMu.Unlock()
	if outstanding != nil {
		t.Fatal("거절하면서 파도를 만들었다 — 아무도 발표하지 않을 파도다")
	}
}

// TestAFailedWaveCarriesNoAssemblyForEitherMarket 은 지도자와 기다린 시장이
// **같은 것**을 받는지 본다.
//
// 수집이 중간까지 채운 값과 오류를 함께 돌려주는 것은 실제로 가능하다 —
// `NewPairedStrategyEntryProductionAssembly` 의 마지막 오류 자리는 assembly 를
// 다 만든 뒤의 `publishStrategyRuntime` 이다. 그때 비우는 자리가 파도 밖에
// 있으면, 비우는 쪽(지도자)과 안 비우는 쪽(기다린 시장)이 갈린다.
func TestAFailedWaveCarriesNoAssemblyForEitherMarket(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, fake, started := refreshWaveFixture()
		_, wave, _ := c.joinStrategyRefreshWave(started)
		failure := errors.New("projection store refused the wave")

		type outcome struct {
			assembly StrategyEntryProductionAssembly
			err      error
		}
		arrived := make(chan outcome, 1)
		go func() {
			assembly, err := c.refreshPairedStrategyEntryProductionAssembly(context.Background(), fake)
			arrived <- outcome{assembly: assembly, err: err}
		}()
		synctest.Wait()

		leader, err := c.collectStrategyRefreshWave(wave, started, func() (StrategyEntryProductionAssembly, error) {
			// 반쯤 채워진 값 + 오류. 수집의 마지막 단계가 실패한 모양이다.
			return refreshWaveSentinel(), failure
		})
		if !errors.Is(err, failure) {
			t.Fatalf("지도자가 받은 오류=%v", err)
		}
		if leader.Schedule.KR.CalendarVersion != "" {
			t.Fatalf("지도자가 실패한 파도의 반쪽을 받았다: %+v", leader.Schedule.KR)
		}

		synctest.Wait()
		got := <-arrived
		if !errors.Is(got.err, failure) {
			t.Fatalf("기다린 시장이 받은 오류=%v", got.err)
		}
		if got.assembly.Schedule.KR.CalendarVersion != "" {
			t.Fatalf("기다린 시장이 실패한 파도의 반쪽을 받았다 — 지도자와 다른 것을 받았다: %+v",
				got.assembly.Schedule.KR)
		}
	})
}
