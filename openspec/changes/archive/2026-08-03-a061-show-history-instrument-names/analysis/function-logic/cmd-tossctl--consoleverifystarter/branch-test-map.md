# Branch Test Map: `consoleVerifyStarter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | execution flock unavailable | existing console verification exclusion tests | existing | existing |
| B2 | market record path unavailable | existing console verify path tests | existing | existing |
| B3 | prior evidence unreadable | existing corrupt-record tests | existing | existing |
| B4 | profile intent marker path fails | profile path tests | existing | existing |
| B5 | rate budget busy until context ends while broker stays unbuilt | `TestA061RunAndConsoleVerificationDoNotBuildBrokerWhileTheProfileBudgetIsOccupied` | yes | yes |
| B6 | official broker/account resolution fails | existing console verify failure tests | existing | existing |
| B7 | recorder open fails | existing verify record tests | existing | existing |
| B8 | runner construction refuses invalid setup | existing console verification tests | existing | existing |
| B9 | canceled runner preserves recorded work | existing console shutdown/cancellation tests | existing | existing |
| tail | successful supervised runner holds budget through completion | full console verification suite | existing | yes |
