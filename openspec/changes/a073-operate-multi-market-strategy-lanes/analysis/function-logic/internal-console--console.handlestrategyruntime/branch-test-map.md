# Branch Test Map: `Console.handleStrategyRuntime`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | POST/PUT/PATCH/DELETE are refused | `TestStrategyRuntimeStatusIsAuthenticatedGETOnlyAndHasNoInputSurface` | pending | pending |
| B2 | nil reader renders KR+US dormant OFF/UNOBSERVED/UNWIRED | `TestStrategyRuntimeDormantPairIsHonest` | pending | pending |
| B3 | read/validation failure renders typed unknown without raw error | `TestStrategyRuntimeReaderFailureFailsClosedForBothMarkets` | pending | pending |
| B4 | KR current and US unavailable remain independent | `TestStrategyRuntimeMarketsRenderIndependently` | pending | pending |
