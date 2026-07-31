# Branch Test Map: `ReadOnly.AccountExitEvents`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nonpositive limit | existing account-view test | yes | pending |
| B2 | query error | read-only storage test | yes | pending |
| B3 | row scan | v10 account event roundtrip | yes | pending |
| B4 | rows.Err | read-only storage test | yes | pending |
| B5 | newest window ordering | existing account-view ordering | yes | pending |
| B6 | forged arm-suppression enum/evidence | `TestExitEventReadRejectsForgedArmSuppressionEvidence` | yes | yes |
