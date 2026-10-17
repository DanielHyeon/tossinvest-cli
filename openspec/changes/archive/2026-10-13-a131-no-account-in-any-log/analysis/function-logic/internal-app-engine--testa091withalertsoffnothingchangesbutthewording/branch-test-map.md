# Branch Test Map: `TestA091WithAlertsOffNothingChangesButTheWording`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go` (226-310) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 235:3 | `if n := len(h.alerts.events); n != 0 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 247:3 | `if h.alerts.count(obs.EventExitStopSoldNothing) != 0 \|\| h.alerts.count(obs.EventExitProposalCapped) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 251:3 | `if want := "005930 청산이 확정 하한에 걸려 한 주도 나가지 않았다"; a.Title != want {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 260:3 | `if len(lines) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 264:3 | `if l["level"] != "ERROR" \|\| l["cause"] != "floor_unknown" \|\| l["symbol"] != "005930" \|\| l["position_id"]` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B6 | if at 267:3 | `if _, ok := l["account"]; ok \|\| strings.Contains(fmt.Sprint(l), exitAccount) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B7 | if at 270:3 | `if len(r.log.ofEvent(t, obs.EventExitStopSoldNothing)) != 0 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B8 | if at 284:3 | `if a.Body != want {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B9 | range at 288:2 | `for _, cause := range []struct {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B10 | if at 299:4 | `if rows, _ := r.journal.PendingAlerts(context.Background(), 0); len(rows) != 0 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B11 | if at 302:4 | `if rej := r.gate.CheckEntry(); rej != nil {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B12 | if at 305:4 | `if got := r.mode(); got != journal.ModeNormal {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
