# Branch Test Map: `nullable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 문자열은 NULL, 값은 trim되어 저장 | `store_test.go`의 관측 왕복 테스트(부재 필드가 빈 문자열로 돌아온다) | — (동작 무변경) | yes |
