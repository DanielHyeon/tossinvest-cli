# Function Logic Map: `Journal.PendingStrategyPlans`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal | non-nil, open journal | `Journal.Open` | account-required/error, no writes |
| accountRef | non-blank exact account scope | caller + persisted attempt account | account-required error |
| legacy attempt | `PLANNED` or `IN_DOUBT` without a v26 first-leg binding | `strategy_attempt_lineage` | returned for legacy mutation recovery |
| first-leg attempt | attempt is referenced by `strategy_first_leg_bindings` | immutable v26 companion | excluded from legacy mutation-attempt recovery |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil/closed journal or blank account | none | validation error | existing invalid-input tests |
| B2 | account query fails | read only | database error | existing closed-journal coverage |
| B3 | row scan fails | read only | scan error | query contract coverage |
| B4 | matching legacy attempt | append receipt in stable order | pending receipt | existing account-scope/recovery tests |
| B5 | attempt has immutable first-leg binding | none | omitted from result | paired KR/US first-leg reopen test |
| B6 | row iteration finishes/errors | read only | result/rows error | existing recovery suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `db.QueryContext` | enumerate account-scoped recoverable legacy attempts | context cancellation/database errors propagate | AST query branch |
| `rows.Scan` | reconstruct opaque strategy plan receipt | scan failure propagates without mutation | AST loop branch |

## State mutations and fallbacks

- Read-only operation; the only behavior change is a negative membership filter against the immutable
  v26 companion table. First-leg attempts remain `PLANNED` for the future lease path and cannot be
  refused by the older mutation-attempt recovery loop.
- Legacy attempts remain ordered by `created_at,attempt_id` and retain their existing recovery behavior.

## Safety conclusion

- Safe edit boundary: add only the v26 companion exclusion predicate; do not change legacy states,
  ordering, receipt fields, account scoping or `RecoverStrategyDispatch`.
- High-risk impact: yes — without the exclusion, restart recovery can terminally refuse a valid KR/US
  first-leg admission before a dispatch lease is intentionally created.
