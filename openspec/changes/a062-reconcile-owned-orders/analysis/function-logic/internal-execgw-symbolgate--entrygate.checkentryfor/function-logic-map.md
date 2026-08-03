# Function Logic Map: `EntryGate.CheckEntryFor`

Source: `internal/execgw/symbolgate.go`
Function: `EntryGate.CheckEntryFor`
Signature: `EntryGate.CheckEntryFor(params=2, results=1)`
Source SHA-256: `46d559bec6ca59e70876da33face5ae8dadce7d832c5c32e105dd2add8cc8617`
Revision: `current`

## Inputs and invariants

- Inputs are `EntryGate.CheckEntryFor(params=2, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/execgw/symbolgate.go:223 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/execgw/symbolgate.go:224 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/execgw/symbolgate.go:231 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/execgw/symbolgate.go:234 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | range | internal/execgw/symbolgate.go:243 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/execgw/symbolgate.go:244 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/execgw/symbolgate.go:248 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | range | internal/execgw/symbolgate.go:252 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/execgw/symbolgate.go:253 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/execgw/symbolgate.go:256 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `g.mu.Lock`: errors and state follow mapped branches.
- `g.mu.Unlock`: errors and state follow mapped branches.
- `refresh`: errors and state follow mapped branches.
- `err.Error`: errors and state follow mapped branches.
- `g.checkAccountEntry`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `symbolKey`: errors and state follow mapped branches.
- `strings.EqualFold`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 5; return points: 7; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
