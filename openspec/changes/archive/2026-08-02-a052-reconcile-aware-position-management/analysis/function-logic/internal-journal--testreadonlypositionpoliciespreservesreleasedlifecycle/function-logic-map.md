# Function Logic Map: `TestReadOnlyPositionPoliciesPreservesReleasedLifecycle`

- Source: `internal/journal/position_policy_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestReadOnlyPositionPoliciesPreservesReleasedLifecycle(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 174 | `if _, err := j.ApplyPositionPolicy(context.Background(), policyRequest(positionpolicy.ActionRelease, 0)); err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestReadOnlyPositionPoliciesPreservesReleasedLifecycle |
| B2 | `if` at line 178 | `if err != nil \|\| len(states) != 1 \|\| states[0].PositionID != "p-policy" \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestReadOnlyPositionPoliciesPreservesReleasedLifecycle |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournal` | execute the explicit dependency at line 172 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `seedPolicyPosition` | execute the explicit dependency at line 173 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `j.ApplyPositionPolicy` | execute the explicit dependency at line 174 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `context.Background` | execute the explicit dependency at line 174 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `policyRequest` | execute the explicit dependency at line 174 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 175 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `PositionPolicies` | execute the explicit dependency at line 177 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `openTestReadOnly` | execute the explicit dependency at line 177 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `j.Path` | execute the explicit dependency at line 177 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `len` | execute the explicit dependency at line 178 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatalf` | execute the explicit dependency at line 180 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 3 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
