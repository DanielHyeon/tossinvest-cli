# Branch Test Map: `newClosedBar1mHeader`

- Source: `internal/strategyevidence/breakout_bar.go`, SHA-256 `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b`; branch IDs follow `ast.json` (11 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: original delivery RED = build failure on the new symbols against the unmodified file (implementer report 2026-08-16); calendar-day and `:`-symbol rules RED-first in the P1/P2 fix round (2026-08-17).

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 137:2 — market `jp` | `TestClosedBar1mHeaderRefusesInconsistentInput` /unknown market | yes (implementer report 2026-08-16: build-failure RED on the new symbols) | yes |
| B2 | if at 141:2 — symbol `"  "` / `AA:PL` | `TestClosedBar1mHeaderRefusesInconsistentInput` /empty symbol, `TestClosedBarRefusesSymbolWithRecordSeparatorOrLowerCase` | yes (implementer report 2026-08-16; `:` refusal: fix round 2026-08-17, RED-first) | yes |
| B3 | if at 146:2 — `KRX:` session under US | `TestClosedBar1mHeaderRefusesInconsistentInput` /session calendar does not match market | yes (implementer report 2026-08-16) | yes |
| B4 | if at 149:2 — empty calendar version | `TestClosedBar1mHeaderRefusesInconsistentInput` /empty calendar version | yes (implementer report 2026-08-16) | yes |
| B5 | if at 152:2 — KRW under US | `TestClosedBar1mHeaderRefusesInconsistentInput` /currency does not match market | yes (implementer report 2026-08-16) | yes |
| B6 | if at 155:2 — revision 0 | `TestClosedBar1mHeaderRefusesInconsistentInput` /revision zero | yes (implementer report 2026-08-16) | yes |
| B7 | if at 158:2 — zero observed clock | `TestClosedBar1mHeaderRefusesInconsistentInput` /zero observed clock | yes (implementer report 2026-08-16) | yes |
| B8 | if at 163:2 — open_at + 1 s | `TestClosedBar1mHeaderRefusesInconsistentInput` /open_at is not on the minute | yes (implementer report 2026-08-16) | yes |
| B9 | if at 167:2 — `TradingDay` error | not-applicable: unreachable (defensive; market pinned by B1, tzdata embedded) | not-applicable | not-applicable |
| B10 | if at 170:2 — 2026-09-23 / 1996-01-02 / ±1 day inside `US:2026-08-14`; KR 09:00 KST accepted | `TestClosedBar1mHeaderRefusesABarFromAnotherCalendarDay`, `TestClosedBarRequiresTheSessionCalendarDay`, `TestClosedBarAcceptsEveryBarOfTheMarketLocalSessionDay` | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B11 | if at 176:2 — observed at open + 30 s | `TestClosedBar1mHeaderRefusesInconsistentInput` /observed before the bar closed | yes (implementer report 2026-08-16) | yes |

Happy path and determinism: `TestClosedBar1mHeaderCarriesDeterministicBarIdentity` (identity strings, clocks, r1→r2 supersedes), `TestNewClosedBar1mEnvelopeDerivesScaleMinorsAndIdentityTogether` (constructor header == helper header).

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; reproduced by both reviewers (review.md 2026-08-17).
