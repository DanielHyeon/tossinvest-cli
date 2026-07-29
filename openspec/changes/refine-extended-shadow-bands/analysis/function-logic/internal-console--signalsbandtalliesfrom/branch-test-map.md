# Branch Test Map: `signalsBandTalliesFrom`

- Source: `internal/console/signals.go`
- Function: `internal/console/signals.go:signalsBandTalliesFrom`

RED/GREEN is what was actually observed. `no` in the RED column means the branch is
base behaviour this change did not alter and no failing state was manufactured for
it; the test named is the one covering it now.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the extended tally is rendered as a block | TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing | yes — before the `Alarm` field was filled, the block rendered with no alarm in it | yes |
| B2 | a code with no tally is skipped | existing signals tests — only the recorded codes appear on the page | no | yes |
| B3 | one count column per band, in the scale's order | TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing (the counts survive beside the alarm) | no | yes |
| B4 | the not-measured census | existing signals tests | no | yes |

## Mutations run against this function

| Mutation | Result |
|---|---|
| `Alarm` left unset in the block literal | RED — `131 measured records that all produced the same crossings render with no alarm` |
| the `{{if .Alarm}}` block removed from the template | RED — same assertion, proving the field reaches the page and is not merely populated |
