# Branch Test Map: `TestMigrationV27AddsPairedWeeklyAuthorityWithoutChangingV26Rows`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (18) `if` — if err := old.Close(); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B2 | (22) `if` — if err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B3 | (26) `if` — if version, err := current.SchemaVersion(context.Background()); err != nil ||  | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B4 | (30) `range` — for table, want := range before | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B5 | (31) `if` — if after[table] != want | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B6 | (36) `range` — for _, name := range []string{"strategy_weekly_reservation_scopes", "strategy_ | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B7 | (38) `if` — if err := current.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type=' | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |

변이 검증: v30 마이그레이션에 `UPDATE ... SET selector_revision=2`를 넣으면
`TestMigrationV30LeavesExistingQuarantinesUnstamped`가 RED가 된다 (관측 완료).
