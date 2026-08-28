# Branch Test Map: `buildLaneInput`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing`, `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 328:2 | arm entered 1x (package suite); entered by `TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing` |
| B2 | if at 334:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if at 337:2 | arm entered 3x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B4 | if at 338:3 | arm never entered: count 0 in every profile measured for this function |
| B5 | if at 350:3 | arm entered 1x (package suite); entered by `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots` |
| B6 | if at 353:4 | arm never entered: count 0 in every profile measured for this function |
| B7 | if at 360:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if at 365:2 | arm never entered: count 0 in every profile measured for this function |
| B9 | if at 366:3 | arm never entered: count 0 in every profile measured for this function |
| B10 | if at 370:3 | arm never entered: count 0 in every profile measured for this function |
| B11 | if at 374:3 | arm never entered: count 0 in every profile measured for this function |
| B12 | if at 378:3 | arm never entered: count 0 in every profile measured for this function |
| B13 | if at 390:3 | arm never entered: count 0 in every profile measured for this function |
| B14 | if at 392:4 | arm never entered: count 0 in every profile measured for this function |
| B15 | if at 398:3 | arm never entered: count 0 in every profile measured for this function |
| B16 | if at 403:2 | arm never entered: count 0 in every profile measured for this function |
| B17 | if at 407:2 | arm never entered: count 0 in every profile measured for this function |
| B18 | if at 411:2 | arm never entered: count 0 in every profile measured for this function |
| B19 | if at 431:2 | arm never entered: count 0 in every profile measured for this function |
| B20 | if at 433:3 | arm never entered: count 0 in every profile measured for this function |
| B21 | if at 439:2 | arm never entered: count 0 in every profile measured for this function |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
