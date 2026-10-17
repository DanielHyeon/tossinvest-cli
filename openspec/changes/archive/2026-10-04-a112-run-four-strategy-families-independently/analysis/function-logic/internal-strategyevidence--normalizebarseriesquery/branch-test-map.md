# Branch Test Map: `normalizeBarSeriesQuery`

- Source: `internal/strategyevidence/breakout_series.go`, SHA-256 `ece27d1ede03408e1f819f5d65f42fca9a252dd3e693b4cbadf834bee4e9abc5`; branch IDs follow `ast.json` (7 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: original delivery RED = build failure on the new symbols against the unmodified file (implementer report 2026-08-16); the `:`-in-symbol refusal was RED-first in the P1/P2 fix round (2026-08-17).

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 177:2 — market `JP` / `""` | `TestSealBarSeriesRefusesInvalidQuery` /unknown market, /empty market | yes (implementer report 2026-08-16: build-failure RED on the new symbols) | yes |
| B2 | if at 183:2 — symbol `" "` | `TestSealBarSeriesRefusesInvalidQuery` /empty symbol | yes (implementer report 2026-08-16; `:` refusal fix round 2026-08-17 — no query-level `:` fixture, the header/payload fixtures cover the shared predicate: `TestClosedBarRefusesSymbolWithRecordSeparatorOrLowerCase`) | yes |
| B3 | if at 187:2 — session `""` / `KRX:2026-08-14` under US | `TestSealBarSeriesRefusesInvalidQuery` /empty session, /session calendar does not match market | yes (implementer report 2026-08-16) | yes |
| B4 | if at 190:2 — interval 30000 | `TestSealBarSeriesRefusesInvalidQuery` /interval is not one minute | yes (implementer report 2026-08-16) | yes |
| B5 | if at 194:2 — MaxBars 513 / −1 | `TestSealBarSeriesRefusesInvalidQuery` /max bars above the hard cap, /negative max bars | yes (implementer report 2026-08-16; mutant 512→513 killed) | yes |
| B6 | if at 198:2 — MaxBars 0 → 512 | `TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest`, `TestSealBarSeriesRefusesMoreThanFiveHundredTwelveBars` | yes (implementer report 2026-08-16) | yes |
| B7 | if at 201:2 — zero EvaluationAt / zero IngestionCutoff | `TestSealBarSeriesRefusesInvalidQuery` /zero evaluation clock, /zero ingestion cutoff | yes (implementer report 2026-08-16) | yes |

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; reproduced by both reviewers (review.md 2026-08-17).
