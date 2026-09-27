# Branch Test Map: `insertFreshRiskBucketReservations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 517:2 — `for i, cap := range decision.Caps {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B2 | if at 525:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B3 | if at 528:3 — `if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,po` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B4 | if at 532:3 — `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_policies WHERE bucket_dimension=? AND bucket_value=? AND policy_version=?`, string(cap.Key.Dimens` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B5 | if at 539:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B6 | if at 543:3 — `if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,polic` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B7 | if at 547:3 — `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_snapshots WHERE snapshot_id=?`, ref.SnapshotID).Scan(&storedSnapshotDigest); err != nil \ | package suites (pre-edit) | n/a (pre-edit map) | \ at `f2decd0a` |
| B8 | if at 551:3 — `if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_gener` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
