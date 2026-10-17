# Function Logic Map: `BudgetCoordinator.Complete`

- Source: `internal/scheduler/budget.go`
- Source SHA-256: `36879e8856901f349ea2a75856ddc3124520f5bd67b674359dcf854e6e8b6bf8`
- Signature: `BudgetCoordinator.Complete(params=2, results=1)`
- Source range: `412:1`–`414:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.1).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 있는 capability 는 같은 범위로만 완료된다(양방향 교차 replay 금지 — 변이 T01 · T03 · T04 · T05 CAUGHT).

## Branches and early returns

- Exact AST return nodes: `413:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.complete` | 413:9 |

## State mutations and fallbacks

- `complete` 가 완료 순번 · 기록을 쓴다(뮤텍스 아래).

## Safety conclusion

- High-risk 아님(생산 호출자 0). 범위 없는 토큰 · 범위 없는 기록의 완료 판정 불변.
