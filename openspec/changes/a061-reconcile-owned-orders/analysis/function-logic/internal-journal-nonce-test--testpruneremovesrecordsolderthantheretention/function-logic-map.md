# Function Logic Map: `TestPruneRemovesRecordsOlderThanTheRetention`

Source: `internal/journal/nonce_test.go`  
Function: `TestPruneRemovesRecordsOlderThanTheRetention`  
Signature: `TestPruneRemovesRecordsOlderThanTheRetention(params=1, results=0)`  
Source SHA-256: `83fcf17c3cd3758fadd4f23e7f31e675b8e3a2df7d56d3cdd6e70b583e16f5e3`

## Inputs and invariants

- Inputs are the parameters in `TestPruneRemovesRecordsOlderThanTheRetention(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce_test.go:234 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/nonce_test.go:241 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/nonce_test.go:244 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/nonce_test.go:250 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/nonce_test.go:253 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openTestJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `cancelDecision`: returned errors and state follow the mapped branches.
- `boundAttempt`: returned errors and state follow the mapped branches.
- `a.MarkDispatchStarted`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `testIssued`: returned errors and state follow the mapped branches.
- `j.PruneSpentNonces`: returned errors and state follow the mapped branches.
- `issued.Add`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `spentNonceCount`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 8 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
