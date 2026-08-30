# Function Logic Map: `LoadProductionAuthorityBatch`

- Source: `internal/strategyproposal/production.go` (252-330)
- Function: `LoadProductionAuthorityBatch` in package `strategyproposal`
- Signature: `LoadProductionAuthorityBatch(params=4, results=2)`
- File SHA-256: `43ebb628cdfef4f891b652e81dc71c677063d0ad4cbbc9d0d3bc3b3cdcb52236`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 18.
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

Exact AST return positions: 260:3, 263:3, 267:3, 271:3, 275:3, 287:3, 294:3, 329:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 256:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | if | 262:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B3 | if | 266:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B4 | if | 270:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 274:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B6 | range | 280:2 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B7 | if | 281:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 286:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B9 | if | 289:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B10 | if | 293:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | range | 297:2 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B12 | if | 299:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if | 305:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | if | 308:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B15 | if | 312:3 | arm not entered (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B16 | if | 316:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B17 | if | 320:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B18 | if | 324:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalConfig` | 253:11 |
| `productionOwnerUID` | 254:20 |
| `ProductionFileName` | 255:10 |
| `filepath.IsAbs` | 256:46 |
| `filepath.IsAbs` | 256:83 |
| `filepath.IsAbs` | 256:123 |
| `config.ObservedAt.IsZero` | 257:30 |
| `digestValid` | 257:61 |
| `identity` | 257:100 |
| `len` | 257:133 |
| `identity` | 258:4 |
| `identity` | 258:45 |
| `identity` | 258:83 |
| `identity` | 258:123 |
| `identity` | 259:4 |
| `identity` | 259:48 |
| `len` | 259:87 |
| `len` | 259:108 |
| `ctx.Err` | 262:12 |
| `readProductionFile` | 265:15 |
| `filepath.Join` | 265:34 |
| `digest` | 266:19 |
| `decodeManifest` | 269:19 |
| `verifyManifest` | 270:20 |
| `strategyevidence.OpenReadOnly` | 273:24 |
| `marketclock.NewFake` | 273:118 |
| `evidenceStore.Close` | 277:8 |
| `strategyevidence.NewDormantSnapshotReadPort` | 278:10 |
| `isWeeklyLane` | 281:6 |
| `journal.OpenReadOnly` | 282:21 |
| `journalRO.Close` | 290:9 |
| `canonicalTargets` | 292:24 |
| `make` | 296:12 |
| `len` | 296:49 |
| `strategyrouter.RouteSet` | 304:13 |
| `routed.Valid` | 305:52 |
| `target.Approved.CandidateLifeID` | 305:70 |
| `routeSetAdmitsScope` | 308:7 |
| `port.Replay` | 315:20 |
| `buildLaneInput` | 319:24 |
| `strategyflow.Propose` | 323:15 |
| `proposal.ValidProposal` | 324:7 |
| `batchKey` | 327:10 |

## State mutations and fallbacks

- AST assignments: 19. Defers: 2. Goroutine statements: 0.

## Safety conclusion

The widened admission is measured and judged in review.md decision 53: with an active owner the eligible set is exactly one decision, so no second family can enter an owned symbol; with no owner and every lane OFF the set is empty and `RouteSet` refuses. The residual — an un-arbitrated singleton — is task 5.4's coordinator.
