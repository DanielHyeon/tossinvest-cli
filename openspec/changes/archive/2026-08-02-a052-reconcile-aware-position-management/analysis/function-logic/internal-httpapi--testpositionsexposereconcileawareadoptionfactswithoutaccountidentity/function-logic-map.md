# Function Logic Map: `TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity`

- Source: `internal/httpapi/router_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 126 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity |
| B2 | `range` at line 130 | `for _, want := range []string{` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity |
| B3 | `if` at line 137 | `if !strings.Contains(text, want) {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity |
| B4 | `range` at line 141 | `for _, forbidden := range []string{"accountRef", "capability", "token", "command"} {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity |
| B5 | `if` at line 142 | `if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` | execute the explicit dependency at line 114 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `json.Marshal` | execute the explicit dependency at line 125 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 127 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `string` | execute the explicit dependency at line 129 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Contains` | execute the explicit dependency at line 137 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Errorf` | execute the explicit dependency at line 138 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.ToLower` | execute the explicit dependency at line 142 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 4 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
