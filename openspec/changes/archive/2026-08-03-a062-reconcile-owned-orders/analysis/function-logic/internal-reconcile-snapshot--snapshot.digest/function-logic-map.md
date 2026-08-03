# Function Logic Map: `Snapshot.Digest`

Source: `internal/reconcile/snapshot.go`
Function: `Snapshot.Digest`
Signature: `Snapshot.Digest(params=0, results=1)`
Source SHA-256: `827f148d49ae878bd1acb64327dbd5545cebe9a576e130305255e06861e1b8e3`
Revision: `current`

## Inputs and invariants

- Inputs are `Snapshot.Digest(params=0, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/reconcile/snapshot.go:169 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | range | internal/reconcile/snapshot.go:180 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | range | internal/reconcile/snapshot.go:187 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `b.WriteString`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `unknown`: errors and state follow mapped branches.
- `sort.Slice`: errors and state follow mapped branches.
- `less`: errors and state follow mapped branches.
- `brokerOrderIdentity`: errors and state follow mapped branches.
- `fmt.Fprintf`: errors and state follow mapped branches.
- `canonicalDecimal`: errors and state follow mapped branches.
- `b.String`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 4; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
