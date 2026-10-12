# Branch Test Map: `LoadEntries`

- Source: `internal/verifylive/record.go` (461-523) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 463:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B2 | if at 464:3 | `if errors.Is(err, os.ErrNotExist) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B3 | if at 484:3 | `if pending == nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B4 | if at 488:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B5 | for at 496:2 | `for scanner.Scan() {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B6 | if at 499:3 | `if len(trimSpace(raw)) == 0 {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B7 | if at 502:3 | `if err := flush(); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B8 | if at 507:2 | `if err := scanner.Err(); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B9 | if at 511:2 | `if pending != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B10 | if at 513:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B11 | if at 514:4 | `if errors.Is(err, errUnknownFormat) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |

## a131 시험

- `TestA131SealedVerifyProofStillBinds`
