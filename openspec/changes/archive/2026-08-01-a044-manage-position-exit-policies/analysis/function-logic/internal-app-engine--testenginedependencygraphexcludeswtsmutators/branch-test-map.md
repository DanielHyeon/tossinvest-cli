# Branch Test Map: `TestEngineDependencyGraphExcludesWTSMutators`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | inspect forbidden packages | `this test` | yes | yes |
| B2 | fail if a forbidden package is reachable | `this test` | yes | yes |
| B3 | inspect mandatory safety dependencies | `this test` | yes | yes |
| B4 | fail if an authority dependency disappears | `this test` | yes | yes |
