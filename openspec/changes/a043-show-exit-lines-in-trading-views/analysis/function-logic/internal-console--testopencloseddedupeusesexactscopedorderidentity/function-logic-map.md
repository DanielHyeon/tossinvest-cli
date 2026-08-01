# Function Logic Map: `TestOpenClosedDedupeUsesExactScopedOrderIdentity`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| OPEN/CLOSED identity variants | exact overlap, whitespace id, market/day reuse, equal/different malformed time | composite broker-order identity requirement | only exact valid or stable equal fallback overlaps |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate identity cases | isolated test harness per case | none | table covers all variants |
| B2 | total differs | test failure only | show retained rows | this test |
| B3 | closed count differs | test failure only | expose dedupe/count drift | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.orders` | exercise production OPEN/CLOSED merge | no retry; in-memory seam | current AST |

## State mutations and fallbacks

- Test fixtures only. Exact overlap remains one live row; scoped reuse remains two rows.

## Safety conclusion

- Safe edit boundary: OPEN/CLOSED display merge contract.
- High-risk impact: no.
