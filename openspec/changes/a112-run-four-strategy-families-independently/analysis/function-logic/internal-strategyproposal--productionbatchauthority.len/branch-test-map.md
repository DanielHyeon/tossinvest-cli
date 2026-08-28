# Branch Test Map: `ProductionBatchAuthority.Len`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `6cc7474d631e24c1daee677743fdbcc942787e9ae6874ed318cd3550326803b3`. AST branch positions are authoritative.
- Rows carry measured counts. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- No individually-measured test entered any arm of this function; see the per-branch rows for what the package and engine suites did enter.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | body at 80 | not entered by any measured profile |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
