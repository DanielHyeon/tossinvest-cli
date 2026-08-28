# Branch Test Map: `LoadProductionAuthorityBatch`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 233:2 | arm entered 4x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B2 | if at 239:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if at 243:2 | arm never entered: count 0 in every profile measured for this function |
| B4 | if at 247:2 | arm never entered: count 0 in every profile measured for this function |
| B5 | if at 251:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | range at 257:2 | arm entered 3x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B7 | if at 258:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if at 263:2 | arm never entered: count 0 in every profile measured for this function |
| B9 | if at 266:2 | arm never entered: count 0 in every profile measured for this function |
| B10 | if at 270:2 | arm never entered: count 0 in every profile measured for this function |
| B11 | range at 274:2 | arm entered 3x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B12 | if at 276:3 | arm never entered: count 0 in every profile measured for this function |
| B13 | if at 282:3 | arm never entered: count 0 in every profile measured for this function |
| B14 | if at 285:3 | arm never entered: count 0 in every profile measured for this function |
| B15 | if at 289:3 | arm entered 2x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B16 | if at 293:3 | arm never entered: count 0 in every profile measured for this function |
| B17 | if at 297:3 | arm never entered: count 0 in every profile measured for this function |
| B18 | if at 301:3 | arm never entered: count 0 in every profile measured for this function |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
