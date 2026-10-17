# Branch Test Map: `TestRecordCarriesNoUnmaskedAccountNumber`

- Source: `internal/verifylive/record_test.go` (156-177) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 159:2 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 163:2 | `if err != nil {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 166:2 | `if strings.Contains(string(data), "123-45-678901") {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 170:2 | `if strings.Contains(string(data), "8901") {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 174:2 | `if err != nil \|\| len(entries) == 0 \|\| !strings.HasSuffix(entries[0].AccountRef, "8901") {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
