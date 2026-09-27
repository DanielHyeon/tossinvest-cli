# Pre-edit branch table: `internal-journal--loadriskbucketorderauthority` (before a066 5.6.1)

Moved out of `function-logic-map.md` when the map was re-based on the post-edit AST at `b8211926`.

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | if at 462:2 | `if err != nil {` | NOT covered |
| B2 | range at 468:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {` | covered |
| B3 | for at 472:2 | `for rows.Next() {` | covered |
| B4 | if at 475:3 | `if err := rows.Scan(&binding.Dimension, &binding.Value, &binding.PolicyVersion, &binding.ReservationID, &binding.ReservedMinor, &binding.SnapshotID, &binding.SnapshotDige` | NOT covered |
| B5 | if at 479:3 | `if _, duplicate := byDimension[dimension]; !required[dimension] \|\| duplicate \|\| binding.PolicyDigest != policyDigest \|\| rowQuote == "" \|\| rowBase == "" \|\| (quot` | NOT covered |
| B6 | if at 485:2 | `if err := rows.Err(); err != nil {` | NOT covered |
| B7 | if at 488:2 | `if len(byDimension) != len(required) {` | NOT covered |
| B8 | range at 492:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {` | covered |
| B9 | if at 494:3 | `if !ok {` | NOT covered |
| B10 | if at 500:2 | `if err != nil {` | NOT covered |
