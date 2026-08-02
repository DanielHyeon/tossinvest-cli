# Function Logic Map: `TestAnAdoptedHoldingRendersAsManagedWithItsBasis`

- Source: `internal/console/portfolio_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestAnAdoptedHoldingRendersAsManagedWithItsBasis(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 874 | `if !strings.Contains(page, "편입 기록") {` | local/read-model state only; see AST assignments | continues through the function contract | TestAnAdoptedHoldingRendersAsManagedWithItsBasis |
| B2 | `if` at line 877 | `if !strings.Contains(page, "진입 결정") {` | local/read-model state only; see AST assignments | continues through the function contract | TestAnAdoptedHoldingRendersAsManagedWithItsBasis |
| B3 | `if` at line 882 | `if strings.Contains(row, "관리 외(미편입)") {` | local/read-model state only; see AST assignments | continues through the function contract | TestAnAdoptedHoldingRendersAsManagedWithItsBasis |
| B4 | `if` at line 885 | `if !strings.Contains(row, "엔진 관리") {` | local/read-model state only; see AST assignments | continues through the function contract | TestAnAdoptedHoldingRendersAsManagedWithItsBasis |
| B5 | `if` at line 890 | `if !strings.Contains(row, "원장 기록 · 실효 미확인") \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestAnAdoptedHoldingRendersAsManagedWithItsBasis |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness` | execute the explicit dependency at line 869 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `seedEngineJournal` | execute the explicit dependency at line 870 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.authenticate` | execute the explicit dependency at line 871 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.page` | execute the explicit dependency at line 873 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Contains` | execute the explicit dependency at line 874 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Error` | execute the explicit dependency at line 875 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rowFor` | execute the explicit dependency at line 881 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Errorf` | execute the explicit dependency at line 883 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 3 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
