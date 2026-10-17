# Function Logic Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go`
- Current-base source SHA-256: `9610abb34ee350fa3f19ad06cb41d279ebe3c19df44f0a60336dd83276c6c289`
- Signature: `strategyDispatchCycle.dispatch(params=2, results=2)`
- Source range: `75:1`–`211:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Measurement regime (8.7.2 편집 뒤): 몸통 진입 count. engine tagged suite 바이너리(`-coverpkg=./internal/app/engine,./internal/strategyrouter`, -trimpath 없이)를 `systemd-run … MemoryMax=16G` 안에서 실행, 스위트 PASS; 전체 시험 509 개를 하나씩 돈 per-test 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh` · `a872_attribute.py`). 모든 행에서 시험별 합 == 스위트(ATTRIBUTION MISMATCH 0).

Exact AST return positions: 79:3, 82:3, 88:3, 94:3, 98:3, 116:4, 137:3, 141:3, 145:3, 149:3, 153:3, 157:3, 164:3, 170:3, 183:3, 189:3, 193:3, 195:2, 206:5, 209:4.

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

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `delivered.Result` | 76:12 |
| `validateStrategyFirstLegResult` | 77:23 |
| `errors.New` | 79:28 |
| `errors.New` | 82:28 |
| `StrategyMarket` | 84:12 |
| `cycle.schedule.forMarket` | 85:18 |
| `cycle.fx.forMarket` | 85:52 |
| `errors.New` | 88:28 |
| `strategyFirstLegPlaceIntent` | 93:15 |
| `cycle.gateway.ObserveStrategyProtection` | 96:21 |
| `strings.ToLower` | 96:66 |
| `string` | 96:82 |
| `familyActivation` | 114:19 |
| `cycle.proposals.forMarket` | 114:19 |
| `activation.Verified` | 114:73 |
| `protection.Generation` | 115:6 |
| `activation.ProtectionReadyMinGeneration` | 115:32 |
| `fmt.Errorf` | 116:29 |
| `protection.Generation` | 118:5 |
| `activation.ProtectionReadyMinGeneration` | 118:30 |
| `familyActivation` | 135:12 |
| `cycle.proposals.forMarket` | 135:12 |
| `family.Verified` | 136:5 |
| `errors.New` | 137:28 |
| `family.LeaseCeiling` | 139:23 |
| `cycle.clockNow` | 139:43 |
| `cycle.gateway.ObserveStrategyEntryGate` | 143:25 |
| `strings.ToLower` | 143:69 |
| `string` | 143:85 |
| `cycle.dispatchOwner` | 147:16 |
| `cycle.firstLeg.admit` | 151:14 |
| `fmt.Errorf` | 153:28 |
| `cycle.journal.LookupDecision` | 155:19 |
| `errors.New` | 157:28 |
| `uint64` | 161:24 |
| `bundle.Generation` | 162:20 |
| `cycle.risk.forMarket` | 162:20 |
| `errors.New` | 164:28 |
| `schedule.restore.Activation.Generation` | 166:26 |
| `schedule.restore.Activation.ExpiresAt` | 167:25 |
| `activationExpiresAt.IsZero` | 169:34 |
| `now.IsZero` | 169:66 |
| `now.Before` | 169:83 |
| `errors.New` | 170:28 |
| `journal.StrategyDispatchMarket` | 172:63 |
| `protection.Generation` | 175:25 |
| `strconv.FormatUint` | 175:68 |
| `protection.Generation` | 175:87 |
| `protection.Digest` | 175:135 |
| `reconciliation.Generation` | 176:29 |
| `reconciliation.Digest` | 176:80 |
| `strategyRuntimeBuildDigest` | 177:94 |
| `min` | 178:9 |
| `activationExpiresAt.Sub` | 178:27 |
| `cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease` | 179:16 |
| `cycle.journal.ClaimStrategyDispatchLease` | 185:18 |
| `strategyFirstLegPlaceIntent` | 191:17 |
| `cycle.gateway.PlaceClaimedStrategy` | 195:9 |
| `cycle.revalidateSchedule` | 205:14 |
| `family.LeaseCeiling` | 208:14 |
| `cycle.clockNow` | 208:34 |

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

## 2026-09-27 — 태스크 8.7.2 편집 전 측정과 계획

