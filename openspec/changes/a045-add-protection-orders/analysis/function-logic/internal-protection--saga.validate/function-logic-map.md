# Function Logic Map: `Saga.Validate`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Untrusted/persisted saga. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 common identity/instrument/numeric/client/time/state checks; existing code has no per-state field invariants. | No mutation. Persisted invalid state must fail before classification or update. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `validState`, string checks; pure. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Persisted invalid state must fail before classification or update.

## Safety conclusion

- Safe edit boundary: Add explicit invariants for PLANNED/REGISTERING/ACTIVE/REPLACING/TRIGGERED/CLOSED/RECONCILE/IN_DOUBT.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
