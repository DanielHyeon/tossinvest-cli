# Branch Test Map: `evaluateWith`

- Source: `internal/strategyflow/flow.go`; file SHA-256 `c4e9738af8202122e48460436ce5cf7717b8ec8af4495b1b581171114dfe06ce`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage`, `TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms`, `TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage`, `TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence`, `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped`, `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 29:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | if at 34:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if at 41:2 | arm entered 1x (package suite); entered by `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped` |
| B4 | if at 51:2 | arm entered 1x (package suite); entered by `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped` |
| B5 | if at 56:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | if at 69:2 | arm entered 1x (package suite); entered by `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed` |
| B7 | if at 75:2 | arm never entered: count 0 in every profile measured for this function |
| B8 | if at 82:2 | arm entered 1x (package suite); entered by `TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence` |
| B9 | if at 88:2 | arm entered 1x (package suite); entered by `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed` |
| B10 | if at 93:2 | arm entered 1x (package suite); entered by `TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage` |
| B11 | if at 105:2 | arm entered 35x (package suite); entered by `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` |
| B12 | if at 114:2 | arm entered 1x (package suite); entered by `TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
