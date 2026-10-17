# Pre-edit branch table: `internal-riskbucket--loadproductionriskentries` (before a066 5.6.1)

Moved out of `function-logic-map.md` when the map was re-based on the post-edit AST at `b8211926`.

| Branch | Position | Condition (AST source line at `f2decd0a`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | if at 340:2 | `if !ok {` | NOT covered |
| B2 | if at 343:2 | `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | NOT covered |
| B3 | if at 352:2 | `if err != nil {` | NOT covered |
| B4 | if at 357:2 | `if err := db.PingContext(ctx); err != nil {` | NOT covered |
| B5 | if at 361:2 | `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \|\| version != productionRiskJournalSchema {` | NOT covered |
| B6 | if at 365:2 | `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,` | NOT covered |
| B7 | if at 376:2 | `if authorityObserved.After(scope.AsOf) \|\| authorityFresh.Before(scope.AsOf) {` | NOT covered |
| B8 | range at 380:2 | `for _, dimension := range requiredDimensions {` | covered |
| B9 | if at 382:3 | `if err != nil {` | NOT covered |
| B10 | if at 386:3 | `if err != nil {` | covered |
| B11 | if at 395:3 | `if err != nil {` | NOT covered |
| B12 | if at 405:3 | `if err != nil {` | NOT covered |
| B13 | if at 414:3 | `if err != nil {` | NOT covered |
