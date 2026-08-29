# Branch Test Map: `TestReleasedV25BuildRefusesV26Journal`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (129) `if` — if err := current.Close(); err != nil | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |
| B2 | (133) `if` — if !errors.Is(err, ErrSchemaTooNew) | `TestMigrationV26...`, `TestMigrationV27...`, `TestMigrationV30LeavesExistingQuarantinesUnstamped` | yes | yes |

변이 검증: v30 마이그레이션에 `UPDATE ... SET selector_revision=2`를 넣으면
`TestMigrationV30LeavesExistingQuarantinesUnstamped`가 RED가 된다 (관측 완료).
