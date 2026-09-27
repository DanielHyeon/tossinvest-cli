# Function Logic Map: `pairedStrategyDispatchCycleFixture`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go` (226-279)
- Function: `pairedStrategyDispatchCycleFixture` in package `engine`
- File SHA-256: `ffbd3a816543468aad30bd6d2264c5fbb834be2e6e3385f77d80d5944b4ac281`
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

Exact AST return positions: 275:108, 278:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 233:2 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B2 | if | 237:3 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B3 | if | 243:3 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B4 | if | 252:3 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B5 | else | 254:10 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B6 | if | 262:2 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B7 | if | 268:2 | test-file helper — 시험 파일은 커버리지 계측 밖이라 진입 수를 잴 수 없다. 이 fixture 를 부르는 시험 16 개가 전부 PASS: `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached`, `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`, `TestARefusedHandoffLeavesTheWorkerDormant`, `TestASingleSelectedScopeIsTheValueThatCrossesTheSeam`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheLaneStageOnItsOwnCallsTheGatewayZeroTimes`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |

8.7.2 편집: 끝에 `cycle.now = fakeClock.Now` 한 줄(생산 조립과 같은 모양). 분기는 바뀌지 않았다(편집 전·후 7 개, 좌표 불변).

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 227:2 |
| `newStrategyRiskLoaderFixture` | 228:17 |
| `riskFixture.loader.collect` | 230:14 |
| `context.Background` | 230:41 |
| `riskFixture.results.forMarket` | 234:13 |
| `strategyproposal.ProductionBatchAuthorityForTest` | 235:12 |
| `string` | 235:80 |
| `batch.For` | 236:20 |
| `t.Fatal` | 238:4 |
| `strategyaccount.AuthorityForTest` | 250:15 |
| `now.Add` | 250:77 |
| `now.Add` | 250:100 |
| `strings.Repeat` | 251:15 |
| `pairedDispatchSchedule` | 258:14 |
| `clock.NewFake` | 259:15 |
| `journal.Open` | 260:12 |
| `context.Background` | 260:25 |
| `filepath.Join` | 260:69 |
| `t.TempDir` | 260:83 |
| `journal.FixedFSProber` | 261:13 |
| `t.Fatal` | 263:3 |
| `t.Cleanup` | 265:2 |
| `j.Close` | 265:25 |
| `execgw.NewRiskGuardian` | 266:19 |
| `risk.DefaultPolicy` | 267:11 |
| `costs.DefaultModel` | 267:40 |
| `t.Fatal` | 269:3 |
| `newProductionStrategyFirstLegAuthorityLoader` | 271:12 |
| `newStrategyFirstLegAdmissionBridge` | 272:14 |
| `newStrategyDispatchCycle` | 274:11 |

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
