# Function Logic Map: `refreshPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`
- Signature: `Context.refreshPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `494:1`–`514:2`
- AST evidence: `ast.json`, regenerated after the L5 5.2.1 edit.
- Risk scan: `risk-pattern-report.md`.

## What the L5 5.2.1 edit changed and why

편집 전 본문은 이랬다.

```go
c.strategyRefreshMu.Lock()
defer c.strategyRefreshMu.Unlock()
now := clk.Now().UTC()
if 캐시가 창 안 { return 캐시 }
fresh, err := c.NewPairedStrategyEntryProductionAssembly(ctx, clk)   // 원격
c.strategyRefreshAt = now
```

두 가지를 쟀다.

1. **원격 권한 수집이 공유 잠금 안에서 돌았다.** `NewPairedStrategyEntryProductionAssembly`
   는 official 달력(`TypedMarketCalendar`)과 official FX 를 타고 후보 DB·저널·evidence DB 를
   읽는다. CodeGraph 로 이 함수의 호출자는 `runProductionStrategyMarketCycle` 하나이고
   그것이 시장마다 하나씩, 즉 goroutine 둘이다. 그래서 KR 의 느린 official 응답이 US
   주기 전체를 세웠다 — 그 주기는 5.1.2.1 이후 레인 평가까지 함께 들고 있다.
2. **1초 창이 파도의 시작 시각을 쟀다.** `strategyRefreshAt` 에 들어가는 값은 수집이
   끝난 시각이 아니라 시작한 시각이다. 그래서 3초짜리 파도가 끝나는 순간 캐시는 이미
   3초 묵었고, 잠금을 물려받은 두 번째 시장은 창 밖이라 **자기 파도를 다시 돌았다.**
   원격이 느릴수록 합치기가 정확히 꺼졌다.

편집 뒤에는 잠금이 상태 전이만 지키고 파도는 잠금 밖에서 돈다. 창의 의미(**시작**
시각 기준 1초)는 바꾸지 않았다 — 완료 시각으로 옮기면 캐시 수명이 파도 길이만큼
늘어나는 완화가 되고, 이 태스크는 그것을 요구하지 않는다. 합치기는 창이 아니라
파도가 한다.

## Inputs and invariants

- Inputs/results are the exact AST signature above.
- 잠금은 이 함수 본문에 더 이상 나타나지 않는다. `joinStrategyRefreshWave` 와
  `publishStrategyRefreshWave` 두 곳뿐이고, 그 완전성은
  `a112_refresh_wave_test.go` 가 패키지의 모든 비시험 파일을 훑어 지킨다.
- 실패한 파도는 캐시에 들어가지 않고, 지도자와 기다린 시장에게 **같은** 값(빈
  assembly + 같은 오류)이 간다.

## Branches and early returns

- Exact AST return nodes: `527:3, 535:3, 538:3, 542:2, 543:3`.

| Branch | AST kind | Source location | Test disposition |
|---|---|---|---|
| B1 | if— `c` 또는 시계가 없다 | 526:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestARefreshWithoutAClockRefusesBeforeItCanMintAWave` |
| B2 | if— 창 안의 캐시가 있다 | 534:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B3 | if— 지도자가 아니다 (파도에 합류) | 537:2 | arm entered 5x (engine tagged suite); arm entered 5x (engine untagged suite); `TestAFailedWaveCarriesNoAssemblyForEitherMarket`, `TestAFailedWaveReachesEveryMarketAndIsNeverCached`, `TestAMarketWaitingOnTheWaveLeavesWhenItsOwnCycleIsCancelled`, `TestAPanickingWaveNeverStrandsTheMarketsWaitingOnIt`, `TestTwoMarketsRideOneAuthorityWaveInsteadOfTakingTurns` |

## Calls and live bindings

| Callee expression | Source location | Note |
|---|---|---|
| errors.New | 527:45 | B1 의 거절 |
| UTC | 529:9 | 창을 재는 시각을 UTC 로 고정한다 |
| clk.Now | 529:9 | 창을 재는 유일한 시각. 잠금 **밖**에서 읽는다 |
| c.joinStrategyRefreshWave | 533:26 | 잠금 안에서 하는 일의 전부 |
| awaitStrategyRefreshWave | 538:10 | 채널에서 기다린다 — `ctx` 를 함께 본다 |
| c.collectStrategyRefreshWave | 542:9 | 지도자의 자리. 잠금 밖 |
| c.NewPairedStrategyEntryProductionAssembly | 543:10 | 원격 파도 본체. 클로저 안이므로 `collect` 인자로 넘어간다 |

## State mutations and fallbacks

- 이 함수는 `Context` 의 어떤 칸도 직접 쓰지 않는다. `strategyRefresh`,
  `strategyRefreshAt`, `strategyRefreshWave` 는 전부 위 두 헬퍼 안에서만 바뀐다.
- 취소: 기다리는 시장은 `ctx.Done()` 으로 빠져나온다. 편집 전에는 `Lock()` 에 걸려
  자기 주기가 취소되어도 원격이 끝날 때까지 서 있었다.
- 패닉: 지도자가 수집 중 패닉하면 `collectStrategyRefreshWave` 의 defer 가 오류를
  발표하고 다시 던진다. 발표를 거르면 기다린 시장이 자기 주기 마감까지 선다.

## Safety conclusion

- 주문·손절·사이징 경로는 이 편집에 없다. 바뀐 것은 권한 수집의 **동시성 모양**이다.
- 합치기가 늘리는 것은 기다린 시장이 받는 assembly 의 나이(최대 파도 한 번)다. 그
  나이가 주문으로 새지 않는 근거는 둘이고 둘 다 읽은 것이다: `strategy_dispatch_cycle.go:114`
  가 assembly 자신의 관측 시각(`cycle.schedule.observedAt`)으로 서명된 활성화 수명을
  검사하고, 같은 파일 `:147` 의 `FinalAuthorityCheck` 가 broker 에 내기 **직전**
  `revalidateSchedule` 로 일정 권위를 다시 수집해 세대·달력 버전·매니페스트 digest 가
  움직였으면 거절한다.
- 새 동작 하나: 실패한 파도가 두 시장 모두에게 같은 오류로 간다(전에는 두 번째 시장이
  자기 파도를 다시 돌았다). 원격이 고장 났을 때 부하가 절반이 되는 방향이다.
- 반증 16개 전부 CAUGHT, 대조군 2개 SURVIVED(의도). 그중 M10b(발표가 `done` 을 먼저 닫고
  값을 나중에 쓴다)는 `-race` **없이는 초록**이고 `-race` 로만 빨갛다 — 그래서 이 로트가
  `make test-race` 에 이 패키지의 동시성 시험을 이름으로 배선했다.
