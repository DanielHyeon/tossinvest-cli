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

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.6.1 post-edit) |
|---|---|---|---|---|
| B1 | range at 523:2 | `for i, cap := range decision.Caps {`; then `bucket, ref := plan.Admission.Buckets[i], plan.Snapshots[i]` (line last changed by `a37d97f5`) | a066 commit | covered |
| B2 | if at 527:3 | `if err != nil {`; then `return err` (line last changed by `a37d97f5`) | a066: 5.6.1 F2: storeRiskBucketPolicyRecord fails (record not stored / digest error) → refuse; the key-only collision is gone | NOT covered |
| B3 | if at 534:3 | `if err != nil {`; then `return err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B4 | if at 538:3 | `if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,policy_version,limit_minor,filled_minor,held_minor,snapshot_version,policy_digest,observed_at,fresh_until,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, ref.SnapshotID, ref.SnapshotDigest, bound.SnapshotEvidence.Source, snapshotRecordDigest, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, bucket.LimitMinor, bucket.FilledMinor, bucket.HeldMinor, ref.SnapshotVersion, ref.PolicyDigest, canonicalRiskTime(snapshotObserved), canonicalRiskTime(snapshotFresh), canonicalRiskTime(plan.CreatedAt)); err != nil {`; then `return err` (line last changed by `8022f578`) | not a066 | NOT covered |
| B5 | if at 542:3 | `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_snapshots WHERE snapshot_id=?`, ref.SnapshotID).Scan(&storedSnapshotDigest); err != nil \|\| storedSnapshotDigest != snapshotRecordDigest {`; then `return fmt.Errorf("%w: immutable snapshot collision", ErrRiskBucketSnapshotMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B6 | if at 546:3 | `if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,created_at,updated_at,policy_record_digest) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,'0','0','HELD',?,?,?)`, reservationID, plan.DecisionID, plan.ExistingReservationID, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, ref.SnapshotID, cap.ReservationAtFinal, cap.ReservationAtFinal, canonicalRiskTime(plan.CreatedAt), canonicalRiskTime(plan.CreatedAt), policyRecordDigest); err != nil {`; then `return fmt.Errorf("journal: insert %s q_final reservation: %w", cap.Key.Dimension, err)` (line last changed by `b8211926`) | not a066 | NOT covered |

5.6.1 post-edit (HEAD `b8211926`): 8 → 6 branches: pre-edit B2–B4 (digest, parent insert, key-only collision) became one `storeRiskBucketPolicyRecord` call (new B2); the reservation insert gained `policy_record_digest`. Pre-edit table: `analysis/pre-edit/5.6.1/internal-journal--insertfreshriskbucketreservations.md`.

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
