# Function Logic Map: `TestManagedPositionPrimaryRowShowsTheProtectionDecision`

- Source: `internal/console/trading_views_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestManagedPositionPrimaryRowShowsTheProtectionDecision(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 80 | `if start < 0 {` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B2 | `if` at line 84 | `if end < 0 {` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B3 | `range` at line 88 | `for _, want := range []string{` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B4 | `if` at line 98 | `if !strings.Contains(row, want) {` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B5 | `range` at line 102 | `for _, raw := range []string{"HALF_RISK", "intent-77"} {` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B6 | `if` at line 103 | `if strings.Contains(row, raw) {` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B7 | `if` at line 107 | `if !strings.Contains(row, "원장 기준선 <strong>69500</strong>") \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |
| B8 | `if` at line 112 | `if !strings.Contains(page, '<caption>보유 종목과 보호 상태</caption>') \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestManagedPositionPrimaryRowShowsTheProtectionDecision |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness` | execute the explicit dependency at line 74 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `seedJournal` | execute the explicit dependency at line 75 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.authenticate` | execute the explicit dependency at line 76 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.page` | execute the explicit dependency at line 78 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Index` | execute the explicit dependency at line 79 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 81 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Contains` | execute the explicit dependency at line 98 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Errorf` | execute the explicit dependency at line 99 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Error` | execute the explicit dependency at line 110 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 5 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
