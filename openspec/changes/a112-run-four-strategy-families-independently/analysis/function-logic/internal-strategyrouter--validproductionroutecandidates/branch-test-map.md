# Branch Test Map: `validProductionRouteCandidates`

- Source: `internal/strategyrouter/production.go`; file SHA-256 `dbf4e5afdfefcc6210a870d5c5e1952d3531eb119181be452e704964759bbcd8`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyrouter/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityFailureIsMarketLocal`, `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 376:2 | arm entered 2x (package suite); entered by `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` |
| B2 | range at 380:2 | arm entered 164x (package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityFailureIsMarketLocal` |
| B3 | if at 382:3 | arm entered 164x (package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently`, `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`, `TestProductionRouteAuthorityFailureIsMarketLocal` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
