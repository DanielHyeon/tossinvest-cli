# Function Logic Map: `ExitDecisionProvenance.zero`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance tuple | all fields absent or some present | typed judgement/proposal | boolean only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | every field trims empty | none | true | legacy journal tests |
| B2 | any field present | none | false | provenance mismatch test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | normalize optional fields | no error | AST |

## State mutations and fallbacks

- Zero is accepted only for pre-a042 legacy callers.

## Safety conclusion

- Safe edit boundary: pure predicate.
- High-risk impact: yes — compatibility branch.
