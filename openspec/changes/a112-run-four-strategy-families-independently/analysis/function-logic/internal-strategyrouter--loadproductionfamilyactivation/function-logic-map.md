# Function Logic Map: `LoadProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go` (427-478)
- Function: `LoadProductionFamilyActivation` in package `strategyrouter`
- File SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (post-edit).
- AST evidence: `ast.json` — AST branches 9.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 뒤 (태스크 8.7.2, 2026-09-27).** 편집 전 번들은 커밋 `c1d1e295`(validate 는 `cb378a63`) 에 있다 — 이 파일은 GREEN 뒤 현재 소스의 AST 와 측정이다.

편집: 맨 앞 B1 — 빈 핀(`strings.TrimSpace` 뒤) → `ErrProductionFamilyActivationUndeclared`. ctx·결속 값보다 **앞**이다: 핀 없는
시장의 답이 주기의 사정에 흔들리지 않게.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite (and the strategyrouter tagged suite for the two router bundles, same flags): `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 433:3, 436:3, 439:3, 451:3, 456:3, 460:3, 466:3, 470:3, 473:3, 476:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 432:2 | arm entered 12x (strategyrouter tagged suite, post-edit); `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared` |
| B2 | if | 435:2 | arm not entered (strategyrouter tagged suite, post-edit); no per-test profile entered it |
| B3 | if | 438:2 | arm entered 1x (strategyrouter tagged suite, post-edit); `TestACancelledContextPromotesNothing` |
| B4 | if | 445:2 | arm entered 1x (strategyrouter tagged suite, post-edit); `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared` |
| B5 | if | 455:2 | arm entered 6x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestBytesThatChangedAfterTheDeploymentPinnedThemPromoteNothing`, `TestGoldenBytesThatDriftFromThePinPromoteNothing`, `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared`, `TestWithNoActivationFileNothingIsPromoted` |
| B6 | if | 459:2 | arm entered 4x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseBytesAreNotCanonicalPromotesNothing` |
| B7 | if | 465:2 | arm entered 2x (strategyrouter tagged suite, post-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared` |
| B8 | if | 469:2 | arm entered 29x (strategyrouter tagged suite, post-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing`, `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing`, `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant`, `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared`, `TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing` |
| B9 | if | 472:2 | arm not entered (strategyrouter tagged suite, post-edit); no per-test profile entered it |

표는 strategyrouter 태그 스위트(전체 시험 71 개 per-test)다. 엔진 스위트(전체 509 개 per-test)에서의 진입:

- 엔진 B1: arm entered 41x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestALaneRefusesALineageThatRenamedItselfIntoAnotherLane`, `TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`, `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAnUndeclaredMarketStaysUndeclaredWhateverItsCalibrationSays`, `TestEntriesComeBackInOwnerScopeOrderNotRouteOrder`, `TestEveryLaneStaysDormantOnAProposalItActuallyOwns`, `TestExactlyOneLaneOwnsEachSealedProposal`, `TestRollingBackLeavesALatchedLaneLatched`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated`, `TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`
- 엔진 B2: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B3: arm entered 1x (engine tagged suite, post-edit); `TestADeclaredMarketWhoseCycleIsCancelledRollsBack`
- 엔진 B4: arm entered 2x (engine tagged suite, post-edit); `TestADeclaredMarketWhoseCalibrationDoesNotAgreeIsRolledBack`
- 엔진 B5: arm entered 8x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`, `TestARolledBackGateClosesTheMarketEvenWithoutLanes`, `TestRollingBackLeavesALatchedLaneLatched`, `TestTheProductionActivationLoaderRunsAndFindsNoManifest`
- 엔진 B6: arm not entered (engine tagged suite, post-edit); no per-test profile entered it
- 엔진 B7: arm entered 2x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`
- 엔진 B8: arm entered 2x (engine tagged suite, post-edit); `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`
- 엔진 B9: arm not entered (engine tagged suite, post-edit); no per-test profile entered it

**통과 이유가 바뀐 시험.** 편집 전 엔진 스위트에서 이 함수의 진입은 B3(형식 검사) 31 회였다 — 핀을 두지 않은 엔진 시험들이
"핀 형식 무효" 로 거절돼 영값 관문을 받았다. 편집 뒤 같은 시험들은 B1(미선언)으로 41 회 들어간다. 결과(영값 관문 → 기존 경로)는
같고 이유가 바뀌었다; 되돌림 판별이 그 이유에 기대므로 이 이동이 곧 의도다. 이제 B4(형식 검사)에 엔진이 들어가는 것은 보정
불일치 시험 하나뿐이고 그 문을 M18 이 잰다.

구멍: B2(ctx nil + 핀 있음)는 두 스위트 모두 진입 0 — 편집 전에도 0 이었다. B9(끝 ctx 재확인)도 0.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strings.TrimSpace` | 432:5 |
| `ctx.Err` | 438:12 |
| `filepath.Clean` | 441:21 |
| `strings.TrimSpace` | 441:36 |
| `strings.TrimSpace` | 442:26 |
| `productionRouteOwnerUID` | 443:20 |
| `ProductionFamilyActivationFileName` | 444:10 |
| `filepath.IsAbs` | 445:32 |
| `config.ObservedAt.IsZero` | 445:68 |
| `productionRouteDigestValid` | 446:4 |
| `productionRouteIdentity` | 447:4 |
| `productionRouteDigestValid` | 448:4 |
| `productionRouteDigestValid` | 449:4 |
| `productionRouteIdentity` | 450:4 |
| `productionRouteIdentity` | 450:56 |
| `readProductionRouteFile` | 453:15 |
| `filepath.Join` | 453:39 |
| `productionRouteDigest` | 455:19 |
| `decodeProductionFamilyActivation` | 458:19 |
| `validateProductionFamilyActivation` | 468:16 |
| `ctx.Err` | 472:12 |
| `productionRouteTime` | 475:16 |

## State mutations and fallbacks

파일 하나를 읽기만 한다. 쓰기 없음.

## Safety conclusion

- 변이 M1(미선언 삭제)·M2(미선언을 ctx 검사 뒤로) CAUGHT.
- 새 갈래는 **오류만** 돌려준다 — 승격 경로를 넓히지 않는다.
- High-risk impact: yes.
