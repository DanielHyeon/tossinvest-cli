# Branch Test Map: `TestReEnteringAnActiveScopeKeepsTheFirstObservation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | first enter error/state | this test | existing | yes |
| B2 | second enter error | this test | existing | yes |
| B3 | second enter unexpectedly new | this test | existing | yes |
| B4 | first observation changed | this test | existing | yes |
| B5 | active read error | this test | existing | yes |
| B6 | active count mismatch | this test | existing | yes |
