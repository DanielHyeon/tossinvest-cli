# Branch Test Map: `TestOpenReadOnlyMissingDatabaseCreatesNothing`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing DB returns typed error and no partial reader | this test | read-only opener absent at `948e721` | PASS |
| B2 | missing DB attempt creates no parent directory | this test | read-only opener absent at `948e721` | PASS |
