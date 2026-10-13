# Branch Test Map: `TestA090R4AJudgementEndsTheStreakWithOneReleaseLine`

- Source: `internal/app/engine/a090_unobserved_position_test.go` (742-769) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 749:2 | `if cycle.Unobserved != 0 \|\| cycle.Judged != 2 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 753:2 | `if len(lines) != 2 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 757:2 | `if release["position_id"] != redact.String(target.ID) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 760:2 | `if secs, ok := release["unobserved_seconds"].(float64); !ok \|\| secs < 15 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 766:2 | `if got := len(f.lines()); got != 3 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
