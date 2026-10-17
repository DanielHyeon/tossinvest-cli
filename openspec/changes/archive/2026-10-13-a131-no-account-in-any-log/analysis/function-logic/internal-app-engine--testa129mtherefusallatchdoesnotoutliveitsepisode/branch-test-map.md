# Branch Test Map: `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode`

- Source: `internal/app/engine/a129_exit_order_log_test.go` (762-786) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 780:2 | `if len(f.submit.places) == placesBefore {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 783:2 | `if got := len(a129Of(f.lines(), obs.EventOrderRejected)); got != 2 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
