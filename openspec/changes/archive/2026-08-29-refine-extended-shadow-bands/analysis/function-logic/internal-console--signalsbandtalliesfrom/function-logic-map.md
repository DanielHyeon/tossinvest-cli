# Function Logic Map: `signalsBandTalliesFrom`

- Source: `internal/console/signals.go`
- Function: `internal/console/signals.go:signalsBandTalliesFrom`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

**Modified** in this change (§5.2, independent review R2). One field is filled that
was not filled before: `Alarm`. The four branches are base structure and none of them
moved — the diff is the single call `signalsBandAlarm(tally)` inside the block
literal built at B1.

Why it had to change: this function and `cmd/tossctl`'s scan report build from the
**same** `candidate.BandTally`, and only the scan report knew about the collapse. The
spec delta names no surface — *"보고 표면은 그것을 정상 수치로 표시하지 않고 경보로
표시해야 한다"* — so a collapsed KR tally rendering here as an ordinary row is an
approved requirement being false about a live screen.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in map[VetoCode]BandTally | may be nil, may be missing a code | `candidate.TallyVerdicts`, via `SignalsMarket.Bands` | a missing code is skipped at B2 rather than rendered as an empty block |
| `candidate.VetoCodes` | D3's fixed order | `internal/candidate` | the page does not reorder itself between refreshes the way a map range would |
| `candidate.BandsFor(code)` | may be empty (near_high) | `internal/candidate/band.go` | B3 then adds no count columns, and `Collapsed()` refuses to call that state collapsed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over `candidate.VetoCodes` | appends to `out` | no return | TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing |
| B2 | this code has no tally | none | `continue` — absent is not zero, and a block of zeroes would say the code was counted | existing signals tests: only the two recorded codes render |
| B3 | range over the code's bands | appends one count per band, in the scale's order | no return | TestTheSignalsScreenSaysSoWhenAShadowScaleResolvedNothing asserts the counts survive |
| B4 | range over the not-measured census | fills a rename map | no return | existing signals tests |

The `Alarm` assignment is inside B1's body and has no branch of its own: it is a
field of the literal, and the branch is inside `signalsBandAlarm`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `signalsBandAlarm` | **new** — the collapse sentence for this tally | pure, cannot error, empty on the ordinary state | ast.json calls, line 841 |
| `candidate.BandsFor` | the scale, so the columns are the scale's and in its order | pure, may return empty | ast.json calls, line 843 |
| `sortedSignalsCounts` | the not-measured census, biggest first | pure | ast.json calls, line 852 |
| `string` conversions | VetoCode and VetoUnmeasured to display strings | n/a | ast.json calls, lines 840, 850 |

No I/O, no clock, no config read. Nothing here can fail.

## State mutations and fallbacks

- `out` and `block` are local. `counts` is a local map rebuilt per code.
- No fallback: a code with no tally is skipped, not defaulted.

## Safety conclusion

- Safe edit boundary: replacing the counts with the alarm, or dropping the counts
  when the alarm is set. The alarm goes beside them because they are its evidence —
  the same rule `signalsTallyAlarm` states for the veto tally.
- Second boundary: computing the collapse condition here. It comes from
  `candidate.BandTally.Collapsed` so the terminal and the browser cannot disagree
  about it; that disagreement is the defect `TallyVerdicts`' own doc records having
  happened once already with a fourth shadow band.
- High-risk impact: no. Read-only display assembly. No order, stop, sizing, ledger,
  auth or fill path is reachable, no threshold is read, and no toggle is written.
