# Function Logic Map: `adoptPage`

- Source: `internal/officialbars/producer.go`
- Source SHA-256: `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `adoptPage(page official.StrictMinutePage) ([]observedBar, error)`
- Source range: `415:1`–`434:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 3, returns 3, calls 14, assignments 4, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Input is one validated `official.StrictMinutePage`; output is the same candles in the same order (newest first) as `observedBar` values, each carrying the page's response-bound provenance.
- **Decision 30 — this is the single place where the broker's clock convention is converted.** `candle.Timestamp` is parsed as the bar's **close** instant and `openAt = closeAt − barInterval` (`barInterval` = `strategyevidence.ClosedBar1mIntervalMS` ms = 60 s, `producer.go:138`). The measurement is the human-run US probe of **2026-08-18 03:29 KST** (review.md): at wall clock 03:29:14 the newest label was `03:30:00` with volume 251, and 26 seconds later the *same* label `03:30:00` carried 1,089 while the label `03:29:00` stayed frozen at 2,997 — so the bar growing during `[03:29, 03:30)` was the one labelled `03:30`, i.e. the label names the close. The openapi schema ("봉 시작 시각"), `candle_reads.go`'s comment and the a047 KR precedent all say the opposite and are all wrong on this point.
- The **stored** contract is unchanged: `bar_label = "open_at"` remains L1a's strategy-facing payload semantics and `internal/strategyevidence` was not touched. Only the meaning of the wire value moved, so the conversion lives here and nowhere else.
- The source comment at 402–414 states the rule and adds a deliberate instruction: **do not "fix" this subtraction back**. Reverting it shifts every stored bar one minute, and the regular-session window and every `successor_open_at_ms` shift with it.
- Every bar built here inherits `page.ReadAt` (the `BodyReadComplete` instant of the used attempt) and `page.BodyDigest` (the `sha256:` of the exact response bytes). That is what makes `source_observed_at_ms` and `source_response_digest` in the stored payload response-bound rather than clock-bound — ruling 29's provenance test asserts it end to end.
- `openAt` is normalised to UTC before the subtraction (`closeAt.UTC().Add(-barInterval)`). The broker's label is a fixed-offset KST literal; the payload stores the converted instant, and the raw label survives verbatim inside `candle.Timestamp`.
- Invariant relied on by every later rule: through `official.Client` each `candle.Timestamp` already matched `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,3})?[+-]\d{2}:\d{2}$`, parsed, and (ruling 26) started a minute. This function re-parses rather than trusting, and re-asserts minute alignment on the *converted* side (B3); both refusals are unreachable behind the real reader.
- Test-helper note for every row below and in the sibling bundles: `usBar`/`krBar` (`producer_test.go:115,131`) take the bar's **open** instant and synthesise `Timestamp = open + 1m`, so the re-based tests still read as open instants; the new `usBarLabelled`/`krBarLabelled` (`:122,:136`) speak in raw broker labels and are used where the boundary itself is the subject.

## Branches and early returns

Exact AST return nodes: `420`, `427`, `433`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | range | 417:2 | walk the page's candles in order, converting each broker close label to an open instant and attaching `ReadAt` and `BodyDigest` | `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant` (three raw labels in, three stored/reported opens exactly one minute earlier), `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollTakesObservedAtFromTheResponseNotThePollInstant` (`ReadAt = PollAt − 10 min` shows up as `source_observed_at_ms`), `TestPollNeverAdmitsTheNewestObservedBar` |
| B2 | if | 419:3 | a candle timestamp that `time.Parse` cannot read → `PAGE_INVALID`, nothing adopted | not-applicable: through `official.Client` the reader refuses any candle whose timestamp fails the grammar, fails to parse, or does not start a minute (ruling 26), so this arm is a defence for a foreign `CandleReader`; unlike its sibling at 298 it carries no unreachability comment in the source, and no interface-guard test drives it — recorded gap |
| B3 | if | 426:3 | decision 30's post-conversion assertion: the converted `openAt` does not land on a whole minute → `PAGE_INVALID`, nothing adopted | not-applicable: subtracting an exact 60 s from a minute-aligned instant is minute-aligned, and ruling 26 (`internal-official--strictminutecandle` B10) already refused any label with non-zero seconds or a fraction before the producer sees it — so behind `official.Client` this arm is unreachable and no fixture can drive it. Kept as the cheap nail the source comment at 424–425 describes, because if it ever fires bar identity itself has broken. The untaken (aligned) side is pinned on every converted bar by `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant` and `TestPollTreatsTheClosingLabelAsTheLastRegularBar` — recorded gap for the taken side |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `make([]observedBar, 0, len(page.Candles))` | 416 | exact-capacity local slice; no growth beyond the page |
| `time.Parse(time.RFC3339, candle.Timestamp)` | 418 | parse of the broker's **close** label — an instant the reader already validated; `internal-official--strictminutecandle` B9/B10 is the first parse |
| `closeAt.UTC().Add(-barInterval)` | 423 | **the decision-30 conversion**: UTC normalisation then minus one bar interval (60 s, from `strategyevidence.ClosedBar1mIntervalMS`); `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant` |
| `openAt.Second()`, `openAt.Nanosecond()` | 426 | minute alignment re-asserted on the converted side |
| `refuse(RefusalPageInvalid, …)`, `strconv.Itoa`, `strconv.Quote`, `openAt.Format(time.RFC3339Nano)` | 420–421, 427–429 | two typed refusals, each naming the offending index and the offending literal |
| `append` | 431 | the `observedBar` carries the converted `openAt` and the untouched `candle` |

## State mutations and fallbacks

- Locals only (4 AST assignments): `out`, `closeAt`/`err`, `openAt`. No package state, no store access, no goroutines, no defers, no clock read — the only arithmetic is the fixed `−60 s`.
- No fallback: a single unreadable timestamp or a single off-minute converted instant discards the whole page (`return nil, …`) and, through `PollClosedBars` B19, the whole poll — consistent with the pages-first rule that a refusal appends nothing.

## Safety conclusion

- Pure transformation with no I/O. It now carries two safety duties: keep each bar bound to the response it came from (so L3 can tell two observations of the same minute apart), and convert the broker's close label into the open instant the stored contract promises. It fails closed on both things it can judge.
- Decision-30 blast radius, and why the "do not revert" comment is load-bearing: `openAt` is the bar's identity (`OpenAtMS`), the successor binding (`SuccessorOpenAtMS`), the regular-window test in `PollClosedBars` B36 and the gap arithmetic in `minuteGaps`. Removing the subtraction shifts all four by one minute in the same direction, which is why the correction had to land before any `setup_id` was minted from `Bars[0].ID` (review.md decision 30(c); nothing is wired, so nothing was).
- High-risk adjacency is indirect: the values it stamps become `source_observed_at_ms` and `source_response_digest` in stored evidence. Recorded residual (review.md 2026-08-17): the reader's uncapped `doRequest` read with a post-read 2 MiB cap bounds how large a page reaching this function can be.
- Recorded residual (review.md 2026-08-18): KR's own labelling convention is still formally unmeasured — the KR probe stands for 2026-08-18 09:00–15:30 KST. The conversion is applied to both markets on the strength of the US measurement plus the corroborating KR 20:00 closing-print observation.
