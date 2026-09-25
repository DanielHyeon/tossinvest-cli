# Branch Test Map: `activeScopeWhere`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | account-wide query remains global | `TestAccountWideAndSymbolScopesAreSeparate` | existing | PASS at HEAD 2026-09-25 |
| B2 | symbol global selects only NULL market | `TestGlobalReconcileScopeBlocksMarketEntryWithoutBeingReleasedByIt` | Wave 1E | PASS at HEAD 2026-09-25 |

Fall-through return (market set → exact `scope_market = ?` predicate) is not an AST branch; it was
listed as `B3` until Wave 2A (2026-09-25), when the checker began refusing rows for branch IDs the AST
does not produce. Its coverage: `TestMarketScopedReconcilesEnterReadAndReleaseIndependently`,
`TestAtomicMarketReleaseDoesNotCrossIntoPeerMarket` (both PASS at HEAD 2026-09-25).
