# Branch Test Map: `OpenRecorder`

- Source: `internal/soak/record.go` (179-195) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 181:2 | `if err := os.MkdirAll(dir, 0o700); err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B2 | if at 185:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B3 | if at 190:2 | `if err != nil {` | `TestA131MissingSealKeyFailsClosed`(봉인 줄 + 키 없음 → 거절) | 편집 영향 |

## a131 시험

- `TestA131OpenRecorderRefusesAMissingKeyForASealedRecord`
