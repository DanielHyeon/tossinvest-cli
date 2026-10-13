# Branch Test Map: `TestA090R1AZeroQuoteSiblingIsCountedAndLoggedAtNormalGrade`

- Source: `internal/app/engine/a090_unobserved_position_test.go` (246-269) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 252:2 | `if cycle.Err != nil \|\| cycle.Judged != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 255:2 | `if cycle.Unobserved != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 259:2 | `if len(lines) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 262:2 | `if lines[0]["position_id"] != redact.String(target.ID) \|\| lines[0]["cause"] != "no_quote" \|\|` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 266:2 | `if rows := f.outageRows(); len(rows) != 0 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
