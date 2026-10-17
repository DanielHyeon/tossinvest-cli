# Function Logic Map: `TestMarketScopedReconcilesEnterReadAndReleaseIndependently`

- Source: `internal/journal/reconcile_states_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| KR/US same account+symbol | independent exact scopes | v24 reconcile contract | assertions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | enter KR/US, re-enter KR, release KR, inspect US, refuse second KR release | test DB writes | assertions | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| reconcile APIs | verify exact scope lifecycle | no live effects | AST |

## State mutations and fallbacks

- Test-only durable rows.

## Safety conclusion

- Safe edit boundary: exact market behavior coverage.
- High-risk impact: no production mutation.
