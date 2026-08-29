# Function Logic Map: `TallyBands`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:TallyBands`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: the loop also collects the measured values and the tally carries their quantiles. B1-B4 and B6-B9 are base's branches unchanged; B5 is new.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| code VetoCode | a code with a shadow scale, or any other | TallyVerdicts | a code with no scale gives an empty Crossed map and counts every record as unmeasured |
| in []ShadowBand | any slice, including zero values | TallyVerdicts | an unassigned ShadowBand is unmeasured by construction and lands under NOT_EVALUATED |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over BandsFor(code) seeding the Crossed map | Crossed gets exactly the code's bands, at zero | no return | TestTheTallyHasOneColumnPerBandAndCountsThemRight |
| B2 | range over the input records | the whole tally | no return | TestTheBandTallyAccountsForEveryCandidate |
| B3 | the record is unmeasured or belongs to another code | NotMeasured[why]++ - no value, no crossings | continue | TestABandOfAnotherCodeIsNotCountedAsMeasured, TestAnUnmeasuredTallyHasNoQuantiles |
| B4 | a measured record of the wrong code, inside B3 | its reason becomes NOT_EVALUATED: nobody measured *this* code for it | no return | TestABandOfAnotherCodeIsNotCountedAsMeasured |
| B5 | NEW - the record's rendered Value parses as a rational | appends to the quantile input; a value that did not parse would shorten QuantileBase and leave Measured alone, so the discrepancy reaches the screen | no return | TestTheTallyCarriesTheDistributionAndNotOnlyTheCrossings |
| B6 | range clearing the per-record `seen` set | resets duplicate suppression | no return | TestTheBandTallyAccountsForEveryCandidate |
| B7 | range over the record's crossings | counts them | no return | TestTheTallyHasOneColumnPerBandAndCountsThemRight |
| B8 | the crossing names a band this code does not shadow, or a duplicate | skipped | continue | TestABandOfAnotherCodeIsNotCountedAsMeasured |
| B9 | the crossing is true | Crossed[band]++ | no return | TestTheTallyHasOneColumnPerBandAndCountsThemRight |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| BandsFor | the code's scale, which is also the Crossed key set | total | ast.json calls |
| big.Rat.SetString | parses the rendered value back for the quantiles | returns ok=false rather than panicking; B5 handles it | ast.json calls |
| bandQuantiles | the distribution positions | total; nil for an empty input | ast.json calls; band.go:bandQuantiles |
| b.Reason | names the missing input of an unmeasured record | never empty for an unmeasured record | band.go:Reason |

## State mutations and fallbacks

- Local BandTally only; the input slice is not modified. bandQuantiles sorts the slice this function built, never the caller's.
- Total == Measured + sum(NotMeasured) still holds: B5 does not change which half a record lands in.

## Safety conclusion

- Safe edit boundary: Letting the quantile parse failure fall into NotMeasured, which would break the invariant above by moving a record between halves after it was counted. QuantileBase is reported separately for that reason.
- High-risk impact: no - an aggregate over shadow records that decide nothing.
