# Function Logic Map: `pairedStrategyDispatchCycleFixture`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go` (244-298)
- Function: `pairedStrategyDispatchCycleFixture` in package `engine`
- File SHA-256: `b0b9734d75c5e4fafa2b2033bd9c3d9660af813aa7d485b1840d2a1f8ebcd958`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

(편집 전 반환 좌표는 git 이력 — 현재 목록은 아래 Branches 절.)

## Inputs and invariants

시험 픽스처다. 생산 코드가 아니지만 이 change 가 편집했으므로 증거를 남긴다.

두 시장의 제안·계좌·위험·일정 권한을 세우고 공유 dispatch 주기를 만든다.
태스크 8.8.2 가 그 생성자에 `proposals` 를 넘기도록 한 줄을 바꿨다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | non-nil | 시험 런타임 | `t.Fatal` |
| 배선의 보호 세대 | 상수 9 (`strategyDispatchGatewaySpy`) | 같은 파일의 spy | 이 값이 바뀌면 ProtectionReady 하한 시험이 다른 것을 잰다 |

**이 픽스처가 증명하지 못하는 것.** 활성화는 여기서 영값이다. 하한 결속을 재는
시험은 `kr.activation` 을 자기가 채우고 `cycle.proposals` 를 다시 넣는다 —
픽스처가 대신 채우면 모든 시험이 활성화된 상태만 보게 된다.

## Branches and early returns

Exact AST return positions: 294:108, 297:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 251:2 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B2 | if | 255:3 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B3 | if | 261:3 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B4 | if | 271:3 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B5 | else | 273:10 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B6 | if | 281:2 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B7 | if | 287:2 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |

8.7.2 편집: 끝에 `cycle.now = fakeClock.Now` 한 줄(생산 조립과 같은 모양). 분기는 바뀌지 않았다(편집 전·후 7 개, 좌표 불변).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 245:2 |
| `newStrategyRiskLoaderFixture` | 246:17 |
| `riskFixture.loader.collect` | 248:14 |
| `context.Background` | 248:41 |
| `riskFixture.results.forMarket` | 252:13 |
| `strategyproposal.ProductionBatchAuthorityForTest` | 253:12 |
| `string` | 253:80 |
| `batch.For` | 254:20 |
| `t.Fatal` | 256:4 |
| `strategyaccount.AuthorityForTest` | 268:15 |
| `now.Add` | 268:77 |
| `now.Add` | 268:100 |
| `strings.Repeat` | 269:15 |
| `a112ScopedAccount` | 270:13 |
| `pairedDispatchSchedule` | 277:14 |
| `clock.NewFake` | 278:15 |
| `journal.Open` | 279:12 |
| `context.Background` | 279:25 |
| `filepath.Join` | 279:69 |
| `t.TempDir` | 279:83 |
| `journal.FixedFSProber` | 280:13 |
| `t.Fatal` | 282:3 |
| `t.Cleanup` | 284:2 |
| `j.Close` | 284:25 |
| `execgw.NewRiskGuardian` | 285:19 |
| `risk.DefaultPolicy` | 286:11 |
| `costs.DefaultModel` | 286:40 |
| `t.Fatal` | 288:3 |
| `newProductionStrategyFirstLegAuthorityLoader` | 290:12 |
| `newStrategyFirstLegAdmissionBridge` | 291:14 |
| `newStrategyDispatchCycle` | 293:11 |

## State mutations and fallbacks

- 임시 디렉터리에 원장을 열고 `t.Cleanup` 으로 닫는다.
- fallback 없음. 조립에 실패하면 시험이 멈춘다.

## Safety conclusion

- Safe edit boundary: 시험 전용. 생산 바이너리에 들어가지 않는다.
- High-risk impact: no — 다만 이 픽스처가 재는 대상은 High-risk 경로이므로,
  배선의 상수를 바꾸면 그것을 인용하는 시험이 다른 것을 재게 된다.

## 2026-09-27 — 태스크 8.7.2 편집 전 (현재 AST 와 SHA 일치 확인)

편집은 한 줄: 생산 조립과 같은 모양으로 `cycle.now = fakeClock.Now` 를 둔다. 시험이 fake clock 을 파도 뒤로 전진시켜 "수집 때 유효 ·
SUBMITTING 때 만료" 를 재려면 dispatch 주기가 그 시계를 봐야 한다. 분기는 바뀌지 않는다.
