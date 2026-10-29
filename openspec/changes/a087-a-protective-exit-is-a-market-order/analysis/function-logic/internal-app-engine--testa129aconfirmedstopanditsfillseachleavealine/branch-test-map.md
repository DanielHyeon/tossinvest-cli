# Branch Test Map: `TestA129AConfirmedStopAndItsFillsEachLeaveALine`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 제출 수량·cycle 오류 | 이 함수 | source 키 없음 | observed source와 기존 가격 |
| B2 | 확정 줄 개수 | 이 함수 | source 키 없음 | 정확히 한 줄 |
| B3 | 확정 줄 필드 | 이 함수 | price_source 없음 | observed key |
| B4 | 로그 가격과 intent 가격 | 이 함수 | 사본 분기 가능성 | 값 동일 |
| B5 | order id 형식 | 이 함수 | 기존 | 기존 유지 |
| B6 | follow 시작 | 이 함수 | 기존 | 한 건 |
| B7 | 부분 체결 줄 | 이 함수 | 기존 | 한 줄 |
| B8 | 같은 누적 체결 | 이 함수 | 기존 | 추가 줄 없음 |
| B9 | 전량 체결 줄 | 이 함수 | 기존 | terminal true |
| B10 | follow 종료 | 이 함수 | 기존 | 0건 |
| B11 | 종결 뒤 관측 | 이 함수 | 기존 | 추가 줄 없음 |
