# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | enumerate routes | self/static suite | protection route unlisted | pass |
| B2 | classify route verb | self/static suite | protection route unclassified | pass |
| B3 | read-only route case | self/static suite | existing coverage | pass |
| B4 | mutation route case | self/static suite | missing CSRF classification | pass |
| B5 | enumerate CSRF seams | self/static suite | protection seam missing | pass |
| B6 | require mutation membership | self/static suite | bypass detected | pass |
