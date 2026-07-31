# Function Logic Map: `TestEvidenceQueryUsesOnlyRowsRemainingAfterFilter`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one linked live row plus 257 hidden closed rows | local live filter | broker/journal fixtures | visible link survives and counts do not change |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | hidden-row setup, link assertion, count assertion | test only | fail test | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.orders` | prove scopes follow filter | explicit test failure | AST + named test |

## State mutations and fallbacks

- Hidden rows exceed the journal's maximum scope count.

## Safety conclusion

- Safe edit boundary: presentation/query-bound regression test.
- High-risk impact: no.
