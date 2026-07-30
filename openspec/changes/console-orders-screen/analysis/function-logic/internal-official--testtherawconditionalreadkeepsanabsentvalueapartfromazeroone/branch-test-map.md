# Branch Test Map: `TestTheRawConditionalReadKeepsAnAbsentValueApartFromAZeroOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 경로별 응답 분기 | 자체 실행 | yes (`ConditionalOrdersRaw` 부재로 컴파일 실패) | yes |
| B2 | 토큰 발급 | 자체 실행 | yes | yes |
| B3 | null 필드를 포함한 페이지 응답 | 자체 실행 | yes | yes |
| B4 | 예상 밖 경로 보고 | 자체 실행 | yes | yes |
| B5 | 읽기가 오류 없이 돌아온다 | 자체 실행 | yes | yes |
| B6 | 조건주문 1건 | 자체 실행 | yes | yes |
| B7 | 필드 기대표 순회 | 자체 실행 | yes | yes |
| B8 | `orderPrice` null → `""`, 실제 값은 그대로 | 자체 실행 | yes | yes |
| B9 | 페이지 경계(`hasNext`, `nextCursor`) 보존 | 자체 실행 | yes | yes |
