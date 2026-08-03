# Function Logic Map: `fillMarketFromCurrency`

Source: `internal/filldetect/payload.go`
Function: `fillMarketFromCurrency`
Signature: `fillMarketFromCurrency(params=1, results=1)`
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`
Revision: `current`

## Inputs and invariants

- Inputs are `fillMarketFromCurrency(params=1, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/filldetect/payload.go:134 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | case | internal/filldetect/payload.go:135 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | case | internal/filldetect/payload.go:137 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | case | internal/filldetect/payload.go:139 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `strings.ToUpper`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 3; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
