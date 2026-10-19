# Branch Test Map: `assertConsoleOnlyCallsDormantDescriptor`

| Branch | Scenario | Test | RED evidence | GREEN evidence |
|---|---|---|---|---|
| B1 | Non-call AST nodes are ignored. | A047 repository-wide focused test | See preserved gate1/import-guard RED; not every historical branch was individually induced. | `verification/import-guard-green.log` |
| B2 | Non-selector calls are ignored. | A047 repository-wide focused test | See preserved gate1/import-guard RED; not every historical branch was individually induced. | `verification/import-guard-green.log` |
| B3 | Calls not using strategyengine identifier are ignored. | A047 repository-wide focused test | See preserved gate1/import-guard RED; not every historical branch was individually induced. | `verification/import-guard-green.log` |
| B4 | Non-descriptor call reports a test error. | A047 repository-wide focused test | See preserved gate1/import-guard RED; not every historical branch was individually induced. | `verification/import-guard-green.log` |
