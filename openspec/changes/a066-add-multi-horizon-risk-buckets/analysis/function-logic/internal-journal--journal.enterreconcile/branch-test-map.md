# Branch Test Map: `Journal.EnterReconcile`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reject invalid/market-without-symbol requests | `TestReconcileMarketScopeValidation` | pending | pending |
| B2 | legacy global row blocks KR/US entry | `TestGlobalReconcileScopeBlocksMarketEntryWithoutBeingReleasedByIt` | pending | pending |
| B3 | KR and US rows coexist, same-market re-entry is idempotent | `TestMarketScopedReconcilesEnterReadAndReleaseIndependently` | pending | pending |
| B4 | v24 indexes/triggers reject overlapping direct writes | migration/index tests | pending | pending |
| B5 | evidence validation | existing evidence test | existing | yes |
| B6 | transaction begin | journal storage contract | existing | yes |
| B7 | existing-row scan switch | re-entry tests | yes | yes |
| B8 | existing row | global/same-market tests | yes | yes |
| B9 | no row | first-entry tests | yes | yes |
| B10 | scan error | query contract | existing | yes |
| B11 | derive optional ID | KR/US distinct-ID assertion | yes | yes |
| B12 | insert error | overlap trigger tests | yes | yes |
| B13 | commit error | transaction contract | existing | yes |
