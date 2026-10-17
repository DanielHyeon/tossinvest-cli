# Branch Test Map: `canonicalDescriptor`

- Source: `internal/strategyflow/registry.go`; file SHA-256 `c7cfd15029a18c87f4de9ff2cb2730280cd1345a6d182b0eee687a11348cbdda`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | range at 114:2 | arm entered 146x (package suite); entered by `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` |
| B2 | if at 115:3 | arm entered 39x (package suite); entered by `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
