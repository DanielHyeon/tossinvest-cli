# Function Logic Map: `JournalLedger.Apply`

Source: `internal/filldetect/ledger.go`
Function: `JournalLedger.Apply`
Signature: `JournalLedger.Apply(params=2, results=2)`
Source SHA-256: `75966bf1507d412f48b88efddab42f045fff31f45c76bc9cf43dfc0a634242c4`
Revision: `current`

## Inputs and invariants

- Inputs are `JournalLedger.Apply(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger.go:37 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/ledger.go:42 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/filldetect/ledger.go:45 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/filldetect/ledger.go:48 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/filldetect/ledger.go:76 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/filldetect/ledger.go:81 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/filldetect/ledger.go:84 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/filldetect/ledger.go:85 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/filldetect/ledger.go:97 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/filldetect/ledger.go:102 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/filldetect/ledger.go:106 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `fmt.Errorf`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `decimalString`: errors and state follow mapped branches.
- `journal.RFC3339`: errors and state follow mapped branches.
- `l.Journal.RecordFill`: errors and state follow mapped branches.
- `l.Refresh`: errors and state follow mapped branches.
- `time.Parse`: errors and state follow mapped branches.
- `ts.UTC`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 12; return points: 7; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
