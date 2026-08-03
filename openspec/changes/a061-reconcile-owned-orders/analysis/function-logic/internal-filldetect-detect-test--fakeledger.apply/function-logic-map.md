# Function Logic Map: `fakeLedger.Apply`

Source: `internal/filldetect/detect_test.go`  
Function: `fakeLedger.Apply`  
Signature: `fakeLedger.Apply(params=2, results=2)`  
Source SHA-256: `7fe5825a894d212e278325c39d6b369d975ef46f006b913627daa8c7264e2e26`

## Inputs and invariants

- Inputs are the parameters in `fakeLedger.Apply(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect_test.go:193 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/detect_test.go:209 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `l.mu.Lock`: returned errors and state follow the mapped branches.
- `l.mu.Unlock`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `strings.ToLower`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `l.clk.Now`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 7 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
