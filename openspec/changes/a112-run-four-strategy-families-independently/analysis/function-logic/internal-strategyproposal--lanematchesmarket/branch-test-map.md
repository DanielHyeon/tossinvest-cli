# Branch Test Map: `laneMatchesMarket`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 547:2 | arm entered 15x (package suite); entered by `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix` |
| B2 | if at 550:2 | arm entered 9x (package suite); entered by `TestLaneMatchesMarketCoversTheBreakoutFamily`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
