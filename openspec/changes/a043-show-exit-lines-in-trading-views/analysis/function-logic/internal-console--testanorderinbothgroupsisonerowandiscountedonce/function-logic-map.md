# Function Logic Map: `TestAnOrderInBothGroupsIsOneRowAndIsCountedOnce`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one PARTIAL_FILLED record returned by OPEN and CLOSED | byte-identical scoped identity | broker API overlap contract | OPEN copy survives once and remains live |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | total is not one | test failure only | report duplicated row | this test |
| B2 | live count is not one | test failure only | report exposure miscount | this test |
| B3 | closed count is not zero | test failure only | report duplicate finished count | this test |
| B4 | surviving row is not live | test failure only | report wrong copy won | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.orders` | execute the production overlap projection | no retry; in-memory fixture | current AST |

## State mutations and fallbacks

- Test-only broker seam and console view; no production mutation.

## Safety conclusion

- Safe edit boundary: duplicate projection regression assertion.
- High-risk impact: no.
