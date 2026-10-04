# Function Logic Map: `PollClosedBars`

- Source: `internal/officialbars/producer.go`
- Source SHA-256: `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `PollClosedBars(ctx context.Context, reader CandleReader, store BarStore, in BarPollInput) (BarPollResult, error)`
- Source range: `169:1`–`400:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 42, returns 22, calls 86, assignments 56, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- One poll of one symbol: crawl the official 1-minute pages newest-first, validate every page, then append the closed bars as `official_closed_bar_1m` evidence. Decisions 16/17 fix the package boundary (stdlib plus `internal/clock`, `internal/official`, `internal/scheduler`, `internal/strategyevidence`; import allowlist pinned by `TestProductionImportsStayInsideTheAllowlist`).
- `BarPollInput`: `Market` must parse as KR or US; `Symbol` must be non-blank and whitespace-free (the market grammar itself belongs to the reader, the envelope constructor and `SealBarSeries` — ruling 29 removed the duplicated regexp here); `Calendar.Market` must equal `Market`; `PollAt` is required and is the sole clock (no `time.Now` in this package — grep of `internal/officialbars/producer.go` finds the word only in the package doc comment that forbids it); `LowerBound` defaults to the regular open; `MaxPages` is 0 (meaning `defaultMaxPages` = 4) or 1..8.
- Calendar gate (ruling 24, amending decision 17(a)): `ValidityAt` refuses **only** `CalendarClockSkew` — a producer records evidence and must be able to poll a 6 h 30 m session through a 6 h freshness window. Instead the snapshot is bound to the instant by `Market.TradingDay(PollAt) == Calendar.Today.Date`, and both edges of `Today.Regular` must fall on that same market-local day. Which snapshot was used stays visible to L3 through `calendar_version` in every payload.
- **Decision 30 (2026-08-18) — two clock spaces inside one function.** The broker labels each 1-minute candle with its **close** instant (human-run US probe, 2026-08-18 03:29 KST; review.md). `adoptPage` converts once, so every `observedBar.openAt` this function touches — `bars[i].openAt`, `NewestObserved`, the `lowerBound` comparison, the merged-order walk, the window filter, `OpenAt` and `SuccessorOpenAt` — is **already an open instant**. The `before` literal and the `NextBefore` cursor are the exception: they stay in the broker's **label (close) space** and are compared against each other (B27), never against a converted bar. The source comment at 301–302 states that deliberately: an upper bound is compared with an upper bound.
- Consequence of the inclusive close bound: `before = PollAt` already excludes the forming bar, because the forming bar's close instant has not arrived yet. The "never admit `bars[0]`" successor rule (decision 6) is kept anyway — decision 6 requires an *observed* successor as proof — at the recorded cost of one poll of latency per bar.
- Page contract, per decision 17(c) as amended by ruling 27: pages are newest-first and strictly descending inside the reader; a following page must not start newer than the previous page's oldest bar; an equal-instant overlap bar is dropped exactly once and the merged list must still be strictly descending (`RefusalPageOrder`). All of these run on converted open instants.
- Decision 6 invariant: `bars[0]` (the newest observed bar) is never admitted — the only evidence a minute is closed is that its successor was observed. Storage is regular-window only (`Today.Regular.Open <= open_at < Close`, design §5) and, under decision 30, that test compares **true opens**: for US the bar *labelled* 16:00 ET opens at 15:59 and is the **last regular** bar, while the bar *labelled* 09:30 ET opens at 09:29 and is **pre-market** — the inverse of what the brief assumed before the probe. KRX reads the same way against 09:00–15:30 KST. Bars outside the window are still observed, counted and used as successors.
- Test-helper note, stated once for the whole bundle: the re-based tests express open instants through `usBar`/`krBar` (`producer_test.go:115,131`), which now synthesise `Timestamp = open + 1m`; the new `usBarLabelled`/`krBarLabelled` (`:122,:136`) speak in raw broker labels and are used where the label itself is the subject.
- Pages-first invariant: every page is fetched and validated before the first `Append`. Any contract refusal returns with zero rows written.
- De-duplication horizon (ruling 25): the `SealBarSeries` read uses `EvaluationAt = IngestionCutoff = PollAt + knownReadHorizon` (`366 * 24h`) so it sees every row already stored for the session, whatever instant the store stamped it with. Replay cutoffs are L3's concern.

