# Branch Test Map: `ProductionBatchAuthority.For`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- No individually-measured test entered any arm of this function; see the per-branch rows for what the package and engine suites did enter.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 102:2 | arm never entered: count 0 in every profile measured for this function |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
