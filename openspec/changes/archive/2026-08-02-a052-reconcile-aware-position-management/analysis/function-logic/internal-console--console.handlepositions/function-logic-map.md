# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `Console.handlePositions(params=2, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 49 | `if c.opts.PositionPolicies != nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B2 | `if` at line 55 | `if states, err := c.opts.PositionPolicies.List(r.Context()); err == nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B3 | `range` at line 57 | `for _, state := range states {` | local/projection assignment | continues through contract | go test ./internal/console |
| B4 | `if` at line 62 | `if c.opts.Settings != nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B5 | `if` at line 63 | `if block, _, err := c.opts.Settings.Load(); err == nil {` | local/projection assignment | continues through contract | go test ./internal/console |
| B6 | `range` at line 66 | `for i := range page.Snap.Rows {` | local/projection assignment | continues through contract | go test ./internal/console |
| B7 | `if` at line 72 | `if runtimeAttempted {` | local/projection assignment | continues through contract | go test ./internal/console |
| B8 | `range` at line 73 | `for i := range page.Snap.Rows {` | local/projection assignment | continues through contract | go test ./internal/console |
| B9 | `if` at line 77 | `if row.InJournal {` | local/projection assignment | continues through contract | go test ./internal/console |
| B10 | `if` at line 79 | `if !ok {` | local/projection assignment | continues through contract | go test ./internal/console |
| B11 | `else` at line 81 | `} else {` | local/projection assignment | continues through contract | go test ./internal/console |
| B12 | `if` at line 90 | `if row.Management.Block != nil {` | local/projection assignment | continues through contract | go test ./internal/console |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.positions` | explicit dependency at line 43 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.Context` | explicit dependency at line 43 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.opts.PositionPolicies.Runtime` | explicit dependency at line 54 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.opts.PositionPolicies.List` | explicit dependency at line 55 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `make` | explicit dependency at line 56 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `len` | explicit dependency at line 56 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strings.TrimSpace` | explicit dependency at line 58 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.opts.Settings.Load` | explicit dependency at line 63 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `block.Included` | explicit dependency at line 67 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `block.Excludes` | explicit dependency at line 68 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicy.ProjectManagement` | explicit dependency at line 86 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `row.Managed` | explicit dependency at line 88 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newReconcileBlockView` | explicit dependency at line 91 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.now` | explicit dependency at line 91 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `attachPositionExitLines` | explicit dependency at line 95 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `c.render` | explicit dependency at line 96 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 18 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
