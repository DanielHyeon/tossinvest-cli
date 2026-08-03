# Function Logic Map: `TestPruneRemovesRecordsOlderThanTheRetention`

Source: `internal/journal/nonce_test.go`
Function: `TestPruneRemovesRecordsOlderThanTheRetention`
Signature: `TestPruneRemovesRecordsOlderThanTheRetention(params=1, results=0)`
Source SHA-256: `83fcf17c3cd3758fadd4f23e7f31e675b8e3a2df7d56d3cdd6e70b583e16f5e3`
Revision: `base`

## Inputs and invariants

- Inputs are `TestPruneRemovesRecordsOlderThanTheRetention(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce_test.go:234 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/nonce_test.go:241 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/nonce_test.go:244 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/nonce_test.go:250 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/nonce_test.go:253 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `openTestJournal`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `cancelDecision`: errors and state follow mapped branches.
- `boundAttempt`: errors and state follow mapped branches.
- `a.MarkDispatchStarted`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `testIssued`: errors and state follow mapped branches.
- `j.PruneSpentNonces`: errors and state follow mapped branches.
- `issued.Add`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `spentNonceCount`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 8; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