## Branches and early returns

Exact AST return nodes: `172`, `176`, `180`, `183`, `187`, `196`, `199`, `210`, `217`, `220`, `225`, `236`, `260`, `263`, `269`, `275`, `299`, `304`, `327`, `346`, `392`, `399`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 171:2 | a nil reader or a nil store → `DEPENDENCY_MISSING` before anything else | untested: every test passes a fake reader and a real temp store; reachable from a caller, recorded gap |
| B2 | if | 175:2 | `ParseMarket` refused the market → `MARKET_INVALID`, no request | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `unknown market`) |
| B3 | if | 179:2 | blank or whitespace-carrying symbol → `SYMBOL_INVALID`, no request | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtests `empty symbol`, `symbol with a space`) |
| B4 | if | 182:2 | the calendar belongs to another market → `CALENDAR_MARKET_MISMATCH` | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `calendar market mismatch`) |
| B5 | if | 186:2 | zero `PollAt` → `POLL_AT_INVALID` (the poll carries its own clock) | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `poll instant missing`) |
| B6 | if | 195:2 | ruling 24 calendar gate: refuse only `CalendarClockSkew`, accept VALID/STALE/REFRESH_TOO_LATE | taken: `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `calendar fetched after the poll`); untaken: `TestPollAcceptsAStaleOrLateCalendarThroughTheWholeSession` (open+6h and open+6h29m both admit 2 bars) |
| B7 | if | 198:2 | today carries no regular session → `NO_REGULAR_SESSION` | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `no regular session`) |
| B8 | range | 204:2 | walk both edges (open, close) of the regular window | `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay` and every accepting poll |
| B9 | if | 209:3 | ruling 24: an edge whose market-local day is not `Today.Date` → `CALENDAR_INVALID` | `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay` (hand-built snapshot; zero reader calls) |
| B10 | if | 216:2 | `TradingDay(PollAt)` itself failed → `CALENDAR_DAY_MISMATCH` | not-applicable: `TradingDay` only fails when the market zone cannot be loaded, and `time/tzdata` is embedded in the binary |
| B11 | if | 219:2 | ruling 24: the poll instant is on another market-local day → `CALENDAR_DAY_MISMATCH` | taken: `TestPollRefusesInvalidInputBeforeAnyRequest` (subtest `poll instant on another trading day`); untaken across the date line: `TestPollLabelsOvernightUSBarsWithTheirEasternDate` (01:00 KST → `US:2026-08-14`) |
| B12 | if | 224:2 | `MaxPages` negative or above 8 → `MAX_PAGES_INVALID` | `TestPollRefusesInvalidInputBeforeAnyRequest` (subtests `max pages above the bound`, `max pages negative`) |
| B13 | if | 230:2 | `MaxPages` 0 → the documented default of 4 | taken: `TestPollDefaultsToFourPages`; untaken: `TestPollStopConditions` (subtest `max pages`, explicit 2) |
| B14 | if | 234:2 | the `Asia/Seoul` zone is unavailable → `DEPENDENCY_MISSING` | not-applicable: unreachable defence, stated in the source comment at 235 — `time/tzdata` is compiled in, so the lookup cannot fail |
| B15 | if | 243:2 | zero `LowerBound` → the regular open | taken: every poll that leaves `LowerBound` unset, e.g. `TestPollStopConditions` (subtest `lower bound`); untaken: `TestPollStopsAtAnExplicitLowerBound` |
| B16 | for | 254:2 | the crawl loop: count the page, request it, validate, merge | `TestPollCrawlsPagesInTheMeasuredShape` (2 measured-shape pages) |
| B17 | if | 259:3 | reader failure → `READER_ERROR`, `Pages` preserved (ruling 29), zero appends | `TestPollReturnsAReaderErrorWithoutAppending`, `TestPollKeepsThePageCountWhenTheReaderFails` |
| B18 | if | 262:3 | ruling 29 page identity: the page carries another market/symbol → `PAGE_IDENTITY_MISMATCH` | `TestPollRefusesAPageForAnotherMarketOrSymbol` |
| B19 | if | 268:3 | `adoptPage` refused a candle timestamp, or its converted open instant missed a whole minute → propagate `PAGE_INVALID` | not-applicable: through `official.Client` every candle label already matched the reader's grammar, parsed and (ruling 26) started a minute, so neither `adoptPage` B2 nor its decision-30 sibling B3 can fire (see `internal-officialbars--adoptpage`); no interface-guard test exists for this arm |
| B20 | if | 272:3 | a previous page and a non-empty current page exist → run the cross-page rules | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` |
| B21 | if | 274:4 | `checkOverlap` refused → return `OVERLAP_MISMATCH`, zero appends | not-applicable through `official.Client` (ruling 27 unreachability comment at 438–443); pinned as an interface guard by `TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded` and `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical` |
| B22 | if | 280:4 | ruling 27: drop exactly the one equal-instant overlap bar (no silent `seen` dedupe) | `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` (Observed 5, Admitted 4) |
| B23 | if | 285:3 | `page.Terminal` → stop, `Terminal` true | `TestPollNeverAdmitsTheNewestObservedBar`, `TestPollCrawlsPagesInTheMeasuredShape` |
| B24 | if | 289:3 | an empty page (still carrying a cursor) → stop | `TestPollStopConditions` (subtest `empty page`) |
| B25 | if | 292:3 | the page's oldest bar is older than `LowerBound` → `ReachedLowerBound`, stop | `TestPollStopConditions` (subtest `lower bound`), `TestPollStopsAtAnExplicitLowerBound` |
| B26 | if | 297:3 | the cursor is not a timestamp → `PAGE_INVALID` | not-applicable: unreachable defence, stated in the source comment at 298 — the reader already parsed the cursor under the same grammar |
| B27 | if | 303:3 | D4 loop guard: the new cursor is not strictly older than the bound this request carried → `CURSOR_LOOP`; both values stay in the broker's label (close) space and are never converted (source comment at 301–302) | `TestPollRefusesACursorThatDoesNotMoveBackwards`, `TestPollTruncatesTheSubSecondPollInstant` (a cursor equal to the truncated bound is refused) |
| B28 | if | 307:3 | `Pages` reached `maxPages` → `Truncated`, stop | `TestPollStopConditions` (subtest `max pages`), `TestPollDefaultsToFourPages` |
| B29 | if | 311:3 | the advisory rate budget is exhausted → `RateGated`, stop before the next request | `TestPollStopConditions` (subtest `rate gated`, 1 reader call) |
| B30 | for | 325:2 | walk the merged list to prove it is strictly newest-first | `TestPollRefusesAMergedListThatIsNotStrictlyDescending` and every accepting poll |
| B31 | if | 326:3 | ruling 27: the merged list is not strictly descending → `PAGE_ORDER_INVALID`, zero appends | not-applicable through `official.Client` (unreachability comment at 320–324); pinned as an interface guard by `TestPollRefusesAMergedListThatIsNotStrictlyDescending` and `TestPollInterfaceGuardRefusesAnAdjacentDuplicateInstant` |
| B32 | if | 333:2 | at least one bar observed → publish `NewestObserved`, itself a converted open instant | taken: `TestPollNeverAdmitsTheNewestObservedBar` and every crawl test; `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant` asserts `NewestObserved == label − 1 min`; untaken (a first page with zero bars) is untested — recorded gap |
| B33 | if | 345:2 | the de-duplication read failed → `STORE_ERROR` before any append | untested: the test doubles fail `Append`, never `SealBarSeries`; recorded gap |
| B34 | range | 350:2 | index the stored rows by `Payload.OpenAtMS` (D2 — the query already pins market/symbol/session/interval) | `TestPollRepolledWithIdenticalContentAppendsNothing`, `TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant` |
| B35 | for | 357:2 | decision 6: admit from index 1 — the newest observed bar is never admitted | `TestPollNeverAdmitsTheNewestObservedBar` (3 bars → 2 envelopes), `TestPollAdmitsThePreviousNewestBarOnceASuccessorExists` |
| B36 | if | 359:3 | design §5: skip bars whose **converted open** falls outside `[regularOpen, regularClose)` — they still serve as successors. Under decision 30 the bar *labelled* 16:00 ET opens at 15:59 and is admitted as the last regular bar, while the bar *labelled* 09:30 ET opens at 09:29 and is skipped as pre-market | `TestPollTreatsTheClosingLabelAsTheLastRegularBar` (labels 16:01/16:00/15:59/09:31/09:30 → 5 observed, 3 admitted, stored opens 09:30, 15:58, 15:59), `TestPollTreatsTheKoreanClosingLabelAsTheLastRegularBar` (the same boundary on KRX 09:00–15:30 KST), `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (open-instant helpers; 5 observed, 2 admitted; the 15:59 bar's successor is the bar opening at 16:00, after hours) |
| B37 | if | 368:3 | this minute is already stored (ruling 25 horizon makes every stored row visible) | `TestPollRepolledWithIdenticalContentAppendsNothing`, `TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant`, `TestPollRetryAfterAStoreErrorDoesNotConflict` |
| B38 | if | 369:4 | identical raw → `Unchanged++` and skip; different raw → revision+1 correction | equal: `TestPollRepolledWithIdenticalContentAppendsNothing`; different: `TestPollWritesACorrectionAsARevisionAndLeavesTheEarlierReplayAlone` (r2 with `SupersedesRevisionIdentity` r1) |
| B39 | if | 383:3 | the envelope constructor refused this bar → record a `BarRefusal`, keep the other bars | `TestPollRecordsAConstructorRefusalAndKeepsTheOtherBars` |
| B40 | if | 387:3 | `store.Append` returned an error | `TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision`, `TestPollReturnsStoreErrorWithTheCountsSoFar` |
| B41 | if | 388:4 | `ErrRevisionConflict` → `Conflicts++` and continue; any other error → `STORE_ERROR` with the counts so far | conflict: `TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision`; other: `TestPollReturnsStoreErrorWithTheCountsSoFar` |
| B42 | if | 395:3 | the admitted bar was a correction → `Corrections++` | `TestPollWritesACorrectionAsARevisionAndLeavesTheEarlierReplayAlone` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `refuse` | 172:18 |
| `marketclock.ParseMarket` | 174:17 |
| `string` | 174:41 |
| `refuse` | 176:18 |
| `strconv.Quote` | 176:57 |
| `string` | 176:71 |
| `marketCode` | 178:10 |
| `checkSymbol` | 179:12 |
| `refuse` | 183:18 |
| `strconv.Quote` | 184:30 |
| `string` | 184:44 |
| `string` | 184:81 |
| `in.PollAt.IsZero` | 186:5 |
| `refuse` | 187:18 |
| `in.Calendar.ValidityAt` | 195:17 |
| `refuse` | 196:18 |
| `string` | 196:73 |
| `refuse` | 199:18 |
| `market.TradingDay` | 208:19 |
| `refuse` | 210:19 |
| `strconv.Quote` | 211:68 |
| `strconv.Quote` | 212:39 |
| `market.TradingDay` | 215:18 |
| `refuse` | 217:18 |
| `refuse` | 220:18 |
| `refuse` | 225:18 |
| `strconv.Itoa` | 226:43 |
| `strconv.Itoa` | 227:27 |
| `marketclock.MarketKR.Location` | 233:16 |
| `refuse` | 236:18 |
| `sessionCalendar` | 239:15 |
| `lowerBound.IsZero` | 243:5 |
| `Format` | 250:12 |
| `In` | 250:12 |
| `in.PollAt.Truncate` | 250:12 |
| `in.PollAt.Truncate` | 251:19 |
| `reader.StrictMinuteCandles` | 258:16 |
| `refuse` | 260:19 |
| `strconv.Itoa` | 260:62 |
| `refuse` | 263:19 |
| `strconv.Quote` | 264:21 |
| `strconv.Quote` | 265:15 |
| `adoptPage` | 267:19 |
| `len` | 272:6 |
| `len` | 272:27 |
| `len` | 273:29 |
| `checkOverlap` | 274:14 |
| `current.openAt.Equal` | 280:7 |
| `append` | 284:10 |
| `len` | 289:6 |
| `current.openAt.Before` | 292:6 |
| `len` | 292:14 |
| `time.Parse` | 296:18 |
| `refuse` | 299:19 |
| `strconv.Quote` | 299:56 |
| `cursor.Before` | 303:7 |
| `refuse` | 304:19 |
| `strconv.Quote` | 305:15 |
| `page.Budget.Exhausted` | 311:6 |
| `len` | 325:26 |
| `bars.openAt.Before` | 326:7 |
| `refuse` | 327:19 |
| `bars.openAt.Format` | 328:57 |
| `len` | 332:20 |
| `len` | 333:5 |
| `minuteGaps` | 336:16 |
| `store.SealBarSeries` | 339:17 |
| `in.PollAt.Add` | 342:17 |
| `in.PollAt.Add` | 342:67 |
| `refuse` | 346:18 |
| `make` | 349:11 |
| `len` | 349:61 |
| `len` | 357:26 |
| `bar.openAt.Before` | 359:6 |
| `bar.openAt.Before` | 359:41 |
| `uint64` | 366:15 |
| `uint64` | 368:29 |
| `bar.openAt.UnixMilli` | 368:36 |
| `strategyevidence.NewClosedBar1mEnvelope` | 376:20 |
| `append` | 384:21 |
| `err.Error` | 384:83 |
| `store.Append` | 387:16 |
| `errors.Is` | 388:7 |
| `refuse` | 392:19 |
| `Format` | 392:80 |
| `bar.openAt.UTC` | 392:80 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `marketclock.ParseMarket`, `marketCode`, `sessionCalendar` | 174, 178, 239 | market code and `KRX:`/`US:` session prefix; `TestPollUsesTheKoreanCalendarAndSessionPrefix` (`KRX:2026-08-14`), `TestPollLabelsOvernightUSBarsWithTheirEasternDate` (`US:2026-08-14`) |
| `checkSymbol` | 179 | nil/blank check only — ruling 29 removed the duplicated market regexp |
| `in.Calendar.ValidityAt` | 195 | ruling 24: only `CalendarClockSkew` refuses |
| `market.TradingDay` ×2 | 208, 215 | window-edge day and poll-instant day binding (ruling 24); `internal-clock--market.tradingday` |
| `marketclock.MarketKR.Location` | 233 | the `before` literal is formatted in `Asia/Seoul` for both markets (measured request shape) |
| `in.PollAt.Truncate` / `In` / `Format` | 250–251 | page-1 bound `2006-01-02T15:04:05.000-07:00`, an inclusive upper bound on the bar **close** — so `before = PollAt` already excludes the forming bar (decision 30(a)); `TestPollTruncatesTheSubSecondPollInstant`, `TestPollUsesTheKoreanCalendarAndSessionPrefix` |
| `reader.StrictMinuteCandles` | 258 | **external** — the only outbound call; production binding is `(*official.Client).StrictMinuteCandles`, which sits on the ordinary `c.get → send → doRequest` token path. Wired end to end by `TestPollEndToEndThroughTheRealOfficialClient` (httptest, 2 pages, real clock store) |
| `adoptPage` | 267 | page → `observedBar`, converting the broker's close label into `openAt` (decision 30) and attaching the response-bound `ReadAt`/`BodyDigest`; see `internal-officialbars--adoptpage` |
| `checkOverlap` | 274 | interface-level cross-page guard; see `internal-officialbars--checkoverlap` |
| `time.Parse(time.RFC3339, page.NextBefore)`, `cursor.Before` | 296, 303 | cursor decoding and the D4 loop guard, both in the broker's label (close) space — deliberately not converted, since the comparison is bound-to-bound (source comment at 301–302) |
| `page.Budget.Exhausted` | 311 | advisory rate gate between pages only; never a retry |
| `minuteGaps` | 336 | informational gap report; see `internal-officialbars--minutegaps` |
| `store.SealBarSeries` | 339 | **external** — one de-duplication read per poll at the ruling 25 horizon (`PollAt + 366*24h` on both cutoffs, `MaxBars` 512, `RegularSessionOnly` false); `TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant` |
| `strategyevidence.NewClosedBar1mEnvelope` | 376 | **external** — the only evidence constructor reachable from this package (`TestProductionWritesEvidenceOnlyThroughTheCombinedConstructor` forbids `NewEnvelope` and a bare `Header{` literal) |
| `store.Append` | 387 | **external** — the only write; idempotency and quarantine belong to `Store.Append` |
| `errors.Is(err, strategyevidence.ErrRevisionConflict)` | 388 | conflict classification; `TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision` |

## State mutations and fallbacks

- Locals only (56 AST assignments): `result`, `bars`, `previous`, `before`/`beforeInstant`, `maxPages`, `lowerBound`, `known`. `bars` holds converted open instants; `before`/`beforeInstant` hold broker labels. No package-level state, no goroutines (`go_statements` null), no defers, no logging, no clock read — `PollAt` is the sole instant and `time.Now` does not appear in the file.
- The only persistent mutation is `store.Append` at 387, one envelope per admitted bar, after every page has been fetched and validated. A contract refusal at B17–B21, B26, B27, B31 or B33 returns before the first append; a store failure at B41 returns with the rows already appended left in place and the counts reported, because appended rows are legitimate evidence.
- Per-bar fallbacks are deliberate and narrow: a constructor refusal (B39) records a `BarRefusal` and continues; a revision conflict (B41) increments `Conflicts` and continues. `Gaps` never suppresses an observed bar — refusing a session for a missing minute is L3's contiguity rule (decision 17(g)).

## Safety conclusion

- High-risk adjacency: the reader this producer holds sits on the official client GET/token path (`c.get → send → doRequest`), so a poll can drive the shared credential's ≤2 refresh-on-401; the producer itself writes evidence rows. Both fail closed — an invalid argument or calendar sends no request at all (asserted by `len(reader.calls) == 0` in `TestPollRefusesInvalidInputBeforeAnyRequest` and `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay`), and any page-contract refusal appends nothing.
- Decision 30's blast radius runs through this function even though it edits none of its own arithmetic: bar identity, successor binding, the regular-window filter and the gap report all read `openAt`. The conversion is done once in `adoptPage`; the source comment there forbids reverting it, and this function's own cursor comparison is the one place that must stay in label space.
- No order, stop-loss, sizing, guardian or toggle surface is touched; the package's import allowlist forbids `net/http`, engine, journal, execgw, guardian, router and toggle packages (`TestProductionImportsStayInsideTheAllowlist`), and nothing in production calls `PollClosedBars` yet (L5 wiring, human-approved).
- Recorded residuals (review.md 2026-08-17, not defects of this lot): `doRequest` reads the body uncapped and the 2 MiB cap is applied after the read; an absent `nextBefore` is refused although the documented schema lists only `candles` as required (fail-closed until a terminal page is measured); the reader allocates `[]json.RawMessage` before the count bound, inside the 2 MiB cap; ruling 26 makes a single off-minute bar refuse the whole page (availability traded for successor integrity); and under ruling 25 a pre-existing foreign row of the same identity is absorbed as a correction, so `Conflicts` signals only a writer racing between the de-duplication read and the append — L5 must not read `Conflicts == 0` as "no foreign writer".
