# Function Logic Map: `TestTheWholeExitPathEndToEnd`

Source: `internal/app/engine/exit_e2e_test.go`
Function: `TestTheWholeExitPathEndToEnd`
Signature: `TestTheWholeExitPathEndToEnd(params=1, results=0)`
Source SHA-256: `8cc6877572d4602364bcaf443e33aff26036804732de2e6fc0ace8a60aecdc2d`
Revision: `current`

## Inputs and invariants

- Inputs are `TestTheWholeExitPathEndToEnd(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/exit_e2e_test.go:390 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/exit_e2e_test.go:393 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/exit_e2e_test.go:400 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/app/engine/exit_e2e_test.go:403 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/app/engine/exit_e2e_test.go:411 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/app/engine/exit_e2e_test.go:414 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/app/engine/exit_e2e_test.go:421 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/app/engine/exit_e2e_test.go:425 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/app/engine/exit_e2e_test.go:428 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/app/engine/exit_e2e_test.go:431 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/app/engine/exit_e2e_test.go:435 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/app/engine/exit_e2e_test.go:445 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/app/engine/exit_e2e_test.go:448 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | if | internal/app/engine/exit_e2e_test.go:451 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | if | internal/app/engine/exit_e2e_test.go:459 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | if | internal/app/engine/exit_e2e_test.go:465 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/app/engine/exit_e2e_test.go:469 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | if | internal/app/engine/exit_e2e_test.go:472 | Preserve the condition, error propagation, and fail-closed behavior. |
| B19 | if | internal/app/engine/exit_e2e_test.go:481 | Preserve the condition, error propagation, and fail-closed behavior. |
| B20 | if | internal/app/engine/exit_e2e_test.go:485 | Preserve the condition, error propagation, and fail-closed behavior. |
| B21 | if | internal/app/engine/exit_e2e_test.go:488 | Preserve the condition, error propagation, and fail-closed behavior. |
| B22 | if | internal/app/engine/exit_e2e_test.go:493 | Preserve the condition, error propagation, and fail-closed behavior. |
| B23 | if | internal/app/engine/exit_e2e_test.go:496 | Preserve the condition, error propagation, and fail-closed behavior. |
| B24 | if | internal/app/engine/exit_e2e_test.go:499 | Preserve the condition, error propagation, and fail-closed behavior. |
| B25 | if | internal/app/engine/exit_e2e_test.go:506 | Preserve the condition, error propagation, and fail-closed behavior. |
| B26 | range | internal/app/engine/exit_e2e_test.go:510 | Preserve the condition, error propagation, and fail-closed behavior. |
| B27 | if | internal/app/engine/exit_e2e_test.go:511 | Preserve the condition, error propagation, and fail-closed behavior. |
| B28 | if | internal/app/engine/exit_e2e_test.go:523 | Preserve the condition, error propagation, and fail-closed behavior. |
| B29 | if | internal/app/engine/exit_e2e_test.go:530 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `newE2EStack`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `s.entry`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `s.engine.Journal.ExitState`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `s.broker.quote`: errors and state follow mapped branches.
- `s.observe`: errors and state follow mapped branches.
- `s.state`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `s.broker.placed`: errors and state follow mapped branches.
- `fmt.Sprint`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `s.lastBrokerOrder`: errors and state follow mapped branches.
- `s.fill`: errors and state follow mapped branches.
- `afterFill.Pending`: errors and state follow mapped branches.
- `s.position`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- `s.engine.Journal.OpenExitStates`: errors and state follow mapped branches.
- `s.engine.Journal.ExitEvents`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `strings.Join`: errors and state follow mapped branches.
- `s.broker.reads`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 23; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
