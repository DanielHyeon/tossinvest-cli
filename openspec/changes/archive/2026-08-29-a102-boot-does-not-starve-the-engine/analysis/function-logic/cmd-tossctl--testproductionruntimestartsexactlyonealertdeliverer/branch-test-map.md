# Branch Test Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

Source: `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` (20-29). AST 기준 branches **2** / returns 0.

| Branch | 위치 | 지는 테스트 | GREEN 실측 |
|---|---|---|---|
| B1 | `:22` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |
| B2 | `:26` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |

이 함수는 테스트다 — 자기 분기를 자기가 실행한다. a102 GREEN에서
`go test ./cmd/tossctl ./internal/enginelock ./internal/console -count=1`이 전부 통과했고,
이 함수의 단언은 편집 전과 **같은 값을 같은 이유로** 요구한다.
