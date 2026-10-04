# Branch Test Map: `failedFields (새 함수)`

- Source SHA-256: `57123f814187d9201ec6999d6375adc649b579c89beb6c4731f36cc51243b005`; AST branch locations are authoritative.
- Revision: **modified (a112 8.8.4-B, 2026-10-01).** 새 함수(a112 8.8.4 로트 B) — 편집 전 없음. 복합 결속의 진단: 실패한 이름만 순서대로 모은다.
- 편집 전 번들: `analysis/measurements/lot-8.8.4-B/pre-edit/internal-strategyrouter--failedfields/`. 변이 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 635:2 — 검사 순회 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel` | yes — 새 함수; 변이 B2(첫 실패만) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
| B2 | if at 636:3 — 실패한 검사의 이름을 모음 | `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(세 필드 동시 · 두 항목 동시) | yes — 변이 B3(아무것도 안 모음 → 결속 통과) CAUGHT(`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`) | yes |
