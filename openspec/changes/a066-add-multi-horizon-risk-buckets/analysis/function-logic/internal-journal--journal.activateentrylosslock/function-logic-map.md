# Function Logic Map: `Journal.ActivateEntryLossLock`

- Source: `internal/journal/risk_bucket_entry_loss_lock.go`
- AST evidence: `ast.json` (post-edit, 72–117, 10 branches). Pre-edit map (HEAD `1a0027d2`, 72–108, 8 branches) is kept in `analysis/pre-edit/5.5-relaxation/`.
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| lock (account, market, horizon, cause, activated_at) | canonical account, KR/US, SHORT/MEDIUM, cause non-empty, time set | caller (automatic trigger; 0 production callers today) | invalid → `ErrInvalidRequest` |
| open lock for the scope | at most one open row (v35 `one_open` trigger; open = no release row) | `risk_bucket_entry_loss_locks` + `risk_bucket_entry_loss_lock_releases` in a BEGIN IMMEDIATE tx | found → REAFFIRM event written, return it, changed=false |

## Branches and early returns (post-edit, from `ast.json`)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | nil journal (73) | none | `ErrInvalidRequest` | lock tests |
| B2 | invalid scope/cause/time (76) | none | `ErrInvalidRequest` | `TestEntryLossLockSchemaRefusesValuesTheRuleCannotMatch`, `TestEntryLossLockAccountFormIsTheDecisionBuildersForm` |
| B3 | begin error (81) | none | err | storage exit (structural, `TestA066StorageErrorExitsFailClosed`) |
| B4 | open-lock read error (86) | none | err | storage exit (structural) |
| B5 | an **open** lock is present (89) | REAFFIRM event (lock_seq, cause, time) in the same tx | existing, false | `TestA066ReaffirmAfterTheOperatorLookedMakesTheReleaseStale`, `TestEntryLossLockActivationKeepsTheFirstCauseAndIsImmutable` |
| B6 | REAFFIRM insert error (92) | none (rolled back) | err | storage exit (structural) |
| B7 | REAFFIRM commit error (96) | none | err | storage exit (structural) |
| B8 | lock insert error (103) | none | err | storage exit; the v35 `one_open` abort is `TestA066AtMostOneOpenLockPerScopeIsEnforcedByTheSchema` (raw writer) |
| B9 | LastInsertId error (107) | none | err | storage exit (structural) |
| B10 | commit error (110) | none | err | storage exit (structural) |

Fallthrough (no open lock, including after a release): a new lock row → `changed=true` — `TestA066EntryLossLockReleaseIsOperatorApprovedAuditedAndOpensEntry` (activation after release opens a new lock).

## Edit made (5.5, D8)

- B5 was "a lock row is present → nothing written". It is now "an **open** lock is present → REAFFIRM event". B6 and B7 are the two new storage exits of that event. The old B6–B8 are now B8–B10 (same conditions).
- The function never relaxes anything. A repeated tightening still keeps the first cause and returns the same lock.

## Calls and live bindings

| Call | Binding | Note |
|---|---|---|
| `validEntryLossLockScope` | same file | canonical account, KR/US, SHORT/MEDIUM |
| `j.db.BeginTx` | SQLite, `_txlock=immediate` | BEGIN IMMEDIATE — concurrent activations serialize |
| `activeEntryLossLock(ctx, tx, …)` | same tx | open lock = no release row (v35) |
| `tx.ExecContext` REAFFIRM insert | `risk_bucket_entry_loss_lock_events` (v35, append-only triggers) | only when an open lock exists |
| `tx.ExecContext` lock insert | `risk_bucket_entry_loss_locks` (v35 `one_open` trigger) | only when no open lock exists |
| `formatJournalTime` | journal helper | seconds, UTC |

Production callers: 0 (dormant; the trigger lot wires activation). The release is `ReleaseEntryLossLock` in `risk_bucket_relaxation.go`, never this function.

## State mutations and fallbacks

- Open lock present: one REAFFIRM row, committed. No lock row changes. A failure rolls back (`defer tx.Rollback()`), nothing written.
- No open lock: one lock row, committed.
- No fallback relaxes: every error path returns an error and writes nothing.

## Safety conclusion

- Entry path only. The lock is read by admission and revalidation, and never by a risk-reducing path (D7).
- High-risk impact: yes. The release is a relaxation, allowed only through the OPERATOR/approval/audit API of D8.
