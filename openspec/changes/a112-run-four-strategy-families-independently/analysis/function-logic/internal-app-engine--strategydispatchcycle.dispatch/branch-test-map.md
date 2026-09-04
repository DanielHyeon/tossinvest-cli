# Branch Test Map: `dispatch`

- Source SHA-256: `0ce70d7b683d586d4224440b2fe66df7e018caacdb20b7c5ae1f46e7ad98d7b1`; AST branch locations are authoritative.
- L0 did not alter this function and does not claim an existing test covers a branch.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 72:2 | arm entered 1x (tagged); arm not entered (untagged); `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B2 | if at 75:2 | arm entered 19x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B3 | if at 80:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B4 | if at 87:2 | arm entered 19x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B5 | if at 91:2 | arm entered 4x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B6 | if at 108:2 | arm entered 15x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B7 | if at 109:3 | arm entered 1x (tagged); arm not entered (untagged); `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B8 | if at 116:2 | arm entered 4x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B9 | if at 120:2 | arm entered 2x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B10 | if at 124:2 | arm entered 1x (tagged); arm not entered (untagged); `TestTheSameEnvelopeCannotPlaceASecondOrder` | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B11 | if at 128:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B12 | if at 135:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B13 | if at 141:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B14 | if at 154:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B15 | if at 160:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |
| B16 | if at 164:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it | n/a — 8.8.2 는 이 함수의 이 분기를 편집하지 않았다 | 예 |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.
