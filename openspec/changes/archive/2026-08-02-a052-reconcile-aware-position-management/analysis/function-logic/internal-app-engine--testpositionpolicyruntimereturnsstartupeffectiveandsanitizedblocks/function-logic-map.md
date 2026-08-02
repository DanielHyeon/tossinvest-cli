# Function Logic Map: `TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks`

- Source: `internal/app/engine/position_policy_command_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 36 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |
| B2 | `if` at line 39 | `if !got.EffectiveKnown \|\| !got.Effective.Enabled \|\| got.Effective.DefaultStopPct != .03 \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |
| B3 | `if` at line 45 | `if block.Scope != positionpolicy.ScopeAccount \|\| block.Reason != "RECONCILE_PERMANENT" \|\|` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` | execute the explicit dependency at line 25 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `service.Runtime` | execute the explicit dependency at line 35 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `context.Background` | execute the explicit dependency at line 35 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 37 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `len` | execute the explicit dependency at line 40 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatalf` | execute the explicit dependency at line 42 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `block.StartedAt.Equal` | execute the explicit dependency at line 46 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 4 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
