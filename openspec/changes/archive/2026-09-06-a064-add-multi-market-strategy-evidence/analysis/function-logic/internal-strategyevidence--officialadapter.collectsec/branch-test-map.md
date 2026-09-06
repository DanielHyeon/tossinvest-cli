# Branch Test Map: `OfficialAdapter.collectSEC`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | a non-ten-digit CIK makes zero requests | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B2 | a failed first page stops the collection | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B3 | a first page outside the frozen shape is refused | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B4 | unparseable recent filings are refused | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B5 | a body demanding more pages than the policy allows is incomplete | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B6 | the frozen two-page fixture is walked and both pages are requested in order | `TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity` | existing coverage | PASS |
| B7 | five spoofed page names — traversal, another CIK, a suffix, an absolute URL, a short ordinal — are refused before the second request | `TestSECHistoricalPageResourceIsNotChosenByTheRemoteBody` | all five were fetched: the guard only checked for a blank name | PASS |
| B8 | a failed historical page stops the collection | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B9 | pages that together exceed the byte budget are refused | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B10 | an unparseable historical page is refused | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B11 | a historical page whose record count contradicts its declaration is refused | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
