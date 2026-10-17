# Branch Test Map: `Client.accountsLocked`

- Source: `internal/official/reads.go` (38-57) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 40:2 | `if err := c.get(ctx, "/api/v1/accounts", nil, &raw); err != nil {` | a131 무편집 분기 — `go test ./internal/official/` 전체(영수증 analysis/verify) | 무편집 |
| B2 | if at 44:2 | `if selected == 0 && len(raw) > 0 && raw[0].AccountSeq > 0 {` | a131 무편집 분기 — `go test ./internal/official/` 전체(영수증 analysis/verify) | 무편집 |
| B3 | else at 46:9 | `} else if selected > 0 && !c.accountSeqExplicit {` | a131 무편집 분기 — `go test ./internal/official/` 전체(영수증 analysis/verify) | 무편집 |
| B4 | if at 46:9 | `} else if selected > 0 && !c.accountSeqExplicit {` | a131 무편집 분기 — `go test ./internal/official/` 전체(영수증 analysis/verify) | 무편집 |
| B5 | if at 47:3 | `if len(raw) == 0 \|\| raw[0].AccountSeq <= 0 \|\| int64(raw[0].AccountSeq) != selected {` | a131 무편집 분기 — `go test ./internal/official/` 전체(영수증 analysis/verify) | 무편집 |

## a131 시험

- `TestA131AccountsRegistersEveryAccountNumber` — httptest 응답의 계좌 번호가 `redact.String` 에서 치환됨
