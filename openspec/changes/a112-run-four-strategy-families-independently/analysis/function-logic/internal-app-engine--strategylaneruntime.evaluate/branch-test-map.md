# Branch Test Map: `strategyLaneRuntime.evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go` (191-224); file SHA-256 `0526b42f2ba26f101931e4f30425ae64558dd1d7e0e0070fa5f9c9a2e34df104`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- Per-test attribution set: 두 엔진 바이너리의 테스트 **전체**(태그 491 · 무태그 438).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 194:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B2 | if | 200:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | range | 205:2 | arm entered 288x (engine tagged suite); arm entered 280x (engine untagged suite); `TestADroppedTriggerNeverDrivesACycle`, `TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed`, `TestALatchOnlyReopensForAStrictlyNewerVerifiedActivation`, `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`, `TestALedgerThatCannotTakeTheLatchStopsTheCycle`, `TestARestoredLatchKeepsTheFirstCauseAcrossTheRestart`, `TestEveryFamilyLaneIsDormantUntilASignedManifestPromotesIt`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty`, `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther` |
| B4 | range | 207:3 | arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes` |
| B5 | if | 208:4 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes` |
| B6 | if | 218:2 | arm entered 1x (engine tagged suite); arm entered 1x (engine untagged suite); `TestALedgerThatCannotTakeTheLatchStopsTheCycle` |

RED→GREEN: 승격을 인자로 받는 편집은 이 함수의 **행동을 바꾸지 않는다**(영값 승격에서
모든 레인이 DORMANT — 오늘의 값). 바뀐 것을 재는 시험은 관문 쪽이다:
`TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt` 와
`TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`.

B1 · B2 는 진입 0 회다. 구멍으로 남겨 두고 5.6.2 에 넘긴다 — 채우는 척하지 않는다.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
