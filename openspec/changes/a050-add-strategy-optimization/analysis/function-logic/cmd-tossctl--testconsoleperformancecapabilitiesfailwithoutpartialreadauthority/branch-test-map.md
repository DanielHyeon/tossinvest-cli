# Branch Test Map: `TestConsolePerformanceCapabilitiesFailWithoutPartialReadAuthority`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing profile database returns the typed missing error and zero read seams | this test | read-only constructor absent | yes |
| B2 | missing profile database is not created as a side effect | this test | read-only constructor absent | yes |
| B3 | non-directory profile fixture creation succeeds | this test | read-only constructor absent | yes |
| B4 | invalid profile path returns an error and no partial performance/evidence capability | this test | read-only constructor absent | yes |
