# Branch Test Map: `inputFor`

- Source: `internal/strategyflow/flow_test.go`; file SHA-256 `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`. AST branch positions are authoritative.
- No coverage profile can cover a `_test.go` function. Each row states what the arm is, and the run that exercised it: `go test -count=1 -tags tossos_testseams ./internal/strategyflow/`.
- This function is itself the test. The run below is the evidence that it executes and passes.

| Branch | Anchor | Classification | Observed |
|---|---|---|---|
| B1 | switch at 265:2 | path arm — taken on the exercised path | exercised by the named run |
| B2 | case at 266:2 | path arm — taken on the exercised path | exercised by the named run |
| B3 | case at 268:2 | path arm — taken on the exercised path | exercised by the named run |
| B4 | case at 270:2 | path arm — taken on the exercised path | exercised by the named run |
| B5 | case at 272:2 | path arm — taken on the exercised path | exercised by the named run |
| B6 | case at 274:2 | path arm — taken on the exercised path | exercised by the named run |
| B7 | case at 276:2 | path arm — taken on the exercised path | exercised by the named run |
| B8 | case at 278:2 | path arm — taken on the exercised path | exercised by the named run |
| B9 | case at 280:2 | path arm — taken on the exercised path | exercised by the named run |
| B10 | case at 282:2 | path arm — taken on the exercised path | exercised by the named run |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
