# Branch Test Map: `Client.ListWatchlist`

| Branch | Scenario | Test / evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | grouped-read error propagates | TestListWatchlistRejectsInvalidCollections | baseline already errors on legacy-route absence; not this branch RED | named test PASS |
| B2 | flatten groups in mapped order | TestListWatchlistFromFixtures | retrospective baseline FAIL at legacy endpoint | named test PASS |

Named-test PASS is fixture evidence, not instrumented branch coverage. Historical RED terminal output was not retained; retrospective baseline failures and independent guard mutant logs are identified separately in analysis/verification-receipt.md. Static unchanged branches are explicitly not individually proven.
