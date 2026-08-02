# Function Logic Map: `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`

- Source: `internal/console/portfolio_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range` at line 284 | `for _, want := range []string{` | local/read-model state only; see AST assignments | continues through the function contract | TestThePositionsScreenShowsTheExitLineOfAManagedPosition |
| B2 | `if` at line 290 | `if !strings.Contains(page, want) {` | local/read-model state only; see AST assignments | continues through the function contract | TestThePositionsScreenShowsTheExitLineOfAManagedPosition |
| B3 | `range` at line 295 | `for _, staleRaw := range []string{"HALF_RISK", "intent-77"} {` | local/read-model state only; see AST assignments | continues through the function contract | TestThePositionsScreenShowsTheExitLineOfAManagedPosition |
| B4 | `if` at line 296 | `if strings.Contains(row, staleRaw) {` | local/read-model state only; see AST assignments | continues through the function contract | TestThePositionsScreenShowsTheExitLineOfAManagedPosition |
| B5 | `range` at line 300 | `for _, evidence := range []string{"원장 기록 · 실효 미확인", "원장 기준선 <strong>69500</strong>",` | local/read-model state only; see AST assignments | continues through the function contract | TestThePositionsScreenShowsTheExitLineOfAManagedPosition |
| B6 | `if` at line 302 | `if !strings.Contains(row, evidence) {` | local/read-model state only; see AST assignments | continues through the function contract | TestThePositionsScreenShowsTheExitLineOfAManagedPosition |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness` | execute the explicit dependency at line 279 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `seedJournal` | execute the explicit dependency at line 280 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.authenticate` | execute the explicit dependency at line 281 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.page` | execute the explicit dependency at line 283 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Contains` | execute the explicit dependency at line 290 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Errorf` | execute the explicit dependency at line 291 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rowFor` | execute the explicit dependency at line 294 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 3 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
