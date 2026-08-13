# Branch Test Map: `TestAFreshMarkerWithALiveProcessStillRefuses`

Source: `cmd/tossctl/engineproc_test.go` (200-225). AST 기준 branches **5** / returns 0.

| Branch | 위치 | 지는 테스트 | GREEN 실측 |
|---|---|---|---|
| B1 | `:203` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |
| B2 | `:212` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |
| B3 | `:215` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |
| B4 | `:218` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |
| B5 | `:222` if | 이 함수 자신이 그 테스트다 | 통과 (a102 GREEN) |

이 함수는 테스트다 — 자기 분기를 자기가 실행한다. a102 GREEN에서
`go test ./cmd/tossctl ./internal/enginelock ./internal/console -count=1`이 전부 통과했고,
이 함수의 단언은 편집 전과 **같은 값을 같은 이유로** 요구한다.
