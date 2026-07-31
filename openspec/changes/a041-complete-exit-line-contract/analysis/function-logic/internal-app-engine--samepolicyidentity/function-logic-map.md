# Function Logic Map: `samePolicyIdentity`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| two policy identities | complete tuples | evaluator/runtime resolver | false on any semantic difference |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all normalized fields equal | none | true | policy identity tests |
| B2 | any field differs | none | false | reinterpretation refusal test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | normalize stored typed values | no error path | AST |

## State mutations and fallbacks

- Pure comparison; no fallback.

## Safety conclusion

- Safe edit boundary: exact tuple equality.
- High-risk impact: yes — fail-closed gate.