`ast.json` 의 SHA-256 `12f578d4…` 는 편집 전 이 워크트리 파일과 같다(재확인). 아래 표는 **몸통 진입** regime 이다
(분기 좌표 뒤 첫 커버리지 블록). 위 8.8.2 절의 수(예: B2 19x)는 조건 평가를 센 다른 regime 이므로 섞지 않는다.
하네스: `analysis/harness/a872_pertest_cover.sh`(엔진 전체 시험 494 개를 하나씩) + `a872_attribute.py`. 모든 행에서
시험별 합 == 스위트(ATTRIBUTION MISMATCH 0).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 72:2 | arm entered 1x (engine tagged suite, pre-edit); `TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall` |
| B2 | if | 75:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B3 | if | 80:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B4 | if | 87:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B5 | if | 91:2 | arm entered 4x (engine tagged suite, pre-edit); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` |
| B6 | if | 108:2 | arm entered 2x (engine tagged suite, pre-edit); `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` |
| B7 | if | 109:3 | arm entered 1x (engine tagged suite, pre-edit); `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` |
| B8 | if | 116:2 | arm entered 4x (engine tagged suite, pre-edit); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral`, `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` |
| B9 | if | 120:2 | arm entered 2x (engine tagged suite, pre-edit); `TestNoJournalOrGatewayFaultInTheDispatchCycleIsClassifiedCentral` |
| B10 | if | 124:2 | arm entered 1x (engine tagged suite, pre-edit); `TestTheSameEnvelopeCannotPlaceASecondOrder` |
| B11 | if | 128:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B12 | if | 135:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B13 | if | 141:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B14 | if | 154:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B15 | if | 160:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |
| B16 | if | 164:2 | arm not entered (engine tagged suite, pre-edit); no per-test profile entered it |

**편집 계획 (Manager 승인 2026-09-27, 조건 ①~⑤).** lease TTL 의 상한 `30*time.Second`(150:2)를 가족 활성화가 깎은
상한으로 바꾼다. 계산 규칙(검증 안 된 활성화는 상한을 그대로 돌려줌 · 남은 수명 0 이하는 만료 오류 · 결과는 입력 상한을
넘지 않음 = **min 결합만**)은 strategyrouter 의 메서드 하나(`FamilyActivation.LeaseCeiling`)에 두고, 이 함수는 파도 시각과
30초 상한을 넘기고 결과를 `min(…, 스케줄 활성화 남은 수명)` 에 쓰는 배관만 한다.

부르는 자리는 보호 하한 결속(108:2) **바로 뒤, q_final admission(123:2) 앞**이다. 이 함수 머리의 "admission 이 커밋하기 전에
실패할 수 있는 읽기 전용 경계를 전부 끝낸다"(84행 주석)를 따른다. 새 분기는 하나(`err != nil` → 오류 반환)다. 생산에서는
넘기는 시각이 활성화를 검증한 바로 그 파도 시각이라 이 갈래가 닿지 않는다(시험 seam 으로만 닿는다) — 실질적인 보호는 TTL
결합이다: lease 가 가족 활성화 만료를 넘어 살지 못한다.

## 2026-09-27 — 8.7.2 두 번째 편집 전 (Codex P1 수정, Manager 승인)

첫 편집(미커밋, 워크트리 SHA-256 `fc5f6721…`) 뒤의 구문 트리에는 분기가 열일곱이었다: B1 72:2 · B2 75:2 · B3 80:2 · B4 87:2 · B5 91:2 ·
B6 108:2 · B7 109:3 · **B8 125:2(새 것: `LeaseCeiling` 오류 → admission 앞 거절)** · B9 129:2 · B10 133:2 · B11 137:2 · B12 141:2 ·
B13 148:2 · B14 154:2 · B15 167:2 · B16 173:2 · B17 177:2. 첫 편집 전의 B8~B16 이 B9~B17 로 밀렸다 — 조건을 소스와 하나씩
대조했다(옛 116:2 `ObserveStrategyEntryGate` 오류 → 새 129:2, …, 옛 164:2 → 새 177:2).

**Codex 가 연 P1.** `LeaseCeiling` 은 파도 시각 기준 **상대** TTL 이고 journal 은 그 TTL 을 **자기 현재 시각**에 더한다. 파도→발급
사이 δ 만큼 lease 행의 명목 만료가 가족 활성화 만료를 넘는다. 스케줄 활성화는 같은 상대 TTL 을 쓰지만 `FinalAuthorityCheck`
(게이트웨이가 브로커 바이트 **전** 에 부른다, `internal/execgw/gateway.go` 의 `call` 클로저)가 실시계로 다시 적재해 막는다. 가족
활성화에는 그 최종 검사가 없었다.

**두 번째 편집 계획.** (1) `strategyDispatchCycle` 에 `now func() time.Time` 필드. (2) admission 앞 `LeaseCeiling` 호출의 시각을
파도 시각에서 `cycle.now()` 로. (3) `FinalAuthorityCheck` 클로저가 스케줄 재검증에 더해 `LeaseCeiling(cycle.now(), …)` 의 만료
오류를 돌려준다 — 규칙은 여전히 strategyrouter 한 곳. (4) `now` 가 nil 인데 가족 활성화가 검증돼 있으면 거절(fail-closed);
검증 안 된 활성화(미선언·기존 경로)는 `LeaseCeiling` 이 시각을 보지 않으므로 생산 변화 0.
불변식은 이렇게 적는다: **lease 행은 명목상 만료를 최대 δ 넘을 수 있으나, SUBMITTING 은 만료 뒤 최종 검사를 통과할 수 없다.**
