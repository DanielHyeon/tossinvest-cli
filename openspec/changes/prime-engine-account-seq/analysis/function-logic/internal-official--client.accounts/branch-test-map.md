# Branch Test Map: `Client.Accounts`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없는 wrapper가 mutex를 해제하며 helper 결과를 그대로 반환한다 | `TestScopedReadWaitsForAnInflightPublicAccountDiscovery`, `TestCancelledPublicDiscoveryUnlocksTheNextScopedDiscovery` | public/scoped duplicate 429 | yes |
