# Function Logic Map: `at`

- Source: `internal/risk/chain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

New leaf helper added by this change. Copies a step name onto a refusal and returns an ALLOW untouched.

It is a new function with no previous revision, and it is in the required set because the insertion that created it lands at `Evaluate`'s end boundary, which the checker attributes to `Evaluate`. It is placed directly beneath its only caller for that reason and for the ordinary one — a three-line helper reads best next to the code that uses it.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `step` | an `entryChain` rung name, `StepPreflight`, or `StepReduction` | `Evaluate`'s call site | n/a — the caller supplies a literal or `s.name` |
| `d` | a Decision the chain just produced | the rung, preflight, or the reduction branch | returned unchanged when Allowed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`if` @ internal/risk/chain.go:145) | `!d.Allowed` | sets `d.Step` on the local copy | the Decision, stamped or not | `TestAnAllowedVerdictNamesNoStep` (the false arm), `TestEveryRungReportsItsOwnName` (the true arm) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | the function calls nothing | n/a | AST: no call expressions |

## State mutations and fallbacks

- `d` is a value parameter, so the assignment mutates a local copy and the caller's Decision is untouched. The function returns the copy.
- No side effects, no I/O, no fallback.

## Safety conclusion

- Safe edit boundary: total. The function cannot fail and cannot affect `Allowed`, `Reason` or `Detail` — it only ever writes `Step`.
- High-risk impact: **yes by path** (it is in the Guardian's chain package), **no by effect**: an ALLOW passes through byte-identical, which is what keeps §0.9 out of scope.
