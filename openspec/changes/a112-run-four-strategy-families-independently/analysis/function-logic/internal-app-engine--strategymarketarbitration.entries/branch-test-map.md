# Branch Test Map: `strategyMarketArbitration.entries`

- Source: `internal/app/engine/strategy_market_coordinator.go` (106-116); file SHA-256 `91b72a6f52f9be492fc945dc5a3856f4d9cb4b8cdc9cc97832f0f08b698d096a`. AST branch positions are authoritative.
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

M-E3 (`return nil, false` → `continue`) 가 이 메서드를 겨눈 mutation 이며
`collectmarket` 쪽 표에 결과를 적었다. 처음에는 SURVIVED 했고, 그것을 죽이는 테스트를 넣은 뒤 KILLED 다.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range| 108:2 | arm entered 42x (engine tagged suite); arm not entered (engine untagged suite); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B2 | if| 110:3 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
