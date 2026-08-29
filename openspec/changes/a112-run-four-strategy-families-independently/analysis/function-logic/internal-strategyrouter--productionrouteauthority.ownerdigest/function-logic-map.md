# Function Logic Map: `ProductionRouteAuthority.OwnerDigest`

- Source: `internal/strategyrouter/production.go` (110-110)
- Function: `ProductionRouteAuthority.OwnerDigest` in package `strategyrouter`
- Signature: `ProductionRouteAuthority.OwnerDigest(params=0, results=1)`
- File SHA-256: `eafb36f41e2c07b85737692afa20fac968123481c812237f8678ad7a140bb520`
- Pinned revision: `base` — this function's body is **unchanged**. The frozen base
  `a8c3d067470fe9cd00523a7629ee93ee05de8e5c` is pinned because the diff only touches its neighbourhood: the three new
  seal accessors were inserted immediately after it, and a pure insertion after a
  function's last line counts as intersecting that function.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Returns the scalar owner-history digest of a sealed authority. One expression, no
parameters beyond the value receiver, one result.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body executed **2x** under the untagged package suite.

Exact AST return positions: 110:69.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 110:1 | arm entered 2x (untagged package suite); entered by `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| (no call expressions in this function) | — |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

- A branchless value accessor over an unexported string. It cannot mint, mutate or
  widen any authority, and this lot did not change one character of it.
