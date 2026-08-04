# Function Logic Map: `TestReEnteringAnActiveScopeKeepsTheFirstObservation`

- Source: `internal/journal/reconcile_states_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal fixture | active reconcile scope | test contract | fatal assertion |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | enter, re-enter and inspect one active row | test DB writes only | assertions | same test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| journal reconcile API | verify idempotent first observation | fatal on error | AST |

## State mutations and fallbacks

- Test-only mutation; no live side effect.

## Safety conclusion

- Safe edit boundary: unchanged existing test body; map required by insertion boundary detection.
- High-risk impact: no production impact.
