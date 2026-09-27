# Function Logic Map: `LoadProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go` (383-427)
- Function: `LoadProductionFamilyActivation` in package `strategyrouter`
- File SHA-256: `411d612f7444c31744585b53c6128ccb71d0406a3e708efe9b7aee6232ca67d7`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (pre-edit).
- AST evidence: `ast.json` — AST branches 8.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 전 기록 (태스크 8.7.2, 2026-09-27).** 이 번들은 편집 **전** 소스의 AST 다. 8.7.1 이 만든 신규 함수라 번들이 없었고, 8.7.2 가 내부를 바꾸므로 여기서 처음 만든다.

입력: ctx, `FamilyActivationConfig`(설정 디렉터리·시장·핀·관측 시각·다섯 결속 값). 이 함수가 활성화를 발급하는 **유일한** 생산 경로다.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 385:3, 388:3, 400:3, 405:3, 409:3, 415:3, 419:3, 422:3, 425:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 384:2 | arm not entered (strategyrouter tagged suite, pre-edit); no per-test profile entered it |
| B2 | if | 387:2 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestACancelledContextPromotesNothing` |
| B3 | if | 394:2 | arm not entered (strategyrouter tagged suite, pre-edit); no per-test profile entered it |
| B4 | if | 404:2 | arm entered 5x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestBytesThatChangedAfterTheDeploymentPinnedThemPromoteNothing`, `TestGoldenBytesThatDriftFromThePinPromoteNothing`, `TestWithNoActivationFileNothingIsPromoted` |
| B5 | if | 408:2 | arm entered 4x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseBytesAreNotCanonicalPromotesNothing` |
| B6 | if | 414:2 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing` |
| B7 | if | 418:2 | arm entered 26x (strategyrouter tagged suite, pre-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing`, `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing`, `TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing` |
| B8 | if | 421:2 | arm not entered (strategyrouter tagged suite, pre-edit); no per-test profile entered it |

- B1(384:2) ctx nil → Unavailable. B2(387:2) ctx 취소 → ctx 오류.
- B3(394:2) 소유자 UID·파일 이름·절대 경로·관측 시각·핀 형식·결속 값 형식 → Unavailable. **핀이 비어 있어도 여기로 온다** — 편집 전에는 "선언 안 함" 과 "잘못 선언함" 이 같은 값이다.
- B4(404:2) 파일 읽기 실패·핀 불일치 → Unavailable. B5(408:2) 정규 바이트 아님 → Unavailable.
- B6(414:2) 폐기 → Revoked. B7(418:2) 결속·수명·서술자 → validate 의 오류(Expired 포함). B8(421:2) ctx 재확인.

엔진 스위트(편집 전)에서 이 함수의 분기 진입:

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 384:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B2 | if | 387:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B3 | if | 394:2 | arm entered 31x (engine tagged suite, pre-edit); `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheProductionActivationLoaderRunsAndFindsNoManifest`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| B4 | if | 404:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B5 | if | 408:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B6 | if | 414:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B7 | if | 418:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B8 | if | 421:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |

엔진 스위트는 B3 만 들어간다 — 엔진 시험 대부분이 핀을 두지 않으므로(= 오늘 생산) 핀 형식 검사에서 끝난다.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `ctx.Err` | 387:12 |
| `filepath.Clean` | 390:21 |
| `strings.TrimSpace` | 390:36 |
| `strings.TrimSpace` | 391:26 |
| `productionRouteOwnerUID` | 392:20 |
| `ProductionFamilyActivationFileName` | 393:10 |
| `filepath.IsAbs` | 394:32 |
| `config.ObservedAt.IsZero` | 394:68 |
| `productionRouteDigestValid` | 395:4 |
| `productionRouteIdentity` | 396:4 |
| `productionRouteDigestValid` | 397:4 |
| `productionRouteDigestValid` | 398:4 |
| `productionRouteIdentity` | 399:4 |
| `productionRouteIdentity` | 399:56 |
| `readProductionRouteFile` | 402:15 |
| `filepath.Join` | 402:39 |
| `productionRouteDigest` | 404:19 |
| `decodeProductionFamilyActivation` | 407:19 |
| `validateProductionFamilyActivation` | 417:16 |
| `ctx.Err` | 421:12 |
| `productionRouteTime` | 424:16 |

## State mutations and fallbacks

파일 하나를 읽기만 한다(`readProductionRouteFile`, `0400`·현재 UID). 쓰기 없음.

## Safety conclusion

- **편집 계획:** 맨 앞(ctx 검사보다도 앞)에 `strings.TrimSpace(config.ManifestDigest) == ""` → `ErrProductionFamilyActivationUndeclared` 를 둔다. 앞에 두는 이유: 핀이 없는 시장의 답이 ctx 상태·다른 결속 값에 따라 바뀌면 오늘 생산(핀 없음)이 rolled back 으로 읽힐 수 있다. 미선언은 설정 사실 하나로만 정한다.
- 이 편집은 승격 경로를 넓히지 않는다: 새 갈래는 **오류만** 돌려준다.
- High-risk impact: yes.
