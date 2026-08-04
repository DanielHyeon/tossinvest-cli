# Function Logic Map: `TestReconcileMarketScopeValidation`

- Source: `internal/journal/reconcile_states_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| invalid requests | unknown market or market without symbol | v24 API contract | assert invalid request |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | iterate invalid entry scopes and invalid release scope | test journal only | assertions | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| reconcile APIs | exercise validation | no live effects | AST |

## State mutations and fallbacks

- Test-only writes refused.

## Safety conclusion

- Safe edit boundary: validation coverage.
- High-risk impact: no production mutation.
