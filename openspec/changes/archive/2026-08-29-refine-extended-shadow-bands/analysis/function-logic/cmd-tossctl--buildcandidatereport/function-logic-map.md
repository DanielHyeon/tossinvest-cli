# Function Logic Map: `buildCandidateReport`

- Source: `cmd/tossctl/candidate.go`
- Function: `cmd/tossctl/candidate.go:buildCandidateReport`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: the shadow band block is built in BandsFor's order, carries the quantiles and the collapse alarm, and no longer hands out the tally's map. B1-B13 are base's branches unchanged; B14 and B15 are new.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| res candidate.CycleResult | one turn's result | Cycle | a nil Bands map yields no shadow band block, and writeCandidateTable skips it by its `ok` check |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over NotDue source ids | fills Sources.NotDue | no return | existing scan report tests |
| B2 | range over missing sources | fills Sources.Missing with reason and rate-limit flag | no return | TestTheScanOutputNamesTheMissingSourcesAndTheRetreat |
| B3 | range over backoff entries | fills Backoff | no return | same |
| B4 | range over readings in sorted source order | fills Readings deterministically | no return | existing scan report tests |
| B5 | range over per-source sightings | fills Sightings | no return | existing first-sighting tests |
| B6 | range over one source's unmeasured reasons | fills that entry's NotMeasured | no return | same |
| B7 | the passed count is not zero | swaps passedNote for passedUnexpected | no return | existing passed-note tests |
| B8 | range over raised veto codes | fills Veto.Raised | no return | TestTheScanJSONReportsTheCountsAnOperatorActsOn |
| B9 | range over unmeasured veto codes | fills Veto.NotMeasured | no return | same |
| B10 | range over veto reasons | fills Veto.Reasons | no return | same |
| B11 | range over acceleration not-computed reasons | fills ShadowAcceleration.NotComputed | no return | same |
| B12 | range over the shadow band tallies | one shadowBandReport per code | no return | same |
| B13 | range over one tally's unmeasured reasons | fills that block's NotMeasured | no return | same |
| B14 | NEW - range over BandsFor(code) | builds Crossed as an ordered list rather than a map, so encoding/json cannot sort ten numeric edges as strings | no return | TestTheScanJSONReportsTheCountsAnOperatorActsOn, order assertion |
| B15 | NEW - range over the tally's quantiles | copies them into this surface's own type | no return | TestTheReportShowsWhereTheValuesWereAndNotOnlyHowManyCrossed |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| candidate.BandsFor | NEW - the scale order the JSON is emitted in | total; nil for a code with no scale | ast.json calls |
| tally.CollapsedAlarm | NEW - the judgement, taken from the tally rather than reworded here | total; empty when the scale resolved something | ast.json calls |
| candidateTallyAlarms / sortedSourceIDs | unchanged | total | ast.json calls |

## State mutations and fallbacks

- Builds and returns a local candidateReport. Reads res and writes nothing to it.

## Safety conclusion

- Safe edit boundary: Ranging the tally's Crossed map to build the list. Go randomises map order, so two runs of the same scan would differ - the property sortedSourceIDs already exists to protect.
- High-risk impact: no - a read-only rendering of a discovery result.
