# Branch Test Map: `TestTheTransitionLogLineIsCountable`

- Source: `internal/obs/mode_test.go` (89-133) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 95:2 | `if err := n.AnnounceOperatingMode(context.Background(), journal.ModeEntryBlocked,` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 108:2 | `if err := json.Unmarshal([]byte(first), &line); err != nil {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 118:2 | `if _, ok := line[obs.FieldAccount]; ok {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | range at 121:2 | `for field, value := range want {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 123:3 | `if !ok {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B6 | if at 126:3 | `if got != value {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B7 | if at 130:2 | `if reason, _ := line[obs.FieldReason].(string); !strings.Contains(reason, "approved-by: sre-1") {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
