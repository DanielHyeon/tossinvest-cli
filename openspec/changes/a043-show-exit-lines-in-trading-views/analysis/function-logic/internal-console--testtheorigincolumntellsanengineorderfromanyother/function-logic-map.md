# Function Logic Map: `TestTheOriginColumnTellsAnEngineOrderFromAnyOther`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one confirmed engine order and one unmatched order | readable journal | base fixture | exact labels required |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | row and page-label assertions | test only | fail test | named base test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.orders` | verify origin distinction | explicit test failure | base AST + named test |

## State mutations and fallbacks

- Base test retained; new confirmed-state tests narrow the contract.

## Safety conclusion

- Safe edit boundary: regression evidence.
- High-risk impact: no.
