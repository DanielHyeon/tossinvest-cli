# Function Logic Map: `BudgetCoordinator.TryAcquire`

- Source: `internal/scheduler/budget.go`
- Source SHA-256: `36879e8856901f349ea2a75856ddc3124520f5bd67b674359dcf854e6e8b6bf8`
- Signature: `BudgetCoordinator.TryAcquire(params=3, results=1)`
- Source range: `312:1`–`314:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.1).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 용량 판정은 범위와 무관하게 endpoint 하나의 commitment 집합에서 한다(family 가 물리 용량을 복제하지 않음 — 변이 T06 CAUGHT).

## Branches and early returns

- Exact AST return nodes: `313:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.tryAcquire` | 313:9 |

## State mutations and fallbacks

- `tryAcquire` 가 endpoint 의 commitment · 발급 기록을 쓴다(뮤텍스 아래) — 범위 digest 를 기록에 싣는다.

## Safety conclusion

- High-risk 아님(생산 호출자 0). 범위 없는 경로의 새 통과 · 새 거절 0.
