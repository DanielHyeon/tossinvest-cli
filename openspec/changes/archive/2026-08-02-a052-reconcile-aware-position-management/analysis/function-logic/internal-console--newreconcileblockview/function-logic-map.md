# Function Logic Map: `newReconcileBlockView`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `newReconcileBlockView(params=2, results=1)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `switch` at line 138 | `switch block.Scope {` | local/projection assignment | continues through contract | go test ./internal/console |
| B2 | `case` at line 139 | `case positionpolicy.ScopeMarket:` | local/projection assignment | continues through contract | go test ./internal/console |
| B3 | `case` at line 141 | `case positionpolicy.ScopeSymbol:` | local/projection assignment | continues through contract | go test ./internal/console |
| B4 | `if` at line 146 | `if !block.StartedAt.IsZero() {` | local/projection assignment | continues through contract | go test ./internal/console |
| B5 | `if` at line 149 | `if d < 0 {` | local/projection assignment | continues through contract | go test ./internal/console |
| B6 | `switch` at line 152 | `switch {` | local/projection assignment | continues through contract | go test ./internal/console |
| B7 | `case` at line 153 | `case d < time.Minute:` | local/projection assignment | continues through contract | go test ./internal/console |
| B8 | `case` at line 155 | `case d < 24*time.Hour:` | local/projection assignment | early return/error nearby | go test ./internal/console |
| B9 | `case` at line 157 | `default:` | local/projection assignment | early return/error nearby | go test ./internal/console |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToUpper` | explicit dependency at line 140 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strings.TrimSpace` | explicit dependency at line 140 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `block.StartedAt.IsZero` | explicit dependency at line 146 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `Format` | explicit dependency at line 147 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `block.StartedAt.UTC` | explicit dependency at line 147 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `now.Sub` | explicit dependency at line 148 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `fmt.Sprintf` | explicit dependency at line 156 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `int` | explicit dependency at line 156 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `string` | explicit dependency at line 161 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 11 assignment(s), 1 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
