# Branch Test Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

시험 함수 자신. 분기 변화 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:22` 조립 실패 | 이 함수(`cmd/tossctl/a098_the_engine_registers_a_sender_test.go:20`) — 조립 성공 경로만 탐 | no | no |
| B2 | `:28` 보조 실행자 집합 불일치 | 이 함수 — 배선 전 base 에서 실패(`analysis/red/red.md`), 배선 뒤 통과 | yes | yes |
