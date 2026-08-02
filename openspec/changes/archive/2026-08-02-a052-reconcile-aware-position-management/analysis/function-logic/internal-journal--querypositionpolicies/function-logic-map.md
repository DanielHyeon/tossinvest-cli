# Function Logic Map: `queryPositionPolicies`

- Source: `internal/journal/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `queryPositionPolicies(params=2, results=2)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 87 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle |
| B2 | `for` at line 92 | `for rows.Next() {` | local/read-model state only; see AST assignments | early return/error path nearby | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle |
| B3 | `if` at line 94 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle |
| B4 | `if` at line 99 | `if err := rows.Err(); err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration; TestReadOnlyPositionPoliciesPreservesReleasedLifecycle |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `queryer.QueryContext` | execute the explicit dependency at line 85 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `fmt.Errorf` | execute the explicit dependency at line 88 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rows.Close` | execute the explicit dependency at line 90 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rows.Next` | execute the explicit dependency at line 92 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `scanPositionPolicyScope` | execute the explicit dependency at line 93 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `append` | execute the explicit dependency at line 97 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rows.Err` | execute the explicit dependency at line 99 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 4 assignment(s), 4 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
