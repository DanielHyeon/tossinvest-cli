# Branch Test Map: `Journal.ActivateEntryLossLock`

Post-edit, from `ast.json` (72–117, 10 branches).

| Branch | Line | Test | Kind |
|---|---|---|---|
| B1 | 73 | lock tests (nil journal is unreachable from a real handle; guard only) | declared |
| B2 | 76 | `TestEntryLossLockSchemaRefusesValuesTheRuleCannotMatch`, `TestEntryLossLockAccountFormIsTheDecisionBuildersForm` | behavioural |
| B3 | 81 | `TestA066StorageErrorExitsFailClosed` | structural |
| B4 | 86 | `TestA066StorageErrorExitsFailClosed` | structural |
| B5 | 89 | `TestA066ReaffirmAfterTheOperatorLookedMakesTheReleaseStale` (event written, older view stale), `TestEntryLossLockActivationKeepsTheFirstCauseAndIsImmutable` | behavioural |
| B6 | 92 | `TestA066StorageErrorExitsFailClosed` | structural |
| B7 | 96 | `TestA066StorageErrorExitsFailClosed` | structural |
| B8 | 103 | `TestA066AtMostOneOpenLockPerScopeIsEnforcedByTheSchema` (trigger), `TestA066StorageErrorExitsFailClosed` | behavioural + structural |
| B9 | 107 | `TestA066StorageErrorExitsFailClosed` | structural |
| B10 | 110 | `TestA066StorageErrorExitsFailClosed` | structural |
| fallthrough | 117 | `TestA066EntryLossLockReleaseIsOperatorApprovedAuditedAndOpensEntry` (new lock after release) | behavioural |
