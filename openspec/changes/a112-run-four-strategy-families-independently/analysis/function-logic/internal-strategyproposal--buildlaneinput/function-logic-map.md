# Function Logic Map: `buildLaneInput`

- Source: `internal/strategyproposal/production.go` (438-554)
- Function: `buildLaneInput` in package `strategyproposal`
- Signature: `buildLaneInput(params=6, results=3)`
- File SHA-256: `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`
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

Exact AST return positions: 440:3, 446:3, 450:4, 465:5, 467:4, 472:4, 474:3, 478:4, 490:4, 504:5, 506:4, 510:4, 512:3, 515:3, 519:3, 545:4, 547:3, 551:3, 553:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 439:2 | arm entered 2x (untagged proposal suite); arm entered 4x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing`, `TestBuildLaneInputRefusesBreakoutLanesWithTheBreakoutReason` |
| B2 | if | 445:2 | arm entered 1x (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestBuildLaneInputRefusesNonBreakoutLanesForADifferentReason` |
| B3 | if | 448:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B4 | if | 449:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 461:3 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestAHealthyBatchCarriesNoFault`, `TestAnEvaluationRefusalIsAbsenceNotFault`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B6 | if | 464:4 | arm not entered (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestAnEvaluationRefusalIsAbsenceNotFault` |
| B7 | if | 471:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 476:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B9 | if | 477:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B10 | if | 481:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | if | 485:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B12 | if | 489:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if | 501:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | if | 503:4 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B15 | if | 509:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B16 | if | 514:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B17 | if | 518:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B18 | if | 522:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B19 | if | 542:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B20 | if | 544:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B21 | if | 550:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `parseTime` | 442:23 |
| `parseTime` | 443:20 |
| `parseTime` | 444:16 |
| `int64` | 453:86 |
| `continuationlane.BuildProductionKRProposalAuthority` | 463:14 |
| `strategyflow.ContinuationKR` | 467:11 |
| `a.Request` | 467:39 |
| `continuationlane.BuildProductionUSProposalAuthority` | 470:13 |
| `strategyflow.ContinuationUS` | 474:10 |
| `a.Request` | 474:38 |
| `reversalStructure` | 488:21 |
| `time.Duration` | 497:22 |
| `reversallane.BuildProductionKRProposalAuthority` | 502:14 |
| `strategyflow.ReversalKR` | 506:11 |
| `a.Request` | 506:35 |
| `reversallane.BuildProductionUSProposalAuthority` | 508:13 |
| `strategyflow.ReversalUS` | 512:10 |
| `a.Request` | 512:34 |
| `isWeeklyLane` | 514:6 |
| `journalRO.WeeklyMarketReservation` | 517:22 |
| `string` | 517:79 |
| `weeklyvaluelane.BuildProductionKRProposalAuthority` | 543:13 |
| `strategyflow.WeeklyKR` | 547:10 |
| `a.Request` | 547:32 |
| `weeklyvaluelane.BuildProductionUSProposalAuthority` | 549:12 |
| `strategyflow.WeeklyUS` | 553:9 |
| `a.Request` | 553:31 |

## State mutations and fallbacks

- AST assignments: 23. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Unchanged by this lot; the bundle is refreshed because the file hash moved. The breakout refusal remains fail-closed: no proposal is emitted and nothing downstream sees a partially invented snapshot.
