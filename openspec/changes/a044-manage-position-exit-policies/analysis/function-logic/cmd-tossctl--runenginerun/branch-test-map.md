# Branch Test Map: `runEngineRun`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil context defaults safely | `engine command suite` | yes | yes |
| B2 | profile/root resolution fails closed | `engine command suite` | yes | yes |
| B3 | engine lock acquisition failure refuses boot | `engine command suite` | yes | yes |
| B4 | assembly failure closes owned context | `engine command suite` | yes | yes |
| B5 | interlock clause details are detected | `engine interlock suite` | yes | yes |
| B6 | each unmet clause is reported | `engine interlock suite` | yes | yes |
| B7 | unverified automation refuses runtime | `engine interlock suite` | yes | yes |
| B8 | verify-lock path is available | `engine command suite` | yes | yes |
| B9 | fresh verify lock refuses engine | `engine command suite` | yes | yes |
| B10 | marker failure warns and continues | `engine marker suite` | yes | yes |
| B11 | marker success reports lifecycle | `engine marker suite` | yes | yes |
| B12 | runtime factory failure stops before endpoint | `engine command suite` | yes | yes |
| B13 | command service requires engine-owned journal | `TestPositionPolicyCommandServiceRequiresEngineOwnedJournal` | yes | yes |
| B14 | endpoint start failure stops before loops | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` | yes | yes |
