# Function Logic Map: `TestWhyDoesThisPositionExist`

Source: `internal/journal/provenance_test.go`
Function: `TestWhyDoesThisPositionExist`
Signature: `TestWhyDoesThisPositionExist(params=1, results=0)`
Source SHA-256: `3a77145080d4963658125cee4e7ae33db9f1c6c76d329aa893f589748775f301`
Revision: `base`

## Inputs and invariants

- Inputs are `TestWhyDoesThisPositionExist(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/provenance_test.go:47 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/provenance_test.go:53 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/provenance_test.go:56 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/journal/provenance_test.go:60 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/provenance_test.go:68 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/provenance_test.go:77 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/provenance_test.go:80 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/provenance_test.go:83 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `projectingJournal`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `place`: errors and state follow mapped branches.
- `j.RecordFill`: errors and state follow mapped branches.
- `terminalFill`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `currentPosition`: errors and state follow mapped branches.
- `j.PositionProvenance`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `stepOf`: errors and state follow mapped branches.
- `sort.SliceIsSorted`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- `strings.Contains`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 8; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
