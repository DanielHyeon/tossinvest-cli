# Function Logic Map: `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`

- Source: `internal/app/engine/adoption_include_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `for` at line 113 | `for i := 0; i < reconcile.DefaultMaxFailures; i++ {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile |
| B2 | `if` at line 118 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile |
| B3 | `if` at line 121 | `if i+1 == reconcile.DefaultMaxFailures && !out.Permanent {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile |
| B4 | `if` at line 126 | `if rejected := h.tracker.EntryAllowed("us", "AAPL"); rejected == nil \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile |
| B5 | `if` at line 133 | `if cycle.Folded != 1 \|\| cycle.Adopted != 0 \|\| cycle.Unmanaged != 0 \|\| h.prices.calls != 0 {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile |
| B6 | `if` at line 137 | `if p.Adopted() {` | local/read-model state only; see AST assignments | continues through the function contract | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDriverHarness` | execute the explicit dependency at line 110 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `includeOnly` | execute the explicit dependency at line 111 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.tracker.Observe` | execute the explicit dependency at line 114 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Context` | execute the explicit dependency at line 114 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 119 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatalf` | execute the explicit dependency at line 122 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.clk.Advance` | execute the explicit dependency at line 124 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.tracker.EntryAllowed` | execute the explicit dependency at line 126 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.holdsMarket` | execute the explicit dependency at line 130 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.cycle` | execute the explicit dependency at line 132 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.positionMarket` | execute the explicit dependency at line 136 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `p.Adopted` | execute the explicit dependency at line 137 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 8 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
