# Function Logic Map: `Journal.ClaimStrategyDispatchLease`

- Source: `internal/journal/strategy_dispatch_runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Journal` / DB | non-nil, open schema v26 journal | `Journal.db` | invalid request, no mutation |
| CAS lease identity | canonical `lease_id`, positive expected revision/owner epoch, canonical fence | caller plus durable lease row | invalid request, no mutation |
| current dispatch owner | exact durable epoch and fencing token | `strategy_dispatch_owner_current` | fenced, no mutation |
| lease state/revision | exact `ISSUED` and requested revision | `strategy_dispatch_leases` | replay/stale claim fails closed; terminal lease is not revived |
| expiry | `journal clock < expires_at` | lease row + `Journal.clk` | atomic `REFUSED + RELEASED` |
| market authority | exact account/market/symbol/revision/digest | `strategy_dispatch_market_authorities` | atomic `REFUSED + RELEASED` |
| first-leg binding and authorities | exact binding, q_final holds, campaign/claim/leg/owner/strategy lineage remain current; attempt entry identity, activation manifest and client order repeat the live joins | v26 tables and bound v20-v26 rows | atomic `REFUSED + RELEASED` |
| operation/router | `lease.operation_id == attempt.client_order_id`; lease router exactly equals immutable binding router | v26 binding + attempt + lease | mismatch refuses before transport |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | malformed receiver/CAS identity | none | `ErrInvalidRequest` | invalid claim |
| B2 | transaction/current owner lookup fails or owner is stale | none | `ErrStrategyDispatchFenced` or storage error | stale owner |
| B3 | lease missing | none | typed terminal refusal result cannot be created; fail closed | missing lease |
| B4 | lease already consumed or expected revision differs | none; original lease/disposition unchanged | replay/stale error | replayed claim |
| B5 | lease owner does not equal caller/current owner | none | fenced | stale-owner claim |
| B6 | lease expired | lease + exact aggregate/five holds atomically `REFUSED + RELEASED`; outcome audit row appended | refused lease | expiry KR+US |
| B7 | exact current authority/binding/lineage/holds mismatch, including cross-market, attempt-entry, manifest, client-order, operation or router drift | same refusal transaction as B6 | refused lease | drift matrix KR+US |
| B8 | all durable authority is exact/current | lease `ISSUED -> CLAIMED`, revision +1; outcome audit row appended; holds remain HELD | claimed lease | paired KR+US success |
| B9 | refusal cannot prove exactly one bound aggregate in RELEASED state and exactly five scoped dimensions in RELEASED/zero-hold state | transaction rollback; lease/audit remain ISSUED/unwritten | `ErrStrategyDispatchLeaseUnavailable` | partial/cross-scope release KR+US |
| B10 | claim/refusal helper or outcome append fails | transaction rollback | wrapped storage error | synthetic late outcome failure |
| B11 | transaction commit fails | no successful receipt; SQLite transaction remains uncommitted | wrapped commit error | journal transaction fail-closed contract |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requireCurrentStrategyDispatchOwner` | fence stale process before authority validation | no retry or fallback | CodeGraph + AST |
| `loadStrategyDispatchLease` | load exact durable lease within transaction | any load error aborts transaction | CodeGraph + AST |
| SQLite authority queries | re-read exact market authority, first-leg binding and mutable authorities | one immediate transaction, fail closed | schema v25/v26 + AST |
| refusal/claim update helpers | atomically normalize lease/outcome/real reservations | lease CAS is one-row; UPDATE row-count errors fail, then aggregate=1 and five distinct released/zero-hold buckets are proven; partial/cross-scope rolls back | schema transition trigger + adversarial tests |

## State mutations and fallbacks

- Current pre-edit implementation is a dormant stub and performs no writes.
- Target implementation performs no Gateway/broker call and grants only durable `CLAIMED`, never `SUBMITTING`.
- Validation refusal is irreversible and normalizes only the lease-bound aggregate and five monetary holds in the same transaction. Already-RELEASED exact rows are idempotent; missing, cross-scope or non-releasable rows roll back without lease/audit mutation.
- Missing/replayed/stale-owner requests cannot mutate a different or terminal lease.
- KR and US use one market-parameterized code path and are tested in the same wave.

## Safety conclusion

- Safe edit boundary: journal-only fenced claim/revalidation CAS; no execution transport, activation, approval, toggle, or Gateway capability.
- High-risk impact: yes — exposure-raising dispatch authority and risk reservations; fail-closed atomic tests are mandatory.
