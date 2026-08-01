# Branch Test Map: `ReadOnly.Close`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil close succeeds; initialized close marks capability closed and later read fails | cmd capability lifecycle and read-only lifecycle tests | lifecycle absent at `948e721` | PASS |
