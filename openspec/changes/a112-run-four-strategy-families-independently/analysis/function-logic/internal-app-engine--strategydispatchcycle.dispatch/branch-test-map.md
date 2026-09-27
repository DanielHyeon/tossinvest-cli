# Branch Test Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go` (75-211); file SHA-256 `9610abb34ee350fa3f19ad06cb41d279ebe3c19df44f0a60336dd83276c6c289`. AST branch positions are authoritative.

- Measurement regime (8.7.2 편집 뒤): 몸통 진입 count. engine tagged suite 바이너리(`-coverpkg=./internal/app/engine,./internal/strategyrouter`, -trimpath 없이)를 `systemd-run … MemoryMax=16G` 안에서 실행, 스위트 PASS; 전체 시험 509 개를 하나씩 돈 per-test 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh` · `a872_attribute.py`). 모든 행에서 시험별 합 == 스위트(ATTRIBUTION MISMATCH 0).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 78:2 | arm entered 1x (engine tagged suite, post-edit); `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall` |
| B2 | if | 81:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B3 | if | 86:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B4 | if | 93:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B5 | if | 97:2 | arm entered 4x (engine tagged suite, post-edit); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` |
| B6 | if | 114:2 | arm entered 10x (engine tagged suite, post-edit); `TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`, `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` |
| B7 | if | 115:3 | arm entered 1x (engine tagged suite, post-edit); `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` |
| B8 | if | 136:2 | arm entered 1x (engine tagged suite, post-edit); `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission` |
| B9 | if | 140:2 | arm entered 2x (engine tagged suite, post-edit); `TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission`, `TestTheOrderLeaseCannotOutliveTheFamilyActivation` |
| B10 | if | 144:2 | arm entered 4x (engine tagged suite, post-edit); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` |
| B11 | if | 148:2 | arm entered 2x (engine tagged suite, post-edit); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral` |
| B12 | if | 152:2 | arm entered 1x (engine tagged suite, post-edit); `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B13 | if | 156:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B14 | if | 163:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B15 | if | 169:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B16 | if | 182:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B17 | if | 188:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B18 | if | 192:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B19 | if | 205:4 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |

8.7.2 가 더한 분기: B8(시계 없음 + 검증된 가족 활성화 → admission 앞 거절), B9(admission 앞 `LeaseCeiling` 만료 오류). 최종 검사 클로저(`FinalAuthorityCheck`) 안의 분기는 하나(B19 205:4, 스케줄 재검증 오류)이고 가족 만료는 분기 없이 마지막 `LeaseCeiling` 의 오류를 그대로 돌려준다. FuncLit 안이라 이 함수의 B 목록에 잡힌다 — 클로저는 게이트웨이가 부르므로 이 스위트(스파이 게이트웨이)에서는 시험이 직접 불러 잰다(`TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires`·`TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder`). 첫 편집 전 B8~B16 은 B10~B17 로 밀렸다 — 조건을 소스와 하나씩 대조했다.

옛 표(8.8.2 까지의 조건-평가 regime 과 편집 전 좌표)는 이 번들의 git 이력에 있다 — 이 파일은 현재 소스만 적는다.

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
