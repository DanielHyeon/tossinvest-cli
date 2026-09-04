# Branch Test Map: `strategyLaneRuntime.runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go` (252-274); file SHA-256 `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- Per-test attribution set: 두 엔진 바이너리의 테스트 **전체**(태그 491 · 무태그 438).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 256:2 | arm entered 170x (engine tagged suite); arm entered 170x (engine untagged suite); `TestADroppedTriggerNeverDrivesACycle`, `TestALatchOnlyReopensForAStrictlyNewerVerifiedActivation`, `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`, `TestALedgerThatCannotTakeTheLatchStopsTheCycle`, `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart`, `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`, `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther` |
| B2 | if | 268:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

B2 는 오늘 생산에서 도달 불가다(생산 `Step` 이 오류를 내지 않는다). 구멍으로 남겨
두고 5.6.2 에 넘긴다 — 채우는 척하지 않는다.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
