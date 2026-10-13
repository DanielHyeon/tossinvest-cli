# Branch Test Map: `TestA090R2AnAbsentSiblingIsCountedLikeAZeroQuote`

- Source: `internal/app/engine/a090_unobserved_position_test.go` (271-289) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 277:2 | `if cycle.Err != nil \|\| cycle.Judged != 1 \|\| cycle.Unobserved != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 281:2 | `if len(lines) != 1 \|\| lines[0]["position_id"] != redact.String(target.ID) \|\| lines[0]["cause"] != "no_quot` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 286:2 | `if got := len(f.lines()); got != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
