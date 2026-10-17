# Branch Test Map: `OpenRecorder`

- Source: `internal/verifylive/record.go` (327-343) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 329:2 | `if err := os.MkdirAll(dir, 0o700); err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B2 | if at 333:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B3 | if at 338:2 | `if err != nil {` | `TestA131OpenRecorderRefusesAMissingKeyForASealedRecord` | 편집 영향 |

## a131 시험

- `TestA131OpenRecorderRefusesAMissingKeyForASealedRecord`
