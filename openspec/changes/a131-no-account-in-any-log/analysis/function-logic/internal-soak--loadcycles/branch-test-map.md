# Branch Test Map: `LoadCycles`

- Source: `internal/soak/record.go` (230-296) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 232:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B2 | if at 233:3 | `if errors.Is(err, os.ErrNotExist) {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B3 | if at 256:3 | `if pending == nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B4 | if at 260:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B5 | for at 268:2 | `for scanner.Scan() {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B6 | if at 271:3 | `if len(trimSpace(raw)) == 0 {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B7 | if at 274:3 | `if err := flush(); err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B8 | if at 279:2 | `if err := scanner.Err(); err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B9 | if at 283:2 | `if pending != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B10 | if at 285:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B11 | if at 286:4 | `if errors.Is(err, errUnknownFormat) {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |

## a131 시험

- `TestA131SealedSoakRecordStillAttestsTheSameAccount`
- `TestA131LegacyAndSealedLinesReadTheSame`
- `TestA131MissingSealKeyFailsClosed`
