# Branch Test Map: `scanReconcileState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no active row | `TestReleasingNothingIsNotAnError` | existing | pending |
| B2 | select/scan shape includes scope market | market read/history test | pending | pending |
| B3 | time parse failure | existing journal corruption coverage | existing | pending |
| B4 | released and active rows decode | release/history tests | existing | pending |
| B5 | released timestamp parse | release/history tests | existing | yes |
