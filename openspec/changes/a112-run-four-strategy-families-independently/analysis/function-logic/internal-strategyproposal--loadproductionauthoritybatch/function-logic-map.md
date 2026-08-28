# Function Logic Map: `LoadProductionAuthorityBatch`

- Source: `internal/strategyproposal/production.go` (229-307)
- Function: `LoadProductionAuthorityBatch` in package `strategyproposal`
- Signature: `LoadProductionAuthorityBatch(params=4, results=2)`
- File SHA-256: `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 18.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Reads the signed manifest and evidence, routes with `RouteSet`, replays the snapshot, builds a lane input and composes a proposal — once per manifest scope. L3 changed three things: `Route` became `RouteSet`, eligibility became an exact-identity search over the candidate set (`routeSetAdmitsScope`), and the batch key became (symbol, laneID) so one symbol can carry several families.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **4x** under the package suite.

Exact AST return positions: 237:3, 240:3, 244:3, 248:3, 252:3, 264:3, 271:3, 306:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 233:2 | arm entered 4x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B2 | if | 239:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if | 243:2 | arm never entered: count 0 in every profile measured for this function |
| B4 | if | 247:2 | arm never entered: count 0 in every profile measured for this function |
| B5 | if | 251:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | range | 257:2 | arm entered 3x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B7 | if | 258:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if | 263:2 | arm never entered: count 0 in every profile measured for this function |
| B9 | if | 266:2 | arm never entered: count 0 in every profile measured for this function |
| B10 | if | 270:2 | arm never entered: count 0 in every profile measured for this function |
| B11 | range | 274:2 | arm entered 3x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B12 | if | 276:3 | arm never entered: count 0 in every profile measured for this function |
| B13 | if | 282:3 | arm never entered: count 0 in every profile measured for this function |
| B14 | if | 285:3 | arm never entered: count 0 in every profile measured for this function |
| B15 | if | 289:3 | arm entered 2x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B16 | if | 293:3 | arm never entered: count 0 in every profile measured for this function |
| B17 | if | 297:3 | arm never entered: count 0 in every profile measured for this function |
| B18 | if | 301:3 | arm never entered: count 0 in every profile measured for this function |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalConfig` | 230:11 |
| `productionOwnerUID` | 231:20 |
| `ProductionFileName` | 232:10 |
| `filepath.IsAbs` | 233:46 |
| `filepath.IsAbs` | 233:83 |
| `filepath.IsAbs` | 233:123 |
| `config.ObservedAt.IsZero` | 234:30 |
| `digestValid` | 234:61 |
| `identity` | 234:100 |
| `len` | 234:133 |
| `identity` | 235:4 |
| `identity` | 235:45 |
| `identity` | 235:83 |
| `identity` | 235:123 |
| `identity` | 236:4 |
| `identity` | 236:48 |
| `len` | 236:87 |
| `len` | 236:108 |
| `ctx.Err` | 239:12 |
| `readProductionFile` | 242:15 |
| `filepath.Join` | 242:34 |
| `digest` | 243:19 |
| `decodeManifest` | 246:19 |
| `verifyManifest` | 247:20 |
| `strategyevidence.OpenReadOnly` | 250:24 |
| `marketclock.NewFake` | 250:118 |
| `evidenceStore.Close` | 254:8 |
| `strategyevidence.NewDormantSnapshotReadPort` | 255:10 |
| `isWeeklyLane` | 258:6 |
| `journal.OpenReadOnly` | 259:21 |
| `journalRO.Close` | 267:9 |
| `canonicalTargets` | 269:24 |
| `make` | 273:12 |
| `len` | 273:49 |
| `strategyrouter.RouteSet` | 281:13 |
| `routed.Valid` | 282:52 |
| `target.Approved.CandidateLifeID` | 282:70 |
| `routeSetAdmitsScope` | 285:7 |
| `port.Replay` | 292:20 |
| `buildLaneInput` | 296:24 |

(3 further call sites omitted; `ast.json` carries all 43.)

## State mutations and fallbacks

- AST assignments: 19. Defers: 2. Goroutine statements: 0.
- Fills one local map and returns it sealed with the manifest digest. Opens the journal read-only and closes it; no write path exists in this function.

## Safety conclusion

- Every failure `continue`s rather than substituting a value, so a scope that cannot be composed simply produces no proposal. It never selects between families — selection is the coordinator's, which is the whole point of task 4.3.1.
