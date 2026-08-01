# Branch Test Map: `buildCommonPolicies`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all common profiles receive stable identities | common policy descriptor contract test | no | yes |
| B2 | invalid policy identity is refused | policy identity validation tests | no | yes |
| B3 | table changes under pinned digest panic/fail test | pinned semantic digest test | yes | yes |
