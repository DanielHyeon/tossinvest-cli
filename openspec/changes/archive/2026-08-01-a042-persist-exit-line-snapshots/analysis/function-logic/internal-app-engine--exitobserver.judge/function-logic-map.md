# Function Logic Map: `ExitObserver.judge`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed state and quote | valid identity/state and one policy kind | journal + a041 evaluator | refuse and hold |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | identity error, break-even error, ladder/ratchet dispatch | alert or delegate | nil/error | existing exitloop tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| policy-specific judge | evaluate one immutable snapshot | fail closed | CodeGraph + AST |

## State mutations and fallbacks

- No planned body edit; retained as impact evidence for the changed working-set classification.

## Safety conclusion

- Safe edit boundary: unchanged.
- High-risk impact: yes.
