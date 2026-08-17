# Function Logic Map: `checkOverlap`

- Source: `internal/officialbars/producer.go`
- Source SHA-256: `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `checkOverlap(previousLast, currentFirst observedBar) error`
- Source range: `444:1`–`457:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 3, returns 3, calls 7, assignments 0, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs are two adopted bars: the oldest bar of page N and the newest bar of page N+1. The function judges only their relative position and, when they name the same minute, their byte identity.
- Decision 30 (2026-08-18): both `openAt` values arrive **already converted** by `adoptPage` (`openAt = broker label − 60 s`), so every comparison here is open-instant against open-instant. Nothing in this function reads a cursor or a `before`, which are the two values that stay in the broker's label (close) space — so the correction moved both operands by the same minute and left the rule itself untouched. The `candle` compared for byte identity still carries the raw broker `Timestamp`.
- Ruling 27 fixes the rule and its reachability. The measured contract is cursor = last bar − 1 min and page N+1's first bar ≤ the cursor, so equality is *permitted* (an inclusive bound) but never *required*: an extended-hours minute may have no trades, and demanding `first == cursor` would refuse legitimate crawls. Manager decision 17(c) as first written ("page N+1 first == page N last, byte-identical") was withdrawn as a mis-statement from memory rather than from the receipt.
- Unreachability is declared in the source comment at 438–443: through `official.Client` the reader enforces "cursor strictly older than this page's oldest bar" and "every bar ≤ `before`", so page N+1 always starts strictly older than page N ended. Both arms below therefore exist for a foreign `CandleReader` implementation — the same posture as L1a's record-id subsumption precedent.
- Byte identity is compared as the whole `official.RawMinuteCandle` value (all seven decoded fields), not as a re-encoded body.

## Branches and early returns

Exact AST return nodes: `447`, `453`, `456`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | switch | 445:2 | the two-arm guard; falling through means the pages line up and the caller proceeds | `TestPollCrawlsPagesInTheMeasuredShape` (measured shape, no overlap bar), `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` (equal instants, identical bars) |
| B2 | case | 446:2 | the next page starts *newer* than the previous page's oldest bar → `OVERLAP_MISMATCH` | not-applicable through `official.Client` (ruling 27 unreachability comment at 438–443); pinned as an interface guard by `TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded` |
| B3 | case | 450:2 | the two pages carry the same minute with differing fields → `OVERLAP_MISMATCH` | not-applicable through `official.Client` (same comment); pinned as an interface guard by `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical`, with the accepting side (same minute, identical fields) pinned by `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `currentFirst.openAt.After(previousLast.openAt)` | 446 | forward-walk violation test (ruling 27 form D1) |
| `refuse(RefusalOverlapMismatch, …)`, `…openAt.Format(time.RFC3339)` | 447–449, 453–454 | typed refusal naming both instants |
| `currentFirst.openAt.Equal(previousLast.openAt)` | 450 | equal-minute detection; the field comparison beside it is a plain struct equality on `official.RawMinuteCandle` |

## State mutations and fallbacks

- None. No assignments (AST `assignments` 0), no state, no I/O, no goroutines, no defers, no clock read; the function only returns an error or nil.
- Fallback behaviour lives in the caller: `PollClosedBars` B21 turns a refusal into a whole-poll `OVERLAP_MISMATCH` with zero appends, and B22 drops exactly one equal-instant bar when this function stays silent.

## Safety conclusion

- Defensive cross-page integrity guard. It cannot admit anything; its only outcomes are "refuse the poll" or "say nothing", so a foreign reader cannot smuggle a page that walks forward or that contradicts an already-observed minute.
- Both branches are declared not-applicable behind the production reader and are still exercised through the `CandleReader` interface by two tests renamed `TestPollInterfaceGuard…` in the fix round, so the declaration is documented rather than merely asserted.
- Recorded residual (review.md 2026-08-17): the measured contiguity of the crawl is enforced by the reader's cursor invariants rather than here, and ruling 27 explicitly declines to require `first == cursor`; the merged-list order check at `producer.go:325-330` is the sibling guard that catches a repeated minute.
