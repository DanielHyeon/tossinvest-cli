# Function Logic Map: `strategyProposalAuthorityLoader.familyGateFor`

- Source: `internal/app/engine/strategy_family_activation.go` (93-112)
- Function: `strategyProposalAuthorityLoader.familyGateFor` in package `engine`
- File SHA-256: `4a1ac0a98c8a7c598c5850a3c006e81821c56bb9e11aad113d9b943b4ec59022`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (pre-edit).
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 전 기록 (태스크 8.7.2, 2026-09-27).** 이 번들은 편집 **전** 소스의 AST 다. 8.7.1 이 만든 신규 함수라 번들이 없었고, 8.7.2 가 내부를 바꾸므로 여기서 처음 만든다.

입력: 로더(`loader.lanes`·`loader.loadActivation` seam), 시장, 스케줄·경로 권한, 관측 시각. 활성화는 **이 주기에** 읽는다(묵은 ON 금지).

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 97:3, 109:3, 111:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 96:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B2 | if | 104:2 | arm entered 30x (engine tagged suite, pre-edit); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheProductionActivationLoaderRunsAndFindsNoManifest`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B3 | if | 108:2 | arm entered 32x (engine tagged suite, pre-edit); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheProductionActivationLoaderRunsAndFindsNoManifest`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestWithoutAVerifiedActivationCoordinationIsUnchanged` |

- B1(96:2) `loader == nil` → 영값 관문. 생산에서는 닿지 않는다(`collectMarket` 이 이 로더의 메서드).
- B2(104:2) seam 이 nil 이면 생산 구현 `loadFamilyActivation`.
- B3(108:2) `err != nil || !activation.Verified()` → R 109:3 **영값 관문**. 오류의 종류를 가리지 않는다 — Unavailable·Revoked·Expired 가 전부 여기로 온다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `load` | 107:21 |
| `activation.Verified` | 108:20 |
| `loader.lanes.lanesFor` | 111:59 |

## State mutations and fallbacks

없음. 관문 값을 만들어 돌려준다. 레인 목록은 `loader.lanes.lanesFor` 가 RLock 아래 복사한다.

## Safety conclusion

- **편집 전 사실:** B3 가 모든 로드 오류를 영값 관문(= 관문 없음 = 기존 경로)으로 접는다. 그래서 사람이 끈 가족이 활성화 만료·폐기 뒤 기존 경로로 되살아난다(넓히는 방향). 이것이 8.7.2 가 닫는 구멍이다.
- **편집 계획(Manager 승인, 사람 결정 전달 2026-09-27):** B3 를 둘로 가른다. `ErrProductionFamilyActivationUndeclared`(핀 없음 = 미배포)만 영값 관문(오늘과 같음)이고, 그 밖의 오류·미검증은 `rolledBack` 관문(네 레인 영값 → DORMANT → 범위 소멸 → FAMILY_GATE_CLOSED)이다. 판별 규칙은 strategyrouter 한 곳.
- High-risk impact: yes.
