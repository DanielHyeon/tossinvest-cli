# Function Logic Map: `Journal.BeginStrategyDispatchSubmitting`

- Source: `internal/journal/strategy_dispatch_runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal / CAS shape | non-nil DB, canonical lease/fence identities, positive revision and owner epoch | caller shape only; never authority | `ErrInvalidRequest`, zero writes |
| current dispatch owner | request epoch/token exactly equal durable `strategy_dispatch_owner_current` | journal transaction | `ErrStrategyDispatchFenced`, zero writes |
| lease state/revision | exact `CLAIMED` and requested revision | `strategy_dispatch_leases` | non-`CLAIMED` replay returns consumed/unavailable without revival; stale revision on a current claimed lease is terminal refusal |
| lease time | `issued_at <= now < expires_at` | journal clock plus durable lease | terminal `REFUSED + RELEASED` before transport |
| market authority | exact account/market/symbol/revision/digest | `strategy_dispatch_market_authorities` | terminal `REFUSED + RELEASED` before transport; revision prevents A->B->A revival |
| v26 first-leg binding | exact decision/reservation/account/market/symbol/candidate/evidence/lane/router/campaign/leg | immutable `strategy_first_leg_bindings` | terminal `REFUSED + RELEASED` before transport |
| live first-leg graph | exact q_final, risk intent/client order, aggregate HELD, attempt/manifest/router, strategy, campaign/claim/leg, risk owner, five distinct HELD buckets | current journal rows joined inside the same transaction | terminal `REFUSED + RELEASED`; inability to prove exact releases rolls the entire transaction back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid receiver or CAS | none | `ErrInvalidRequest` | invalid/missing CAS table |
| B2 | transaction cannot begin | none | wrapped storage error | journal unavailable/closed-DB behavior is covered by package DB failure suites |
| B3 | current owner read fails or request is not current | none | `ErrStrategyDispatchFenced` | old owner after replacement |
| B4 | lease row is absent | none | `ErrStrategyDispatchLeaseUnavailable` | missing lease |
| B5 | lease load or timestamp parse fails for another reason | none; transaction rolls back | wrapped load/parse error | package malformed durable-state/load error contract |
| B6 | lease state is not `CLAIMED` | none | `ErrStrategyDispatchLeaseConsumed` | ISSUED/SUBMITTING/REFUSED replay; terminal row unchanged |
| B7 | current-authority query fails | none; transaction rolls back | wrapped authority error | late outcome/storage rollback test plus package journal query-failure contract |
| B8 | validator returns a refusal code | exact aggregate and five buckets become `RELEASED`; lease becomes terminal `REFUSED`; immutable outcome appended, all in one transaction | refused lease or storage/release-proof error | paired KR/US expiry, stale revision, owner recovery, A->B->A, cross-market, live-join and release-cardinality matrix |
| B9 | validator returns no refusal | one CAS records `transport_started_at`, `CLAIMED -> SUBMITTING`, revision increment and immutable transition | durable `SUBMITTING` lease | paired KR/US current and concurrent CAS matrix |
| B10 | refusal/success mutation helper fails | whole transaction rolls back; lease remains `CLAIMED`, no marker, holds remain HELD | wrapped storage/fence/release-proof error | injected late outcome and cross-scope release-proof failure |
| B11 | transaction commit fails | no successful return; SQLite transaction durability decides all-or-none | wrapped commit error | package crash/durability suites; no function-local commit fault seam |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireCurrentStrategyDispatchOwner` | fence stale process before any lease mutation | exact durable equality; no retry or owner inference | CodeGraph + current source |
| `loadStrategyDispatchLease` | load authoritative lease in transaction before branch selection and after mutation | missing is typed unavailable; parse/storage failure aborts | CodeGraph + current source |
| `validateStrategyDispatchClaimAuthority` | reuse the exact time, market, v26 binding and current live-join validator used by `ISSUED -> CLAIMED` | returns typed refusal code; query error aborts | current source lines 258-366 |
| `refuseClaimedStrategyDispatchSubmittingTx` | release exact real holds and append irreversible refusal from CLAIMED | exact 1 aggregate and 5 distinct scoped buckets must be proven or all writes roll back | paired release and rollback tests |
| `beginStrategyDispatchSubmittingTx` | atomically set transport marker/revision/state and append transition | exact one-row CAS; transition failure rolls back | paired success/race/rollback tests |

## State mutations and fallbacks

- Success writes only journal durability: `state=SUBMITTING`, `revision+1`, `transport_started_at=now`, `updated_at=now`, and one immutable outcome. It does not call Gateway or broker transport.
- Authority drift writes only an exact pre-transport terminal proof: lease `REFUSED`, disposition `RELEASED`, aggregate `RELEASED`, and five scoped bucket rows `RELEASED/held_minor=0` plus one immutable outcome.
- There is no fallback rate, inferred market, resealed authority, lease mint, activation/toggle mutation, broker call or terminal revival.
- A stale process is fenced without touching the old claimed lease. A new current owner may terminally refuse the old bound-owner lease using its own current CAS, but cannot submit it.

## Safety conclusion

- Safe edit boundary: replace only the dormant `BeginStrategyDispatchSubmitting` leaf and add private transaction helpers in `strategy_dispatch_runtime.go`; leave schema, lease minting, Gateway, broker and recovery settlement outside this task.
- High-risk impact: yes — this is the final durable pre-transport boundary. Completion requires paired KR/US RED/GREEN, race, rollback and post-edit AST/map refresh.
