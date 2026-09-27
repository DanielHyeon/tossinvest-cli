# Branch Test Map: `loadProductionRiskEntries`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 340:2 — `if !ok {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B2 | if at 343:2 — `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B3 | if at 352:2 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B4 | if at 357:2 — `if err := db.PingContext(ctx); err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B5 | if at 361:2 — `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \ | package suites (pre-edit) | n/a (pre-edit map) | \ at `f2decd0a` |
| B6 | if at 365:2 — `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B7 | if at 376:2 — `if authorityObserved.After(scope.AsOf) \ | package suites (pre-edit) | n/a (pre-edit map) | \ at `f2decd0a` |
| B8 | range at 380:2 — `for _, dimension := range requiredDimensions {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B9 | if at 382:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B10 | if at 386:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | covered at `f2decd0a` |
| B11 | if at 395:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B12 | if at 405:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
| B13 | if at 414:3 — `if err != nil {` | package suites (pre-edit) | n/a (pre-edit map) | NOT covered at `f2decd0a` |
