# Function Logic Map: `LoadProductionAuthorityBatch`

- Source: `internal/strategyproposal/production.go` (237-315)
- Function: `LoadProductionAuthorityBatch` in package `strategyproposal`
- Signature: `LoadProductionAuthorityBatch(params=4, results=2)`
- File SHA-256: `9fae1db65477dfe421a1e96e3437ff2909cc8439c1b987029a534d9aded9db94`
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

Exact AST return positions: 245:3, 248:3, 252:3, 256:3, 260:3, 272:3, 279:3, 314:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 241:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | if | 247:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B3 | if | 251:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B4 | if | 255:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 259:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B6 | range | 265:2 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B7 | if | 266:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 271:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B9 | if | 274:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B10 | if | 278:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | range | 282:2 | arm not entered (untagged proposal suite); arm entered 3x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B12 | if | 284:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if | 290:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | if | 293:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B15 | if | 297:3 | arm not entered (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B16 | if | 301:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B17 | if | 305:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B18 | if | 309:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalConfig` | 238:11 |
| `productionOwnerUID` | 239:20 |
| `ProductionFileName` | 240:10 |
| `filepath.IsAbs` | 241:46 |
| `filepath.IsAbs` | 241:83 |
| `filepath.IsAbs` | 241:123 |
| `config.ObservedAt.IsZero` | 242:30 |
| `digestValid` | 242:61 |
| `identity` | 242:100 |
| `len` | 242:133 |
| `identity` | 243:4 |
| `identity` | 243:45 |
| `identity` | 243:83 |
| `identity` | 243:123 |
| `identity` | 244:4 |
| `identity` | 244:48 |
| `len` | 244:87 |
| `len` | 244:108 |
| `ctx.Err` | 247:12 |
| `readProductionFile` | 250:15 |
| `filepath.Join` | 250:34 |
| `digest` | 251:19 |
| `decodeManifest` | 254:19 |
| `verifyManifest` | 255:20 |
| `strategyevidence.OpenReadOnly` | 258:24 |
| `marketclock.NewFake` | 258:118 |
| `evidenceStore.Close` | 262:8 |
| `strategyevidence.NewDormantSnapshotReadPort` | 263:10 |
| `isWeeklyLane` | 266:6 |
| `journal.OpenReadOnly` | 267:21 |
| `journalRO.Close` | 275:9 |
| `canonicalTargets` | 277:24 |
| `make` | 281:12 |
| `len` | 281:49 |
| `strategyrouter.RouteSet` | 289:13 |
| `routed.Valid` | 290:52 |
| `target.Approved.CandidateLifeID` | 290:70 |
| `routeSetAdmitsScope` | 293:7 |
| `port.Replay` | 300:20 |
| `buildLaneInput` | 304:24 |
| `strategyflow.Propose` | 308:15 |
| `proposal.ValidProposal` | 309:7 |
| `batchKey` | 312:10 |

## State mutations and fallbacks

- AST assignments: 19. Defers: 2. Goroutine statements: 0.

## Safety conclusion

The widened admission is measured and judged in review.md decision 53: with an active owner the eligible set is exactly one decision, so no second family can enter an owned symbol; with no owner and every lane OFF the set is empty and `RouteSet` refuses. The residual — an un-arbitrated singleton — is task 5.4's coordinator.
