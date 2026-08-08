# Branch Test Map: `buildGateway`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid gateway inputs refuse | existing gateway construction tests | baseline | pass |
| B2 | restore failure refuses | existing reconciliation recovery tests | baseline | pass |
| B3 | `execgw.New` failure refuses | existing gateway tests | baseline | pass |
