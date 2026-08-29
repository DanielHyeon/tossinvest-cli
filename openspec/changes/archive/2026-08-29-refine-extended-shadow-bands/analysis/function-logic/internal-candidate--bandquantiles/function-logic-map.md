# Function Logic Map: `bandQuantiles`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:bandQuantiles`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

New in this change. Nearest-rank positions over the measured values, so every figure printed is a value some candidate actually had.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| values []*big.Rat | any length including zero; may contain negatives | TallyBands | an empty slice returns nil, never a zero-valued median |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no values | none | nil - an unmeasured tally carries no arithmetic | TestAnUnmeasuredTallyHasNoQuantiles |
| B2 | range over BandQuantilePoints | appends one position per point, in declared order | no return | TestTheTallyCarriesTheDistributionAndNotOnlyTheCrossings |
| B3 | the computed rank is below 1 | clamped to 1 | no return | same test, the min position |
| B4 | the computed rank is past the end | clamped to n | no return | same test, the max position |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sort.Slice with Cmp | orders on the exact rationals, not on the strings | total | ast.json calls |
| formatDecimal | renders the chosen value | total; truncates towards zero at metricScale | ast.json calls; metrics.go:formatDecimal |

## State mutations and fallbacks

- Sorts the slice it is given, in place. Its only caller builds that slice and does not reuse the order, which is why the sort is here rather than on a copy.

## Safety conclusion

- Safe edit boundary: Interpolating between two neighbours. This is the surface a threshold gets chosen from and an interpolated figure is a number no candidate had.
- High-risk impact: no.
