# Branch Test Map: `TestConditionalOrdersIntegration`

본문 무변경이므로 RED 없음. GREEN은 `go test ./internal/official/...` 통과.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 경로별 응답 분기 | 자체 실행 | — (동작 무변경) | yes |
| B2 | 토큰 발급 | 자체 실행 | — | yes |
| B3 | 조건주문 페이지 응답 | 자체 실행 | — | yes |
| B4 | `status=OPEN`이 와이어에 실린다 | 자체 실행 | — | yes |
| B5 | 예상 밖 경로를 보고한다 | 자체 실행 | — | yes |
| B6 | 읽기가 오류 없이 돌아온다 | 자체 실행 | — | yes |
| B7 | 건수·id·SINGLE의 second nil·HasNext | 자체 실행 | — | yes |
| B8 | 첫 다리 trigger가 사상된다 | 자체 실행 | — | yes |
