# Branch Test Map: `Console.render`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | successful pages receive the restrictive CSP; invalid template still fails before commit | `TestMarketScheduleIsAuthenticatedReadOnlyAndHasNoFreeFormControls` + existing console render suite | yes (CSP absent) | yes |
