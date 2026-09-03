package engine

import (
	"context"
	"fmt"
	"time"
)

// 이 파일은 태스크 5.2 의 한 절이다: **remote evidence refresh 를 공유 assembly
// mutex 밖에 둔다.**
//
// 고치기 전 모양과 그것이 왜 결함인지.
//
//	c.strategyRefreshMu.Lock()
//	defer c.strategyRefreshMu.Unlock()
//	now := clk.Now().UTC()
//	if 캐시가 1초 안이면 { return 캐시 }
//	fresh, err := c.NewPairedStrategyEntryProductionAssembly(ctx, clk)  // 원격
//	c.strategyRefreshAt = now
//
// 잰 것 두 가지.
//
//  1. **파도가 잠금 안에서 돈다.** `NewPairedStrategyEntryProductionAssembly` 는
//     official 달력(`TypedMarketCalendar`)과 official FX 를 타고, 후보 DB·저널·
//     evidence DB 를 읽는다. 두 시장의 주기 goroutine 이 이 잠금 하나를
//     공유하므로, KR 의 느린 official 응답이 US 주기 전체를 세운다. 그 주기는
//     5.1.2.1 이후 레인 평가까지 들고 있고, 위에는 30초 감시견이 있다.
//
//  2. **1초 창은 파도가 1초보다 오래 걸리면 아무도 태우지 못한다.**
//     `strategyRefreshAt` 에 넣는 값은 파도가 **끝난** 시각이 아니라 **시작**
//     시각(`now`)이다. 그래서 3초짜리 파도가 끝나면 캐시는 태어나자마자 3초
//     묵은 것이 되고, 잠금을 물려받은 두 번째 시장은 창을 벗어나 **자기 파도를
//     다시 돈다.** 즉 원격이 느릴수록 — 합치는 것이 가장 필요한 때 — 합치기가
//     정확히 꺼진다. 두 시장은 원격을 두 번 타고 그 둘은 직렬이다.
//
// 그래서 잠금이 하는 일을 **상태 전이**로만 줄인다. 잠금 안에서 하는 일은
// 캐시를 보고, 진행 중인 파도가 있으면 거기 합류하고, 없으면 지도자가 되는 것
// 뿐이다. 파도 자체는 잠금 **밖에서** 돌고, 기다리는 시장은 mutex 가 아니라
// 채널에서 기다린다 — 그래서 자기 주기가 취소되면 빠져나올 수 있다. 오늘
// `Lock()` 에 걸린 시장은 ctx 를 보지 못한다.
//
// **1초 창의 의미는 건드리지 않는다.** `strategyRefreshAt` 은 여전히 파도의
// 시작 시각이다. 창을 완료 시각으로 옮기면 캐시의 수명이 파도 길이만큼
// 늘어나는데, 그것은 이 태스크가 요구하지 않은 완화다. 합치기는 창이 아니라
// 파도가 한다.

// strategyRefreshWave 는 진행 중인 권한 수집 **한 번**이다.
//
// 값은 지도자만 채우고, 채운 뒤에 done 을 닫는다. 그래서 done 을 받은 쪽이
// 읽는 필드는 이미 쓰기가 끝난 것이다 — 채널 닫힘이 그 순서를 보증한다.
type strategyRefreshWave struct {
	// done 은 이 파도가 끝났음을 알리는 유일한 신호다. 닫히기만 하고 값을
	// 보내지 않는 이유는 기다리는 쪽이 몇인지 지도자가 알 필요가 없어서다.
	done chan struct{}
	// assembly 와 err 은 지도자가 done 을 닫기 **전에** 채운다.
	assembly StrategyEntryProductionAssembly
	err      error
}

// joinStrategyRefreshWave 는 공유 잠금 안에서 하는 **유일한** 일이다.
//
// 세 답 중 하나를 돌려준다.
//   - 신선한 캐시가 있다: cached 가 non-nil. 파도는 없다.
//   - 진행 중인 파도가 있다: wave 가 non-nil, leader 는 false. 기다린다.
//   - 아무것도 없다: wave 가 non-nil, leader 는 true. 이 호출자가 돈다.
//
// 여기서 원격·파일·저널을 부르지 않는다. 그 사실을 주석이 아니라
// a112_refresh_wave_test.go 의 호출 목록 셈이 지킨다.
func (c *Context) joinStrategyRefreshWave(now time.Time) (cached *StrategyEntryProductionAssembly,
	wave *strategyRefreshWave, leader bool) {
	c.strategyRefreshMu.Lock()
	defer c.strategyRefreshMu.Unlock()
	if c.strategyRefresh != nil && !now.Before(c.strategyRefreshAt) && now.Sub(c.strategyRefreshAt) < time.Second {
		return c.strategyRefresh, nil, false
	}
	if c.strategyRefreshWave != nil {
		return nil, c.strategyRefreshWave, false
	}
	c.strategyRefreshWave = &strategyRefreshWave{done: make(chan struct{})}
	return nil, c.strategyRefreshWave, true
}

