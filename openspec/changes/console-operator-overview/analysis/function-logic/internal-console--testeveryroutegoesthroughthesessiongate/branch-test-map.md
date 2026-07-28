# Branch Test Map: `TestEveryRouteGoesThroughTheSessionGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표 전체 순회 | 이 테스트 자신 + `TestEveryRouteRefusesARequestWithoutTheSessionToken`(런타임) | — | yes |
| B2 | session0 없는 등록 | 래퍼 제거 변이 | yes | yes |
| B3 | 추출기가 표를 덜 읽음 | 파일 범위 축소 변이(16으로 떨어진다) | yes | yes |
