# Function Logic Map: `Journal.ActivateEntryLossLock`

- Source: `internal/journal/risk_bucket_entry_loss_lock.go`
- AST evidence: `ast.json` (pre-edit, extracted 2026-09-28 at HEAD `1a0027d2`, 72–108, 8 branches; copy in `analysis/pre-edit/5.5-relaxation/`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| lock (account, market, horizon, cause, activated_at) | canonical account, KR/US, SHORT/MEDIUM, cause non-empty, time set | caller (automatic trigger; 0 production callers today) | invalid → `ErrInvalidRequest` |
| existing lock for the scope | at most one row (v33 first-cause-wins trigger) | `risk_bucket_entry_loss_locks` in a BEGIN IMMEDIATE tx | found → return it, changed=false, **nothing written** |

## Branches and early returns (pre-edit)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | nil journal (73) | none | `ErrInvalidRequest` | lock tests |
| B2 | invalid scope/cause/time (76) | none | `ErrInvalidRequest` | `TestA066*` padded-account and invalid-scope tests |
| B3 | begin error (81) | none | err | storage exit |
| B4 | read error (86) | none | err | storage exit |
| B5 | lock already present (89) | **none** — a repeated activation leaves no trace | existing, false | first-cause-wins tests |
| B6 | insert error (94) | none | err | storage exit (the trigger aborts are storage errors here) |
| B7 | LastInsertId error (98) | none | err | storage exit |
| B8 | commit error (101) | none | err | storage exit |

## Planned edit

Planned 5.5 edit (D8):
- B5 becomes "an **open** lock is present". Inside the same transaction it writes a `REAFFIRM` event (cause, time) and still returns the open lock with changed=false. The event is what lets a later release that the operator approved before this tightening lose (stale).
- Before B5, the open-lock read goes through `activeEntryLossLock`, which now means "no release row".
- The function never relaxes anything.

## Safety conclusion

- Entry path only. The lock is read by admission and revalidation, and never by a risk-reducing path (D7).
- High-risk impact: yes. The release is a relaxation, allowed only through the OPERATOR/approval/audit API of D8.
