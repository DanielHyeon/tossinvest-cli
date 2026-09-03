# Branch Test Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go` (252-376); file SHA-256 `653c9fa1a7f9e24754fde6c7d7c56414fc540afe7f0992343795acb6f533314b`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
  묶지 않고 돌린 이 패키지가 커널 OOM 으로 데스크톱을 세 번 죽였기 때문이다(`engine.test`, anon-rss 약 36GB).
  원인은 이 lot 이 고친 조정자 용량이며, 측정 방법이 아니라 측정 대상이 문제였다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 73.8% of statements.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS, 63.6% of statements.
- Per-test attribution set: 같은 태그 바이너리를 `-test.run '^<Test>$'` 로 하나씩 돌린 열 개의 프로파일.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 테스트별 진입 수의 합이 스위트 전체 진입 수와
  정확히 같다. 이 집합 밖의 테스트가 어느 arm 이든 들어갔다면 그 등식이 깨진다. 깨진 행은
  `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

Mutation receipts (production source mutated, 빌드·실행 후 mutation 전에 뜬 pristine 사본에서 복원,
복원은 SHA-256 대조로 확인):

| # | mutation | result | killed by |
|---|---|---|---|
| M-C1 | `NewMarketCoordinator` 의 용량을 `Capacity` 에서 `1<<30` 으로 되돌림 | KILLED | `TestTheProductionCoordinatorUsesTheServerOwnedCapacity` |
| M-E1 | `if outcome.Overflow` 를 `if false` 로 바꿔 넘침을 흘려보냄 | KILLED | `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` |
| M-E2 | 넘침 결과에서 `QueueDropCount` 대입을 삭제 | KILLED | `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` |
| M-E3 | `entries()` 의 `return nil, false` 를 `continue` 로 바꿔 못 찾은 자리를 건너뜀 | KILLED | `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList` |

M-E3 는 이 lot 에서 **처음에 SURVIVED** 했다. 그 가드의 진입 수가 0 이라는 위 측정과 같은 사실이었고,
그래서 가드를 지우지 않고 그것을 죽이는 테스트를 새로 넣었다.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if| 265:2 | arm entered 5x (engine tagged suite); arm entered 2x (engine untagged suite); `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B2 | if| 268:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | if| 271:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B4 | if| 276:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B5 | if| 280:2 | arm entered 19x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B6 | range| 286:2 | arm entered 10059x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B7 | if| 288:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | if| 300:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | if| 305:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol` |
| B10 | if| 326:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B11 | if| 335:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` |
| B12 | if| 343:2 | arm entered 4x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal` |
| B13 | if| 353:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B14 | if| 361:2 | arm entered 6x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B15 | range| 369:2 | arm entered 41x (engine tagged suite); arm not entered (engine untagged suite); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
