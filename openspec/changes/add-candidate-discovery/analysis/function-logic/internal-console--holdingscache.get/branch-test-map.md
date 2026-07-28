# Branch Test Map: `holdingsCache.get`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 브로커 seam 미배선 콘솔에서 포지션 화면을 연다 | `TestThePositionsScreenRendersWithEitherSourceMissing` | — | yes |
| B2 | TTL 내 재요청 0콜 / TTL 경과 시 1콜 / 검증 중 0콜 / 실패 후 TTL 내 재시도 없음 | `TestRefreshingInsideTheTTLCostsNoBrokerCall`, `TestARefreshIsExactlyOneHoldingsCall`, `TestAVerificationInProgressSuspendsTheRefresh`, `TestAFailingBrokerIsNotRetriedOnEveryPageLoad` | — | yes |
