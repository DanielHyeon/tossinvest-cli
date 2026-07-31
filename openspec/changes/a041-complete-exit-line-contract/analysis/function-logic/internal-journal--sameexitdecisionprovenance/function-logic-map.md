# Function Logic Map: `sameExitDecisionProvenance`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement/proposal provenance | two complete normalized tuples | exit snapshot | false on any difference |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all fields equal | none | true | concurrent engine test |
| B2 | any field differs | none | false | journal mismatch test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | normalize typed inputs | no error | AST |

## State mutations and fallbacks

- Exact equality; no policy registry lookup or recomputation.

## Safety conclusion

- Safe edit boundary: pure comparison.
- High-risk impact: yes — arm provenance integrity.
