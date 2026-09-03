# Function Logic Map: `strategyMarketArbitration.entries`

- Source: `internal/app/engine/strategy_market_coordinator.go` (106-116)
- Function: `strategyMarketArbitration.entries` in package `engine`
- Signature: `strategyMarketArbitration.entries(params=0, results=2)`
- File SHA-256: `91b72a6f52f9be492fc945dc5a3856f4d9cb4b8cdc9cc97832f0f08b698d096a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

조정자가 고른 순서 그대로 레인 권한 목록을 만든다. 태스크 5.4.2 가 만든 새 메서드다.

입력은 `strategyMarketArbitration` 이 들고 있는 두 가지다 — 조정자가 낸 `outcome.Selections`(봉인된 계보 신원)와
`byIdentity`(그 신원에서 이 프로세스의 레인 권한으로 되돌리는 색인). 불변식은 하나다:
**선택 하나마다 되돌릴 자리가 정확히 하나 있어야 한다.**

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
  묶지 않고 돌린 이 패키지가 커널 OOM 으로 데스크톱을 세 번 죽였기 때문이다(`engine.test`, anon-rss 약 36GB).
  원인은 이 lot 이 고친 조정자 용량이며, 측정 방법이 아니라 측정 대상이 문제였다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 73.8% of statements.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS, 63.6% of statements.
- **표본 수 주의(5.5-fix3 정정):** 위 두 값은 각각 **한 번** 잰 것이다. 같은 스위트를 3~5회 재측정하면 태그는 2912·2909·2909 of 3941, 무태그는 2498~2501 of 3936 으로 흔들린다. 즉 소수점 첫째 자리는 안정적이지 않으므로, 두 값을 다른 로트의 값과 자릿수까지 견주지 않는다.
- Per-test attribution set: 같은 태그 바이너리를 `-test.run '^<Test>$'` 로 하나씩 돌린 열 개의 프로파일.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 테스트별 진입 수의 합이 스위트 전체 진입 수와
  정확히 같다. 이 집합 밖의 테스트가 어느 arm 이든 들어갔다면 그 등식이 깨진다. 깨진 행은
  `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

Exact AST return positions: 111:4, 115:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range| 108:2 | arm entered 42x (engine tagged suite); arm not entered (engine untagged suite); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutASignedActivationCoordinationIsUnchanged` |
| B2 | if| 110:3 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 107:12 |
| `len` | 107:54 |
| `append` | 113:12 |

## State mutations and fallbacks

- AST assignments: 3. Defers: 0. Goroutine statements: 0.

## Safety conclusion

못 찾은 자리를 건너뛰면 그 종목이 조용히 사라진 채 목록만 짧아진다. 사라진 종목은 아무 기록도
남기지 않고, 남은 것들은 정상인 척 다음 관문을 통과한다. 그래서 건너뛰지 않고 `false` 를 돌려 시장을 닫는다.
