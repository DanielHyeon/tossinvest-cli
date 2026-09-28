# Branch Test Map: `Journal.ReleaseReconcile`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid market/account-wide market refused | `TestReconcileMarketScopeValidation` | pending | pending |
| B2 | KR request cannot release US or global | cross-market/global release tests | pending | pending |
| B3 | expected cause mismatch preserves row | `TestExpectCauseReleasesOnlyYourOwnState` | existing | pending |
| B4 | exact market releases only its row | `TestMarketScopedReconcilesEnterReadAndReleaseIndependently` | pending | pending |
| B5 | evidence validation | existing release-evidence test | existing | yes |
| B6 | transaction begin | storage contract | existing | yes |
| B7 | exact row absent | cross-market refusal | yes | yes |
| B8 | scan error | query contract | existing | yes |
| B9 | expected cause mismatch | existing cause ownership test | existing | yes |
| B10 | update failure | transaction contract | existing | yes |
| B11 | commit failure | transaction contract | existing | yes |
