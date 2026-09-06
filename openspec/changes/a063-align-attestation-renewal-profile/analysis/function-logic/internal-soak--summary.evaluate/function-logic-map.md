# Function Logic Map: `Summary.Evaluate`

- Source: `internal/soak/attest.go`; E-based current S snapshot, SHA-256 `d693c9dbd9583789b6372aa63a126e7e4817df874c74ff76c597633eb60d29d5` (retrospective, not pre-edit evidence).
- AST evidence: `ast.json`.

## Inputs and invariants

`s` is the summarized soak record; `now` and `c` are forwarded to `evaluateIssues` (which applies criteria defaults). A qualifying summary must return `(true, nil)`, not an empty reasons slice; a nonqualifying summary returns every issue message rather than only the first.

## Branches and early returns

| B-id | Source condition and effect | Focused test evidence |
|---|---|---|
| B1 | `for _, issue := range issues` appends each `issue.Message`; empty issues perform zero iterations. | `TestEvaluateRefusesAnEmptyRecord` establishes a refusal with reasons, but the zero-iteration case is unmeasured by focused test. |
| B2 | `len(reasons) == 0` returns `true, nil`; otherwise the final return is `false, reasons`. | `TestEvaluatePreservesNilReasonsForAQualifyingSoak` establishes true/nil; `TestEvaluateRefusesAnEmptyRecord` establishes false/nonempty. |

## Calls and live bindings

The substantive call is `s.evaluateIssues(now, c)`; `make`, `len`, and `append` construct an in-memory slice. There is no I/O, retry, timeout, config/path binding, or live broker effect.

## State mutations and fallbacks

The only mutation is in-memory slice construction. The false/reasons return is the caller-facing refusal fallback.

## Safety conclusion

There is no durable mutation or live broker effect.
