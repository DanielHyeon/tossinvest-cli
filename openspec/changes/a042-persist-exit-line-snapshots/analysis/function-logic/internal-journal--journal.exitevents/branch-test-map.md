# Branch Test Map: `Journal.ExitEvents`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query failure | storage fault test | yes | pending |
| B2 | row scan success/failure | v10 event roundtrip | yes | pending |
| B3 | rows.Err | storage fault test | yes | pending |
| B4 | ordered successful read | event ordering test | yes | pending |
| B5 | unknown or partial arm-suppression evidence | `TestExitEventReadRejectsForgedArmSuppressionEvidence` | yes | yes |
| B6 | reason deleted, source forged, effective missing/swapped, nonorderable suppression or armed action forged | event forgery subtest matrix | yes | yes |
| B7 | one flattened v10 column on legacy lifecycle event | `TestLegacyEventRequiresEveryV10ColumnToBeNull` all-column table | yes | yes |
| B8 | required flattened evaluation field NULL or any scalar mismatched | `TestEvaluatedEventRequiresExactFlattenedTuple` | yes | yes |
