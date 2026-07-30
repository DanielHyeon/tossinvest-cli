# Branch Test Map: `Client.accountsLocked`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | HTTP 429/cancellation/malformed 2xx leaves cache unresolved and later discovery can proceed | `TestFailedPublicDiscoveryDoesNotPrimeTheSequence`, `TestMalformedPublicDiscoveryDoesNotPrimeTheSequence`, `TestCancelledPublicDiscoveryUnlocksTheNextScopedDiscovery` | behavior characterized | yes |
| B2 | Zero cache is primed from one positive first record; explicit positive is preserved | `TestAccountsPrimesTheSequenceForTheNextScopedRead`, `TestAccountsPreservesAnExplicitPositiveSequence` | duplicate 429 in RED | yes |
| B3 | Non-priming paths preserve explicit/invalid state or validate implicit state | `TestAccountsPreservesAnExplicitPositiveSequence`, `TestExplicitNegativeSequenceIsNeverSent`, `TestImplicitAccountSequenceDriftIsRejected` | provenance was not represented | yes |
| B4 | Implicit positive selection enters later-response validation | `TestImplicitAccountSequenceDriftIsRejected` | drift returned a conflicting list | yes |
| B5 | Later public discovery cannot be empty, invalid, or different from an implicit selection | `TestImplicitAccountSequenceDriftIsRejected` table (`different_positive`, `zero`, `negative`, `empty`) | invalid drift variants missing | yes |
