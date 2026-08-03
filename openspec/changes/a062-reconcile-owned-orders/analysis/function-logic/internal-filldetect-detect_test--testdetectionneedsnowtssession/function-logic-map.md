# Function Logic Map: `TestDetectionNeedsNoWTSSession`

Source: `internal/filldetect/detect_test.go`
Function: `TestDetectionNeedsNoWTSSession`
Signature: `TestDetectionNeedsNoWTSSession(params=1, results=0)`
Source SHA-256: `7fe5825a894d212e278325c39d6b369d975ef46f006b913627daa8c7264e2e26`
Revision: `current`

## Inputs and invariants

- Inputs are `TestDetectionNeedsNoWTSSession(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect_test.go:858 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/detect_test.go:861 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `clock.NewFake`: errors and state follow mapped branches.
- `Format`: errors and state follow mapped branches.
- `pollStart.Add`: errors and state follow mapped branches.
- `newPager`: errors and state follow mapped branches.
- `page`: errors and state follow mapped branches.
- `filled.json`: errors and state follow mapped branches.
- `newLedger`: errors and state follow mapped branches.
- `d.PollOnce`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
