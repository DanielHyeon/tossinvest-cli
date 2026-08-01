# Function Logic Map: `readRateBudget`

- Source: `internal/official/ratebudget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| path | raw request path | official HTTP request | normalized with `budgetKey` |
| rate headers | decimal limit/remaining and verbatim reset | Toss Open API response | missing/invalid numeric fields remain absent values; any header still makes the budget reported |
| now | response-completion instant | official client clock | normalized to UTC and used as the sole delta/plausibility anchor |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no limit, remaining, or reset header | none | unreported budget | no-header test |
| B2 | reset header absent | records `ResetAbsent` | reported budget without a derived reset | reset-only/reported-budget tests |
| B3 | reset raw parses below the exact epoch threshold and derives a plausible instant | records `ResetDelta` and derived UTC instant | reported parsed budget | delta/boundary tests |
| B4 | reset raw parses at or above the exact epoch threshold and derives a plausible instant | records `ResetEpoch` and Unix UTC instant | reported parsed budget | epoch/threshold tests |
| B5 | reset is malformed, negative, arithmetic-unsafe, or outside [-1m,+24h] around response completion | preserves raw, records `ResetUnparsed`, clears derived reset | reported but untrusted budget | implausible/overflow/boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ParseRateBudgetReset` | centralize exact raw-kind derivation, overflow checks, and plausibility bounds for official and downstream validation | pure; never retries or mutates input | CodeGraph + AST + reset parser tests |
| `headerInt` | parse optional decimal allowance headers | malformed/absent is `(0,false)` | existing header tests |
| `budgetKey` | normalize per-resource paths to endpoint budget keys | pure | budget-key tests |

## State mutations and fallbacks

- The function mutates only the returned value. Reset parsing is delegated to one exported read-only helper so downstream consumers cannot drift from official semantics.
- Invalid reset evidence stays reported with raw provenance, but carries `ResetUnparsed` and a zero derived instant.

## Safety conclusion

- Safe edit boundary: official rate-header parsing and its pure reset derivation helper.
- High-risk impact: yes, because scheduler admission trusts this provenance.
