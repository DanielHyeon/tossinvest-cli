# Branch Test Map: `lockedRecord.appendLine`

- Source: `internal/verifylive/record.go` (425-451) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 426:2 | `if len(line) == 0 \|\| line[len(line)-1] != '\n' {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B2 | if at 430:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B3 | if at 433:2 | `if st.Size() != judged {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B4 | if at 437:2 | `if err != nil \|\| n != len(line) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B5 | if at 440:2 | `if err := recordSync(l.f); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B6 | if at 444:2 | `if _, err := l.f.ReadAt(back, judged); err != nil && !errors.Is(err, io.EOF) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B7 | if at 447:2 | `if !bytes.Equal(back, line) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |

## a131 시험

- `TestA131ReconcileLineIsSealedAndStillBinds`
