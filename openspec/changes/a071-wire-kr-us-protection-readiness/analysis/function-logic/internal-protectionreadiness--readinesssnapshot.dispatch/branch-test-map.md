# Branch Test Map: `ReadinessSnapshot.Dispatch`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | valid exact KR/US scope | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | existing account/market-only behavior | pending |
| B2 | quantity below/above signed range | new dispatch substitution test | currently accepted | pending |
| B3 | order/session/trigger/replace substitution | new dispatch substitution test | currently accepted | pending |
| B4 | capability digest substitution | new dispatch substitution test | currently accepted | pending |
| B5 | corrupt KR leaves valid US unchanged | `TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket` | existing | pending |
