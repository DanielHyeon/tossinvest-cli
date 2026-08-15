# Risk Pattern Report: `TestProductionAPIExportsNoAuthorityMintingFunction`

Source: `internal/protectionlifecycle/external_api_test.go`.

This is a structural boundary guard. A new exported lifecycle function can make the worker's authority available outside the runtime composition and bypass the intended gate/recovery topology. The guard must remain green with no exception for A100.
