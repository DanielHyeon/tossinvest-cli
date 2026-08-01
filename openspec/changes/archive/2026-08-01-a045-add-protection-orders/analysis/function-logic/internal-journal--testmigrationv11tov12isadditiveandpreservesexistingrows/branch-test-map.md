# Branch Test Map: `TestMigrationV11ToV12IsAdditiveAndPreservesExistingRows`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | establish v11 fixture | self | existing coverage | pass |
| B2 | migrate exactly to v12 | self | current schema leaked | pass |
| B3 | verify preserved legacy row | self | existing coverage | pass |
| B4 | enumerate v12 objects | self | existing coverage | pass |
| B5 | query catalog | self | existing coverage | pass |
| B6 | require each additive object | self | existing coverage | pass |
