# Branch Test Map: `TestAFailingExitHookRollsBackTheProjectionToo`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | setup journal | self | existing | yes |
| B2 | bind hooks | self | existing | yes |
| B3 | record fill error | self | existing | yes |
| B4 | snapshot rollback | self | existing | yes |
| B5 | Position rollback | self | existing | yes |
| B6 | campaign/exit rollback | self | yes | yes |
| B7 | cleanup/assertion | self | existing | yes |
