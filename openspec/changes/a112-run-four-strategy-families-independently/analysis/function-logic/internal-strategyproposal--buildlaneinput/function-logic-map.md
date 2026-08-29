# Function Logic Map: `buildLaneInput`

- Source: `internal/strategyproposal/production.go` (335-451)
- Function: `buildLaneInput` in package `strategyproposal`
- Signature: `buildLaneInput(params=6, results=3)`
- File SHA-256: `9fae1db65477dfe421a1e96e3437ff2909cc8439c1b987029a534d9aded9db94`
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

Exact AST return positions: 337:3, 343:3, 347:4, 362:5, 364:4, 369:4, 371:3, 375:4, 387:4, 401:5, 403:4, 407:4, 409:3, 412:3, 416:3, 442:4, 444:3, 448:3, 450:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 336:2 | arm entered 2x (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing`, `TestBuildLaneInputRefusesBreakoutLanesWithTheBreakoutReason` |
| B2 | if | 342:2 | arm entered 1x (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestBuildLaneInputRefusesNonBreakoutLanesForADifferentReason` |
| B3 | if | 345:2 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B4 | if | 346:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 358:3 | arm not entered (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B6 | if | 361:4 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B7 | if | 368:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 373:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B9 | if | 374:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B10 | if | 378:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | if | 382:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B12 | if | 386:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if | 398:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | if | 400:4 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B15 | if | 406:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B16 | if | 411:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B17 | if | 415:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B18 | if | 419:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B19 | if | 439:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B20 | if | 441:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B21 | if | 447:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `parseTime` | 339:23 |
| `parseTime` | 340:20 |
| `parseTime` | 341:16 |
| `int64` | 350:86 |
| `continuationlane.BuildProductionKRProposalAuthority` | 360:14 |
| `strategyflow.ContinuationKR` | 364:11 |
| `a.Request` | 364:39 |
| `continuationlane.BuildProductionUSProposalAuthority` | 367:13 |
| `strategyflow.ContinuationUS` | 371:10 |
| `a.Request` | 371:38 |
| `reversalStructure` | 385:21 |
| `time.Duration` | 394:22 |
| `reversallane.BuildProductionKRProposalAuthority` | 399:14 |
| `strategyflow.ReversalKR` | 403:11 |
| `a.Request` | 403:35 |
| `reversallane.BuildProductionUSProposalAuthority` | 405:13 |
| `strategyflow.ReversalUS` | 409:10 |
| `a.Request` | 409:34 |
| `isWeeklyLane` | 411:6 |
| `journalRO.WeeklyMarketReservation` | 414:22 |
| `string` | 414:79 |
| `weeklyvaluelane.BuildProductionKRProposalAuthority` | 440:13 |
| `strategyflow.WeeklyKR` | 444:10 |
| `a.Request` | 444:32 |
| `weeklyvaluelane.BuildProductionUSProposalAuthority` | 446:12 |
| `strategyflow.WeeklyUS` | 450:9 |
| `a.Request` | 450:31 |

## State mutations and fallbacks

- AST assignments: 23. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Unchanged by this lot; the bundle is refreshed because the file hash moved. The breakout refusal remains fail-closed: no proposal is emitted and nothing downstream sees a partially invented snapshot.
