# Branch Test Map: `TestOrdersFilterEmpty`

base revision의 evidence다. 단언 8개는 HEAD의 후신에 그대로 남아 있으므로 GREEN은
후신 테스트의 통과로 확인한다(`go test ./internal/official/...`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 경로별 응답 분기 | HEAD 후신 `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne` | — (단언 무변경) | yes |
| B2 | 토큰 발급 | 동상 | — | yes |
| B3 | 빈 주문 페이지 응답 | 동상 | — | yes |
| B4 | 여섯 키 순회 | 동상 | — | yes |
| B5 | 어떤 키도 클라이언트가 지어내지 않는다 | 동상 | — | yes |
| B6 | 예상 밖 경로는 404 | 동상 | — | yes |
| B7 | `Orders`가 오류 없이 돌아온다 | 동상 | — | yes |
| B8 | 결과 0건 | 동상 | — | yes |
