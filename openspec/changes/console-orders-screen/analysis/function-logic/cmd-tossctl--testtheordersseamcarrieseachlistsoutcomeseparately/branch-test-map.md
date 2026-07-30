# Branch Test Map: `TestTheOrdersSeamCarriesEachListsOutcomeSeparately`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 라우팅 switch | 이 테스트 | no (fixture) | yes |
| B2 | 토큰 응답 | 동일 | no (fixture) | yes |
| B3 | OPEN 성공 | 동일 | yes | yes |
| B4 | CLOSED 성공(hasNext) | 동일 | yes | yes |
| B5 | 조건주문 429 | 동일 | yes | yes |
| B6 | 404 기본 | 동일 | no (fixture) | yes |
| B7 | 반쪽 판독에도 함수 에러 없음 | 동일 | yes (하나로 접는 구현에서 FAIL) | yes |
| B8 | 미체결 생존 | 동일 | yes | yes |
| B9 | `OpenError` 비어 있음 | 동일 | yes | yes |
| B10 | 종결 생존 | 동일 | yes | yes |
| B11 | `ClosedTruncated` 유지 | 동일 | yes | yes |
| B12 | `ConditionalError` 채워짐 | 동일 | yes (삼키는 구현에서 FAIL) | yes |
| B13 | null이 0이 아니라 부재로 도착 | 동일 | yes (`parseDecimal` 경유 시 FAIL) | yes |
