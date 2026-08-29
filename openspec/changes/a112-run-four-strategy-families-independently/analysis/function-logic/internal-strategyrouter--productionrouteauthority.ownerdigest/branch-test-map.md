# Branch Test Map: `ProductionRouteAuthority.OwnerDigest`

- Source: `internal/strategyrouter/production.go`; file SHA-256 `eafb36f41e2c07b85737692afa20fac968123481c812237f8678ad7a140bb520` at frozen base `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`. AST branch positions are authoritative.
- Rows carry measured counts. untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | body at 110:1 | arm entered 2x (untagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
