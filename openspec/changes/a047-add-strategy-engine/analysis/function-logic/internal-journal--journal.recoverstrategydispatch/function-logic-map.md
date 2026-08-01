# Function Logic Map: `Journal.RecoverStrategyDispatch`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| pending receipt | account scoped PLANNED/IN_DOUBT exact receipt | `PendingStrategyPlans` | mismatch error |
| core attempts | bounded `LIMIT 2` rows for strategy intent id | execgw journal | exact terminal classification |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 0 attempt | durable REFUSED reason/revision | typed recovery error | zero test |
| B2 | >1 attempt | durable IN_DOUBT if planned | ambiguity error | cardinality test |
| B3 | exactly one CONFIRMED + broker id | DISPATCHED + links | nil | confirmed test |
| B4 | exactly one NOT_DISPATCHED/FAILED_CONFIRMED | durable REFUSED | recovery error | terminal table |
| B5 | RECORDED/DISPATCH_STARTED/ACKED/IN_DOUBT/UNRESOLVED or malformed confirmed | durable/current IN_DOUBT | recovery error | state table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| bounded core query | detect 0/1/>1 without unbounded scan | DB errors propagate | AST + tests |
| terminal recorders | exact receipt CAS | never guess success | AST |

## State mutations and fallbacks

- Recovery performs no broker call and cannot create a second attempt.

## Safety conclusion

- Safe edit boundary: startup reconciliation of already durable plans.
- High-risk impact: yes, ambiguous exposure remains blocked.
