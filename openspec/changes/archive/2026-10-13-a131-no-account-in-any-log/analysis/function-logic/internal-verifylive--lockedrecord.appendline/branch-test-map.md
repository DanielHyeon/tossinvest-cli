# Branch Test Map: `lockedRecord.appendLine`

- Source: `internal/verifylive/record.go` (467-501) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 468:2 | `if len(line) == 0 \|\| line[len(line)-1] != '\n' {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B2 | if at 473:2 | `if err != nil {` | 성공 경로 `TestA131ReconcileLineIsSealedAndStillBinds` — 키 준비 실패 경로 시험 없음(잔재) | 편집 영향 |
| B3 | if at 476:2 | `if line, err = redact.SealRecordLine(line, key); err != nil {` | `TestA131ReconcileLineIsSealedAndStillBinds` | 편집 영향 |
| B4 | if at 480:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B5 | if at 483:2 | `if st.Size() != judged {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B6 | if at 487:2 | `if err != nil \|\| n != len(line) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B7 | if at 490:2 | `if err := recordSync(l.f); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B8 | if at 494:2 | `if _, err := l.f.ReadAt(back, judged); err != nil && !errors.Is(err, io.EOF) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B9 | if at 497:2 | `if !bytes.Equal(back, line) {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |

## a131 시험

- `TestA131ReconcileLineIsSealedAndStillBinds`
