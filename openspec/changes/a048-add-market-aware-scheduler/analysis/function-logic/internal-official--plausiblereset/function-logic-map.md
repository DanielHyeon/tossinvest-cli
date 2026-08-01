# Function Logic Map: `plausibleReset`

- Source: `internal/official/ratebudget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| reset | derived non-zero candidate instant | official parser | zero is false |
| now | non-zero response-completion instant | official client | zero is false |
| bounds | inclusive one minute behind through 24 hours ahead | official constants | outside is false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | reset or now is zero | none | false | zero-time tests |
| B2 | saturating `Sub` result lies inside inclusive bounds | none | true | parser boundary tests |
| B3 | difference lies outside bounds, including saturated extremes | none | false | implausible/extreme tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Time.Sub` | compare signed difference without negating `MinInt` or adding to extreme timestamps | saturates safely; outside small bounds is false | AST + boundary tests |

## State mutations and fallbacks

- Pure predicate. It avoids absolute-duration negation and retains inclusive boundary semantics.

## Safety conclusion

- Safe edit boundary: official reset plausibility.
- High-risk impact: yes, because implausible reset confidence can stall or overspend polling.
