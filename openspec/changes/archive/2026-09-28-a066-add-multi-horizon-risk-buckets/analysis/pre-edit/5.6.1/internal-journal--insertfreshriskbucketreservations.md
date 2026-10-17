# Pre-edit branch table: `internal-journal--insertfreshriskbucketreservations` (before a066 5.6.1)

Moved out of `function-logic-map.md` when the map was re-based on the post-edit AST at `b8211926`.

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
