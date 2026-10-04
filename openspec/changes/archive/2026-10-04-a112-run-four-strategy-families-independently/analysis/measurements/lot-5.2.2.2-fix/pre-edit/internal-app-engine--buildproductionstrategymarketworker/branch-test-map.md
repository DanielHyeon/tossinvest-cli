# Branch Test Map: `buildProductionStrategyMarketWorker`

- Source SHA-256: `9e24e93028b2728071d71d1d6ccea2c2a83fe768f6efe2dc09a57906c435a373`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 6 분기 → 8: 편집 전 B2(준비 + handoff 승인)에서 handoff 조건을 떼어 범위 순회(B3)와 범위별 승인 · 유효(B4)로 옮기고, 편집 전 B3(봉인 깨짐)을 B4 에 합쳤다. 편집 전 B4 · B5(보호 · 진입 관문 관측 실패 → dormant)는 B5 · B6(그 범위만 건너뜀)이 되고, 승격 근거가 된 범위가 없으면 B7 이 dormant. 편집 전 B6 은 B8(불변).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--buildproductionstrategymarketworker/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 438:2 — 배선 미완/nil → dormant(편집 전 B1 불변) | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` | no — 편집 없음 | yes |
| B2 | if at 451:2 — 시장 권한(일정 · 후보 · 경로 · 환율 · 위험 · 계좌) 준비 미완 → dormant. **편집: handoff 조건을 뺐다**(B3~B4 로) | `TestARefusedHandoffLeavesTheWorkerDormant` | no — 준비 조건 불변 | yes |
| B3 | range at 460:2 — **(새)** 주문 경로와 같은 handoff 목록(`dispatchHandoffs`) 순회 — 활성화 없는 시장은 시장 단위 하나(오늘), 서명 활성화 시장은 범위마다 하나 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | yes — 편집 전 활성화 두 범위 시장 dormant(FAIL 관측), 변이 X16 CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B4 | if at 462:3 — **(새, 편집 전 B2 의 handoff 절반 + 편집 전 B3)** 그 handoff 가 거절했거나 봉인 깨진 제안 → 그 범위는 승격 근거가 못 됨(continue) | `TestARefusedHandoffLeavesTheWorkerDormant`(시장 단위 상한 거절 → 모든 범위 건너뜀 → B7 dormant) | no — 동작 보존(거절 handoff 는 편집 전에도 dormant) | yes |
| B5 | if at 468:3 — 보호 관측 실패 → **그 범위만** 건너뜀(편집 전 B4 는 시장 dormant — 범위 하나면 B7 로 같은 결과) | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` | no — 범위 하나에서는 동작 동일 | yes |
| B6 | if at 471:3 — 진입 관문 관측 실패 → **그 범위만** 건너뜀(J3 — 한 범위의 거절이 다른 범위의 승격을 굶기지 않음) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`(005930 차단 → 000660 으로 승격, 둘 다 차단 → dormant) | yes — 변이 X15(`return dormant`) · X17(관문 무시) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B7 | if at 477:2 — **(새)** 승격 근거가 된 범위 없음 → dormant | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`(모든 범위 차단) · `TestARefusedHandoffLeavesTheWorkerDormant` | yes — 변이 X17 CAUGHT | yes |
| B8 | if at 498:2 — digest / revision / 만료 → dormant(편집 전 B6 불변) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
