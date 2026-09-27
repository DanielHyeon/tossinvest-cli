# Function Logic Map: `strategyProposalAuthorityLoader.loadFamilyActivation`

- Source: `internal/app/engine/strategy_family_activation.go` (124-156)
- Function: `strategyProposalAuthorityLoader.loadFamilyActivation` in package `engine`
- File SHA-256: `4a1ac0a98c8a7c598c5850a3c006e81821c56bb9e11aad113d9b943b4ec59022`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (pre-edit).
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 전 기록 (태스크 8.7.2, 2026-09-27).** 이 번들은 편집 **전** 소스의 AST 다. 8.7.1 이 만든 신규 함수라 번들이 없었고, 8.7.2 가 내부를 바꾸므로 여기서 처음 만든다.

입력: 시장(→ 핀 env 이름), 경로 권한(→ 보정 digest·경로 매니페스트 digest), 스케줄 권한(→ 달력), 관측 시각, 위험 정책 핀 env.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 133:3, 149:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 128:2 | arm entered 14x (engine tagged suite, pre-edit); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B2 | if | 132:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B3 | if | 146:2 | arm entered 14x (engine tagged suite, pre-edit); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |

- B1(128:2) US 면 US 핀 이름.
- B2(132:2) 경로 항목들의 보정 digest 가 하나로 합의되지 않으면 R 133:3 Unavailable — **핀을 보기 전**이다. 편집 전 측정: 진입 0.
- B3(146:2) US 면 US 위험 정책 env.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyMarketCalibrationDigest` | 131:25 |
| `strategyrouter.LoadProductionFamilyActivation` | 149:9 |
| `strategyRouterMarket` | 150:40 |
| `strings.TrimSpace` | 151:19 |
| `loader.getenv` | 151:37 |
| `strategyRuntimeBuildDigest` | 153:60 |
| `strings.TrimSpace` | 154:21 |
| `loader.getenv` | 154:39 |

## State mutations and fallbacks

없음. env 두 개를 읽고 strategyrouter 로더를 부른다.

## Safety conclusion

- **편집 계획:** B2 의 조기 반환을 지운다. 그대로 두면 핀이 없는(= 오늘 생산) 시장에서도 보정 불일치가 Unavailable 로 나가 새 판별에서 **rolled back** 으로 읽히고, 그것은 생산 동작 변화다. 보정 값은 그대로 넘기고(불일치면 빈 문자열) 판정은 strategyrouter 가 한다 — 핀 없음이 먼저, 그다음 `productionRouteIdentity("")` 가 거절(파일을 열기 전).
- High-risk impact: yes.
