# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go` (279-427)
- Function: `strategyProposalAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- File SHA-256: `913050cc0cc0763295af577e49fbb4ccb7d4e838fbfc3408f0f33057fbbe2418`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 16.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 시장의 경로 항목들을 봉인된 제안 항목으로 바꾼다. 태스크 5.4.2 이후 한 종목의 여러 가족 제안을
그 자리에서 훑지 않는다 — `coordinateMarketProposals` 가 시장 조정자에 넣고, 조정자가 소유자 범위마다 하나를 고른다.
목록의 순서도 조정자가 정한다(소유자 범위 사전순이라 종목 오름차순). 그 순서가 `ProposalSetDigest` 에 그대로 들어간다.

조정자가 한 범위를 닫으면 그 종목만 빼지 않고 **시장 전체를 닫는다** — 하나를 빼면 목록이 둘에서 하나로 줄어
아래 세 소비자가 공유하는 `len(entries) != 1` 관문이 오히려 만족되고, 막으려던 것과 상관없는 다른 종목이 대신 풀린다.

**이 불변식은 5.4.2 까지 중재 결함에만 성립했다. 태스크 5.4.3 이 나머지 절반을 채웠다.**
`batch.LanesFor` 가 빈 목록을 돌려주는 일은 "이 종목은 원래 제안이 없다"와 "이 종목의 제안이
고장으로 사라졌다"를 구별하지 못했고, 그래서 고장 하나가 목록을 줄이고 줄어든 목록이
`len(entries) != 1` 관문을 오히려 만족시켜 상관없는 다른 종목을 풀어 줬다(5.4.2 리뷰 P1-1,
시연으로 확인). 지금은 `strategyproposal` 이 그 둘을 갈라 형(型)으로 들고 나오고, 이 함수가
배치를 받자마자 `batch.Fault()` 를 먼저 본다 — 고장이면 조정에 넣지 않고 시장을 닫는다.

고장의 기준은 "제안이 안 나왔다"가 아니라 **"약속받은 봉인된 입력을 얻지 못했다"** 다.
시장 상태 때문에 평가가 거절한 것(스프레드·호가·사이징·보호적이지 않은 목표가)은 고장이
아니라 매일 일어나는 정상 거절이므로 예전처럼 거절로 세고 시장은 열어 둔다. 그 경계는
처음에 잘못 그었고 `TestAnEvaluationRefusalIsAbsenceNotFault` 가 잡았다.

큐가 넘치면 중재 거절 코드를 빌려 쓰지 않고 `PROPOSAL_QUEUE_OVERFLOW` 라는 엔진 자신의 이름으로 보고한다.
동결 골든 `refusal_enums` 에 큐 코드가 없으므로, 빌려 쓰면 큐가 넘친 일이 봉인이 깨진 일로 보고된다.

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

Exact AST return positions: 294:3, 299:3, 302:3, 305:3, 310:3, 322:4, 334:3, 343:3, 374:3, 382:3, 391:3, 400:3, 409:3, 416:3, 422:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 298:2 | arm entered 50x (engine tagged suite); arm entered 2x (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B2 | if | 301:2 | arm entered 45x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B3 | if | 304:2 | arm entered 45x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B4 | if | 309:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B5 | if | 313:2 | arm entered 21x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B6 | range | 319:2 | arm entered 10065x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B7 | if | 321:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | if | 333:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | if | 338:2 | arm entered 44x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B10 | if | 369:2 | arm entered 7x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestTheGateSaysWhichKindOfStopItMade` |
| B11 | if | 376:2 | arm entered 36x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B12 | if | 385:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` |
| B13 | if | 393:2 | arm entered 35x (engine tagged suite); arm not entered (engine untagged suite); `TestAClosedMarketStillCarriesTheGatesActivation`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B14 | if | 403:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B15 | if | 411:2 | arm entered 31x (engine tagged suite); arm not entered (engine untagged suite); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B16 | range | 419:2 | arm entered 43x (engine tagged suite); arm not entered (engine untagged suite); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheGateSaysWhichKindOfStopItMade`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 295:90 |
| `len` | 298:31 |
| `fail` | 299:10 |
| `fail` | 302:10 |
| `fail` | 305:10 |
| `strings.TrimSpace` | 307:13 |
| `loader.getenv` | 307:31 |
| `DecodeString` | 308:14 |
| `base64.StdEncoding.Strict` | 308:14 |
| `base64.StdEncoding.EncodeToString` | 309:19 |
| `len` | 309:72 |
| `fail` | 310:10 |
| `strings.TrimSpace` | 316:12 |
| `loader.getenv` | 316:30 |
| `make` | 317:13 |
| `len` | 317:58 |
| `make` | 318:14 |
| `len` | 318:59 |
| `entry.approved.Symbol` | 320:13 |
| `bySymbol.approved.Valid` | 321:22 |
| `fail` | 322:11 |
| `append` | 325:13 |
| `entry.route.Request` | 325:97 |
| `loader.load` | 327:16 |
| `strategyrouter.Market` | 328:42 |
| `strings.TrimSpace` | 328:111 |
| `loader.getenv` | 328:129 |
| `ed25519.PublicKey` | 329:15 |
| `strings.TrimSpace` | 332:23 |
| `loader.getenv` | 332:41 |
| `batch.ManifestDigest` | 333:19 |
| `fail` | 334:10 |
| `batch.Fault` | 338:22 |
| `fail` | 339:13 |
| `absence.String` | 341:37 |
| `loader.familyGateFor` | 357:9 |
| `coordinateMarketProposals` | 358:26 |
| `len` | 359:30 |
| `distinctGatedOutcomes` | 359:54 |
| `fail` | 370:13 |
| `fail` | 377:13 |
| `fail` | 386:13 |
| `fail` | 394:13 |
| `string` | 396:40 |
| `arbitration.entries` | 402:23 |
| `fail` | 404:13 |
| `len` | 411:5 |
| `fail` | 412:13 |
| `sha256.New` | 418:7 |
| `h.Write` | 420:10 |
| `(unnamed)` | 420:18 |
| `entry.route.approved.Symbol` | 420:25 |
| `entry.authority.Proposal` | 420:66 |
| `len` | 424:17 |
| `len` | 424:53 |
| `hex.EncodeToString` | 425:35 |
| `h.Sum` | 425:54 |

