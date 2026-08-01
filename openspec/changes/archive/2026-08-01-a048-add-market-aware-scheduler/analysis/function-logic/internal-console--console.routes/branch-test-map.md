# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | untrusted remote route set still adds login/logout and all pages remain session gated | existing remote suite + `TestMarketScheduleIsAuthenticatedReadOnlyAndHasNoFreeFormControls` | existing | yes |
| B2 | remote security wrapper remains the final handler and preserves the fixed CSP | existing remote CSP/static suite | existing | yes |
| B3 | market-schedule route is GET/HEAD only, session-gated, and not mutating | `TestMarketScheduleIsAuthenticatedReadOnlyAndHasNoFreeFormControls` + full static route capability suite | yes | yes |
