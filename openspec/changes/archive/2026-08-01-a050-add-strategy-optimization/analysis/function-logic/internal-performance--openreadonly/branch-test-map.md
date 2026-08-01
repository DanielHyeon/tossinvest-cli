# Branch Test Map: `OpenReadOnly`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing path and active WAL refuse capability without creating artifacts | `TestOpenReadOnlyMissingDatabaseCreatesNothing`, `TestOpenReadOnlyRefusesActiveWALRatherThanIgnoringIt` | safe opener absent at `948e721` | PASS |
| B2 | malformed/non-queryable DB refuses capability | read-only malformed fixture coverage | safe opener absent at `948e721` | PASS |
| B3 | exact current schema is accepted while incompatible versions select refusal cases | query success plus schema mismatch tests | safe opener absent at `948e721` | PASS |
| B4 | older schema is rejected and not migrated | `TestOpenReadOnlyRejectsOldSchemaWithoutMigrating` | safe opener absent at `948e721` | PASS |
| B5 | newer schema is rejected | performance newer-schema compatibility coverage | safe opener absent at `948e721` | PASS |
| B6 | WAL/main-file identity drift during validation is rejected | active WAL and immutable identity tests | safe opener absent at `948e721` | PASS |
