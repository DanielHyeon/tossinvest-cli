# Function Logic Map: `TestPositionManagementDoesNotOfferReadoptForIneligibleHolding`

- Source: `internal/console/position_policy_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestPositionManagementDoesNotOfferReadoptForIneligibleHolding(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 294 | `if strings.Contains(page, "새 generation 재편입") {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionManagementDoesNotOfferReadoptForIneligibleHolding |
| B2 | `if` at line 297 | `if !strings.Contains(page, "관리 외(운영자 해제)") \|\| !strings.Contains(page, "OPERATOR_RELEASED") {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionManagementDoesNotOfferReadoptForIneligibleHolding |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `managedPolicyState` | execute the explicit dependency at line 285 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `policyHarness` | execute the explicit dependency at line 291 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.authenticate` | execute the explicit dependency at line 292 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `h.page` | execute the explicit dependency at line 293 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Contains` | execute the explicit dependency at line 294 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 295 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 8 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
