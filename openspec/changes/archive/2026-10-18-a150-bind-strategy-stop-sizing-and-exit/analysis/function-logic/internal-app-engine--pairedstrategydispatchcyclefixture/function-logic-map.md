# Function Logic Map: `pairedStrategyDispatchCycleFixture`

Source: `internal/app/engine/strategy_dispatch_cycle_test.go`; test-only function. Current AST: `ast.json`.

## Inputs and invariants

The fixture uses the same account, market, symbol and admission values as before; new supported first-leg admission now supplies a valid v4 policy snapshot. The v2 compatibility test additionally checks exact historical bytes. No production authority bypass is added.

## Branches and early returns

Existing checks and error branches remain in the source; fixture creation and v4 verification errors fail the test. The function-specific AST enumerates branches.

## Calls and live bindings

Calls use existing test fixtures and real validation. Tests use local journals and fake network interfaces.

## State mutations and fallbacks

Test setup writes temporary journal state only; no live operations.

## Safety conclusion

Test-only compatibility adjustment; original assertions remain. Focused receipts and full gate are recorded separately.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
