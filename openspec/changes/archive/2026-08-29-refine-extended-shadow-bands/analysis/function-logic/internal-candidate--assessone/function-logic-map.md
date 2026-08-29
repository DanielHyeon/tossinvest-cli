# Function Logic Map: `assessOne`

- Source: `internal/candidate/watch.go`
- Function: `internal/candidate/watch.go:assessOne`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: the two statements that built the shadow bands left for shadowBandsFor. The Verdict assembly, the Chase and the acceleration loop are untouched and the branch structure is identical to base (B1-B3 unchanged).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| summary Summary | one live candidate with its stored firsts | Assess's Summaries query | a summary of another market is filtered by Assess, not here |
| rows []Observation | this symbol's readings inside DefaultAssessHistory | Assess's ObservationsSince | empty rows make every measurement unmeasured with a named reason |
| at time.Time | the assessment instant | AssessOptions.At | zero propagates to the age checks as NO_INSTANT |
| th VetoThresholds | the veto's numbers; two of three are absent (D18) | AssessOptions.Thresholds | absent thresholds produce THRESHOLD_ABSENT, never a pass |
| window time.Duration | the acceleration window, already defaulted by Assess | AssessOptions.AccelerationWindow | zero cannot arrive: Assess replaces it before calling |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over rows, grouping observations by source | fills bySource; appends to order on first sight of a source id | no return | existing acceleration coverage through Assess |
| B2 | first sight of a source id inside B1 | appends the id to order so the sort is over distinct ids | no return | same |
| B3 | range over the sorted source ids | appends one Acceleration per source | no return | same |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| MeasureFirstSighting / MeasureExpansion / MeasureRangePosition | the three inputs the verdict is computed from | total; each returns an unmeasured value carrying its reason | ast.json calls |
| AssessChase | the three-state verdict | total; absent thresholds give THRESHOLD_ABSENT | ast.json calls; veto.go:AssessChase |
| shadowBandsFor | NEW - the two shadow records, replacing two direct Measure*Band calls | total | ast.json calls; watch.go:shadowBandsFor |
| NewSourceSeries | one series per source; the error is deliberately dropped because the refusal is also in the value | answers MIXED_SERIES in the value | unchanged from base |
| Accelerate / sort.Slice / append | the acceleration pass and its deterministic order | total | unchanged from base |

## State mutations and fallbacks

- Local Verdict `v` only. No pointer parameter, no package state, no I/O.
- The two band fields are now written by one assignment from shadowBandsFor's results and are never read back here. Reading either one fails TestNoFunctionThatProducesAVerdictCanSeeAShadowBand.

## Safety conclusion

- Safe edit boundary: Reading v.SeenLateBand or v.ExtendedBand here, or naming a band type, constructor or the Crossed/Crossings vocabulary. The 2026-07-28 review made a band decide a veto from inside this function in one line and nothing failed; the check that missed it now covers Verdict results and the read vocabulary.
- High-risk impact: no - discovery assessment only. Reads no order, stop, sizing, ledger or auth path and writes nothing.
