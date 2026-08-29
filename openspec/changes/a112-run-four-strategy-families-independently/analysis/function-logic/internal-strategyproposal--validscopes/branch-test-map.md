# Branch Test Map: `validScopes`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `9fae1db65477dfe421a1e96e3437ff2909cc8439c1b987029a534d9aded9db94`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 538:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | range at 542:2 | arm not entered (untagged proposal suite); arm entered 10x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |
| B3 | if at 546:3 | arm not entered (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
