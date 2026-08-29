# Function Logic Map: `Gateway.refuse`

- Source: `internal/execgw/gateway.go`
- Qualified function: `Gateway.refuse`
- AST evidence: `ast.json` (base revision; removed by the strategy atomic-refusal refactor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The base function accepted one already journalled attempt, an outcome shell and a typed Gateway refusal.
It was valid only for attempts proven never to have reached broker transport.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Attempt.Settle(NOT_DISPATCHED)` fails | settlement transaction may roll back | wrapped close error | existing refusal settlement failure suite |
| success | settlement commits | attempt becomes NOT_DISPATCHED | original typed refusal | ordinary Gateway refusal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Attempt.Settle` | close an ordinary provably-not-sent attempt | one transaction; no broker call/retry | base AST and ordinary refusal regressions |

## State mutations and fallbacks

- The base helper mutated only the core attempt. It did not own strategy lease or six risk holds.
- It was removed/replaced because a prepared strategy refusal must close core attempt, CLAIMED lease,
  aggregate and five bucket holds atomically; ordinary refusal behavior remains covered separately.

## Safety conclusion

- Safe edit boundary: never reuse the base core-only helper for strategy first-leg refusal.
- High-risk impact: yes — split refusal settlement could leave a reusable lease or held capacity.
