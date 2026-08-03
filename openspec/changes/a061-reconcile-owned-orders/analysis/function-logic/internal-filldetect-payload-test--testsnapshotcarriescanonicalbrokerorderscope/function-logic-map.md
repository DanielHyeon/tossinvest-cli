# Function Logic Map: `TestSnapshotCarriesCanonicalBrokerOrderScope`

Source: `internal/filldetect/payload_test.go`  
Function: `TestSnapshotCarriesCanonicalBrokerOrderScope`  
Signature: `TestSnapshotCarriesCanonicalBrokerOrderScope(params=1, results=0)`  
Source SHA-256: `efd2573ee79ff75e82219f73f6fdc135d5491b408ab180b28237e4b710d9061c`

## Inputs and invariants

- Inputs are the parameters represented by `TestSnapshotCarriesCanonicalBrokerOrderScope(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload_test.go:57 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `pollOne`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 1 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
