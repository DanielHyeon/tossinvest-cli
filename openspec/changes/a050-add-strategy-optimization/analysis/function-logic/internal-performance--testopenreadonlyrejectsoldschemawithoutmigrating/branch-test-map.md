# Branch Test Map: `TestOpenReadOnlyRejectsOldSchemaWithoutMigrating`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | old-schema fixture DB opens | this test | read-only seam absent at `948e721` | PASS |
| B2 | fixture `user_version=0` is written | this test | read-only seam absent at `948e721` | PASS |
| B3 | fixture closes before read-only open | this test | read-only seam absent at `948e721` | PASS |
| B4 | product opener returns typed old-schema error and nil reader | this test | read-only seam absent at `948e721` | PASS |
| B5 | raw mode-ro verifier opens | this test | read-only seam absent at `948e721` | PASS |
| B6 | persisted user version query succeeds | this test | read-only seam absent at `948e721` | PASS |
| B7 | persisted table count query succeeds | this test | read-only seam absent at `948e721` | PASS |
| B8 | opener performed neither version migration nor table creation | this test | read-only seam absent at `948e721` | PASS |
