# Function Logic Map: `shadowBandsFor`

- Source: `internal/candidate/watch.go`
- Function: `internal/candidate/watch.go:shadowBandsFor`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

New in this change. It exists so that the function assembling a Verdict has no band construction in it, and so that the function building the bands has no Chase in scope to write a veto into.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| summary Summary | any; only summary.Candidate is read | assessOne | a zero Candidate makes MeasureExtendedBand answer NO_CANDIDATE rather than measure |
| sighting Sighting | measured or not | MeasureFirstSighting | an unmeasured sighting yields an unmeasured band carrying its reason |
| expansion Expansion | measured or not | MeasureExpansion | expansionBandReason names the missing input; no number is recorded |
| at time.Time | the assessment instant | AssessOptions.At | a zero instant becomes NO_INSTANT through inputAgeReason |
| th VetoThresholds | only inputAge() is read; ExtendedGainPct is not | AssessOptions.Thresholds | a zero MaxInputAge takes the package default, never 'no ceiling' |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no branch: a single unconditional return | none | the two shadow records | TestAVetoWithNoThresholdStillLeavesAShadowRecord, TestABandNamesTheSameMissingInputTheVetoWould, TestNoFunctionThatProducesAVerdictCanSeeAShadowBand |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| MeasureSeenLateBand | the seen_late shadow record | total; returns an unmeasured band instead of failing | ast.json calls; band.go:MeasureSeenLateBand |
| MeasureExtendedBand | the extended shadow record | total; same | ast.json calls; band.go:MeasureExtendedBand |

## State mutations and fallbacks

- None. No pointer parameter, no package state, no I/O. Both results are values.

## Safety conclusion

- Safe edit boundary: Adding a Chase, a VetoState or a Verdict to this signature. That is the whole point of the function: with no verdict in scope there is nothing for a crossing to be written into, and a *Verdict parameter would have moved the one-line transition rather than removed it.
- High-risk impact: no - candidate discovery only; no order, stop, sizing, ledger, auth or fill path.
