# Function Logic Map: `Measure`

- Source: `internal/performance/model.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| trade | validated BUY/SELL trade with positive entry/quantity and non-negative cost | persisted journal-derived `Trade` | unsupported side/invalid decimals stay unmeasured or are rejected by caller validation |
| observations | caller-owned rows for the exact position and holding interval | caller; no poll capability | mismatched/out-of-window rows are ignored |
| side-adjusted returns | BUY=(price-entry)/entry, SELL=(entry-price)/entry | `Trade.Side` | no long-only fallback for SELL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate caller observations | none | each row evaluated | model tests |
| B2 | exact position and holding interval match | appends local filtered copy | foreign/out-of-window row excluded | `TestMeasureNeverUsesAnotherPositionOrInventsLineage` |
| B3 | equal observation instants during stable sort | none | ID tie-break | tolerance test |
| B4 | iterate filtered observations | appends pure markout input | continue | tolerance test |
| B5 | iterate fixed 5/15/30 measurements | appends result metric | continue | markout tests |
| B6 | measurement present in +60s window | derives metric | complete or unmeasured if side unsupported | markout tests |
| B7 | BUY/SELL side supported | sets side-adjusted gross | continue | side test |
| B8 | normalized cost available | subtracts cost from side-adjusted gross | continue | side test |
| B9 | selected instant available | searches exact supplied row | continue | provenance tests |
| B10 | iterate filtered rows for selected identity | none | continue | provenance tests |
| B11 | instant and price match selected measurement | sets observation/source/version | break | provenance tests |
| B12 | side-aware decision/entry slippage available | sets slippage | continue | side test |
| B13 | lifetime extrema available | sets side-aware MFE/MAE | snapshot | side/excursion tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `markout.Measure` | selects first observation in 5/15/30m +60s windows | pure; no retry/poll | current HEAD + AST |
| decimal helpers | exact rational side/cost arithmetic | invalid data fails closed to unmeasured | current HEAD + tests |

## State mutations and fallbacks

- Pure derivation only; no storage, broker, configuration, polling, or LIVE side effect.
- Gross markout, cost-adjusted markout, slippage, MFE and MAE all use the same explicit trade side.

## Safety conclusion

- Safe edit boundary: derived analytics math and provenance only.
- High-risk impact: no trading authority; financial reporting correctness requires BUY/SELL symmetry tests.
