# Branch Test Map: `LoadEntries`

- Source: `internal/verifylive/record.go` (511-582) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 513:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B2 | if at 514:3 | `if errors.Is(err, os.ErrNotExist) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B3 | if at 536:3 | `if pending == nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B4 | if at 540:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B5 | if at 543:3 | `if e, err = openEntryAccount(e, opener); err != nil {` | `TestA131UnopenableVerifyLineIsAnError` | 편집 영향 |
| B6 | for at 551:2 | `for scanner.Scan() {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B7 | if at 554:3 | `if len(trimSpace(raw)) == 0 {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B8 | if at 557:3 | `if err := flush(); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B9 | if at 562:2 | `if err := scanner.Err(); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B10 | if at 566:2 | `if pending != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B11 | if at 568:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B12 | if at 569:4 | `if errors.Is(err, errUnknownFormat) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B13 | if at 576:3 | `if e, err = openEntryAccount(e, opener); err != nil {` | `TestA131UnopenableVerifyLineIsAnError` | 편집 영향 |

## a131 시험

- `TestA131SealedVerifyProofStillBinds`
- `TestA131UnopenableVerifyLineIsAnError`
