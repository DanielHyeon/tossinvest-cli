# Branch Test Map: `TestMigrationV26AddsPairedFirstLegAuthorityWithoutChangingV25Rows`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (21) `if` — if err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B2 | (26) `if` — if err := insertStrategyDispatchLease(old, kr, StrategyDispatchLeaseClaimed, " | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B3 | (29) `if` — if err := insertStrategyDispatchLease(old, us, StrategyDispatchLeaseIssued, "" | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B4 | (32) `if` — if _, err := old.db.Exec(`INSERT INTO strategy_dispatch_outcomes( | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B5 | (41) `if` — if err := old.Close(); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B6 | (47) `if` — if err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B7 | (51) `if` — if version, err := current.SchemaVersion(context.Background()); err != nil ||  | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B8 | (55) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM strategy_dispatch_market_a | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B9 | (61) `range` — for table, target := range map[string]*int | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B10 | (66) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(target); e | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B11 | (70) `if` — if leases != 2 || outcomes != 1 || qFinal != 2 || ownerEpochs != 1 || currentO | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B12 | (76) `if` — if err := current.db.QueryRow(`SELECT state,revision FROM strategy_dispatch_le | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B13 | (81) `range` — for table, want := range beforeRows | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B14 | (82) `if` — if got := afterRows[table]; got != want | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B15 | (86) `range` — for _, object := range []struct{ kind, name string } | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B16 | (94) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type=? | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |

변이 검증: v30 마이그레이션에 `UPDATE ... SET selector_revision=2`를 넣으면
`TestMigrationV30LeavesExistingQuarantinesUnstamped`가 RED가 된다 (관측 완료).
