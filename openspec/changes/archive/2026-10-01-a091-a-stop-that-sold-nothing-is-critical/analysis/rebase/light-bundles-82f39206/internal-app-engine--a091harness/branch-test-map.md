# Branch Test Map: `a091Harness`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go`

> 시험 함수의 분기는 그 시험이 돌 때 지나간다 — Test 열은 그 시험(또는 도우미를 부르는 시험)이다.

| Branch | 조건 | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:119` `if floor != nil {` | `TestA091AProtectiveZeroIsACriticalRow` | n/a | yes |
| B2 | `:122` `if mutate != nil {` | `TestA091AProtectiveZeroIsACriticalRow` | n/a | yes |
