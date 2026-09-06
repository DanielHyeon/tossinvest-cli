# Branch Test Map: `NewOfficialAdapter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | an unverified or unsealed policy yields no adapter | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B2 | a missing transport or budget yields no adapter | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
| B3 | OpenDART without a credential provider yields no adapter | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` | existing coverage | PASS |
