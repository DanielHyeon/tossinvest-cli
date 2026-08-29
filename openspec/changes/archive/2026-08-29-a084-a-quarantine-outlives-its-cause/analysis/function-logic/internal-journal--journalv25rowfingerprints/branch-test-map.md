# Branch Test Map: `journalV25RowFingerprints`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (183) `if` — if tables == nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B2 | (187) `if` — if err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B3 | (190) `for` — for rows.Next() | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B4 | (192) `if` — if err := rows.Scan(&table); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B5 | (198) `if` — if err := rows.Close(); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B6 | (203) `range` — for _, table := range tables | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B7 | (206) `if` — if err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B8 | (210) `if` — if err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B9 | (226) `for` — for rows.Next() | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B10 | (229) `range` — for index := range values | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B11 | (232) `if` — if err := rows.Scan(targets...); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B12 | (237) `range` — for _, value := range values | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B13 | (238) `type-switch` — switch typed := value.(type) | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B14 | (239) `case` — case nil: | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B15 | (241) `case` — case int64: | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B16 | (243) `case` — case float64: | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B17 | (245) `case` — case string: | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B18 | (247) `case` — case []byte: | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B19 | (249) `case` — default: | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B20 | (255) `if` — if err := rows.Close(); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |

변이 검증: v30 마이그레이션에 `UPDATE ... SET selector_revision=2`를 넣으면
`TestMigrationV30LeavesExistingQuarantinesUnstamped`가 RED가 된다 (관측 완료).
