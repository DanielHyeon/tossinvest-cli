# Function Logic Map: `minuteGaps`

- Source: `internal/officialbars/producer.go`
- Source SHA-256: `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `minuteGaps(bars []observedBar, regularOpen, regularClose time.Time) []MinuteGap`
- Source range: `460:1`–`481:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 5, returns 1, calls 11, assignments 8, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Input is the merged, strictly newest-first list of observed bars plus the regular window edges; output is a report of the minutes that lie *between two observed bars* and inside the regular window.
- Decision 30 (2026-08-18): the `openAt` values arrive **already converted** by `adoptPage` (`openAt = broker label − 60 s`), so gaps are counted in open-instant space, not in the broker's label space. Both operands and both window edges moved together, so the arithmetic here is unchanged; `TestPollReportsGapsOnConvertedOpenInstants` is the test that pins it — labels 09:35/09:33/09:31 become opens 09:34/09:32/09:30 and the two reported one-minute holes are the minutes **opening** at 09:33 and 09:31.
- Deliberate scope, stated in the `MinuteGap` doc comment at 58–62: the region outside the observed bars (newer than the newest, older than the oldest) is **not** counted — that is unseen territory, not a hole. Each gap is clamped to `[regularOpen, regularClose)`, so a hole that straddles the open or the close is reported only for its in-window part.
- `To` is inclusive of the last missing minute (`end.Add(-time.Minute)`), while `Minutes` counts the half-open span; both are normalised to UTC.
- Gap policy (decision 17(g)): the producer **never** suppresses an observed bar because of a gap. Refusing a session for a missing minute is L3's contiguity rule, because a late-published bar must remain appendable and visible under a later `IngestionCutoff` (dual-cutoff replay).

## Branches and early returns

Exact AST return node: `480`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | for | 462:2 | walk consecutive observed pairs newest→oldest, taking the half-open span `(older+1m, newer]` on converted open instants | `TestPollReportsGapsOnConvertedOpenInstants` (labels 09:35/09:33/09:31 → two one-minute holes opening at 09:33 and 09:31), `TestPollGapsAreClampedToTheRegularWindow` (3 gaps), `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (one 388-minute gap) |
| B2 | if | 465:3 | clamp the start up to the regular open (a pre-session hole is reported only from the open) | `TestPollGapsAreClampedToTheRegularWindow` (`Gaps[2]` starts exactly at the open, 1 minute) |
| B3 | if | 468:3 | clamp the end down to the regular close (an after-hours hole stops at the close) | `TestPollGapsAreClampedToTheRegularWindow` (`Gaps[0]` = 386 minutes, `To` = close − 1 min) |
| B4 | if | 471:3 | after clamping the window is empty → report nothing for this pair | every adjacent-minute pair, e.g. `TestPollCrawlsPagesInTheMeasuredShape` and `TestPollNeverAdmitsTheNewestObservedBar` (no gaps reported) |
| B5 | if | 475:3 | a positive but sub-minute span → report nothing | not-applicable: ruling 26 forces every bar instant onto a minute boundary and the calendar's window edges come from `AdaptOfficialCalendar` at whole minutes, so after B4 the span is always at least one whole minute; defensive |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `older.Add(time.Minute)` | 464 | the first possibly-missing minute is the one after the older bar |
| `start.Before(regularOpen)`, `end.After(regularClose)` | 465, 468 | the two clamps; `TestPollGapsAreClampedToTheRegularWindow` asserts both edges |
| `start.Before(end)` | 471 | emptiness test after clamping |
| `int(end.Sub(start) / time.Minute)` | 474 | whole-minute count |
| `append(gaps, MinuteGap{…})`, `start.UTC()`, `end.Add(-time.Minute).UTC()` | 478 | inclusive `To`, UTC normalisation |

## State mutations and fallbacks

- Locals only (8 AST assignments): `gaps`, `newer`/`older`, `start`/`end`, `minutes`. No package state, no store access, no goroutines, no defers, no clock read. `bars` is read, never written.
- No fallback and no refusal path: the function cannot fail. A pair that cannot produce an in-window gap is skipped (B4, B5) and the walk continues.

## Safety conclusion

- Report-only. It cannot hide, reorder or refuse a bar; the admission loop in `PollClosedBars` never consults its output, so a bug here can only mis-report `Gaps` for L5 logging, never change what becomes evidence.
- The clamping rules keep the report honest about the window it claims to describe, which matters because L3's contiguity rule — the rule that *can* refuse a session — is a separate read-side decision.
- Recorded residual (review.md 2026-08-17): the receipt shows 27 non-one-minute gaps in a single measured crawl, so gaps are normal in extended hours; that is exactly why the producer reports rather than refuses.