## State mutations and fallbacks

- AST assignments: 49. Defers: 0. Goroutine statements: 0.

## Safety conclusion

이 함수는 거절을 느슨하게 한 적이 없다. 5.4.2 가 닫히는 경우를 셋 늘렸고 — 계보 신원
충돌(B10), 큐 넘침(B11), 되돌릴 자리 없음(B13) — 5.4.3 이 넷째를 더했다: 배치가 받아들인
스코프의 제안을 잃음(B9). 넷 다 이전에는 존재하지 않던 닫힘이며 전부 보수 방향이다.
B10 은 아직 진입 0 회다. 이것은 통과가 아니라 커버리지 구멍이며 그대로 남겨 보고한다.

B9 를 **조정 앞**에 둔 이유: 조정까지 가면 잃은 종목은 그냥 없는 종목처럼 보이고,
짧아진 목록이 아래 관문을 오히려 만족시킨다. 닫으려면 줄어들기 전에 봐야 한다.

`ProductionFault` 는 어느 종목의 어느 레인이 무엇 때문에 사라졌는지를 한 줄로 싣는다.
`QueueDropCount` 와 마찬가지로 **아직 운영자에게 닿지 않는다** — 투영은 태스크 7.3 이다.

닫힘 가지 넷은 이제 `RefusedCount` 에 `RoutedCount` 를 그대로 싣는다. 시장이 닫히면 경로에 오른
종목 전부가 못 나간 것이기 때문이다. 앞선 판본의 `refused + 1` 은 10,001 종목이 하나도 못 나간
주기를 "1건 거절"로 보고했다.

`QueueDropCount` 는 지금 배선에서 **언제나 0** 이다. 계수기가 오르는 두 곳(접힘·넘침)이 둘 다
구조적으로 도달 불가이기 때문이다 — 종목 중복이 위(B7)에서 먼저 거절되고 매니페스트가
(종목, 레인)마다 하나만 실어 같은 칸이 두 번 오지 않으며, 큐 상한이 매니페스트 상한과 같아
정상 매니페스트는 넘칠 수 없다. 그러므로 이 수를 "조용한 유실이 없음의 증거"로 인용하지 않는다.

## 2026-09-04 — 태스크 8.8.1 이 더한 판정 (B10)

B10 (`arbitration.erasedScopes != 0`, 369:2) 이 이 로트가 더한 분기이고, **함수
가운데에 들어갔으므로 옛 B10~B15 가 B11~B16 으로 밀렸다.** 밀린 여섯은 조건을
소스와 하나씩 대조해 확인했다(옛 B10 collision → 새 B11, … 옛 B15 → 새 B16).
레이블만 보고 산문을 옮기면 이 자리에서 정확히 틀린다.

**측정에 대한 정직한 주의.** B10 은 태그 스위트에서 7 회 진입하는데 그중 다섯은
KR 이 아니라 **US** 시장이다. `collectUnderGate` 의 활성화 stub 이 시장을 무시하고
언제나 KR 활성화를 돌려주므로, US 레인은 그 활성화에서 자기 열쇠를 못 찾아 전부
DORMANT 가 되고 US 의 유일한 범위(AAPL)가 통째로 지워진다. 즉 이 수는 "가드가
KR 에서 일곱 번 발화했다"가 아니다. KR 쪽 진입은
`TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve` 의 두 번 중 한 번뿐이고,
그 시험이 KR 권위만 읽어 판정하므로 결론은 영향받지 않는다. fixture 의 이 성질을
고치는 일은 8.8.4 에 있다.
