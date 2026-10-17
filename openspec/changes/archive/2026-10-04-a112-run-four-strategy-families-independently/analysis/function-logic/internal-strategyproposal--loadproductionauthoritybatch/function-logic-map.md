# Function Logic Map: `LoadProductionAuthorityBatch`

- Source: `internal/strategyproposal/production.go` (307-418)
- Function: `LoadProductionAuthorityBatch` in package `strategyproposal`
- Signature: `LoadProductionAuthorityBatch(params=4, results=2)`
- File SHA-256: `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 21.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Rebuilds sealed proposals from one signed manifest, immutable evidence and the sealed route request. Since task 4.3.1 the admission test is membership in the eligible set (`routeSetAdmitsScope`) rather than equality with a single raw-score winner, and the result map is keyed by (symbol, lane) so two families for one symbol cannot silently overwrite each other.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was not executed (untagged proposal suite); executed 4x (tagged proposal suite); not executed (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 315:3, 318:3, 322:3, 326:3, 330:3, 342:3, 349:3, 362:4, 416:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 311:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | if | 317:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B3 | if | 321:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B4 | if | 325:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 329:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B6 | range | 335:2 | arm not entered (untagged proposal suite); arm entered 9x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestAHealthyBatchCarriesNoFault`, `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault`, `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault`, `TestAnEvaluationRefusalIsAbsenceNotFault`, `TestAnUnusableRouteAuthorityIsAFault`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B7 | if | 336:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 341:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B9 | if | 344:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B10 | if | 348:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | if | 361:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B12 | range | 366:2 | arm not entered (untagged proposal suite); arm entered 9x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestAHealthyBatchCarriesNoFault`, `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault`, `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault`, `TestAnEvaluationRefusalIsAbsenceNotFault`, `TestAnUnusableRouteAuthorityIsAFault`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B13 | if | 368:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | if | 375:3 | arm not entered (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestAnUnusableRouteAuthorityIsAFault` |
| B15 | if | 380:3 | arm not entered (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault` |
| B16 | if | 384:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B17 | if | 389:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B18 | if | 393:3 | arm not entered (untagged proposal suite); arm entered 1x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault` |
| B19 | if | 398:3 | arm not entered (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault`, `TestAnEvaluationRefusalIsAbsenceNotFault` |
| B20 | if | 406:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B21 | if | 409:4 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalConfig` | 308:11 |
| `productionOwnerUID` | 309:20 |
| `ProductionFileName` | 310:10 |
| `filepath.IsAbs` | 311:46 |
| `filepath.IsAbs` | 311:83 |
| `filepath.IsAbs` | 311:123 |
| `config.ObservedAt.IsZero` | 312:30 |
| `digestValid` | 312:61 |
| `identity` | 312:100 |
| `len` | 312:133 |
| `identity` | 313:4 |
| `identity` | 313:45 |
| `identity` | 313:83 |
| `identity` | 313:123 |
| `identity` | 314:4 |
| `identity` | 314:48 |
| `len` | 314:87 |
| `len` | 314:108 |
| `ctx.Err` | 317:12 |
| `readProductionFile` | 320:15 |
| `filepath.Join` | 320:34 |
| `digest` | 321:19 |
| `decodeManifest` | 324:19 |
| `verifyManifest` | 325:20 |
| `strategyevidence.OpenReadOnly` | 328:24 |
| `marketclock.NewFake` | 328:118 |
| `evidenceStore.Close` | 332:8 |
| `strategyevidence.NewDormantSnapshotReadPort` | 333:10 |
| `isWeeklyLane` | 336:6 |
| `journal.OpenReadOnly` | 337:21 |
| `journalRO.Close` | 345:9 |
| `canonicalTargets` | 347:24 |
| `make` | 351:12 |
| `len` | 351:49 |
| `strategyrouter.RouteSet` | 374:13 |
| `routed.Valid` | 375:52 |
| `fault` | 377:4 |
| `target.Approved.CandidateLifeID` | 380:6 |
| `routeSetAdmitsScope` | 384:7 |
| `port.Replay` | 392:20 |
| `fault` | 394:4 |
| `buildLaneInput` | 397:24 |
| `strategyflow.Propose` | 405:15 |
| `proposal.ValidProposal` | 406:7 |
| `fault` | 410:5 |
| `batchKey` | 414:10 |

## State mutations and fallbacks

- AST assignments: 22. Defers: 2. Goroutine statements: 0.

## Safety conclusion

The widened admission is measured and judged in review.md decision 53: with an active owner the eligible set is exactly one decision, so no second family can enter an owned symbol; with no owner and every lane OFF the set is empty and `RouteSet` refuses. The residual — an un-arbitrated singleton — is task 5.4's coordinator.
