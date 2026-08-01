# Branch Test Map: `Journal.OpenExitStateResults`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query failure is wrapped | `exit snapshot persistence suite` | yes | yes |
| B2 | each current managed row is scanned | `TestPositionPolicyReleaseRemovesExactGenerationFromWorkingSet` | yes | yes |
| B3 | corrupt row fails closed | `exit snapshot corruption suite` | yes | yes |
| B4 | iterator error propagates | `exit snapshot persistence suite` | yes | yes |
