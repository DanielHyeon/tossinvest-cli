# Branch Test Map: `NewLogger`

- Source: `internal/obs/log.go` (98-131) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 100:2 | `if w == nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B2 | if at 104:2 | `if clk == nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B3 | if at 114:4 | `if len(groups) == 0 && a.Key == slog.TimeKey {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B4 | if at 122:2 | `if opts.JSON {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B5 | else at 124:9 | `} else {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B6 | if at 127:2 | `if len(opts.Base) > 0 {` | `TestA131EngineLoggerDropsAccountAndScrubsValues` · `TestA131BaseAttrsAreScrubbed` | 편집 영향 |

## a131 시험

- `TestA131EngineLoggerDropsAccountAndScrubsValues` — 센티널 계좌 필드·감싼 오류 원문·With 속성·마스크·pos- id 가 JSON 줄에 없음
- `TestA131BaseAttrsAreScrubbed`
- `TestA131TheSinksCallTheRule` — 구조 핀(go/parser): NewLogger 가 redact.Handler 를 부름
