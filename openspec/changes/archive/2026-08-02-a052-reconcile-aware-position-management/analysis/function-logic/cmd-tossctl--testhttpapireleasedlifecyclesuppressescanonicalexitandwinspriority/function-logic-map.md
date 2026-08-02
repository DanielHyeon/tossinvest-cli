# Function Logic Map: `TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority`

- Source: `cmd/tossctl/httpapi_reader_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 156 | `if out.ExitLine.CurrentProtection != "195" {` | local/projection assignment | continues through contract | TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority |
| B2 | `if` at line 166 | `if out.AdoptionStatus != "UNMANAGED" \|\| out.AdoptionReason != "OPERATOR_RELEASED" \|\| out.Eligible \|\|` | local/projection assignment | continues through contract | TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` | explicit dependency at line 143 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `now.Format` | explicit dependency at line 151 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyStoredPosition` | explicit dependency at line 155 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `t.Fatalf` | explicit dependency at line 157 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `lifecycleFlags` | explicit dependency at line 160 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyReleasedExitTruth` | explicit dependency at line 161 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicy.NewAdoptionSettings` | explicit dependency at line 163 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyManagementProjection` | explicit dependency at line 165 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 6 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
