# Branch Test Map: `TestTheUSAdvisorySaysItsClosedResponseIsUnmeasured`

본문 무변경이므로 RED 없음. GREEN은 `go test ./internal/verifylive/...` 통과.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | US 권고문이 KR의 측정 코드를 인용하지 않는다 | 자체 실행 | — (동작 무변경) | yes |
| B2 | US 권고문이 미측정임을 말한다 | 자체 실행 | — | yes |
| B3 | KR 권고문은 측정 코드를 유지한다 | 자체 실행 | — | yes |
