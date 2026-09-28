# Branch Test Map: `TestMigrationV21AddsOnlyNullableConsumedSnapshotReference`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | open historical schema | this test | existing | yes |
| B2 | seed failure | this test | existing | yes |
| B3 | close failure | this test | existing | yes |
| B4 | migrate failure | this test | yes | yes |
| B5 | version mismatch | this test | yes | yes |
| B6 | column query error | this test | existing | yes |
| B7 | column loop | this test | existing | yes |
| B8 | scan error | this test | existing | yes |
| B9 | rows error | this test | existing | yes |
| B10 | nullable/default mismatch | this test | yes | yes |
