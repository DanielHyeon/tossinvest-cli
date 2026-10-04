# Function Logic Map: `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped`

- Source: `internal/strategyflow/flow_test.go` (88-110)
- Function: `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped` in package `strategyflow`
- Signature: `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped(params=1, results=0)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Pins that a router refusal never reaches a lane and that an unsupported binding gets a typed code rather than a generic one.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 93:88, 96:3, 105:3.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 98:2 | no coverage block maps to this position |
| B2 | if | 107:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Descriptors` | 89:16 |
| `approvedFixture` | 90:14 |
| `strategyrouter.NewOwnerKey` | 91:12 |
| `approved.Symbol` | 91:66 |
| `registryForTest` | 93:14 |
| `evaluateWith` | 95:13 |
| `inputFor` | 95:107 |
| `string` | 98:60 |
| `t.Fatalf` | 99:3 |
| `evaluateWith` | 102:17 |
| `inputFor` | 102:111 |
| `routeDecision` | 103:15 |
| `t.Fatalf` | 108:3 |

## State mutations and fallbacks

- AST assignments: 10. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
