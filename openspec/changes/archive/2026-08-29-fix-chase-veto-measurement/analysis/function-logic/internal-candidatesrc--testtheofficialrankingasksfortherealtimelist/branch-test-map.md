# Branch Test Map: `TestTheOfficialRankingAsksForTheRealtimeList`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 생성 실패 | 자체 실행 | yes (인자 하나가 늘어 컴파일 실패) | yes |
| B2 | 읽기 실패 | 자체 실행 | yes | yes |
| B3 | duration이 realtime | 자체 실행 | — (기존 동작) | yes |
| B4 | type과 market 전달 | 자체 실행 | — (기존 동작) | yes |
