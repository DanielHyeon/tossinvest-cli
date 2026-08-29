# Branch Test Map: `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 토큰 교환은 예산에 세지 않는다 | 이 테스트 | no (fixture) | yes |
| B2 | 라우팅 switch | 동일 | no (fixture) | yes |
| B3 | OPEN 응답 | 동일 | yes | yes |
| B4 | CLOSED 응답(hasNext) | 동일 | yes | yes |
| B5 | status 누락 요청은 400 — 첫 구현이 도달하던 모양 | 동일 | yes (첫 구현에서 실제로 도달) | yes (도달하지 않음) |
| B6 | 조건주문 빈 목록 | 동일 | no (fixture) | yes |
| B7 | 미지정 경로 404 | 동일 | no (fixture) | yes |
| B8 | Orders가 에러 없이 반환 | 동일 | yes | yes |
| B9 | 새로고침 1회 = 브로커 콜 3회 | 동일 | yes (1콜 구현에서 FAIL) | yes |
| B10 | 라이브 콜에 `status=OPEN` | 동일 | yes (D2 개정 전 FAIL) | yes |
| B11 | 라이브 콜에 limit/cursor 없음 | 동일 | yes (`limit=100`만 보내던 구현에서 FAIL) | yes |
| B12 | 두 번째 콜에 `status=CLOSED` | 동일 | yes | yes |
| B13 | CLOSED에 limit=100 | 동일 | yes | yes |
| B14 | 세 번째 콜은 조건주문 | 동일 | yes | yes |
| B15 | 미체결 1건이 그대로 도착 | 동일 | yes | yes |
| B16 | 미체결은 truncated 아님 | 동일 | yes | yes |
| B17 | 종결 1건이 그대로 도착 | 동일 | yes | yes |
| B18 | 종결의 hasNext가 살아남는다 | 동일 | yes | yes |
