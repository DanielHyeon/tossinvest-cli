# Function Logic Map: `fakeLedger.Apply`

Source: `internal/filldetect/detect_test.go`
Function: `fakeLedger.Apply`
Signature: `fakeLedger.Apply(params=2, results=2)`
Source SHA-256: `7fe5825a894d212e278325c39d6b369d975ef46f006b913627daa8c7264e2e26`
Revision: `current`

## Inputs and invariants

- Inputs are `fakeLedger.Apply(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect_test.go:193 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/detect_test.go:209 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `l.mu.Lock`: errors and state follow mapped branches.
- `l.mu.Unlock`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `strings.ToLower`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `l.clk.Now`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 7; return points: 3; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
