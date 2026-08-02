# Function Logic Map: `TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease`

- Source: `cmd/tossctl/httpapi_reader_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease(params=1, results=0)` | current Go signature | failure remains explicit/unknown; no effective state is invented |
| Runtime/persisted facts | only values supplied by owning engine/journal | a052 contract + current HEAD | missing or ambiguous facts fail closed |
| Safety boundary | read-only projection/test | TossOS invariants | no live order, toggle, or reconcile resolution |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 176 | `if managed \|\| released {` | local/projection assignment | continues through contract | TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease |
| B2 | `if` at line 183 | `if out.AdoptionStatus != "UNMANAGED" \|\| out.AdoptionReason != "NOT_SELECTED" {` | local/projection assignment | continues through contract | TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `lifecycleFlags` | explicit dependency at line 175 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `t.Fatalf` | explicit dependency at line 177 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicy.NewAdoptionSettings` | explicit dependency at line 181 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyManagementProjection` | explicit dependency at line 182 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 4 assignment(s), 0 return(s), and 0 goroutine launch(es).
- Missing or ambiguous operational truth remains unknown/unmanaged; no desired value substitutes for effective runtime state.
- No reconciliation resolution or live order authority is added.

## Safety conclusion

- Safe edit boundary: read/projection/test behavior only.
- High-risk impact: indirect; lifecycle and exit truth remain fail-closed.
