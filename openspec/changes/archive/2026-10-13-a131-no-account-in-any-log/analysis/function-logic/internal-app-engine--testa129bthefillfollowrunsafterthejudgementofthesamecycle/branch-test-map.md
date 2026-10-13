# Branch Test Map: `TestA129BTheFillFollowRunsAfterTheJudgementOfTheSameCycle`

- Source: `internal/app/engine/a129_exit_order_log_test.go` (302-331) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | range at 317:2 | `for i, l := range lines {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 318:3 | `if l["event"] == string(obs.EventOrderConfirmed) && l["symbol"] == "000660" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 321:3 | `if l["event"] == string(obs.EventFillObserved) && l["order_id"] == first {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 325:2 | `if confirmAt < 0 \|\| fillAt < 0 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 328:2 | `if confirmAt > fillAt {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
