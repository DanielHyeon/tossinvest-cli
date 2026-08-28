# Branch Test Map: `ProductionBatchAuthorityForTest`

- Source: `internal/strategyproposal/production_testseam.go`; file SHA-256 `854454d6d04e8527260f0f6148ac72660dd5871ba1667917f4bf5048aff4156b`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- No individually-measured test entered any arm of this function; see the per-branch rows for what the package and engine suites did enter.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | range at 9:2 | arm entered 21x (engine suite) |
| B2 | if at 10:3 | arm never entered: count 0 in every profile measured for this function |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
