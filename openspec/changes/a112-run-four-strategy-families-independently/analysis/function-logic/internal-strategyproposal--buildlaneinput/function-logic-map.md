# Function Logic Map: `buildLaneInput`

- Source: `internal/strategyproposal/production.go` (350-466)
- Function: `buildLaneInput` in package `strategyproposal`
- Signature: `buildLaneInput(params=6, results=3)`
- File SHA-256: `43ebb628cdfef4f891b652e81dc71c677063d0ad4cbbc9d0d3bc3b3cdcb52236`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 21.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Builds one lane input from a signed scope and replayed evidence. The breakout families fail closed at the top (decision 49) because the pure core needs four derived per-bar/per-snapshot members that no evidence record stores and no producer computes.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was executed 3x (untagged proposal suite); executed 7x (tagged proposal suite); not executed (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 352:3, 358:3, 362:4, 377:5, 379:4, 384:4, 386:3, 390:4, 402:4, 416:5, 418:4, 422:4, 424:3, 427:3, 431:3, 457:4, 459:3, 463:3, 465:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 351:2 | arm entered 2x (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing`, `TestBuildLaneInputRefusesBreakoutLanesWithTheBreakoutReason` |
| B2 | if | 357:2 | arm entered 1x (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestBuildLaneInputRefusesNonBreakoutLanesForADifferentReason` |
| B3 | if | 360:2 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B4 | if | 361:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 373:3 | arm not entered (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B6 | if | 376:4 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B7 | if | 383:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 388:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B9 | if | 389:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B10 | if | 393:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | if | 397:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B12 | if | 401:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if | 413:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | if | 415:4 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B15 | if | 421:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B16 | if | 426:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B17 | if | 430:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B18 | if | 434:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B19 | if | 454:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B20 | if | 456:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B21 | if | 462:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `parseTime` | 354:23 |
| `parseTime` | 355:20 |
| `parseTime` | 356:16 |
| `int64` | 365:86 |
| `continuationlane.BuildProductionKRProposalAuthority` | 375:14 |
| `strategyflow.ContinuationKR` | 379:11 |
| `a.Request` | 379:39 |
| `continuationlane.BuildProductionUSProposalAuthority` | 382:13 |
| `strategyflow.ContinuationUS` | 386:10 |
| `a.Request` | 386:38 |
| `reversalStructure` | 400:21 |
| `time.Duration` | 409:22 |
| `reversallane.BuildProductionKRProposalAuthority` | 414:14 |
| `strategyflow.ReversalKR` | 418:11 |
| `a.Request` | 418:35 |
| `reversallane.BuildProductionUSProposalAuthority` | 420:13 |
| `strategyflow.ReversalUS` | 424:10 |
| `a.Request` | 424:34 |
| `isWeeklyLane` | 426:6 |
| `journalRO.WeeklyMarketReservation` | 429:22 |
| `string` | 429:79 |
| `weeklyvaluelane.BuildProductionKRProposalAuthority` | 455:13 |
| `strategyflow.WeeklyKR` | 459:10 |
| `a.Request` | 459:32 |
| `weeklyvaluelane.BuildProductionUSProposalAuthority` | 461:12 |
| `strategyflow.WeeklyUS` | 465:9 |
| `a.Request` | 465:31 |

## State mutations and fallbacks

- AST assignments: 23. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Unchanged by this lot; the bundle is refreshed because the file hash moved. The breakout refusal remains fail-closed: no proposal is emitted and nothing downstream sees a partially invented snapshot.
