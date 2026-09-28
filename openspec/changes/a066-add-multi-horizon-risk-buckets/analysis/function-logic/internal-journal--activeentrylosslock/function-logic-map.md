# Function Logic Map: `activeEntryLossLock`

- Source: `internal/journal/risk_bucket_entry_loss_lock.go`
- AST evidence: `ast.json` (post-edit, 121–141, 3 branches). Pre-edit map (HEAD `1a0027d2`, 115–134) is kept in `analysis/pre-edit/5.5-relaxation/`.
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| scope | canonical | caller | — |
| open lock row | newest lock of the scope with no release row (`NOT EXISTS release`, `ORDER BY lock_seq DESC LIMIT 1`; at most one by the v35 trigger) | `risk_bucket_entry_loss_locks` + `risk_bucket_entry_loss_lock_releases` | none → found=false; read/parse error → err |

## Branches and early returns (post-edit, from `ast.json`)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | no open row (129) | none | found=false | `TestA066EntryLossLockReleaseIsOperatorApprovedAuditedAndOpensEntry` (after the release entry opens), unlocked control cases |
| B2 | read error (132) | none | err (the caller `refuseEntryUnderLossLock` turns it into EntryBlocked) | `TestRefuseEntryUnderLossLockFailsClosedOnUnknownScopeAndReadError`, structural storage exit |
| B3 | activated_at unparseable (136) | none | err | structural storage exit |

## Edit made (5.5, D8)

Only the query changed: "in force" is now "has no release row". Branches are unchanged. Before any release the result is identical (one row per scope). After a release, admission and revalidation stop refusing — the approved relaxation. `TestMigrationV34ToV35KeepsExistingLocksInForceAndSwapsTheTrigger` pins that v34 locks stay in force after the migration.

## Calls and live bindings

| Call | Binding | Note |
|---|---|---|
| `q.QueryRowContext` | `riskBucketQueryer` — a `*sql.Tx` (activation, admission, release) or `*sql.DB` (revalidation) | the `NOT EXISTS` release filter is the whole edit |
| `time.Parse(time.RFC3339Nano, …)` | stdlib | stored form from `formatJournalTime` |

Callers (non-test): `ActivateEntryLossLock`, `refuseEntryUnderLossLock` (called by CRA, commitFresh and Revalidate), `ReleaseEntryLossLock` (reads the open lock inside its tx and requires it to be the lock the operator saw).

## State mutations and fallbacks

- Read only. No write.
- Read or parse error returns an error. `refuseEntryUnderLossLock` turns it into EntryBlocked (fail closed). `ReleaseEntryLossLock` returns it and releases nothing.

## Safety conclusion

- Entry path only. The lock is read by admission and revalidation, and never by a risk-reducing path (D7).
- High-risk impact: yes. The release is a relaxation, allowed only through the OPERATOR/approval/audit API of D8.
