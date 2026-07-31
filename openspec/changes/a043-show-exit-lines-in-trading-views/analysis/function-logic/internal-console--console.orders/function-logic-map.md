# Function Logic Map: `Console.orders`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `choice` | server-defined query link values; unknown values normalize to all | `filterChoiceFrom` / filter helpers | invalid choice never reaches broker and excludes nothing unexpectedly |
| broker reading | each OPEN/CLOSED/conditional list independently measured or failed | `OrdersReader` cache | partial failure never becomes zero or a combined total |
| journal order evidence | broker order id maps through a concrete mutation attempt intent id to the exit event's proposed intent id | read-only journal query | unreadable journal => origin/evidence unknown; absent exact join => `근거 미연결` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | OPEN is known; iterate rows and remember non-empty pending IDs | append exact-id evidence; count every OPEN row | none | existing order list tests + linked exit fixture |
| B4-B6 | CLOSED is known; iterate and skip exact IDs already in OPEN | append remaining rows with exact-id evidence | none | existing partial-filled duplicate test |
| B7-B10 | conditional list is known; iterate rows and candidate explicit IDs, stopping only on exact evidence | append watching rows; never fuzzy join | none | existing conditional tests + unlinked fixture |

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
- New journal query is read-only and keyed by `mutation_attempts.broker_order_id -> mutation_attempts.intent_id -> exit_events.proposed_intent_id`; symbol, price, and time are absent from the join by construction.
- No state mutation or order call is introduced.

## Safety conclusion

- Safe edit boundary: enrich already-built rows with persisted evidence while preserving counts, order state, origin, filtering, and rate budget.
- High-risk impact: no runtime decision/write change. Journal/order data is observed only; malformed evidence degrades to unknown.
