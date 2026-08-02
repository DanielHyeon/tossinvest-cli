# Function Logic Map: `applyManagementProjection`

- Source: `cmd/tossctl/httpapi_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `applyManagementProjection(params=5, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | branchless happy path | execute `applyManagementProjection(params=5, results=0)` | deterministic effects only | normal result | go test ./cmd/tossctl |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicy.ProjectManagement` | explicit dependency at line 192 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `httpapi.AdoptionStatus` | explicit dependency at line 196 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `httpapi.AdoptionReason` | explicit dependency at line 199 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strings.ToLower` | explicit dependency at line 203 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strings.ReplaceAll` | explicit dependency at line 203 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `string` | explicit dependency at line 203 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `reconcileBlockFrom` | explicit dependency at line 204 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 10 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
