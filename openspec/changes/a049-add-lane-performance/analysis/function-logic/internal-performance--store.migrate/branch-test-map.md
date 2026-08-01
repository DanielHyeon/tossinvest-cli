# Branch Test Map: `Store.migrate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | newer schema refused unchanged | `TestStoreRefusesANewerDerivedSchema` | existing | existing |
| B2 | current schema reopen is no-op | replay/restart tests | yes | yes |
| B3 | SIGKILL after DDL/version but before commit leaks nothing | `TestPerformanceMigrationSIGKILLPhasesAreAllOrNone` | yes | yes |
| B4 | successful migration publishes schema+version | `TestStoreSchemaIsSeparateAppendOnlyAndVersioned` | existing | existing |
| B5 | DDL failure/SIGKILL leaks no table | failed migration + phase crash tests | yes | yes |
| B6 | version phase SIGKILL leaks no version/table | phase crash test | yes | yes |
| B7 | commit publishes schema and version together | migration tests | no — existing | yes |
