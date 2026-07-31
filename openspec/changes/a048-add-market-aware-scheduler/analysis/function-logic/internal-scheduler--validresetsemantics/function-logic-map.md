# Function Logic Map: `validResetSemantics`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| budget reset raw/kind/instant | exact output of official reset derivation for `ObservedAt` | `official.ParseRateBudgetReset` | any mismatch is false/fail-closed |
| observed-at | non-zero response completion instant | official `RateBudget` | invalid or implausible derived reset is false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | official derivation returns epoch/delta and both kind and instant equal the supplied budget | none | true | parser-equivalence table |
| B2 | raw is absent/unparsed, kind mismatches threshold, instant differs, arithmetic would wrap, or reset is implausible | none | false | overflow, threshold, plausibility, raw-kind mismatch tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `official.ParseRateBudgetReset` | reuse the only authoritative reset parser | pure and read-only; invalid evidence returns `ResetUnparsed`/zero | CodeGraph + AST + official/scheduler tests |

## State mutations and fallbacks

- No state mutation. The scheduler compares against official derivation rather than duplicating constants or duration arithmetic.

## Safety conclusion

- Safe edit boundary: reset provenance admission only.
- High-risk impact: yes, because false positives can reset issuance accounting.
