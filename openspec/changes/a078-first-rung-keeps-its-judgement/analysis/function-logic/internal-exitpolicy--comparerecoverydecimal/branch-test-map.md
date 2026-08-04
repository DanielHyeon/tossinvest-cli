# Branch Test Map: `compareRecoveryDecimal`

a062는 이 함수를 편집하지 않았다(줄 범위 인접). 기존 동작을 고정하는 행이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 재계산 쪽 값이 읽히지 않으면 오류가 전파된다 | `internal/exitpolicy/recovery_validation_test.go` | n/a | pass |
| B2 | 저장 쪽 값이 읽히지 않으면 오류가 전파된다 | 같음 | n/a | pass |
