# Function Logic Map: `signalsBandAlarm`

- Source: `internal/console/signals.go`
- Function: `internal/console/signals.go:signalsBandAlarm`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

New in this change (§5.2, independent review R2). The Korean sentence this surface
reports a collapsed shadow tally with. The judgement is not here: it is
`candidate.BandTally.Collapsed`, and this function only asks it.

Its sibling is `signalsTallyAlarm`, and the split is the one that function's doc
argues for — a rule implemented twice eventually disagrees with itself, and the
disagreement shows up as a page that looks calm. Each surface owns its wording
because the readers differ; neither owns the arithmetic.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| t candidate.BandTally | any tally, including the zero value | `candidate.TallyVerdicts` via `SignalsMarket.Bands` | a zero tally has `Measured == 0`, so `Collapsed()` is false and B1 returns empty |
| t.Measured | >= 0 | `candidate.TallyBands` | rendered into the sentence as-is; it is the evidence the sentence is about |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!t.Collapsed()` | none | `""` — the ordinary state. Empty is what the template branches on, so no alarm block renders | TestAShadowScaleThatSeparatedSomethingRaisesNoAlarm |
| — | fallthrough | none | the sentence, carrying `t.Measured` | TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Collapsed` | the judgement, which this package does not re-implement | pure, cannot error | ast.json calls, line 568 |
| `strconv.Itoa` | the measured count into the sentence | pure, cannot error | ast.json calls, line 571 |

No I/O, no clock, no config read, no live binding. Nothing here can fail.

## State mutations and fallbacks

- None. Value parameter, no writes, no package state, no fallback.

## Safety conclusion

- Safe edit boundary: re-deriving the collapse condition here instead of calling
  `Collapsed()`. That is the one edit that must not happen — it is how the terminal
  and the browser come to disagree about whether a scale resolved anything, and the
  disagreement would be invisible because both pages would still render.
- Second boundary: rendering this sentence *instead of* the counts. It goes beside
  them; the counts are what the alarm is about, and a page without them has hidden
  its own evidence.
- High-risk impact: no. Display only. No order, stop, sizing, ledger, auth or fill
  path is reachable from here, and it reads no threshold.
