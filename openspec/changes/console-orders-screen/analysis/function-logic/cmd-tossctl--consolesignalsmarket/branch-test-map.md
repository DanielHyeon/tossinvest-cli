# Branch Test Map: `consoleSignalsMarket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Assess 실패 시장이 목록에서 사라지지 않고 사유와 함께 남는다 | /signals 렌더 테스트 + `TestTheSignalsSeamReadsTheStoreAndCallsNoSource` | yes | yes |

tally 복제 금지는 `TestTheConsoleDoesNotBuildTheDiscoveryTalliesItself`가 정적으로 고정한다(§5 리뷰 P2 ④).
