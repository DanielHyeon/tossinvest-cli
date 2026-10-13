# Branch Test Map: `LoadCycles`

- Source: `internal/soak/record.go` (272-347) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 274:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B2 | if at 275:3 | `if errors.Is(err, os.ErrNotExist) {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B3 | if at 300:3 | `if pending == nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B4 | if at 304:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B5 | if at 307:3 | `if c, err = openCycleAccount(c, opener); err != nil {` | `TestA131UnopenableSealedLineIsAnError` · `TestA131MissingSealKeyFailsClosed` | 편집 영향 |
| B6 | for at 315:2 | `for scanner.Scan() {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B7 | if at 318:3 | `if len(trimSpace(raw)) == 0 {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B8 | if at 321:3 | `if err := flush(); err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B9 | if at 326:2 | `if err := scanner.Err(); err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B10 | if at 330:2 | `if pending != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B11 | if at 332:3 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B12 | if at 333:4 | `if errors.Is(err, errUnknownFormat) {` | a131 무편집 분기 — `go test ./internal/soak/` 전체 | 무편집 |
| B13 | if at 341:3 | `if c, err = openCycleAccount(c, opener); err != nil {` | `TestA131UnopenableSealedLineIsAnError`(마지막 줄 — 찢김 관용 없음) | 편집 영향 |

## a131 시험

- `TestA131SealedSoakRecordStillAttestsTheSameAccount`
- `TestA131LegacyAndSealedLinesReadTheSame`
- `TestA131MissingSealKeyFailsClosed`
- `TestA131UnopenableSealedLineIsAnError`
