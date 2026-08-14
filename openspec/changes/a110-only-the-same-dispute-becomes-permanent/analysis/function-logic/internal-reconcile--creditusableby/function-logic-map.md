# Function Logic Map: `creditUsableBy`

- Source: `internal/reconcile/mismatch.go`
- AST evidence: `ast.json` — branches 2
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| credit comparison stamp | RFC3339 or unorderable | `AdjustmentApplied` | unorderable credit is not usable |
| observed comparison stamp | RFC3339 or unorderable | current Diff | unorderable observation cannot spend/refute credit |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | credit stamp cannot be parsed | none | false | a083 timestamp tests |
| B2 | observed stamp cannot be parsed | none | false | a083 timestamp tests |
| Return | both parse | none | true only when observation is strictly later | a083 equal/earlier/later tests; F12 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `creditStampAt` | parse the adjustment credit's comparison stamp | malformed/missing returns false | AST B1 |
| `time.Parse` | parse current observation ordering stamp | malformed returns false | AST B2 |

## State mutations and fallbacks

- Pure ordering predicate; equal time is deliberately not a re-read.
- It is shared by normal release accounting and pre-authority refutation, keeping both paths on one rule.

## Safety conclusion

- Safe edit boundary: strictly-later comparison only; never widen malformed/equal/earlier evidence.
- High-risk impact: yes. A stale credit can release an ordinary exposure block without convergence.
