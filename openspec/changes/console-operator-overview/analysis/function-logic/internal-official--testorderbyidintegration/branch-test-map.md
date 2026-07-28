# Branch Test Map: `TestOrderByIDIntegration`

본문 무변경이므로 RED 없음. GREEN은 `go test ./internal/official/...` 통과.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 경로별 응답 분기 | 자체 실행 | — (동작 무변경) | yes |
| B2 | 토큰 발급 | 자체 실행 | — | yes |
| B3 | 단건 주문 경로 + 헤더 기록 | 자체 실행 | — | yes |
| B4 | 예상 밖 경로는 404 | 자체 실행 | — | yes |
| B5 | 읽기가 오류 없이 돌아온다 | 자체 실행 | — | yes |
| B6 | ID 사상 | 자체 실행 | — | yes |
| B7 | Symbol 사상 | 자체 실행 | — | yes |
| B8 | Price 사상 | 자체 실행 | — | yes |
| B9 | 계좌 헤더가 실린다 | 자체 실행 | — | yes |
