# Function Logic Map: `Journal.RecoverClaimedStrategyDispatchLease`

- Source: `internal/journal/strategy_dispatch_runtime.go`
- Function: `Journal.RecoverClaimedStrategyDispatchLease`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal/DB | non-nil open v26 journal | `Journal.db` | invalid request, no write |
| recovery CAS | canonical lease ID, positive revision/current owner epoch, canonical fence | request plus current-owner row | invalid or fenced, no write |
| lease | exact old `CLAIMED + RESERVED` revision, no `transport_started_at` | `strategy_dispatch_leases` | consumed/stale, no write |
| recovery owner | current epoch/fence and epoch strictly newer than lease issue owner | `strategy_dispatch_owner_current` | fenced, no write |
| first-leg authority | exact v26 strategy/campaign/risk binding, zero mapped orders | immutable/current journal rows | unavailable, transaction rollback |
| monetary holds | one normalizable aggregate plus five exact dimensions | aggregate and risk-bucket reservation rows | unavailable, transaction rollback |
| optional prepared core | zero or one exact matching `RECORDED` attempt with no dispatch/order/settlement | mutation attempt plus intent/strategy binding | consumed/integrity refusal, rollback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid receiver or malformed CAS | none | `ErrInvalidRequest` | invalid recovery |
| B2 | transaction begin fails | none | wrapped storage error | transaction fail-closed contract |
| B3 | current owner proof fails | none | fenced/storage error | stale owner KR/US |
| B4 | lease is missing | none | unavailable | missing lease |
| B5 | lease load fails | none | storage error | transaction fail-closed contract |
| B6 | lease is not exact `CLAIMED + RESERVED`, revision differs, or transport marker exists | none | consumed | replay and `SUBMITTING` KR/US |
| B7 | recovery owner is not newer than original issue owner | none | fenced | same/stale epoch |
| B8 | first-leg/hold authority proof fails | none; transaction rollback | unavailable/storage error | integrity refusal matrix |
| B9 | optional prepared-core lookup fails | none; transaction rollback | unavailable/storage error | cardinality/binding mismatch |
| B10 | an exact prepared core exists | core becomes `NOT_DISPATCHED` in current transaction | continue | prepared core KR/US |
| B11 | prepared-core transition fails | all prior writes rollback | transition/storage error | injected rollback KR/US |
| B12 | lease/six-hold refusal fails | core and all holds rollback | refusal/storage error | injected late outcome KR/US |
| B13 | commit fails | no successful recovery receipt | wrapped commit error | journal durability contract |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireCurrentStrategyDispatchOwner` | fence the crashed predecessor and arbitrary callers | no retry/fallback in transaction | AST + stale-owner tests |
| `loadStrategyDispatchLease` | read exact durable pre-transport state | missing/stale fails closed | AST + replay tests |
| `proveClaimedStrategyDispatchPreTransportAuthority` | prove v26 lineage, unmapped risk owner and six normalizable holds | any mismatch aborts | refusal integrity suite |
| `recoverablePreparedStrategyAttempt` | locate zero or one exact core attempt left by post-Prepare crash | non-RECORDED or multiplicity fails closed | prepared-core tests |
| `transitionStrategyAttemptTx` | close the optional core without dispatch | same transaction; no resend | rollback tests |
| `refuseClaimedStrategyDispatchSubmittingTx` | release aggregate/five buckets and terminalize the lease | exact row/dimension proofs before commit | paired success/rollback tests |

## State mutations and fallbacks

- Optional core: `RECORDED → NOT_DISPATCHED` with
  `RECOVERY_CLAIMED_NO_TRANSPORT`.
- Lease: `CLAIMED + RESERVED → REFUSED + RELEASED`, revision +1, immutable
  original issue owner/fence retained.
- Aggregate and five monetary buckets: exact remaining holds become RELEASED in
  the same transaction.
- There is no fallback to `SUBMITTING`, `AMBIGUOUS`, a fresh lease, retry, resend,
  Gateway, broker, activation, toggle, or approval.

## Safety conclusion

This edit only removes stale exposure authority after a newly fenced owner proves
that durable transport never started. It cannot increase authority or send an
order. KR and US share the identical market-parameterized code and paired tests.

