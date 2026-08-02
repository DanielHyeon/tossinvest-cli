# Function Logic Map: `TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct`

- Source: `internal/httpapi/router_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 180 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct |
| B2 | `if` at line 190 | `if !management.Desired.Enabled \|\| management.Desired.DefaultStopPct != 0.03 \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct |
| B3 | `if` at line 195 | `if management.EffectiveKnown \|\| management.Effective != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct |
| B4 | `if` at line 198 | `if !management.AutoEnabledDesired \|\| management.StopDesired != "3%" \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct |
| B5 | `if` at line 206 | `if management.AutoEnabledEffective \|\| management.StopEffective != "5%" \|\| management.Effective == nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Optimization` | execute the explicit dependency at line 179 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `context.Background` | execute the explicit dependency at line 179 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 181 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `OptimizationFrom` | execute the explicit dependency at line 188 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `reflect.DeepEqual` | execute the explicit dependency at line 191 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatalf` | execute the explicit dependency at line 193 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 8 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
