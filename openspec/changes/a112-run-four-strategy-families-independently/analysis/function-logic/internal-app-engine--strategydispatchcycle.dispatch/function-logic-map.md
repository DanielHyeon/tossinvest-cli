# Function Logic Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go`
- Current-base source SHA-256: `0ce70d7b683d586d4224440b2fe66df7e018caacdb20b7c5ae1f46e7ad98d7b1`
- Signature: `strategyDispatchCycle.dispatch(params=2, results=2)`
- Source range: `63:1`–`149:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `73:3, 76:3, 82:3, 88:3, 92:3, 110:4, 117:3, 121:3, 125:3, 129:3, 136:3, 142:3, 155:3, 161:3, 165:3, 167:2, 174:4`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 72:2 | arm entered 1x (tagged); arm not entered (untagged); `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall` |
| B2 | if | 75:2 | arm entered 19x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B3 | if | 80:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |
| B4 | if | 87:2 | arm entered 19x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B5 | if | 91:2 | arm entered 4x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` |
| B6 | if | 108:2 | arm entered 15x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`, `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner`, `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`, `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B7 | if | 109:3 | arm entered 1x (tagged); arm not entered (untagged); `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` |
| B8 | if | 116:2 | arm entered 4x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` |
| B9 | if | 120:2 | arm entered 2x (tagged); arm not entered (untagged); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral` |
| B10 | if | 124:2 | arm entered 1x (tagged); arm not entered (untagged); `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B11 | if | 128:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |
| B12 | if | 135:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |
| B13 | if | 141:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |
| B14 | if | 154:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |
| B15 | if | 160:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |
| B16 | if | 164:2 | arm not entered (tagged); arm not entered (untagged); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `delivered.Result` | 70:12 |
| `validateStrategyFirstLegResult` | 71:23 |
| `errors.New` | 73:28 |
| `errors.New` | 76:28 |
| `StrategyMarket` | 78:12 |
| `cycle.schedule.forMarket` | 79:18 |
| `cycle.fx.forMarket` | 79:52 |
| `errors.New` | 82:28 |
| `strategyFirstLegPlaceIntent` | 87:15 |
| `cycle.gateway.ObserveStrategyProtection` | 90:21 |
| `strings.ToLower` | 90:66 |
| `string` | 90:82 |
| `familyActivation` | 108:19 |
| `cycle.proposals.forMarket` | 108:19 |
| `activation.Verified` | 108:73 |
| `protection.Generation` | 109:6 |
| `activation.ProtectionReadyMinGeneration` | 109:32 |
| `fmt.Errorf` | 110:29 |
| `protection.Generation` | 112:5 |
| `activation.ProtectionReadyMinGeneration` | 112:30 |
| `cycle.gateway.ObserveStrategyEntryGate` | 115:25 |
| `strings.ToLower` | 115:69 |
| `string` | 115:85 |
| `cycle.dispatchOwner` | 119:16 |
| `cycle.firstLeg.admit` | 123:14 |
| `fmt.Errorf` | 125:28 |
| `cycle.journal.LookupDecision` | 127:19 |
| `errors.New` | 129:28 |
| `uint64` | 133:24 |
| `bundle.Generation` | 134:20 |
| `cycle.risk.forMarket` | 134:20 |
| `errors.New` | 136:28 |
| `schedule.restore.Activation.Generation` | 138:26 |
| `schedule.restore.Activation.ExpiresAt` | 139:25 |
| `activationExpiresAt.IsZero` | 141:34 |
| `now.IsZero` | 141:66 |
| `now.Before` | 141:83 |
| `errors.New` | 142:28 |
| `journal.StrategyDispatchMarket` | 144:63 |
| `protection.Generation` | 147:25 |
| `strconv.FormatUint` | 147:68 |
| `protection.Generation` | 147:87 |
| `protection.Digest` | 147:135 |
| `reconciliation.Generation` | 148:29 |
| `reconciliation.Digest` | 148:80 |
| `strategyRuntimeBuildDigest` | 149:94 |
| `min` | 150:9 |
| `activationExpiresAt.Sub` | 150:29 |
| `cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease` | 151:16 |
| `cycle.journal.ClaimStrategyDispatchLease` | 157:18 |
| `strategyFirstLegPlaceIntent` | 163:17 |
| `cycle.gateway.PlaceClaimedStrategy` | 167:9 |
| `cycle.revalidateSchedule` | 174:11 |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

## 2026-09-04 — 태스크 8.8.2 가 더한 두 분기 (B6·B7)

B6 (`activation.Verified()`, 108:2) 과 B7 (`protection.Generation() < 하한`, 109:3)
이 이 로트가 더한 둘이고, **함수 가운데**에 들어갔으므로 옛 B6~B14 가 B8~B16 으로
밀렸다. 밀린 아홉은 조건을 소스와 하나씩 대조해 확인했다(옛 89:2 → 새 116:2, …,
옛 137:2 → 새 164:2). 레이블만 보고 옮기면 이 자리에서 정확히 틀린다.

**왜 이 함수인가.** 보호 세대는 주문을 내려는 순간에만 존재하는 사실이고, 이
함수가 그 사실을 들고 있으면서 주문을 거절할 수 있는 유일한 자리다. 앞 판본은
같은 결속을 `buildProductionStrategyMarketWorker` 에 두었는데 그 서술자는 화면과
승격만 움직이고 이 경로는 읽지 않는다 — 8.5 적대 리뷰가 그것을 값으로 보였다.

**반증 둘.** 가드를 `if false && …` 로 무력화하면 "하한보다 낮으면 거절" 행이
빨개지고, `if true || …` 로 항상 거절하게 만들면 "하한과 같으면 나간다" 행이
빨개진다. 서로 다른 행이 빨개지므로 이 시험은 양방향으로 판별한다 — 한쪽만
빨개지는 시험은 "항상 거절" 판본도 통과시킨다.
