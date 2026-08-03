# Branch Test Map: `runVerifyRun`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `--list` preview uses no broker | existing verify list tests | existing | existing |
| B2 | nil context fallback | existing command tests | existing | existing |
| B3 | engine execution lock refusal | existing verify execution-lock tests | existing | existing |
| B4 | record path failure | existing verify path tests | existing | existing |
| B5 | corrupt/unreadable record | existing verify record tests | existing | existing |
| B6 | implicit rerun of measured steps | existing replay refusal tests | existing | existing |
| B7 | profile intent marker path fails | profile path tests | existing | existing |
| B8 | optional metadata owns budget until verification context ends while broker stays unbuilt | `TestA061RunAndConsoleVerificationDoNotBuildBrokerWhileTheProfileBudgetIsOccupied` | yes | yes |
| B9 | official broker/account failure | existing broker failure tests | existing | existing |
| B10 | no explicit holding symbol | existing holding selection tests | existing | existing |
| B11 | US run replaces KR default symbol | existing US verification tests | existing | existing |
| B12 | recorder open failure | existing record tests | existing | existing |
| B13 | runner construction failure | existing verify options tests | existing | existing |
| B14 | interruption preserves evidence and returns cleanly | existing interrupt tests | existing | existing |
| tail | successful run releases marker, rate budget, and execution flock | full verify command suite | existing | yes |
