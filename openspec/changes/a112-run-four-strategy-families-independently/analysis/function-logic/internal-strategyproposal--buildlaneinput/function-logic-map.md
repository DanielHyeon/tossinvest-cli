# Function Logic Map: `buildLaneInput`

- Source: `internal/strategyproposal/production.go` (327-443)
- Function: `buildLaneInput` in package `strategyproposal`
- Signature: `buildLaneInput(params=6, results=3)`
- File SHA-256: `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 21.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Turns one manifest scope plus replayed evidence into a tagged `LaneInput`. L3 added the breakout arm at the top, which returns `ErrBreakoutEvidenceUnavailable` rather than composing anything (decision 49).

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **4x** under the package suite.

Exact AST return positions: 329:3, 335:3, 339:4, 354:5, 356:4, 361:4, 363:3, 367:4, 379:4, 393:5, 395:4, 399:4, 401:3, 404:3, 408:3, 434:4, 436:3, 440:3, 442:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 328:2 | arm entered 1x (package suite); entered by `TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing` |
| B2 | if | 334:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if | 337:2 | arm entered 3x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B4 | if | 338:3 | arm never entered: count 0 in every profile measured for this function |
| B5 | if | 350:3 | arm entered 1x (package suite); entered by `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B6 | if | 353:4 | arm never entered: count 0 in every profile measured for this function |
| B7 | if | 360:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if | 365:2 | arm never entered: count 0 in every profile measured for this function |
| B9 | if | 366:3 | arm never entered: count 0 in every profile measured for this function |
| B10 | if | 370:3 | arm never entered: count 0 in every profile measured for this function |
| B11 | if | 374:3 | arm never entered: count 0 in every profile measured for this function |
| B12 | if | 378:3 | arm never entered: count 0 in every profile measured for this function |
| B13 | if | 390:3 | arm never entered: count 0 in every profile measured for this function |
| B14 | if | 392:4 | arm never entered: count 0 in every profile measured for this function |
| B15 | if | 398:3 | arm never entered: count 0 in every profile measured for this function |
| B16 | if | 403:2 | arm never entered: count 0 in every profile measured for this function |
| B17 | if | 407:2 | arm never entered: count 0 in every profile measured for this function |
| B18 | if | 411:2 | arm never entered: count 0 in every profile measured for this function |
| B19 | if | 431:2 | arm never entered: count 0 in every profile measured for this function |
| B20 | if | 433:3 | arm never entered: count 0 in every profile measured for this function |
| B21 | if | 439:2 | arm never entered: count 0 in every profile measured for this function |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `parseTime` | 331:23 |
| `parseTime` | 332:20 |
| `parseTime` | 333:16 |
| `int64` | 342:86 |
| `continuationlane.BuildProductionKRProposalAuthority` | 352:14 |
| `strategyflow.ContinuationKR` | 356:11 |
| `a.Request` | 356:39 |
| `continuationlane.BuildProductionUSProposalAuthority` | 359:13 |
| `strategyflow.ContinuationUS` | 363:10 |
| `a.Request` | 363:38 |
| `reversalStructure` | 377:21 |
| `time.Duration` | 386:22 |
| `reversallane.BuildProductionKRProposalAuthority` | 391:14 |
| `strategyflow.ReversalKR` | 395:11 |
| `a.Request` | 395:35 |
| `reversallane.BuildProductionUSProposalAuthority` | 397:13 |
| `strategyflow.ReversalUS` | 401:10 |
| `a.Request` | 401:34 |
| `isWeeklyLane` | 403:6 |
| `journalRO.WeeklyMarketReservation` | 406:22 |
| `string` | 406:79 |
| `weeklyvaluelane.BuildProductionKRProposalAuthority` | 432:13 |
| `strategyflow.WeeklyKR` | 436:10 |
| `a.Request` | 436:32 |
| `weeklyvaluelane.BuildProductionUSProposalAuthority` | 438:12 |
| `strategyflow.WeeklyUS` | 442:9 |
| `a.Request` | 442:31 |

## State mutations and fallbacks

- AST assignments: 23. Defers: 0. Goroutine statements: 0.
- Reads the journal for the weekly reservation binding; no write.

## Safety conclusion

- Fails closed everywhere: every parse or build failure returns an error and no lane input. The breakout arm is the strongest form of that — it refuses rather than invent the four derived metrics no evidence record stores.
