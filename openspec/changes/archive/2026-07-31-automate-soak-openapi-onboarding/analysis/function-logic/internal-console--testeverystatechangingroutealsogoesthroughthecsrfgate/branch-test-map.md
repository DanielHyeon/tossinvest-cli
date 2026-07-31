# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | iterate registered routes | this test | existing | passed |
| B2 | classify gate relationship | this test | existing | passed |
| B3 | state-changing route lacks CSRF | static logic | existing | passed |
| B4 | read route has CSRF | static logic | existing | passed |
| B5 | iterate required routes | this test | existing | passed |
| B6 | required route missing | static logic | existing | passed |
