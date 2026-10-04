# Branch Test Map: `run`

Branch IDs follow the post-amendment `ast.json` (38 branches; B27/B28 are the new crawl-record write/live checks, old B27..B36 are now B29..B38). Pre-existing branches carry the tests recorded at M-B1 acceptance (review.md 2026-08-15); B26 carries the lot 0.7b.3 RED/GREEN (RED first observed live on 2026-08-16, then as a failing Go test against the pre-amendment code).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | caller-claimed build command / relative or missing `--go-binary` / missing required flags | `TestValidateConfigRejectsCallerClaimedBuildCommand`, `TestValidateConfigRequiresExplicitAbsoluteGoBinary` | yes | yes |
| B2 | nil ctx/reader/now/identity/builder | not-applicable: unreachable via `newProductionDependencies`; constructor guard | no | no |
| B3 | injected clock returns zero time | not-applicable: production clock is `time.Now`; guard only | no | no |
| B4 | overall budget already exhausted before receipt open | `TestRunHoldsBeforeReaderWhenLessThanOneRequestBudgetRemains` | yes | yes |
| B5 | production receipt opener chosen when `deps.openReceipt == nil` | `TestRunCollectsTerminalCandleThenSingleOrderbookAndCalendar` (nil opener path) | no | yes |
| B6 | insecure receipt root | `TestRunRejectsInsecureReceiptBeforeReader` | yes | yes |
| B7 | HOLD paths write `TAINTED` (deferred) | `TestRunStopsOnReaderErrorWithoutFallback` and every HOLD test asserting the taint | yes | yes |
| B8 | store exposes `setLive` → live guard on writes | `TestRunHoldsWhenCancelledDuringPostReadFinalization` | yes | yes |
| B9 | live check before pre-request identity fails | `TestRunHoldsBeforeReaderWhenLessThanOneRequestBudgetRemains` | yes | yes |
| B10 | pre-request identity snapshot error | `TestRunRejectsUnsafeSelectedGoBeforeReader` | yes | yes |
| B11 | prescribed binary verification error | `TestRunRejectsTamperedGoExecutionSnapshotBeforeReader` | yes | yes |
| B12 | live check after identity fails | `TestRunHoldsWhenClockExceedsOverallBudgetAfterFinalResponse` (same guard) | yes | yes |
| B13 | preflight sentinel write error | `TestRunRejectsInsecureReceiptBeforeReader` (write capability) | yes | yes |
| B14 | live check after preflight fails | `TestRunHoldsBeforeReaderWhenLessThanOneRequestBudgetRemains` | yes | yes |
| B15 | explicit initial `before` seeds the seen set | `TestRunHoldsWhenFirstCursorRepeatsExplicitInitialBefore` | yes | yes |
| B16 | candle loop bounded to four pages | `TestRunRecordsCapExhaustionAndStillCollectsOrderbookCalendarAndSeals` (exact call list: 4 candle, then orderbook, calendar) | yes | yes |
| B17 | fewer than 15 s left before a candle request | `TestRunHoldsBeforeReaderWhenLessThanOneRequestBudgetRemains` | yes | yes |
| B18 | reader error → HOLD, no fallback | `TestRunStopsOnReaderErrorWithoutFallback` | yes | yes |
| B19 | live check after candle read fails | `TestRunHoldsWhenClockExceedsOverallBudgetAfterFinalResponse` | yes | yes |
| B20 | invalid source result (bad cursor typing) | `TestRunHoldsOnNonStringRawCursorBeforeOrderbook` | yes | yes |
| B21 | candle receipt write error (secret-like body) | `TestRunHoldsOnSecretLikeRawBodyBeforeNextRequest` | yes | yes |
| B22 | live check after candle receipt fails | `TestRunHoldsWhenCancelledDuringPostReadFinalization` | yes | yes |
| B23 | cursor loop / malformed cursor | `TestRunHoldsOnCursorLoopBeforeOrderbook`, `TestRunStillHoldsOnCursorLoopAfterCapChange`, `TestRunHoldsOnNonStringRawCursorBeforeOrderbook` | yes | yes |
| B24 | raw-null terminal → crawl record `terminal=null` with `pages` = pages read, orderbook, calendar, seal | `TestRunCollectsTerminalCandleThenSingleOrderbookAndCalendar`, `TestRunRecordsNullTerminalInCandleCrawlRecord`, `TestRunRecordsPagesBeforeNullTerminal` | yes | yes |
| B25 | prior page remaining < 1 blocks the next page (failing side, including page 4 where no next request follows — intentional strictness) | `TestRunHoldsWhenEarlierPageRateBudgetIsExhausted`, `TestRunHoldsWhenPageFourRateBudgetIsExhausted` (passing side: `TestRunUsesDecodedCursorBytesWithoutNormalization`) | yes (mutants "remove rate gate" and "skip page-4 gate" killed, 2026-08-16) | yes |
| B26 | four non-null pages (cap exhausted): crawl record `terminal=cap_exhausted`, `last_cursor_sha256`, no 5th request, run continues and seals | `TestRunRecordsCapExhaustionAndStillCollectsOrderbookCalendarAndSeals` | yes (pre-amendment code returned HOLD; test failed with that message) | yes |
| B27 | crawl record write error → HOLD + taint, no orderbook/calendar, no manifest | `TestRunHoldsAndTaintsWhenCandleCrawlRecordWriteFails` | yes (mutant "swallow writeJSON error" killed, 2026-08-16) | yes |
| B28 | live check after crawl record fails | not-applicable: redundant with the write-internal post-write live guard (`receipt_unix.go` `write` rechecks `live` after every write), so deleting it is not observable through `run()` (adversary mutation h survives); kept as belt-and-braces | no | no |
| B29 | orderbook first error aborts | `TestRunStopsOnReaderErrorWithoutFallback` | yes | yes |
| B30 | calendar first error aborts | `TestRunStopsOnReaderErrorWithoutFallback` | yes | yes |
| B31 | live check before post identity fails | `TestRunHoldsWhenCancelledDuringPostReadFinalization` | yes | yes |
| B32 | post identity snapshot error | `TestRunHoldsWhenIdentityDriftsAfterNetwork` | yes | yes |
| B33 | pre ≠ post identity drift | `TestRunHoldsWhenIdentityDriftsAfterNetwork` | yes | yes |
| B34 | live check after post identity fails | `TestRunHoldsWhenClockExceedsOverallBudgetAfterFinalResponse` | yes | yes |
| B35 | seal error (unexpected entry) | `TestRunHoldsWhenUnexpectedReceiptEntryAppearsBeforeSeal` | yes | yes |
| B36 | live check after seal fails | `TestRunHoldsWhenCancelledDuringPostReadFinalization` | yes | yes |
| B37 | sealed verification error | `TestSealedReceiptDetectsPostSealModeDowngrade` | yes | yes |
| B38 | live check before success fails | `TestRunHoldsWhenClockExceedsOverallBudgetAfterFinalResponse` | yes | yes |
