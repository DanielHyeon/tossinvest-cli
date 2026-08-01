# Branch Test Map: `secureSQLiteFiles`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | chmod failure | filesystem fault path | n/a | n/a |
| B2 | optional sidecar missing | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
