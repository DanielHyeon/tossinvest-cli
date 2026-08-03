# Function Logic Map: `TestSnapshotCarriesTheFilledAmount`

Source: `internal/filldetect/payload_test.go`
Function: `TestSnapshotCarriesTheFilledAmount`
Signature: `TestSnapshotCarriesTheFilledAmount(params=1, results=0)`
Source SHA-256: `14be87b64e23ea392eff45b6daf15ee72d8307260b6d46185aac6ed748d5d9c2`
Revision: `base`

## Inputs and invariants

- Inputs are `TestSnapshotCarriesTheFilledAmount(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload_test.go:48 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `pollOne`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 1; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
