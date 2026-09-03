# Function Logic Map: `strategyLaneRuntime.runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go` (252-274)
- Function: `strategyLaneRuntime.runLane` in package `engine`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- File SHA-256: `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

레인 하나의 유계 사이클 한 번이고, 그 관측 하나를 돌려준다. 태스크 8.7.1 이 인자를
하나 늘렸다: `promotion strategyrouter.FamilyActivation` — 관문이 이 주기에 읽은
그 값을 그대로 받는다. 여기서 다시 읽지 않는 이유는 만료다. 두 자리에서 따로 읽으면
그 사이에 수명이 지났을 때 관문은 통과시킨 제안을 관측은 DORMANT 로 적는다.

투입을 **먼저** 넣는다(`lane.Offer()`). 레인의 관문(`RunBounded`)은 투입이 없으면
사이클을 열지 않으므로, 넣지 못했다면 그 이유가 곧 이번 주기의 관측이다 — 잠겨서 못
받았는지(DISABLED)와 칸이 차서 버렸는지(FULL)는 운영자가 할 조치가 다르다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- engine tagged / untagged suite: `go test -c [-tags tossos_testseams] -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/` 뒤 `-test.coverprofile`;
  스위트 실행은 `systemd-run --user --scope -p MemoryMax=12G -p MemorySwapMax=0` cgroup 안.
- Per-test attribution set: 두 바이너리의 테스트 **전체**(태그 491 · 무태그 438).
- **귀속 완전성은 측정이다.** 아래 두 분기에서 테스트별 진입 수의 합이 스위트 전체
  진입 수와 같다. 어긋난 행은 `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 없다.

Exact AST return positions: 258:3, 273:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 256:2 | arm entered 170x (engine tagged suite); arm entered 170x (engine untagged suite); `TestADroppedTriggerNeverDrivesACycle`, `TestALatchOnlyReopensForAStrictlyNewerSignedActivation`, `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`, `TestALedgerThatCannotTakeTheLatchStopsTheCycle`, `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart`, `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`, `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther` |
| B2 | if | 268:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

**B1 의 수가 곧 유계 사이클이 실제로 돈 횟수의 여집합이다** — 이 arm 은 투입이
들어가지 **못한** 경우이므로. 오늘 여덟은 전부 DORMANT 이고 생산 주기가 같은 레인에
반복 투입하므로 이 arm 이 잦다.

**B2 는 진입 0 회다. 커버리지 구멍이며 통과가 아니다.** 그 arm 은 유계 사이클이
오류를 돌려준 경우(`bounded.Err != nil`)이고, 오늘 그 오류를 만들 수 있는 것은
`Step` 이 오류를 내는 경우뿐이다. 생산 `strategyFamilyLaneStep` 은 오류를 절대
내지 않으므로(위 번들의 산문) 이 arm 은 오늘 생산에서 도달 불가다. 5.6.2 가 여덟
lane 위에서 고장 범위를 다시 증명할 때 그 자리를 가져간다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 255:46 |
| `lane.Offer` | 255:67 |
| `lane.Health` | 257:24 |
| `lane.RunBounded` | 260:20 |
| `strategyFamilyLaneStep` | 260:48 |
| `bounded.Err.Error` | 269:25 |
| `lane.Health` | 271:23 |

## State mutations and fallbacks

- AST assignments: 13. Defers: 0. Goroutine statements: 0.
- 대입은 전부 지역 `observation` 필드다. 레인 상태를 바꾸는 것은 `lane.Offer` 와
  `lane.RunBounded` 안이고, 그 둘은 `internal/strategyworker` 의 잠금 아래에 있다.

## Safety conclusion

- Safe edit boundary: 관측을 만드는 함수다. 진입을 여는 방법이 없다.
- High-risk impact: no — 이 함수는 판정하지 않는다. 판정은 관문
  (`coordinateMarketProposals`)과 레인이 한다.
