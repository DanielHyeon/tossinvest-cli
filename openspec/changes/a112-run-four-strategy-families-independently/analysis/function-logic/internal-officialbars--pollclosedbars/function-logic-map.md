# Function Logic Map: `PollClosedBars`

- Source: `internal/officialbars/producer.go`
- Source SHA-256: `83960410b7b870ca60fad568002060de49ebc7271d72b94d46f665ff274b29b1` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `PollClosedBars(ctx context.Context, reader CandleReader, store BarStore, in BarPollInput) (BarPollResult, error)`
- Source range: `166:1`–`395:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 42, returns 22, calls 86, assignments 56, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- One poll of one symbol: crawl the official 1-minute pages newest-first, validate every page, then append the closed bars as `official_closed_bar_1m` evidence. Decisions 16/17 fix the package boundary (stdlib plus `internal/clock`, `internal/official`, `internal/scheduler`, `internal/strategyevidence`; import allowlist pinned by `TestProductionImportsStayInsideTheAllowlist`).
- `BarPollInput`: `Market` must parse as KR or US; `Symbol` must be non-blank and whitespace-free (the market grammar itself belongs to the reader, the envelope constructor and `SealBarSeries` — ruling 29 removed the duplicated regexp here); `Calendar.Market` must equal `Market`; `PollAt` is required and is the sole clock (no `time.Now` in this package — grep of `internal/officialbars/producer.go` finds the word only in the package doc comment that forbids it); `LowerBound` defaults to the regular open; `MaxPages` is 0 (meaning `defaultMaxPages` = 4) or 1..8.
- Calendar gate (ruling 24, amending decision 17(a)): `ValidityAt` refuses **only** `CalendarClockSkew` — a producer records evidence and must be able to poll a 6 h 30 m session through a 6 h freshness window. Instead the snapshot is bound to the instant by `Market.TradingDay(PollAt) == Calendar.Today.Date`, and both edges of `Today.Regular` must fall on that same market-local day. Which snapshot was used stays visible to L3 through `calendar_version` in every payload.
- Page contract, per decision 17(c) as amended by ruling 27: pages are newest-first and strictly descending inside the reader; a following page must not start newer than the previous page's oldest bar; an equal-instant overlap bar is dropped exactly once and the merged list must still be strictly descending (`RefusalPageOrder`).
- Decision 6 invariant: `bars[0]` (the newest observed bar) is never admitted — the only evidence a minute is closed is that its successor was observed. Storage is regular-window only (`Today.Regular.Open <= open_at < Close`, design §5); bars outside the window are observed, counted and used as successors.
- Pages-first invariant: every page is fetched and validated before the first `Append`. Any contract refusal returns with zero rows written.
- De-duplication horizon (ruling 25): the `SealBarSeries` read uses `EvaluationAt = IngestionCutoff = PollAt + knownReadHorizon` (`366 * 24h`) so it sees every row already stored for the session, whatever instant the store stamped it with. Replay cutoffs are L3's concern.

## Branches and early returns

Exact AST return nodes: `169, 173, 177, 180, 184, 193, 196, 207, 214, 217, 222, 233, 257, 260, 266, 272, 296, 299, 322, 341, 387, 394`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 168:2 | a nil reader or a nil store → `DEPENDENCY_MISSING` before anything else | untested: every test passes a fake reader and a real temp store; reachable from a caller, recorded gap |
| B2 | if | 172:2 | `ParseMarket` refused the market → `MARKET_INVALID`, no request | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `unknown market`) |
| B3 | if | 176:2 | blank or whitespace-carrying symbol → `SYMBOL_INVALID`, no request | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtests `empty symbol`, `symbol with a space`) |
| B4 | if | 179:2 | the calendar belongs to another market → `CALENDAR_MARKET_MISMATCH` | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `calendar market mismatch`) |
| B5 | if | 183:2 | zero `PollAt` → `POLL_AT_INVALID` (the poll carries its own clock) | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `poll instant missing`) |
| B6 | if | 192:2 | ruling 24 calendar gate: refuse only `CalendarClockSkew`, accept VALID/STALE/REFRESH_TOO_LATE | taken: `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `calendar fetched after the poll`); untaken: `TestPollAcceptsAStaleOrLateCalendarThroughTheWholeSession` (open+6h and open+6h29m both admit 2 bars) |
| B7 | if | 195:2 | today carries no regular session → `NO_REGULAR_SESSION` | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `no regular session`) |
| B8 | range | 201:2 | walk both edges (open, close) of the regular window | `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay` and every accepting poll |
| B9 | if | 206:3 | ruling 24: an edge whose market-local day is not `Today.Date` → `CALENDAR_INVALID` | `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay` (hand-built snapshot; zero reader calls) |
| B10 | if | 213:2 | `TradingDay(PollAt)` itself failed → `CALENDAR_DAY_MISMATCH` | not-applicable: `TradingDay` only fails when the market zone cannot be loaded, and `time/tzdata` is embedded in the binary |
| B11 | if | 216:2 | ruling 24: the poll instant is on another market-local day → `CALENDAR_DAY_MISMATCH` | taken: `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `poll instant on another trading day`); untaken across the date line: `TestPollLabelsOvernightUSBarsWithTheirEasternDate` (01:00 KST → `US:2026-08-14`) |
| B12 | if | 221:2 | `MaxPages` negative or above 8 → `MAX_PAGES_INVALID` | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtests `max pages above the bound`, `max pages negative`) |
| B13 | if | 227:2 | `MaxPages` 0 → the documented default of 4 | taken: `TestPollDefaultsToFourPages`; untaken: `TestPollStopConditions` (subtest `max pages`, explicit 2) |
| B14 | if | 231:2 | the `Asia/Seoul` zone is unavailable → `DEPENDENCY_MISSING` | not-applicable: unreachable defence, stated in the source comment at 232 — `time/tzdata` is compiled in, so the lookup cannot fail |
| B15 | if | 240:2 | zero `LowerBound` → the regular open | taken: every poll that leaves `LowerBound` unset, e.g. `TestPollStopConditions` (subtest `lower bound`); untaken: `TestPollStopsAtAnExplicitLowerBound` |
| B16 | for | 251:2 | the crawl loop: count the page, request it, validate, merge | `TestPollCrawlsPagesInTheMeasuredShape` (2 measured-shape pages) |
| B17 | if | 256:3 | reader failure → `READER_ERROR`, `Pages` preserved (ruling 29), zero appends | `TestPollReturnsAReaderErrorWithoutAppending`, `TestPollKeepsThePageCountWhenTheReaderFails` |
| B18 | if | 259:3 | ruling 29 page identity: the page carries another market/symbol → `PAGE_IDENTITY_MISMATCH` | `TestPollRefusesAPageForAnotherMarketOrSymbol` |
| B19 | if | 265:3 | `adoptPage` refused a candle timestamp → propagate `PAGE_INVALID` | not-applicable: through `official.Client` every candle instant already matched the reader's grammar and parsed (see `internal-officialbars--adoptpage`); no interface-guard test exists for this arm |
| B20 | if | 269:3 | a previous page and a non-empty current page exist → run the cross-page rules | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` |
| B21 | if | 271:4 | `checkOverlap` refused → return `OVERLAP_MISMATCH`, zero appends | not-applicable through `official.Client` (ruling 27 unreachability comment at 413–418); pinned as an interface guard by `TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded` and `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical` |
| B22 | if | 277:4 | ruling 27: drop exactly the one equal-instant overlap bar (no silent `seen` dedupe) | `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` (Observed 5, Admitted 4) |
| B23 | if | 282:3 | `page.Terminal` → stop, `Terminal` true | `TestPollNeverAdmitsTheNewestObservedBar`, `TestPollCrawlsPagesInTheMeasuredShape` |
| B24 | if | 286:3 | an empty page (still carrying a cursor) → stop | `TestPollStopConditions` (subtest `empty page`) |
| B25 | if | 289:3 | the page's oldest bar is older than `LowerBound` → `ReachedLowerBound`, stop | `TestPollStopConditions` (subtest `lower bound`), `TestPollStopsAtAnExplicitLowerBound` |
| B26 | if | 294:3 | the cursor is not a timestamp → `PAGE_INVALID` | not-applicable: unreachable defence, stated in the source comment at 295 — the reader already parsed the cursor under the same grammar |
| B27 | if | 298:3 | D4 loop guard: the new cursor is not strictly older than the bound this request carried → `CURSOR_LOOP` | `TestPollRefusesACursorThatDoesNotMoveBackwards`, `TestPollTruncatesTheSubSecondPollInstant` (a cursor equal to the truncated bound is refused) |
| B28 | if | 302:3 | `Pages` reached `maxPages` → `Truncated`, stop | `TestPollStopConditions` (subtest `max pages`), `TestPollDefaultsToFourPages` |
| B29 | if | 306:3 | the advisory rate budget is exhausted → `RateGated`, stop before the next request | `TestPollStopConditions` (subtest `rate gated`, 1 reader call) |
| B30 | for | 320:2 | walk the merged list to prove it is strictly newest-first | `TestPollRefusesAMergedListThatIsNotStrictlyDescending` and every accepting poll |
| B31 | if | 321:3 | ruling 27: the merged list is not strictly descending → `PAGE_ORDER_INVALID`, zero appends | not-applicable through `official.Client` (unreachability comment at 315–319); pinned as an interface guard by `TestPollRefusesAMergedListThatIsNotStrictlyDescending` and `TestPollInterfaceGuardRefusesAnAdjacentDuplicateInstant` |
| B32 | if | 328:2 | at least one bar observed → publish `NewestObserved` | taken: `TestPollNeverAdmitsTheNewestObservedBar` and every crawl test; untaken (a first page with zero bars) is untested — recorded gap |
| B33 | if | 340:2 | the de-duplication read failed → `STORE_ERROR` before any append | untested: the test doubles fail `Append`, never `SealBarSeries`; recorded gap |
| B34 | range | 345:2 | index the stored rows by `Payload.OpenAtMS` (D2 — the query already pins market/symbol/session/interval) | `TestPollRepolledWithIdenticalContentAppendsNothing`, `TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant` |
| B35 | for | 352:2 | decision 6: admit from index 1 — the newest observed bar is never admitted | `TestPollNeverAdmitsTheNewestObservedBar` (3 bars → 2 envelopes), `TestPollAdmitsThePreviousNewestBarOnceASuccessorExists` |
| B36 | if | 354:3 | design §5: skip bars outside `[regularOpen, regularClose)` — they still serve as successors | `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (5 observed, 2 admitted; the 15:59 bar's successor is the 16:00 after-hours bar) |
| B37 | if | 363:3 | this minute is already stored (ruling 25 horizon makes every stored row visible) | `TestPollRepolledWithIdenticalContentAppendsNothing`, `TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant`, `TestPollRetryAfterAStoreErrorDoesNotConflict` |
| B38 | if | 364:4 | identical raw → `Unchanged++` and skip; different raw → revision+1 correction | equal: `TestPollRepolledWithIdenticalContentAppendsNothing`; different: `TestPollWritesACorrectionAsARevisionAndLeavesTheEarlierReplayAlone` (r2 with `SupersedesRevisionIdentity` r1) |
| B39 | if | 378:3 | the envelope constructor refused this bar → record a `BarRefusal`, keep the other bars | `TestPollRecordsAConstructorRefusalAndKeepsTheOtherBars` |
| B40 | if | 382:3 | `store.Append` returned an error | `TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision`, `TestPollReturnsStoreErrorWithTheCountsSoFar` |
| B41 | if | 383:4 | `ErrRevisionConflict` → `Conflicts++` and continue; any other error → `STORE_ERROR` with the counts so far | conflict: `TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision`; other: `TestPollReturnsStoreErrorWithTheCountsSoFar` |
| B42 | if | 390:3 | the admitted bar was a correction → `Corrections++` | `TestPollWritesACorrectionAsARevisionAndLeavesTheEarlierReplayAlone` |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `marketclock.ParseMarket`, `marketCode`, `sessionCalendar` | 171, 175, 236 | market code and `KRX:`/`US:` session prefix; `TestPollUsesTheKoreanCalendarAndSessionPrefix` (`KRX:2026-08-14`), `TestPollLabelsOvernightUSBarsWithTheirEasternDate` (`US:2026-08-14`) |
| `checkSymbol` | 176 | nil/blank check only — ruling 29 removed the duplicated market regexp |
| `in.Calendar.ValidityAt` | 192 | ruling 24: only `CalendarClockSkew` refuses |
| `market.TradingDay` ×2 | 205, 212 | window-edge day and poll-instant day binding (ruling 24); `internal-clock--market.tradingday` |
| `marketclock.MarketKR.Location` | 230 | the `before` literal is formatted in `Asia/Seoul` for both markets (measured request shape) |
| `in.PollAt.Truncate` / `In` / `Format` | 247–248 | page-1 bound `2006-01-02T15:04:05.000-07:00`; `TestPollTruncatesTheSubSecondPollInstant`, `TestPollUsesTheKoreanCalendarAndSessionPrefix` |
| `reader.StrictMinuteCandles` | 255 | **external** — the only outbound call; production binding is `(*official.Client).StrictMinuteCandles`, which sits on the ordinary `c.get → send → doRequest` token path. Wired end to end by `TestPollEndToEndThroughTheRealOfficialClient` (httptest, 2 pages, real clock store) |
| `adoptPage` | 264 | page → `observedBar` with the response-bound `ReadAt`/`BodyDigest`; see `internal-officialbars--adoptpage` |
| `checkOverlap` | 271 | interface-level cross-page guard; see `internal-officialbars--checkoverlap` |
| `time.Parse(time.RFC3339, page.NextBefore)`, `cursor.Before` | 293, 298 | cursor decoding and the D4 loop guard |
| `page.Budget.Exhausted` | 306 | advisory rate gate between pages only; never a retry |
| `minuteGaps` | 331 | informational gap report; see `internal-officialbars--minutegaps` |
| `store.SealBarSeries` | 334 | **external** — one de-duplication read per poll at the ruling 25 horizon (`PollAt + 366*24h` on both cutoffs, `MaxBars` 512, `RegularSessionOnly` false); `TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant` |
| `strategyevidence.NewClosedBar1mEnvelope` | 371 | **external** — the only evidence constructor reachable from this package (`TestProductionWritesEvidenceOnlyThroughTheCombinedConstructor` forbids `NewEnvelope` and a bare `Header{` literal) |
| `store.Append` | 382 | **external** — the only write; idempotency and quarantine belong to `Store.Append` |
| `errors.Is(err, strategyevidence.ErrRevisionConflict)` | 383 | conflict classification; `TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision` |

## State mutations and fallbacks

- Locals only (56 AST assignments): `result`, `bars`, `previous`, `before`/`beforeInstant`, `maxPages`, `lowerBound`, `known`. No package-level state, no goroutines (`go_statements` null), no defers, no logging, no clock read — `PollAt` is the sole instant and `time.Now` does not appear in the file.
- The only persistent mutation is `store.Append` at 382, one envelope per admitted bar, after every page has been fetched and validated. A contract refusal at B17–B21, B26, B27, B31 or B33 returns before the first append; a store failure at B41 returns with the rows already appended left in place and the counts reported, because appended rows are legitimate evidence.
- Per-bar fallbacks are deliberate and narrow: a constructor refusal (B39) records a `BarRefusal` and continues; a revision conflict (B41) increments `Conflicts` and continues. `Gaps` never suppresses an observed bar — refusing a session for a missing minute is L3's contiguity rule (decision 17(g)).

## Safety conclusion

- High-risk adjacency: the reader this producer holds sits on the official client GET/token path (`c.get → send → doRequest`), so a poll can drive the shared credential's ≤2 refresh-on-401; the producer itself writes evidence rows. Both fail closed — an invalid argument or calendar sends no request at all (asserted by `len(reader.calls) == 0` in `TestPollRefusesInvalidInputBeforeAnyRequest` and `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay`), and any page-contract refusal appends nothing.
- No order, stop-loss, sizing, guardian or toggle surface is touched; the package's import allowlist forbids `net/http`, engine, journal, execgw, guardian, router and toggle packages (`TestProductionImportsStayInsideTheAllowlist`), and nothing in production calls `PollClosedBars` yet (L5 wiring, human-approved).
- Recorded residuals (review.md 2026-08-17, not defects of this lot): `doRequest` reads the body uncapped and the 2 MiB cap is applied after the read; an absent `nextBefore` is refused although the documented schema lists only `candles` as required (fail-closed until a terminal page is measured); the reader allocates `[]json.RawMessage` before the count bound, inside the 2 MiB cap; ruling 26 makes a single off-minute bar refuse the whole page (availability traded for successor integrity); and under ruling 25 a pre-existing foreign row of the same identity is absorbed as a correction, so `Conflicts` signals only a writer racing between the de-duplication read and the append — L5 must not read `Conflicts == 0` as "no foreign writer".
