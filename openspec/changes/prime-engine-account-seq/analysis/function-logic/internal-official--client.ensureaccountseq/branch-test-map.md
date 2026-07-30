# Branch Test Map: `Client.ensureAccountSeq`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Nonzero atomic fast path avoids mutex and discovery | `TestCachedScopedReadDoesNotWaitForPublicAccountListIO`, `TestExplicitNegativeSequenceIsNeverSent` | slow public I/O blocked cached requests | yes |
| B2 | Fast path distinguishes positive selection from negative refusal | `TestCachedScopedReadDoesNotWaitForPublicAccountListIO`, `TestExplicitNegativeSequenceIsNeverSent` | negative accepted in RED | yes |
| B3 | A concurrent discovery can select while later callers wait | `TestConcurrentScopedFirstUseSharesOneDiscovery`, `TestScopedReadWaitsForAnInflightPublicAccountDiscovery` | duplicate request failed | yes |
| B4 | Post-lock recheck returns the concurrent positive selection | same deterministic waiter tests | duplicate request failed | yes |
| B5 | Account-list 429/error/cancel/malformed response remains a refusal and unlocks | `TestFailedPublicDiscoveryDoesNotPrimeTheSequence`, `TestMalformedPublicDiscoveryDoesNotPrimeTheSequence`, `TestCancelledPublicDiscoveryUnlocksTheNextScopedDiscovery` | contention failed as expected | yes |
| B6 | Empty account list refuses without caching | `TestEmptyDiscoveryCannotProduceAScopedHeader` | baseline semantics | yes |
| B7 | Non-numeric adapted ID refuses | integer API schema makes this unreachable from decoded `apiAccount`; parse guard retained | structural guard | AST verified |
| B8 | Zero or negative discovered first sequence refuses without sending an account header | `TestDiscoveredNonpositiveSequenceIsNeverSent` | yes | yes |
| B9 | Helper selection must equal its adapted first record | structural invariant: `accountsLocked` stores that exact decoded first integer under the same mutex; the guard fails closed if future adaptation changes | not runtime-reachable in current typed API | AST verified |
