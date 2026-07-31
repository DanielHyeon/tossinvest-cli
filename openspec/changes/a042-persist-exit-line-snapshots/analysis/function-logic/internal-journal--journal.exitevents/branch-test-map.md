# Branch Test Map: `Journal.ExitEvents`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query failure | storage fault test | yes | pending |
| B2 | row scan success/failure | v10 event roundtrip | yes | pending |
| B3 | rows.Err | storage fault test | yes | pending |
| B4 | ordered successful read | event ordering test | yes | pending |
| B5 | unknown or partial arm-suppression evidence | `TestExitEventReadRejectsForgedArmSuppressionEvidence` | yes | yes |
