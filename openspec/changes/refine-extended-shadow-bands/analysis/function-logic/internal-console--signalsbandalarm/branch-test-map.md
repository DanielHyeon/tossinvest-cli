# Branch Test Map: `signalsBandAlarm`

- Source: `internal/console/signals.go`
- Function: `internal/console/signals.go:signalsBandAlarm`

RED/GREEN is what was actually observed while implementing §5.2. The function is new,
so its reporting branch was exercised from a failing state.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | a scale that separated 165 from 4 — no alarm | TestAShadowScaleThatSeparatedSomethingRaisesNoAlarm | no — negative control. It passed before this function existed and has to keep passing, which is the whole of what it asserts | yes |
| — | 131 measured records, all with the same crossings | TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing | yes — the page rendered an ordinary row before the wiring existed | yes |

## Mutations run against these branches

| Mutation | Result |
|---|---|
| `Alarm` left unset in `signalsBandTalliesFrom` (the state before this change) | RED — `131 measured records that all produced the same crossings render with no alarm` |
| the `{{if .Alarm}}` block removed from `templates_signals.go` | RED — same assertion. The sentence exists and the page does not carry it |
