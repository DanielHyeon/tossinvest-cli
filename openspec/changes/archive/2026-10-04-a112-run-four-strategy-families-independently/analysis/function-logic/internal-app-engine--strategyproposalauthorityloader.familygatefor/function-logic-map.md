# Function Logic Map: `strategyProposalAuthorityLoader.familyGateFor`

- Source: `internal/app/engine/strategy_family_activation.go` (115-142)
- Function: `strategyProposalAuthorityLoader.familyGateFor` in package `engine`
- File SHA-256: `50c775155bc8a12a5844ddbc30785f72e49b4fab7380146071e458f815d4c377`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (post-edit).
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 뒤 (태스크 8.7.2, 2026-09-27).** 편집 전 번들은 커밋 `c1d1e295`(validate 는 `cb378a63`) 에 있다 — 이 파일은 GREEN 뒤 현재 소스의 AST 와 측정이다.

판별은 하나다: 적재 오류가 `strategyrouter.ErrProductionFamilyActivationUndeclared` 일 때만 영값 관문(기존 경로 — 핀 없는 시장,
오늘 생산). 그 밖의 오류나 미검증 값은 `rolledBack` 관문(레인 목록과 함께). 검증된 값만 승격 관문.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite (and the strategyrouter tagged suite for the two router bundles, same flags): `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 119:3, 133:3, 139:3, 141:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 118:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B2 | if | 125:2 | arm entered 49x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestADeclaredMarketWhoseCalibrationDoesNotAgreeIsRolledBack`, `TestADeclaredMarketWhoseCycleIsCancelledRollsBack`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAnUndeclaredMarketStaysUndeclaredWhateverItsCalibrationSays`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestRollingBackLeavesALatchedLaneLatched`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheProductionActivationLoaderRunsAndFindsNoManifest`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B3 | if | 132:2 | arm entered 41x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAnUndeclaredMarketStaysUndeclaredWhateverItsCalibrationSays`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestRollingBackLeavesALatchedLaneLatched`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |
| B4 | if | 138:2 | arm entered 11x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestADeclaredMarketWhoseCalibrationDoesNotAgreeIsRolledBack`, `TestADeclaredMarketWhoseCycleIsCancelledRollsBack`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestAnUnverifiedActivationWithoutAnErrorStillRollsBack`, `TestRollingBackLeavesALatchedLaneLatched`, `TestTheProductionActivationLoaderRunsAndFindsNoManifest` |

- B1 `loader == nil` → 영값 관문. 생산 미도달(`collectMarket` 이 이 로더의 메서드) — 진입 0 은 구멍으로 적는다.
- B2 seam 이 nil 이면 생산 적재기.
- B3 미선언 → 영값 관문(R 다음 줄). **엔진 시험 대부분이 여기를 지난다** — 핀을 두지 않으므로.
- B4 오류·미검증 → `rolledBack` 관문.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `load` | 128:21 |
| `errors.Is` | 132:5 |
| `loader.lanes.lanesFor` | 135:11 |
| `activation.Verified` | 138:20 |

## State mutations and fallbacks

없음. 관문 값을 만든다. `lanesFor` 는 RLock 아래 복사.

## Safety conclusion

- 변이 M8(모든 오류를 기존 경로로 — 편집 전 동작)·M9(되돌림 표식 빠짐)·M17(무오류·미검증을 기존 경로로)·RB3(취소 ctx 를 기존 경로로) 전부 CAUGHT.
- census 가 호출을 `load`·`errors.Is`·`activation.Verified`·`loader.lanes.lanesFor` 로 못 박는다.
- High-risk impact: yes.

a112 8.5 응답 로트(2026-10-04): 같은 파일의 다른 함수 편집으로 줄만 밀림 — shift_same_file_bundles.py(구조 동일 확인 뒤 좌표 사상)
