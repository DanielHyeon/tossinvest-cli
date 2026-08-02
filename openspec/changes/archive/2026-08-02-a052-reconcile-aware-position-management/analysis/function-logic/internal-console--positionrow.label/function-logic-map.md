# Function Logic Map: `positionRow.Label`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `positionRow.Label(params=0, results=1)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 350 | `if r.HasManagementProjection() {` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B2 | `switch` at line 353 | `switch {` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B3 | `case` at line 354 | `case r.Unknown():` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B4 | `case` at line 356 | `case !r.Managed() && r.Excluded:` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B5 | `case` at line 358 | `case !r.Managed() && r.Designated:` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B6 | `case` at line 360 | `case !r.Managed():` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B7 | `case` at line 362 | `case r.HasExit && r.Exit.Completed:` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B8 | `case` at line 364 | `case r.HasExit:` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |
| B9 | `case` at line 366 | `default:` | none beyond calls listed below | early return/error path nearby | go test ./internal/console |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.HasManagementProjection` | execute the explicit dependency at line 350 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `r.Unknown` | execute the explicit dependency at line 354 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `r.Managed` | execute the explicit dependency at line 356 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 0 assignment(s), 8 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
