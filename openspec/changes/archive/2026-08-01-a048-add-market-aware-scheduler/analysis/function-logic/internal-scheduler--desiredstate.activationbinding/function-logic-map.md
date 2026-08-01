# Function Logic Map: `DesiredState.ActivationBinding`

- Source: `internal/scheduler/desired.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| desired | validated desired state including monotonic revision | persisted operator state | copied exactly; no defaults |
| build digest | nonempty current build at Restore gate | runtime binding | empty is refused before this call |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| straight-line | all approval/scope/version/revision fields copied | none | immutable manifest comparison value | exact restore/activation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | value construction only | no error/retry | Go AST |

## State mutations and fallbacks

- Pure value projection. `DesiredRevision` prevents an activation minted for one persisted revision from authorizing another.

## Safety conclusion

- Safe edit boundary: activation verification input only.
- High-risk impact: yes, because omitted authority fields would enable replay.
