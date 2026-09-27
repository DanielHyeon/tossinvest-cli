# Function Logic Map: `strategyProposalAuthorityLoader.loadFamilyActivation`

- Source: `internal/app/engine/strategy_family_activation.go` (154-190)
- Function: `strategyProposalAuthorityLoader.loadFamilyActivation` in package `engine`
- File SHA-256: `230cc4c84bc3ff2bec2b98caaed10cdec7fd18e46181e3ce995899bbb0ee3492`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (post-edit).
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 뒤 (태스크 8.7.2, 2026-09-27).** 편집 전 번들은 커밋 `c1d1e295`(validate 는 `cb378a63`) 에 있다 — 이 파일은 GREEN 뒤 현재 소스의 AST 와 측정이다.

편집: 보정 합의 실패 시 핀을 보기 전에 Unavailable 을 돌려주던 조기 반환(편집 전 B2 132:2)을 지웠다. 합의가 안 되면 빈 보정 값을
넘기고, 판정은 strategyrouter 가 한다 — 미선언이 먼저, 빈 보정 값은 그다음 결속 형식 검사가 파일을 열기 전에 거절한다.
막던 입력의 행방은 review.md 2026-09-27 절의 표.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite (and the strategyrouter tagged suite for the two router bundles, same flags): `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 183:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 158:2 | arm entered 22x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestRollingBackLeavesALatchedLaneLatched`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B2 | if | 180:2 | arm entered 22x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestRollingBackLeavesALatchedLaneLatched`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |

- B1 US 면 US 핀 이름. B2 US 면 US 위험 정책 env. (조기 반환이 사라져 분기가 셋에서 둘이 되었다.)

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyMarketCalibrationDigest` | 168:20 |
| `strategyrouter.LoadProductionFamilyActivation` | 183:9 |
| `strategyRouterMarket` | 184:40 |
| `strings.TrimSpace` | 185:19 |
| `loader.getenv` | 185:37 |
| `strategyRuntimeBuildDigest` | 187:60 |
| `strings.TrimSpace` | 188:21 |
| `loader.getenv` | 188:39 |

## State mutations and fallbacks

없음. env 두 개를 읽고 strategyrouter 적재기를 부른다.

## Safety conclusion

- 변이 M12(조기 반환 복원 → `TestAnUndeclaredMarketStaysUndeclaredWhateverItsCalibrationSays`)·M18(router 의 보정 형식 검사 삭제 → `TestADeclaredMarketWhoseCalibrationDoesNotAgreeIsRolledBack`) CAUGHT.
- High-risk impact: yes.
