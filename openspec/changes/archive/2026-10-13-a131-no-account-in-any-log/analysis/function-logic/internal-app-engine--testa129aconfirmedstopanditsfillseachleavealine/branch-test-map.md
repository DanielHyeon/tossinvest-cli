# Branch Test Map: `TestA129AConfirmedStopAndItsFillsEachLeaveALine`

- Source: `internal/app/engine/a129_exit_order_log_test.go` (216-287) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 219:2 | `if cycle.Err != nil \|\| len(f.submit.places) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 226:2 | `if len(confirmed) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 237:2 | `if got, want := confirmed[0]["price"], decimalText(place.Intent.Price); got != want {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 240:2 | `if id, _ := confirmed[0]["order_id"].(string); !strings.HasPrefix(id, "O-sell-") {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 243:2 | `if f.observer.FollowedExitOrdersForTest() != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B6 | if at 252:2 | `if len(fills) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B7 | if at 264:2 | `if got := len(a129Of(f.lines(), obs.EventFillObserved)); got != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B8 | if at 272:2 | `if len(fills) != 2 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B9 | if at 276:2 | `if f.observer.FollowedExitOrdersForTest() != 0 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B10 | if at 281:2 | `if got := len(a129Of(f.lines(), obs.EventFillObserved)); got != 2 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B11 | if at 284:2 | `if got := len(a129Of(f.lines(), obs.EventOrderConfirmed)); got != len(f.submit.places) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
