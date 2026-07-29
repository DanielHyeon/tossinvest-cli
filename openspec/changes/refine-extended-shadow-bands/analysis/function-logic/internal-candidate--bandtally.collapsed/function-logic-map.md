# Function Logic Map: `Collapsed`

- Source: `internal/candidate/band.go`
- Function: `internal/candidate/band.go:BandTally.Collapsed`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

New in this change. The spec's alarm condition, computed from the aggregate alone.

**Amended in §5.3** (independent review R3): B1 gained a second disjunct,
`len(t.Crossed) == 0`. Nothing else moved.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| t BandTally (receiver) | any tally | TallyBands | a tally that measured nothing is not collapsed - see B1 |
| t.Crossed | keys are exactly BandsFor(Code), and that may be empty | TallyBands seeds it from the scale | an empty scale is not a collapsed one - see B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nothing was measured, **or the code has no scale at all** | none | false - `measured 0 of N` already says the first, and for the second there is no instrument, so there is no instrument failure. Without the second disjunct the loop below runs zero times and the answer is vacuously true, which would light the alarm permanently on any code whose BandsFor is nil - near_high today and extended on the day issues I3 is carried out | TestATallyThatResolvedNothingSaysSo, third case; TestATallyWithNoScaleIsNotReportedAsCollapsed |
| B2 | range over the per-band counts | none | no return | TestATallyThatResolvedNothingSaysSo |
| B3 | a count is neither 0 nor Measured | none | false - that band separated some records from others, so the scale resolved something | TestATallyThatResolvedNothingSaysSo, second case |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure arithmetic over the tally's own fields | n/a | ast.json calls is empty |

## State mutations and fallbacks

- None. Value receiver, no writes.

## Safety conclusion

- Safe edit boundary: Making this depend on what the scale was *meant* to resolve. It must not: the check has no idea what the edges mean, which is why it keeps working when the edges change.
- High-risk impact: no - a statement about the instrument, not about any candidate. It is on BandTally and not on ShadowBand for that reason; ShadowBand still carries no predicate but Crossed.
