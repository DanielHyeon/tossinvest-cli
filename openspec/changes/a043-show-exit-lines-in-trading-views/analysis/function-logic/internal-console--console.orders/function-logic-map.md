# Function Logic Map: `Console.orders`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `choice` | server-defined query link values; unknown values normalize to all | `filterChoiceFrom` / filter helpers | invalid choice never reaches broker and excludes nothing unexpectedly |
| broker reading | each OPEN/CLOSED/conditional list independently measured or failed | `OrdersReader` cache | partial failure never becomes zero or a combined total |
| journal order evidence | id+resolved account+market-local trading day maps through exact attempt/intent lineage | read-only journal query | invalid identity/unreadable journal => unknown; absent event => `근거 미연결` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | OPEN is known; iterate rows and remember pending IDs | append composite-key evidence; count every OPEN row | none | order list + linked exit fixture |
| B4-B6 | CLOSED is known; iterate and skip IDs already in OPEN | append composite-key evidence | none | partial-filled duplicate test |
| B7-B8 | conditional list is known; use its own id/time only | append watching rows; never guess triggered-order day | none | conditional origin tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ordersCache.get` | one rate-budgeted three-list broker read | TTL/hold/failure contracts unchanged | CodeGraph + existing tests |
| `ledgerOrderEvidence` | read exact broker order provenance and exit snapshot | read-only handle; local SQLite reads, no retry | a042 journal schema + read-only tests |
| `rowFromOrder` / `rowFromConditional` | render broker facts without coercing absence to zero | pure, no error | current HEAD + AST |
| `operatorview.BuildExitLine` via `attachOrderExitEvidence` | canonical display adapter shared with positions/future HTTP | unknown evidence fails closed | adapter tests |
| count/filter helpers | preserve completeness and local-link filtering | pure, no broker call | existing test suite |

## State mutations and fallbacks

- Broker cache is the only external read and retains its existing TTL/hold behavior.
- Journal query is read-only, limited to visible composite identities, and follows only validated attempt/intent references; symbol/price are never fuzzy join keys.
- No state mutation or order call is introduced.

## Safety conclusion

- Safe edit boundary: enrich already-built rows with persisted evidence while preserving counts, order state, origin, filtering, and rate budget.
- High-risk impact: no runtime decision/write change. Journal/order data is observed only; malformed evidence degrades to unknown.
