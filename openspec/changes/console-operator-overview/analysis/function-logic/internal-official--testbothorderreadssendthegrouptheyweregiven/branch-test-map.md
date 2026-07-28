# Branch Test Map: `TestBothOrderReadsSendTheGroupTheyWereGiven`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `OPEN`/`CLOSED` 두 그룹 | 자체 실행 | yes (`OrdersRaw` 부재로 컴파일 실패) | yes |
| B2 | 경로별 응답 분기 | 자체 실행 | yes | yes |
| B3 | 토큰 발급 | 자체 실행 | yes | yes |
| B4 | 서버가 본 `status` 기록 | 자체 실행 | yes | yes |
| B5 | 예상 밖 경로는 404 | 자체 실행 | yes | yes |
| B6 | `Orders`가 성공한다 | 자체 실행 | yes | yes |
| B7 | `OrdersRaw`가 성공한다 | 자체 실행 | yes | yes |
| B8 | 본 값들을 순회한다 | 자체 실행 | yes | yes |
| B9 | 준 그룹이 그대로 실린다(두 읽기 모두) | 자체 실행 | yes | yes |
| B10 | 두 읽기가 정확히 2요청 — 계좌 해석이 공유된다 | 자체 실행 | yes | yes |
