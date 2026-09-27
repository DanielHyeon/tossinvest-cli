# Branch Test Map: `readProductionRiskUsage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 429:2 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B2 | for at 434:2 — `for rows.Next() {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B3 | if at 436:3 — `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
