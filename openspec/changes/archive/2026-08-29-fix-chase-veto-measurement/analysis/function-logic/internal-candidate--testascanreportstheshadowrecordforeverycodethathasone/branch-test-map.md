# Branch Test Map: `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | cycle 실행 | 자체 실행 | — (기존 동작) | yes |
| B2 | 가속 시리즈가 있다 | 자체 실행 | — (기존 동작) | yes |
| B3 | crossing 칸이 전부 있다 | 자체 실행 | — (기존 동작) | yes |
| B4 | 두 code | 자체 실행 | — (기존 동작) | yes |
| B5 | code마다 밴드가 있다 | 자체 실행 | — (기존 동작) | yes |
| B6 | 밴드 total | 자체 실행 | — (기존 동작) | yes |
| B7 | 미측정 census 합산 | 자체 실행 | — (기존 동작) | yes |
| B8 | 합이 total과 맞는다 | 자체 실행 | — (기존 동작) | yes |
| B9 | `seen_late` 밴드가 측정된다 — 저장된 위치 **와** 소스의 보고 둘 다 필요하다 | 자체 실행 | yes (`pricedRow`가 자격을 채우지 않으면 실패한다) | yes |

B9는 `wiring_test.go:pricedRow`가 `RankRequested`와 `NewlyListed`를 채우는 것에 의존한다.
그 fixture가 zero value로 남았다면 이 단언이 `Measured == 0`으로 실패했을 것이고, 그것이
이 change에서 `pricedRow`를 함께 고친 이유다.
