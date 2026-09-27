# Function Logic Map: `insertFreshRiskBucketReservations`

- Source: `internal/journal/risk_bucket_issuance.go`
- AST evidence: `ast.json` (5.6.1 pre-edit, extracted 2026-09-27 at HEAD `f2decd0a` before any 5.6.1 edit)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit source copy (`analysis/harness/test_in_copy.sh`), `go test -coverpkg ./internal/journal,./internal/riskbucket` over the journal package (untagged) and riskbucket (untagged + `tossos_testseams`); rows rendered by `analysis/harness/preedit_rows.py` from `git show f2decd0a:<file>`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| tx / plan / decision | caller-validated q_final admission inside BEGIN IMMEDIATE | commitFreshRiskBucketAdmissionTx | any error returns; caller rolls back |

## Branches and early returns

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | range at 517:2 | `for i, cap := range decision.Caps {` | covered |
| B2 | if at 525:3 | `if err != nil {` | NOT covered |
| B3 | if at 528:3 | `if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,po` | NOT covered |
| B4 | if at 532:3 | `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_policies WHERE bucket_dimension=? AND bucket_value=? AND policy_version=?`, string(cap.Key.Dimens` | NOT covered |
| B5 | if at 539:3 | `if err != nil {` | NOT covered |
| B6 | if at 543:3 | `if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,polic` | NOT covered |
| B7 | if at 547:3 | `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_snapshots WHERE snapshot_id=?`, ref.SnapshotID).Scan(&storedSnapshotDigest); err != nil \|\| stor` | NOT covered |
| B8 | if at 551:3 | `if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_gener` | NOT covered |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `riskBucketRecordDigest` | policy and snapshot record digests | error returns | AST |
| SQL policies/snapshots/reservations | persist the five bindings | INSERT OR IGNORE then re-read; a different record under the same key is a collision (B4) | AST |

## State mutations and fallbacks

- Per cap: policy row (INSERT OR IGNORE), snapshot row (INSERT OR IGNORE), HELD reservation row.

## Safety conclusion

- Safe edit boundary (5.6.1 F2): replace B2–B4 (policy digest + parent insert + key-only collision check) with `storeRiskBucketPolicyRecord`, and bind the reservation to that record digest. The key-only collision (B4) is the measured F2 defect.
- High-risk impact: yes — q_final sizing/admission authority.
