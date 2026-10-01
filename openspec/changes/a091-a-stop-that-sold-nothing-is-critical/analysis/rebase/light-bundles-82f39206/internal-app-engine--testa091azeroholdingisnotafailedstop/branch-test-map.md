# Branch Test Map: `TestA091AZeroHoldingIsNotAFailedStop`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go`

> 시험 함수의 분기는 그 시험이 돌 때 지나간다 — Test 열은 그 시험(또는 도우미를 부르는 시험)이다.

| Branch | 조건 | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:345` `if n := len(r.rows(obs.EventExitStopSoldNothing)); n != 0 {` | `TestA091AZeroHoldingIsNotAFailedStop` | n/a | yes |
| B2 | `:349` `if len(lines) != 1 \|\| !strings.Contains(fmt.Sprint(lines[0]["detail"]), "계좌에 보유가 없다") {` | `TestA091AZeroHoldingIsNotAFailedStop` | n/a | yes |
| B3 | `:353` `if a != "no_holding" {` | `TestA091AZeroHoldingIsNotAFailedStop` | n/a | yes |
