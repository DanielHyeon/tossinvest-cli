# Function Logic Map: `TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder`

Source: `internal/app/engine/exit_e2e_test.go`
Function: `TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder`
Signature: `TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder(params=1, results=0)`
Source SHA-256: `8cc6877572d4602364bcaf443e33aff26036804732de2e6fc0ace8a60aecdc2d`
Revision: `current`

## Inputs and invariants

- Inputs are `TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/exit_e2e_test.go:550 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/exit_e2e_test.go:554 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/exit_e2e_test.go:559 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/app/engine/exit_e2e_test.go:567 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/app/engine/exit_e2e_test.go:571 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/app/engine/exit_e2e_test.go:574 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/app/engine/exit_e2e_test.go:583 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/app/engine/exit_e2e_test.go:590 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/app/engine/exit_e2e_test.go:593 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `newE2EBroker`: errors and state follow mapped branches.
- `isolate`: errors and state follow mapped branches.
- `writeEngineConfig`: errors and state follow mapped branches.
- `writeCredentials`: errors and state follow mapped branches.
- `openE2EStack`: errors and state follow mapped branches.
- `s.entry`: errors and state follow mapped branches.
- `broker.quote`: errors and state follow mapped branches.
- `s.observe`: errors and state follow mapped branches.
- `s.state`: errors and state follow mapped branches.
- `armed.Pending`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `s.lastBrokerOrder`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `broker.placed`: errors and state follow mapped branches.
- `s.engine.Close`: errors and state follow mapped branches.
- `restarted.state`: errors and state follow mapped branches.
- `restarted.engine.Journal.IntentAttempted`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `restarted.observe`: errors and state follow mapped branches.
- `restarted.fill`: errors and state follow mapped branches.
- `settled.Pending`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 11; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
