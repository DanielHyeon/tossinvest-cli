# Function Logic Map: `TestInvalidEvidenceIdentityKeepsOriginUnknown`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| readable journal plus malformed market/time rows | invalid composite identity | broker fixture | origin remains unknown |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | journal and per-row assertions | test only | fail test | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.orders` | verify fail-closed origin | explicit test failure | AST + named test |

## State mutations and fallbacks

- Test-only assertion; no mutation.

## Safety conclusion

- Safe edit boundary: presentation regression test.
- High-risk impact: no.
