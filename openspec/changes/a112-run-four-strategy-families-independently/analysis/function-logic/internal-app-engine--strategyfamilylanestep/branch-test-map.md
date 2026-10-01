# Branch Test Map: `strategyFamilyLaneStep`

- Source: `internal/app/engine/strategy_lane_runtime.go` (172-178); file SHA-256 `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- Per-test attribution set: 두 엔진 바이너리의 테스트 **전체**(태그 491 · 무태그 438).

분기가 없으므로 행은 하나다.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | branchless happy path | 172:1 | arm entered 170x (engine tagged suite); arm entered 170x (engine untagged suite); `TestADroppedTriggerNeverDrivesACycle`, `TestALatchOnlyReopensForAStrictlyNewerVerifiedActivation`, `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`, `TestALedgerThatCannotTakeTheLatchStopsTheCycle`, `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart`, `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`, `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther` |

RED→GREEN: 둘째 인자(서명된 승격)를 더한 편집은 두 열거표를 **실제로 빨갛게** 만들었고
(`TestOnlyThePackageLevelStepEverRunsInsideALane` 의 얼린 철자,
`TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion` 의 앞 판본 개수 단언) 그 둘을 판단해서 고쳤다 —
철자는 갱신하고, 개수 단언은 **타입 목록**으로 바꿨다. 상세는 review.md 의 2026-09-03 절.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
