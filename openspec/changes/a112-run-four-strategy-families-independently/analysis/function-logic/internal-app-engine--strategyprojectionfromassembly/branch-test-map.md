# Branch Test Map: `strategyProjectionFromAssembly`

- Source SHA-256: `95474831b04d24c21d90d72aac7349fe0682d2cbee9beb02c3be6307ac9dc510`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 8 분기 → 10: 편집 전 B7(시장 단위 handoff 거절 또는 봉인 깨짐 → EvidenceStale)을 handoff 목록 순회(B7) · 첫 승인 유효 범위 선택(B8) · 없음(B9)으로 나눴다. 편집 전 B8 → B10.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategyprojectionfromassembly/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 101:2 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 108:3 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | switch at 110:4 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | case at 111:4 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | case at 113:4 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | case at 115:4 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | range at 127:3 — **(새)** 주문 경로와 같은 handoff 목록 순회 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`(projection 단언) | yes — 변이 X19(시장 단위 handoff) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B8 | if at 128:4 — **(새)** 조정자 순서의 첫 승인 · 유효 범위를 보임 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | yes — X19 | yes |
| B9 | if at 133:3 — 승인 범위 없음 → EvidenceStale(편집 전 B7 의 결과) | 진입 0 — 이 로트의 시험 없음(편집 전에도 진입 0) | no | 진입 0 |
| B10 | if at 141:3 — 레인 증거 다이제스트 부재(편집 전 B8) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
