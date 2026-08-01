# Branch Test Map: `verifyAppendOnlyTriggers`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all required trigger names are queried | `TestSchemaIsVersionedAndAppendOnly` | exact verification absent | PASS |
| B2 | legacy pre-install check permits missing trigger only | migration compatibility test | exact verification absent | PASS |
| B3 | post-install missing trigger is rejected | `TestOpenRejectsSameNameNoOpOrDriftedAppendOnlyTrigger` missing case | exact verification absent | PASS |
| B4 | sqlite schema read error propagates | migration error coverage | exact verification absent | PASS |
| B5 | no-op/drifted same-name definition is rejected | `TestOpenRejectsSameNameNoOpOrDriftedAppendOnlyTrigger` | name-only presence could bypass protection | PASS |
