# Branch Test Map: `originOf`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | journal unreadable | TestTheOriginColumnSaysUnknownWhenTheLedgerCouldNotBeRead | existing | yes |
| B2 | exact hit/miss | TestTheOriginColumnTellsAnEngineOrderFromAnyOther | RED after scoped conversion | yes |
