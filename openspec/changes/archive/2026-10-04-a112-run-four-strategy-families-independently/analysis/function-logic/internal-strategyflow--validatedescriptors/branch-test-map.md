# Branch Test Map: `ValidateDescriptors`

- Source: `internal/strategyflow/registry.go`; file SHA-256 `c7cfd15029a18c87f4de9ff2cb2730280cd1345a6d182b0eee687a11348cbdda`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestPairedRegistryCoversAllFourFamiliesInBothMarkets`, `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout`, `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 29:2 | arm entered 1x (package suite); entered by `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |
| B2 | range at 33:2 | arm entered 40x (package suite); entered by `TestPairedRegistryCoversAllFourFamiliesInBothMarkets`, `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout`, `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |
| B3 | range at 37:2 | arm entered 26x (package suite); entered by `TestPairedRegistryCoversAllFourFamiliesInBothMarkets`, `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout`, `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |
| B4 | if at 39:3 | arm entered 3x (package suite); entered by `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
