# Function Logic Map: `coordinateMarketProposals`

- Source: `internal/app/engine/strategy_market_coordinator.go` (57-125)
- Function: `coordinateMarketProposals` in package `engine`
- Signature: `coordinateMarketProposals(params=5, results=2)`
- File SHA-256: `e9a4bde458176c4679f254d71b56439b348110fb6dc1ab2338cd7f2cf376f728`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 8.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 시장의 모든 가족 제안을 조정자에 넣고 중재까지 마친다. 태스크 5.4.2 가 만든 새 함수이며
`arbitrateProposalScope` 를 대체한다(그 함수는 이제 트리에 없다).

여기서는 아무것도 고르지 않는다. 고르는 규칙은 `strategyarbiter` 한 곳에, 받고 접고 줄 세우는 규칙은
`strategycoordinator` 한 곳에 있다. 이 함수는 그 둘에 필요한 값을 모아 주고, 조정자가 돌려준 봉인된 계보
신원을 이 프로세스의 레인 권한으로 되돌릴 색인만 만든다 — 조정자가 엔진 자료구조를 들고 다니면 순수하지 않게 된다.

기대 범위의 종목은 경로 권한이 아니라 승인된 후보에서 읽는다. 권한이 스스로 말한 종목을 그 권한을 검사하는
데 다시 쓰면, 어긋남을 잡아내려던 검사가 언제나 참이 되어 아무것도 잡지 못한다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
  묶지 않고 돌린 이 패키지가 커널 OOM 으로 데스크톱을 세 번 죽였기 때문이다(`engine.test`, anon-rss 약 36GB).
  원인은 이 lot 이 고친 조정자 용량이며, 측정 방법이 아니라 측정 대상이 문제였다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 78.3% of statements.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS, 69.5% of statements.
- **표본 수 주의(5.5-fix3 정정):** 위 두 값은 각각 **한 번** 잰 것이다. 같은 스위트를 3~5회 재측정하면 태그는 2912·2909·2909 of 3941, 무태그는 2498~2501 of 3936 으로 흔들린다. 즉 소수점 첫째 자리는 안정적이지 않으므로, 두 값을 다른 로트의 값과 자릿수까지 견주지 않는다.
- Per-test attribution set: 같은 태그 바이너리를 `-test.run '^<Test>$'` 로 하나씩 돌린 27 개의 프로파일 (2026-09-04 재측정).
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 테스트별 진입 수의 합이 스위트 전체 진입 수와
  정확히 같다. 이 집합 밖의 테스트가 어느 arm 이든 들어갔다면 그 등식이 깨진다. 깨진 행은
  `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

Exact AST return positions: 115:5, 124:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 64:2 | arm entered 10062x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B2 | if | 66:3 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B3 | range | 75:3 | arm entered 10069x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B4 | if | 93:4 | arm entered 11x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestTheGateSaysWhichKindOfStopItMade` |
| B5 | if | 98:4 | arm entered 10058x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B6 | if | 102:4 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B7 | if | 106:4 | arm entered 10056x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B8 | if | 119:3 | arm entered 10059x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyRouterMarket` | 60:18 |
| `strategycoordinator.NewMarketCoordinator` | 61:17 |
| `make` | 62:55 |
| `batch.Len` | 62:103 |
| `batch.LanesFor` | 65:12 |
| `route.approved.Symbol` | 65:27 |
| `len` | 66:6 |
| `route.approved.Symbol` | 71:12 |
| `route.route.Request` | 71:57 |
| `lane.Proposal` | 76:14 |
| `lane.SnapshotDigest` | 80:21 |
| `gate.admit` | 91:29 |
| `lane.SnapshotDigest` | 92:21 |
| `append` | 94:25 |
| `coordinator.Submit` | 101:17 |
| `coordinator.Drops` | 114:33 |
| `len` | 119:22 |
| `coordinator.Arbitrate` | 123:24 |

## State mutations and fallbacks

- AST assignments: 18. Defers: 0. Goroutine statements: 0.

## Safety conclusion

제안을 하나도 내지 않은 종목은 중재 대상이 아니라 거절 수에만 든다(B2).
계보 신원이 두 레인에 겹치면 아무거나 고르지 않고 닫는다(**B7**) — 겹치면 선택을 되돌릴 자리가 하나로 정해지지 않는다.
B7 은 진입 0 회다. 커버리지 구멍이며 통과가 아니다.

**태스크 5.1.2.2 가 분기 둘을 더했고 그래서 옛 B4·B5 가 B6·B7 이 되었다.**
새 둘은 4-가족 관문이다: B4(73:4)는 그 가족의 레인이 이 제안을 조정자로 보내지
않는다는 답이고(DORMANT · LATCHED · REFUSED 를 `arbitration.gated` 에 담는다),
B5(77:4)는 관문이 서 있을 때 레인이 만든 봉투로 갈아 끼우는 자리다. 관문이 서지
않은 시장에서는 B4 가 참이 되지 않고 B5 도 지나가므로 위 봉투가 그대로 들어간다 —
그것이 오늘의 동작이다.

**표의 B 번호는 좌표 순서로 다시 붙었다.** 이 표를 읽는 사람이 옛 리뷰의 "B4" 를
그대로 찾으면 다른 분기를 보게 된다. 그것이 절대 줄 번호로 얼린 열거표의 대가이며,
이 change 는 그 대가를 알고 감수한다(`analysis/` 의 다른 번들도 같은 모양이다).

## 2026-09-04 — 태스크 8.8.1 이 더한 판정 (B8)

B8 (`gatedInScope == len(lanes)`, 119:3) 은 이 로트가 더한 유일한 분기다. B1~B7 은
좌표만 밀렸고 의미는 그대로다 — 새 분기가 함수 **끝**에 붙었으므로 레이블이 밀리지
않았다. 그것을 좌표와 조건으로 대조해 확인했고, 옛 산문을 새 분기에 붙이지 않았다.

세는 것은 "이 소유자 범위가 관문에 **전부** 빼앗겼는가"다. 일부만 빼앗긴 범위는
이웃 가족이 같은 범위에서 이기므로 목록 길이가 그대로이고, 그것이 이 change 가
사려던 가족 단위 격리다. 전부 빼앗긴 범위는 목록을 짧게 만들고, 짧아진 목록이
`strategyhandoff.Capacity = 1` 관문을 오히려 만족시킨다.
