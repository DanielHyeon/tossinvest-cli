# Branch Test Map: `openConsolePerformanceCapabilities`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing/invalid DB open returns no partial authority and creates/migrates nothing | `TestConsolePerformanceCapabilitiesFailWithoutPartialReadAuthority`, performance `OpenReadOnly` tests | `performance.Open` created/migrated the DB | pending |
