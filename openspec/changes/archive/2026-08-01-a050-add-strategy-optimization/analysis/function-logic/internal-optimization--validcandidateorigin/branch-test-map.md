# Branch Test Map: `validCandidateOrigin`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | server preset pair | lifecycle test | existing | existing |
| B2 | evidence pair | evidence lifecycle test | existing | existing |
| B3 | rollback pair | rollback lifecycle test | existing | existing |
| B4 | forged source | `TestApplyFailsClosedForTamperedCandidateMetadata` | yes | yes |