// publishStrategyRefreshWave 는 지도자가 파도를 끝낼 때 부른다.
//
// 실패한 파도는 캐시에 넣지 않는다. 넣으면 다음 주기가 1초 동안 그 실패를
// 성공처럼 되돌려 준다. 대신 기다린 시장에게는 **같은 오류**를 그대로 준다 —
// 실패를 감추고 각자 다시 원격을 타게 하면 고장 난 원격에 부하가 두 배가 된다.
func (c *Context) publishStrategyRefreshWave(wave *strategyRefreshWave, startedAt time.Time,
	assembly StrategyEntryProductionAssembly, err error) {
	c.strategyRefreshMu.Lock()
	defer c.strategyRefreshMu.Unlock()
	wave.assembly, wave.err = assembly, err
	if err == nil {
		c.strategyRefreshAt = startedAt
		c.strategyRefresh = &assembly
	}
	// 이 파도를 자리에서 내린다. 다음 호출자는 캐시를 보거나 새 지도자가 된다.
	if c.strategyRefreshWave == wave {
		c.strategyRefreshWave = nil
	}
	close(wave.done)
}

// collectStrategyRefreshWave 는 지도자가 파도를 도는 자리다. 잠금은 여기서
// 잡지 않는다 — 수집이 무엇을 타든 잠금 밖이다.
//
// collect 를 인자로 받는 이유는 둘이다. 하나는 이 함수가 ctx·clk 를 들 필요가
// 없어지는 것이고, 다른 하나는 **패닉 경로를 잴 수 있게 되는 것**이다. 실제
// 원격 수집은 시험에서 실패시킬 수 없으므로, 인자가 아니면 아래 defer 는
// 시험되지 않는 코드로 남는다 — 이 change 는 그것을 증거로 치지 않는다.
//
// 패닉이 나가도 발표를 거른다. 잠금은 defer 가 풀어 주지만 채널은 아무도 닫아
// 주지 않으므로, 발표 없이 빠져나가면 기다리던 시장이 자기 주기 마감까지 서
// 있는다. 그리고 패닉을 **성공**으로 발표하면 기다린 시장은 빈 assembly 를
// 오류 없이 받고, 그 조합은 `dispatch == nil` 이라 조용히 아무것도 안 하는
// 주기가 된다 — 이 change 가 5.4.3 에서 이미 한 번 없앤 모양이다.
//
// 패닉은 다시 던진다. 오늘 `invokeStrategyCycle` 이 그것을 받아 abnormal 주기로
// 바꾸고, 그 분류는 이 로트가 바꿀 것이 아니다.
func (c *Context) collectStrategyRefreshWave(wave *strategyRefreshWave, startedAt time.Time,
	collect func() (StrategyEntryProductionAssembly, error)) (assembly StrategyEntryProductionAssembly, err error) {
	published := false
	defer func() {
		if published {
			return
		}
		recovered := recover()
		c.publishStrategyRefreshWave(wave, startedAt, StrategyEntryProductionAssembly{},
			fmt.Errorf("engine: paired strategy refresh wave did not finish: %v", recovered))
		panic(recovered)
	}()
	assembly, err = collect()
	if err != nil {
		// 실패한 파도는 **아무것도** 싣지 않는다. 여기서 비우지 않으면 지도자는
		// 빈 값을 받고 기다린 시장은 수집이 중간까지 채운 값을 받는다 — 같은
		// 파도에 대해 두 개의 진실이 생긴다. 비우는 자리를 부르는 쪽에 두면
		// 그 자리는 시험이 닿을 수 없는 갈래가 된다.
		assembly = StrategyEntryProductionAssembly{}
	}
	c.publishStrategyRefreshWave(wave, startedAt, assembly, err)
	published = true
	return assembly, err
}

// awaitStrategyRefreshWave 는 지도자가 아닌 시장이 기다리는 자리다.
//
// ctx 를 함께 본다. 오늘 `Lock()` 에 걸린 시장은 자기 주기가 취소되어도 원격이
// 끝날 때까지 서 있고, 그 시간은 감시견의 마감 시한 안에서 통째로 낭비된다.
func awaitStrategyRefreshWave(ctx context.Context, wave *strategyRefreshWave) (StrategyEntryProductionAssembly, error) {
	select {
	case <-ctx.Done():
		return StrategyEntryProductionAssembly{}, ctx.Err()
	case <-wave.done:
		return wave.assembly, wave.err
	}
}
