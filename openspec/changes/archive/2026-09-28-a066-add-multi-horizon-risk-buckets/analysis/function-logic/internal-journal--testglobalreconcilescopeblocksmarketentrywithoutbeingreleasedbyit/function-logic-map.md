# Function Logic Map: `TestGlobalReconcileScopeBlocksMarketEntryWithoutBeingReleasedByIt`

- Source: `internal/journal/reconcile_states_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| global symbol row + KR/US requests | global blocks entry but exact release cannot clear global | v24 contract | assertions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | enter global, test both market entries/releases, inspect active row | test DB writes | assertions | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| reconcile APIs | global overlap semantics | no live effects | AST |

## State mutations and fallbacks

- Test-only durable rows.

## Safety conclusion

- Safe edit boundary: global authority coverage.
- High-risk impact: no production mutation.
