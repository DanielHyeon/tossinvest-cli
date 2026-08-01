# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | remote-mode branch preserves login/logout registration | existing remote route tests | existing | yes |
| B2 | configured remote security wrapper is returned | existing remote route tests | existing | yes |

The unconditional `/positions` registration is covered by `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` and `TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript`: GET remains available and authenticated POST returns 405.
