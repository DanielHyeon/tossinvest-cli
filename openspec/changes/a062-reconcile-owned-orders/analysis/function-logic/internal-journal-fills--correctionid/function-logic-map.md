# Function Logic Map: `correctionID`

Source: `internal/journal/fills.go`
Function: `correctionID`
Signature: `correctionID(params=1, results=1)`
Source SHA-256: `000918b94c8c3f776b611421c412e4604086fc4cbee2fd0e7c21fe0dd46454c0`
Revision: `current`

## Inputs and invariants

- Inputs are `correctionID(params=1, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/journal/fills.go:1212 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `sha256.New`: errors and state follow mapped branches.
- `fmt.Fprintf`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `hex.EncodeToString`: errors and state follow mapped branches.
- `h.Sum`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 1; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
