# Branch Test Map: `WithAccountSeq`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Positive remains authoritative; negative refuses without discovery/header; default zero stays lazy | `TestAccountsPreservesAnExplicitPositiveSequence`, `TestExplicitNegativeSequenceIsNeverSent`, `TestEnsureAccountSeqLazyResolutionAndCaching` | positive provenance and negative refusal gaps identified in review | yes |
