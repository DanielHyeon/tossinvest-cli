# Function Logic Map: `ensureRiskBucketEntryScopeClean`

- Source: `internal/journal/risk_bucket.go`
- AST evidence: `ast.json` (pre-edit, extracted 2026-09-28 at HEAD `4fb0b26e`, 268–288, 4 branches; copy in `analysis/pre-edit/6.5-fixes/`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| owner key (account, market, symbol) | canonical owner key | caller's admission plan | — |
| owner latches, scope latches, active reconciles | 0 each | `risk_bucket_owners`, `risk_bucket_scope_latches`, `reconcile_states` inside the admission tx | any non-zero → `ErrRiskBucketEntryBlocked` (bare sentinel, no cause named) |

## Branches and early returns (pre-edit)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | owner-latch count query error (270) | none | err | storage exit — structural test |
| B2 | scope-latch count query error (275) | none | err | storage exit — structural test |
| B3 | active-reconcile count query error (279) | none | err | storage exit — structural test |
| B4 | any count ≠ 0 (284) | none | `ErrRiskBucketEntryBlocked` | existing admission latch tests (`TestA066LatchedUsageInASharedBucketBlocksNewEntry`, owner/scope latch admission tests) |

## Calls, callers and the planned edit

Callers (grep internal/journal non-test): `CommitRiskBucketAdmission` (risk_bucket.go:138) and `commitFreshRiskBucketAdmissionTx` (risk_bucket_issuance.go:415), both inside the admission transaction.

Planned 6.5 edit (finding 1): the body moves into one rule function that **names** the cause (owner latch kind, scope latch, reconcile cause). This function keeps its signature and sentinel, and `RevalidateQFinalAdmission` calls the same rule at submit. The querier type widens from `*sql.Tx` to `riskBucketQueryer` (the same three SELECTs), so revalidation can run it on `j.db`. Entry path only.

## Safety conclusion

- Safe edit boundary: the named branches only; entry refusal / fill latch in the conservative direction.
- High-risk impact: yes (admission / submit revalidation / fill path).
