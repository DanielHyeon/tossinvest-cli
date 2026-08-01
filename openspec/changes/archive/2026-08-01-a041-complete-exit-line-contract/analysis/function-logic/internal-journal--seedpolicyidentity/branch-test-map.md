# Branch Test Map: `seedPolicyIdentity`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact legacy omission/claim | runtime preservation test | yes | yes |
| B2 | changed digest under same ID/version | seed mismatch test | yes | yes |
| B3 | adopted RUNNER variant | existing adoption policy tests | existing | yes |
| B4 | legacy resolver error | unknown legacy identity test | yes | yes |
| B5 | partial claimed identity fails validation | seed mismatch test | yes | yes |
| B6 | complete but different tuple fails exact comparison | seed mismatch test | yes | yes |
