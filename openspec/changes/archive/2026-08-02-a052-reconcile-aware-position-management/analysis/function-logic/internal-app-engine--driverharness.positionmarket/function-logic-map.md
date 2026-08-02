# Function Logic Map: `driverHarness.positionMarket`

- Source: `internal/app/engine/reconcileloop_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `driverHarness.positionMarket(params=2, results=1)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 224 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | driverHarness.positionMarket |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `h.t.Helper` | execute the explicit dependency at line 222 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.journal.CurrentPosition` | execute the explicit dependency at line 223 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `context.Background` | execute the explicit dependency at line 223 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.t.Fatalf` | execute the explicit dependency at line 225 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 1 assignment(s), 1 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
