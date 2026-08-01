# Branch Test Map: `Journal.OpenExitStates`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query failure is wrapped | `exit state suite` | yes | yes |
| B2 | each current managed row is scanned | `TestPositionPolicyReleaseRemovesExactGenerationFromWorkingSet` | yes | yes |
| B3 | row scan failure propagates | `exit state suite` | yes | yes |
| B4 | iterator failure propagates | `exit state suite` | yes | yes |
