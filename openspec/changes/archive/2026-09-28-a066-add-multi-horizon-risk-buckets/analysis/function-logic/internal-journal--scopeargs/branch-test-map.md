# Branch Test Map: `scopeArgs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | account-wide | `TestAccountWideAndSymbolScopesAreSeparate` | existing | PASS at HEAD 2026-09-25 |
| B2 | global symbol | `TestGlobalReconcileScopeBlocksMarketEntryWithoutBeingReleasedByIt` | Wave 1E | PASS at HEAD 2026-09-25 |

Fall-through return (market set → `account, symbol, market`) is not an AST branch; it was listed as
`B3` until Wave 2A (2026-09-25). Its coverage: `TestMarketScopedReconcilesEnterReadAndReleaseIndependently`,
`TestAtomicMarketReleaseDoesNotCrossIntoPeerMarket` (both PASS at HEAD 2026-09-25).
