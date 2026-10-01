# Branch Test Map: `BudgetCoordinator.TryAcquire`

- Source SHA-256: `36879e8856901f349ea2a75856ddc3124520f5bd67b674359dcf854e6e8b6bf8`; AST branch locations are authoritative.
- Revision: **modified (a112 7.1, 2026-10-01).** a112 7.1(Manager 판정 Q2): 본문(편집 전 17 분기)을 `tryAcquire(key, class, now, scope)` 로 옮기고 공개 메서드는 범위 0 으로 부르는 위임이 됐다(분기 없음). `tryAcquire` 의 17 분기는 편집 전과 같은 조건 · 같은 순서이고 바뀐 것은 발급 기록에 범위를 싣는 한 줄(`budgetCommitment{…, scope: scope}`)뿐 — 범위 없는 경로의 판정 불변(편집 전 번들 `lot-7.1/pre-edit/`).
- 편집 전 번들: `analysis/measurements/lot-7.1/pre-edit/internal-scheduler--budgetcoordinator.tryacquire/`. 변이 원장 `analysis/measurements/lot-7.1/mutation-7.1.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — `tryAcquire(key, class, now, [32]byte{})` 위임 — 범위 없는 발급의 판정 불변 | `budget_test.go` 기존 시험 전부(scheduler 패키지 PASS — 대조군 pass 146) · `internal/scheduler/a112_strategy_scope_test.go` `TestEveryScopeSharesOnePhysicalCommitmentSet`(범위 없는 API 가 범위 발급과 같은 집합을 봄) | no — 시험 코드 | yes |
