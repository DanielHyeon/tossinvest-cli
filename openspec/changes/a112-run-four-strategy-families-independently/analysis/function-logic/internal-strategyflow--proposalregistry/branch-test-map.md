# Branch Test Map: `proposalRegistry`

- Source: `internal/strategyflow/adapters.go`; file SHA-256 `0f6b4e682e89e6d24c4c3686a5a1ad5ea1f0825e904236ea892b5905029065b6`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- No individually-measured test entered any arm of this function; see the per-branch rows for what the package and engine suites did enter.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | body at 26 | entered 7x |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
