# Function Logic Map: `adoptPage`

- Source: `internal/officialbars/producer.go`
- Source SHA-256: `83960410b7b870ca60fad568002060de49ebc7271d72b94d46f665ff274b29b1` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `adoptPage(page official.StrictMinutePage) ([]observedBar, error)`
- Source range: `398:1`–`409:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 2, returns 2, calls 8, assignments 3, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Input is one validated `official.StrictMinutePage`; output is the same candles in the same order (newest first) as `observedBar` values, each carrying the page's response-bound provenance.
- Every bar built here inherits `page.ReadAt` (the `BodyReadComplete` instant of the used attempt) and `page.BodyDigest` (the `sha256:` of the exact response bytes). That is what makes `source_observed_at_ms` and `source_response_digest` in the stored payload response-bound rather than clock-bound — ruling 29's provenance test asserts it end to end.
- `openAt` is normalised to UTC. The broker's label is a fixed-offset KST literal; the payload stores the instant, and the raw label survives inside `candle.Timestamp`.
- Invariant relied on by every later rule: through `official.Client` each `candle.Timestamp` already matched `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,3})?[+-]\d{2}:\d{2}$`, parsed, and (ruling 26) started a minute. This function re-parses rather than trusting, but the refusal it can raise is unreachable behind the real reader.

## Branches and early returns

Exact AST return nodes: `403, 408`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | range | 400:2 | walk the page's candles in order, attaching `ReadAt` and `BodyDigest` to each | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollTakesObservedAtFromTheResponseNotThePollInstant` (`ReadAt = PollAt − 10 min` shows up as `source_observed_at_ms`), `TestPollNeverAdmitsTheNewestObservedBar` |
| B2 | if | 402:3 | a candle timestamp that `time.Parse` cannot read → `PAGE_INVALID`, nothing adopted | not-applicable: through `official.Client` the reader refuses any candle whose timestamp fails the grammar, fails to parse, or does not start a minute (ruling 26), so this arm is a defence for a foreign `CandleReader`; unlike its sibling at 295 it carries no unreachability comment in the source, and no interface-guard test drives it — recorded gap |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `make([]observedBar, 0, len(page.Candles))` | 399 | exact-capacity local slice; no growth beyond the page |
| `time.Parse(time.RFC3339, candle.Timestamp)` | 401 | second parse of an instant the reader already validated; `internal-official--strictminutecandle` B9/B10 is the first |
| `refuse(RefusalPageInvalid, …)`, `strconv.Itoa`, `strconv.Quote` | 403–404 | typed refusal naming the offending index and literal |
| `append`, `openAt.UTC()` | 406 | UTC normalisation of the fixed-offset broker label |

## State mutations and fallbacks

- Locals only (3 AST assignments): `out`, `openAt`, `err`. No package state, no store access, no goroutines, no defers, no clock read.
- No fallback: a single unreadable timestamp discards the whole page (`return nil, …`) and, through `PollClosedBars` B19, the whole poll — consistent with the pages-first rule that a refusal appends nothing.

## Safety conclusion

- Pure transformation with no I/O; its only safety duty is to keep each bar bound to the response it came from, which is what lets L3 tell two observations of the same minute apart. It fails closed on the one thing it can judge.
- High-risk adjacency is indirect: the values it stamps become `source_observed_at_ms` and `source_response_digest` in stored evidence. Recorded residual (review.md 2026-08-17): the reader's uncapped `doRequest` read with a post-read 2 MiB cap bounds how large a page reaching this function can be.
