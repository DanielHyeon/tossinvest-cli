# Pre-edit branch table: `internal-riskbucket--readproductionriskusage` (before a066 5.6.1)

Moved out of `function-logic-map.md` when the map was re-based on the post-edit AST at `b8211926`.

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | if at 429:2 | `if err != nil {` | NOT covered |
| B2 | for at 434:2 | `for rows.Next() {` | covered |
| B3 | if at 436:3 | `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | NOT covered |
