# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | remote-mode branch preserves login/logout registration | existing remote route tests | existing | existing |
| B2 | static read/write wrapper classification remains valid | existing static route tests | existing | existing |
| B3 | authenticated POST `/positions` returns 405, GET remains 200 | `TestTradingViewsAreInputFreeAndRejectPOST` | no | no |
