# Function Logic Map: `DefaultLadderPolicy`

- Source: `internal/exitpolicy/ladder.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| none | fixed immutable default table | approved exit-policy spec | no runtime fallback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unconditional construction | returns a fresh rung slice with stable id/version | value | default policy tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure literal construction | no error/retry/side effect | CodeGraph + AST |

## State mutations and fallbacks

- No external mutation; each call returns a new slice. Existing numeric values are unchanged.

## Safety conclusion

- Safe edit boundary: attach semantic version only; canonical digest is derived from the unchanged table.
- High-risk impact: yes — policy default identity, no numeric policy change.
