# Function Logic Map: `Console.orders`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `choice` | server-defined query link values; unknown values normalize to all | `filterChoiceFrom` / filter helpers | invalid choice never reaches broker and excludes nothing unexpectedly |
| broker reading | each OPEN/CLOSED/conditional list independently measured or failed | `OrdersReader` cache | partial failure never becomes zero or a combined total |
| journal order evidence | broker order id maps through a concrete mutation attempt decision id only | read-only journal query | unreadable journal => origin/evidence unknown; absent exact join => `근거 미연결` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | OPEN list known | append each broker row and exact-id evidence; count every OPEN row | none | existing order list tests + linked exit fixture |
| B2 | CLOSED list known | skip ids already present in OPEN, append remaining rows | none | existing partial-filled duplicate test |
| B3 | conditional list known | append watching conditional rows; no fuzzy exit join | none | existing conditional tests + unlinked fixture |
| B4 | any list truncated | mark aggregate as lower bound | none | existing pagination tests |
| B5 | filter applied | filter already-built rows locally | none | existing filter matrix |
| B6 | exact broker-order/attempt/decision/event chain exists | attach canonical `ExitLineView` from event effective snapshot | none | `TestOrdersJoinExitEvidenceOnlyByAttemptDecisionID` |
| B7 | no exact chain or corrupt/legacy event | attach unknown/unlinked view; never compare symbol/time | none | unlinked/fuzzy-negative fixtures |
| B8 | conditional list iteration | append each watching conditional without manufacturing a plain-order decision link | none | conditional and unlinked fixtures |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ordersCache.get` | one rate-budgeted three-list broker read | TTL/hold/failure contracts unchanged | CodeGraph + existing tests |
| `ledgerOrderEvidence` (planned) | read exact broker order provenance and exit snapshot | read-only handle; one local SQLite query, no retry | a042 journal schema + planned read-only tests |
| `rowFromOrder` / `rowFromConditional` | render broker facts without coercing absence to zero | pure, no error | current HEAD + AST |
| `operatorview.FromSnapshot` (planned) | canonical display adapter shared with positions/future HTTP | unknown/stale fail closed | adapter tests |
| count/filter helpers | preserve completeness and local-link filtering | pure, no broker call | existing test suite |

## State mutations and fallbacks

- Broker cache is the only external read and retains its existing TTL/hold behavior.
- New journal query is read-only and keyed by `mutation_attempts.broker_order_id -> mutation_attempts.decision_id -> exit_events.decision_id`; symbol, price, and time are absent from the join by construction.
- No state mutation or order call is introduced.

## Safety conclusion

- Safe edit boundary: enrich already-built rows with persisted evidence while preserving counts, order state, origin, filtering, and rate budget.
- High-risk impact: no runtime decision/write change. Journal/order data is observed only; malformed evidence degrades to unknown.
