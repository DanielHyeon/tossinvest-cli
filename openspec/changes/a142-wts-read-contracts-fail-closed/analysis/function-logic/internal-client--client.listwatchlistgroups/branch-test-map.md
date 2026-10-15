# Branch Test Map: `Client.ListWatchlistGroups`

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | session failure | Static preserved branch; no isolated fixture receipt | not claimed | not isolated |
| B2 | bulk GET/decode error | TestListWatchlistRejectsInvalidCollections wrong-type cases | not isolated for this branch | named test PASS |

Named-test PASS is fixture evidence, not instrumented branch coverage. Historical RED terminal output was not retained; retrospective baseline failures and independent guard mutant logs are identified separately in analysis/verification-receipt.md. Static unchanged branches are explicitly not individually proven.

The new helper mapBulkWatchlistGroups has no pre-existing function body to analyze; existing-function pre-edit rule does not create a fictitious baseline for it. Static inspection and TestListWatchlistFromFixtures/TestListWatchlistRejectsInvalidCollections cover its required arrays, ordering and mapping; watchlists/items independent mutants fail. TestWatchlistMutationContract and TestWatchlistMutationSendsXSRF cover their named existing mutation paths, not all mutation operations.
