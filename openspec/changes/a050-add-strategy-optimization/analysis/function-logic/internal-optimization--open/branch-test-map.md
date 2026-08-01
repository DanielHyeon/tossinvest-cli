# Branch Test Map: `Open`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reject blank path | `TestOpenRejectsInvalidOptions` | yes | yes |
| B2 | reject blank actor | `TestOpenRejectsInvalidOptions` | yes | yes |
| B3 | create private parent | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
| B4 | chmod parent 0700 | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
| B5 | default busy timeout | lifecycle tests | existing | existing |
| B6 | SQLite open error | Open fault path | n/a | n/a |
| B7 | ping error | Open fault path | n/a | n/a |
| B8 | initial DB file permissions | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
| B9 | init/migration failure | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
| B10 | defaults for clock/random | lifecycle tests | existing | existing |
| B11 | sidecar permissions | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
| B12 | successful open | `TestAppliedSnapshotAndConsumedCapabilitySurviveProcessReopen` | existing | existing |
