# Function Logic Map: `TestAtomicMarketReleaseDoesNotCrossIntoPeerMarket`

- Source: `internal/journal/reconcile_states_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| two market scopes | exact batch preflight and rollback | release contract | assertions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | seed exact rows, request missing peer scope, assert error and both active | test DB writes | assertions | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| batch reconcile API | verify no cross-market/partial release | no live effects | AST |

## State mutations and fallbacks

- Test-only durable rows.

## Safety conclusion

- Safe edit boundary: atomic cross-market coverage.
- High-risk impact: no production mutation.
