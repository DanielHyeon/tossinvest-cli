# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branchless route assembly registers every route; signed download is session- and CSRF-gated | `TestEveryRouteGoesThroughTheSessionGate`, `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`, `TestSignedReleaseDownloadRequiresSessionPostAndCSRF` | signed download route absent | pass |
