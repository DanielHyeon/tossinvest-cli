# Branch Test Map: `newUpdateCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Invalid output format refuses before discovery | existing update command format tests | baseline | pass |
| B2 | Executable path lookup fails | existing self-update path tests | baseline | pass |
| B3 | Symlink resolution succeeds and replaces the raw executable path | existing install-method tests | baseline | pass |
| B4 | Development build refuses self-replacement | existing development-build test | baseline | pass |
| B5 | Update cache path cannot be resolved | existing update cache tests | baseline | pass |
| B6 | Latest-version spinner returns an error | existing update-check tests | baseline | pass |
| B7 | `--check` returns before the legacy warning or mutation | `TestLegacyUpdateWarnsOnlyAfterCheckOnlyReturnAndBeforeMutation` | warning placement absent | pass |
| B8 | No newer release prints already-current and returns | existing up-to-date test | baseline | pass |
| B9 | Interactive mode enters the confirmation path | existing confirmation tests | baseline | pass |
| B10 | Confirmation result is classified by the switch | existing confirmation tests | baseline | pass |
| B11 | Noninteractive confirmation promotes the run to explicit-yes behavior | existing noninteractive update test | baseline | pass |
| B12 | Confirmation error propagates | existing confirmation-error test | baseline | pass |
| B13 | Operator cancellation returns without replacement | existing cancellation test | baseline | pass |
| B14 | Assumed/confirmed update prints the transition | existing update output test | baseline | pass |
| B15 | Legacy self-update failure is reported and propagated | existing self-update failure test | baseline | pass |
| B16 | Windows binary replacement returns before claiming asynchronous success | existing Windows branch/static coverage | baseline | pass |
