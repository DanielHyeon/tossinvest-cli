# Function Logic Map: `TestPollEndToEndThroughTheRealOfficialClient`

- Source: `internal/officialbars/producer_test.go`
- Source SHA-256: `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `TestPollEndToEndThroughTheRealOfficialClient(t *testing.T)`
- Source range: `1391:1`–`1462:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 13, returns 1, calls 57, assignments 16, defers 2, go statements 0.
- Disposition: **Test function edited by the decision-30 correction (broker label = bar close). Enumerated for gate completeness: the assertions are themselves the evidence; there is no production branch behind these rows.** The edit is one line — the page-2 cursor is now built by the `cursorBefore` helper (`producer_test.go:219`) instead of `brokerTimestamp(t, instants[1].Add(-time.Minute))`. Under decision 30 the measured cursor is "the page's oldest bar's **label** minus one minute", which in open-instant terms is exactly that bar's open, so the helper takes an open instant and the fixture keeps the same wire value it always had.
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- The only test in this package that drives `PollClosedBars` through the **real** `official.Client` rather than a fake `CandleReader`: an `httptest` server answers the token exchange and two `/api/v1/candles` pages, and a real `strategyevidence.Store` on its default (wall) clock receives the appends through a `recordingStore`.
- It runs on the **wall clock** (`time.Now`), which is what makes it an integration probe rather than a table test: the request literal, the token path, the body decode, the digest and the store all have to agree on instants nobody pinned.
- `wideUSCalendar` (`:1362`) makes almost the whole day regular (00:10–23:50 ET on the previous, current and next business day) precisely so the three synthetic bars, which sit 2–4 minutes before `now`, normally land inside the window.
- Fixture bars are built with `usBar` (`:115`), which takes an **open** instant and synthesises `Timestamp = open + 1m` — the broker's close label (decision 30). The two server pages are serialised by `candleJSON` (`:1353`) straight from those `RawMinuteCandle` values, so the bytes the client parses are broker-shaped, not producer-shaped.
- Expected outcome, asserted rather than assumed: 2 pages, 3 observed, 2 admitted, terminal, `SessionID == "US:"+day`, and every appended payload's `SourceObservedAtMS` **strictly** after `PollAt`.

## Branches and early returns

Exact AST return node: `1415` (the token-handler early return inside the server closure).

