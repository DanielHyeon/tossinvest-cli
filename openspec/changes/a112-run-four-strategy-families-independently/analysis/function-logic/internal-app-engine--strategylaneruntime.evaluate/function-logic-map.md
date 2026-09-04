# Function Logic Map: `strategyLaneRuntime.evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go` (191-224)
- Function: `strategyLaneRuntime.evaluate` in package `engine`
- Signature: `strategyLaneRuntime.evaluate(params=5, results=1)`
- File SHA-256: `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 6.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

이 시장의 네 레인에 각자 자기 것인 봉인된 제안을 한 번씩 돌리고 그 관측을 기록한다.
태스크 8.7.1 이 인자를 하나 늘렸다: `promotion strategyrouter.FamilyActivation`.

**두 활성화가 이름이 비슷하지만 다른 것이다.** `activationGeneration uint64` 는
**스케줄러**의 서명 활성화 세대이고 durable latch 의 복구 조건이다(5.3.3).
`promotion` 은 **4-가족** 활성화이고 어느 레인이 켜졌는지를 말한다(8.7.1). 하나로
합치면 복구 조건과 승격이 같은 수에 걸리게 되고, 그때 매니페스트 하나를 다시
서명하는 일이 잠긴 레인을 함께 열게 된다.

승격을 여기서 읽지 않고 받는 이유는 만료다. 관문(`coordinateMarketProposals`)이 이
주기에 읽은 값을 그대로 받아야 관문과 관측이 같은 승격을 본다.

**순서가 계약이다.** 복구 → 사이클 → 잠금 기록. 복구를 사이클 뒤로 미루면 증거가
이미 도착한 레인이 한 주기를 더 잠긴 채로 보낸다.

**관측은 돌려주지 않는다.** 호출자가 버릴 수 있는 답을 만들지 않는다 — 같은 문제를
dispatch 주기에서 `Deliver` 로 푼 것과 같은 답이다. **오류는 돌려준다**: durable
latch 를 원장에 남기지 못했거나 이 빌드에 없는 레인을 가리키는 기록이 남아 있다는
뜻이고, 조용히 넘기면 다음 재시작이 잠긴 레인을 연다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- engine tagged / untagged suite: `go test -c [-tags tossos_testseams] -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/` 뒤 `-test.coverprofile`;
  스위트 실행은 `systemd-run --user --scope -p MemoryMax=12G -p MemorySwapMax=0` cgroup 안.
- Per-test attribution set: 두 바이너리의 테스트 **전체**(태그 491 · 무태그 438).
- **귀속 완전성은 측정이다.** 아래 여섯 분기에서 테스트별 진입 수의 합이 스위트 전체
  진입 수와 같다. 어긋난 행은 `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 없다.

Exact AST return positions: 195:3, 201:3, 219:3, 223:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 194:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B2 | if | 200:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | range | 205:2 | arm entered 288x (engine tagged suite); arm entered 280x (engine untagged suite); `TestADroppedTriggerNeverDrivesACycle`, `TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed`, `TestALatchOnlyReopensForAStrictlyNewerVerifiedActivation`, `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`, `TestALedgerThatCannotTakeTheLatchStopsTheCycle`, `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart`, `TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty`, `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther` |
| B4 | range | 207:3 | arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes` |
| B5 | if | 208:4 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes` |
| B6 | if | 218:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestALedgerThatCannotTakeTheLatchStopsTheCycle` |

**진입 0 회 둘을 이름 붙여 남긴다. 통과가 아니라 구멍이다.**
B1(`runtime == nil`)은 생산에서 도달 불가다 — 부르는 자리
(`runProductionStrategyMarketCycle`)가 `productionStrategyLanes` 의 오류를 먼저
반환하므로 nil 런타임으로는 여기 오지 않는다. B2(복구 오류)는 원장이
`RecoverStrategyLaneLatch` 에서 `ErrStrategyLaneLatchRecoveryEvidence` 가 아닌
오류를 낸 경우이며, 오늘 그 상황을 만드는 시험이 없다. B6(잠금 기록 실패)은
`TestALedgerThatCannotTakeTheLatchStopsTheCycle` 이 채운다 — 같은 모양의 실패를
**쓰기** 쪽에서는 재고 **복구** 쪽에서는 재지 않는다는 뜻이므로, 5.6.2 가 그 자리를
가져간다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.recoverMarketLanes` | 200:12 |
| `runtime.lanesFor` | 203:11 |
| `make` | 204:18 |
| `len` | 204:53 |
| `lane.Owns` | 208:7 |
| `append` | 213:18 |
| `runtime.runLane` | 213:39 |
| `runtime.record` | 215:2 |
| `runtime.persistMarketLatches` | 218:12 |
| `runtime.clk.Now` | 218:76 |
| `runtime.staleLatchError` | 223:9 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.
- 상태를 바꾸는 것은 `runtime.record`(관측 맵) 와 `recoverMarketLanes` ·
  `persistMarketLatches`(원장) 다. 진입을 여는 방법은 없다 — 여는 것은 원장이
  복구를 받아들인 뒤 레인을 **다시 세우는** 것뿐이다(5.3.3).

## Safety conclusion

- Safe edit boundary: 승격은 인자로만 들어온다. 이 함수는 승격을 만들 수 없다.
- High-risk impact: yes — 원장에 잠금을 남기는 경로가 여기 있다. 스키마 v32.
