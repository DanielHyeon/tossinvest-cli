# Branch Test Map: `loadRiskBucketOrderAuthority`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 462:2 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B2 | range at 468:2 — `for _, dimension := range riskbucket.RequiredDimensionOrder() {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B3 | for at 472:2 — `for rows.Next() {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B4 | if at 475:3 — `if err := rows.Scan(&binding.Dimension, &binding.Value, &binding.PolicyVersion, &binding.ReservationID, &binding.ReservedMinor, &binding.SnapshotID, &binding.SnapshotDige` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B5 | if at 479:3 — `if _, duplicate := byDimension[dimension]; !required[dimension] \ | package suites (pre-edit) | n/a (pre-edit map) | \ at `f2decd0a` |
| B6 | if at 485:2 — `if err := rows.Err(); err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B7 | if at 488:2 — `if len(byDimension) != len(required) {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B8 | range at 492:2 — `for _, dimension := range riskbucket.RequiredDimensionOrder() {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B9 | if at 494:3 — `if !ok {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B10 | if at 500:2 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