| Branch | AST kind | Source location | What the guard proves | Test disposition |
|---|---|---|---|---|
| B1 | if | 1395:2 | `marketclock.MarketUS.TradingDay(now)` failed → `t.Fatalf` | assertion guard: the eastern zone must load before anything else is meaningful; unreachable in practice because `time/tzdata` is embedded (the same declaration as `PollClosedBars` B10) |
| B2 | if | 1401:2 | the announced skip window: the oldest fixture bar falls before the regular open **or** the newest is not before the regular close → `t.Skipf` | proves the test refuses to run rather than assert a false negative when the wall clock sits next to the eastern day boundary. With a 00:10–23:50 ET window and bars 2–4 minutes old, that is 00:00–00:14 and 23:52–24:00 ET ≈ 22 minutes a day — the announced residual recorded in review.md 2026-08-17 |
| B3 | if | 1413:3 | inside the server handler: the request path is `/oauth2/token` → serve a bearer token and return | proves the producer reaches the broker over the **ordinary** client path (`c.get → send → doRequest`), token exchange included, not through a bypass |
| B4 | switch | 1417:3 | dispatch on the request's `before` query parameter | proves the crawl is driven by the request literal itself: the server can only answer values the producer actually formats |
| B5 | case | 1418:3 | `before == firstBefore` (the `PollAt`-derived page-1 bound) → two candles plus `nextBefore` = cursor | proves the page-1 bound is `PollAt.Truncate(second)` rendered in `Asia/Seoul` with the measured layout, and that a non-terminal page is followed |
| B6 | case | 1421:3 | `before == cursor` → one candle with `nextBefore: null` | proves the producer sends back exactly the cursor it was given and stops on the terminal page |
| B7 | case | 1423:3 | any other `before` → `t.Errorf` plus HTTP 400 | proves no third or unexpected request is made; a drifted bound fails the test instead of being silently answered |
| B8 | if | 1436:2 | `strategyevidence.Open` failed → `t.Fatalf` | harness guard on the real store; nothing downstream would be evidence if the store never opened |
| B9 | if | 1445:2 | `PollClosedBars` returned an error → `t.Fatalf` | proves the whole path — request literal, token, body contract, decision-30 conversion, envelope construction, append — completes without a refusal |
| B10 | if | 1448:2 | `Pages != 2 \|\| Observed != 3 \|\| Admitted != 2 \|\| !Terminal` → `t.Fatalf` | proves the crawl shape end to end: two pages, three bars seen, decision 6 withholding the newest, terminal reached |
| B11 | if | 1451:2 | `SessionID != "US:"+day` → `t.Fatalf` | proves the session identity is the market-local trading day of the poll instant, not a UTC date |
| B12 | range | 1454:2 | walk every envelope the `recordingStore` captured | proves the provenance assertion is applied to **each** appended row, not just the first |
| B13 | if | 1458:3 | `payload.SourceObservedAtMS <= PollAt.UnixMilli()` → `t.Fatalf` | proves `source_observed_at_ms` comes from the transport's `BodyReadComplete`, not from `PollAt`. The comparison is deliberately strict: an `==` would let the "observed at = poll instant" mistake pass |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `time.Now`, `marketclock.MarketUS.TradingDay` | 1393, 1394 | the wall clock is the poll instant; the trading day comes from the same helper production uses (`internal-clock--market.tradingday`) |
| `wideUSCalendar`, `usBar`, `cursorBefore`, `candleJSON` | 1398, 1407, 1410, 1419–1422 | fixture builders; `usBar` speaks open instants and emits close labels, `cursorBefore` emits the measured cursor for a given oldest open (decision 30) |
| `pollAt.Truncate(time.Second).In(…).Format(beforeLayout)` | 1409 | the test recomputes the page-1 bound the same way the producer does, so B5 can only match if the producer's literal is byte-identical |
| `httptest.NewServer(http.HandlerFunc(…))` | 1412 | **live binding** — a real HTTP server; the client's transport, status classification and body read all run |
| `official.New(… WithBaseURL, WithHTTPClient)` | 1430–1431 | **live binding** — the real `*official.Client`, whose `StrictMinuteCandles` satisfies `CandleReader`; token file lives in `t.TempDir()` |
| `strategyevidence.Open` | 1433 | **live binding** — a real store on a real temp SQLite path with the default (wall) clock, so ingestion instants are not faked |
| `PollClosedBars` | 1442 | the function under test; the only production symbol in this file's e2e path |
| `payloadOf` | 1455 | decodes the appended envelope's typed payload for the provenance assertion |
| `server.Close`, `store.Close` | 1428, 1439 | the two AST defers; ordinary harness teardown |

## State mutations and fallbacks

- Locals only (16 AST assignments): `ctx`, `pollAt`, `day`, `calendar`, `minute`, `instants`, `bars`, `firstBefore`, `cursor`, `server`, `client`, `store`, `recorder`, `result`, `envelope`, `payload`. No package-level state.
- Real side effects, all sandboxed to the test: an HTTP server on a loopback port, a token file under `t.TempDir()`, and a SQLite store under a second `t.TempDir()`. Both are released by the two defers. No network egress leaves the machine, and no production toggle, journal or credential is touched.
- The one fallback is B2's `t.Skipf`: rather than assert against a window the wall clock has left, the test announces the skip. That is a deliberate honesty trade — it is why the ≈22 minutes a day are carried as a recorded residual instead of being hidden by a frozen clock.

## Safety conclusion

- Safe edit boundary: a test file. It writes nothing outside `t.TempDir()`, opens no real broker connection, and places no order. The `official.Client` it builds is pointed at `httptest`, so the shared OpenAPI credential is never used and `send`'s refresh-on-401 never runs against production.
- High-risk impact: no. There is no production branch behind any row above; every row is an assertion or a fixture dispatch. Its value to the gate is that it is the **only** evidence that the producer and the real reader agree on the request literal, the terminal cursor and the response-bound `source_observed_at_ms` — which is exactly what a fake `CandleReader` cannot prove.
- Decision-30 note: the correction changed one fixture line here and none of the assertions, because the fixture helpers were re-based rather than the expectations. The 13 assertion guards below are unchanged in meaning; what changed underneath them is that `usBar` now emits `Timestamp = open + 1m` and the producer subtracts it back.
