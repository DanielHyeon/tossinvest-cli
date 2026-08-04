# Function Logic Map: `Journal.RefuseClaimedStrategyDispatchPreTransport`

- Source: `internal/journal/strategy_dispatch_refusal.go`
- AST evidence: `ast.json` (generated after RED/GREEN)
- Risk scan: `risk-pattern-report.md` (generated after RED/GREEN)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request CAS | exact lease id, CLAIMED revision, current owner epoch/token | journal owner + lease rows | fenced/consumed; zero mutation |
| request binding | complete immutable lease plan equal to loaded plan | durable lease row | unavailable; zero mutation |
| reason | one closed canonical record-only enum | journal enum | invalid request; zero mutation |
| reservations | one exact aggregate plus five canonical buckets, each safely HELD or already RELEASED, with zero fill/overage and no order mapping | risk reservation rows + v26 binding/prospective token | integrity mismatch unavailable; zero mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid journal/request/reason | none | invalid request | invalid reason test |
| B2 | transaction begin fails | none | wrapped storage error | storage contract |
| B3 | current owner/fence differs | none | fenced | ABA paired test |
| B4 | lease absent | none | unavailable | missing lease contract |
| B5 | lease load fails | none | storage error | storage contract |
| B6 | non-CLAIMED, non-RESERVED, transport-started or stale revision | none | consumed | replay/concurrency test |
| B7 | durable plan proof query fails | none | storage error | storage contract |
| B8 | durable plan differs | none | unavailable | cross-attempt/market test |
| B9 | first-leg/prospective/cardinality/integrity proof differs | none | unavailable | integrity/binding tests |
| B10 | aggregate or buckets are a safe HELD/RELEASED mix | preserve prior release metadata; release only remaining HELD rows | terminal REFUSED+RELEASED | paired partial-release tests |
| B11 | atomic release/terminal helper fails | transaction rollback to exact mixed preimage | error | injected rollback test |
| B12 | commit fails | transaction rollback | wrapped error | storage contract |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireCurrentStrategyDispatchOwner` | exact central fence | fail closed, no retry | current code + AST |
| `loadStrategyDispatchLease` | durable lease authority | no caller fallback | current code + AST |
| first-leg/normalization proof query | exact operation/binding/prospective token, 1+5 cardinality, safe state and no prior order mapping | fail closed on FILLED/partial/missing/wrong-scope state | v26 schema + tests |
| `refuseClaimedStrategyDispatchSubmittingTx` | existing atomic release/terminal helper | same transaction; no transport | current code + AST |

## State mutations and fallbacks

- The only mutation is the existing same-transaction `CLAIMED+RESERVED -> REFUSED+RELEASED`
  release helper after all proofs pass. It updates only HELD rows, so prior release metadata survives.
- No fallback, authority repair, reason-based authorization, transport or broker call exists.

## Composite attempt variant

`Attempt.RefuseClaimedStrategyPreTransport` is the post-`Prepare` form. It proves that the core attempt is still
`RECORDED`, then performs the core `NOT_DISPATCHED` transition and the exact lease/aggregate/five-bucket refusal
inside the same transaction. The lease-only journal API remains available only before an attempt exists or while
recovering a claim that cannot be bound to a prepared attempt. An injected failure at any core, reservation or
lease write rolls back the complete KR/US preimage.

## Safety conclusion

- Safe edit boundary: journal-only leaf API and shared proof/transition helpers.
- High-risk impact: yes — irreversible lease/reservation transition; paired rollback/race tests required.
