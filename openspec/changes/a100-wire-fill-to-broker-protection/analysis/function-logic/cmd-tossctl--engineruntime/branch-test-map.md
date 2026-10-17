# Branch Test Map: `engineRuntime`

Source: `cmd/tossctl/engine.go` (649-739), AST branches 6.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil-hint detector path remains constructor-safe | `TestEngineRuntimeB1IsStructurallyUnreachableWithTheHardcodedNilHintPath` | n/a: pre-existing branch | yes |
| B2 | missing reconcile constructor fails closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | n/a: pre-existing branch | yes |
| B3 | exit observer constructor failure fails closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | n/a: pre-existing branch | yes |
| B4 | recovery constructor failure fails closed | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | n/a: pre-existing branch | yes |
| B5 | strategy-entry constructor failure | no current named test; A100 RED required before edit | no | no |
| B6 | alert-deliverer constructor failure | no current named test; A100 RED required before edit | no | no |

Before task 3.9 changes this function, add RED tests for B5/B6 and for the
new worker’s verified-gate, recovery ordering, and cancellation contracts.

> **a100 R0 재동결(2026-10-04) — 좌표 이동.** 옛 · 새 AST 의 분기 (id · 종류) 목록이 같아 ast.json 을 현재 소스 추출로 바꾸고 `(start-end)` 범위 · 파일 SHA-256 만 옮겼다(`analysis/harness/r0_refresh_bundle.py shift`). 산문 · 시험 인용은 손대지 않았다. 옛 ast 는 returns/calls 를 싣지 않던 추출기 판본이라 그 둘은 대조 대상이 아니다.
