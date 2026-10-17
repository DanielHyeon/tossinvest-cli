# Branch Test Map: `BudgetCoordinator.Complete`

- Source SHA-256: `36879e8856901f349ea2a75856ddc3124520f5bd67b674359dcf854e6e8b6bf8`; AST branch locations are authoritative.
- Revision: **modified (a112 7.1, 2026-10-01).** a112 7.1(Manager 판정 Q2): 본문(편집 전 7 분기)을 `complete(key, token, scope)` 로 옮기고 공개 메서드는 범위 0 으로 부르는 위임이 됐다(분기 없음). `complete` 는 편집 전 기록 대조 조건에 `record.scope != scope` 를 더했다 — 범위 있는 commitment 는 범위 없는 완료로 끝나지 않는다(교차 replay 금지).
- 편집 전 번들: `analysis/measurements/lot-7.1/pre-edit/internal-scheduler--budgetcoordinator.complete/`. 변이 원장 `analysis/measurements/lot-7.1/mutation-7.1.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — `complete(key, token, [32]byte{})` 위임 — 범위 없는 완료 | `budget_test.go` 기존 완료 시험 · `internal/scheduler/a112_strategy_scope_test.go` `TestScopedAndUnscopedCapabilitiesCannotCompleteEachOther`(범위 있는 토큰의 범위 없는 완료 거절 — 변이 T03 CAUGHT) | no — 시험 코드 | yes |
