# Function Logic Map: `activeEntryLossLock`

- Source: `internal/journal/risk_bucket_entry_loss_lock.go`
- AST evidence: `ast.json` (pre-edit, extracted 2026-09-28 at HEAD `1a0027d2`, 115–134, 3 branches; copy in `analysis/pre-edit/5.5-relaxation/`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| scope | canonical | caller | — |
| lock row | first row of the scope (`ORDER BY lock_seq LIMIT 1`) | `risk_bucket_entry_loss_locks` | none → found=false; read/parse error → err |

## Branches and early returns (pre-edit)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | no row (122) | none | found=false | the unlocked control cases |
| B2 | read error (125) | none | err (the caller `refuseEntryUnderLossLock` turns it into EntryBlocked) | storage |
| B3 | activated_at unparseable (129) | none | err | — |

## Planned edit

Callers (grep, non-test): `ActivateEntryLossLock` (:85) and `refuseEntryUnderLossLock` (:144). The latter is called by CRA (:142), commitFresh (:419) and Revalidate (:630).

Planned 5.5 edit (D8): "in force" becomes "has no release row". The query adds `NOT EXIST(release WHERE lock_seq=…)` and returns the newest open lock (there is at most one, by the v35 trigger). Branches are unchanged. The effect: after a release, admission and revalidation stop refusing, which is the approved relaxation. Before any release, the result is identical.

## Safety conclusion

- Entry path only. The lock is read by admission and revalidation, and never by a risk-reducing path (D7).
- High-risk impact: yes. The release is a relaxation, allowed only through the OPERATOR/approval/audit API of D8.
