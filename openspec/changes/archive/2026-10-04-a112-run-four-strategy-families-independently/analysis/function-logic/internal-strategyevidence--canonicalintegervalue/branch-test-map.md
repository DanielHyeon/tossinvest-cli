# Branch Test Map: `canonicalIntegerValue`

- Source: `internal/strategyevidence/breakout_bar.go`, SHA-256 `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b`; branch IDs follow `ast.json` (8 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: original delivery RED = build failure on the new symbols (implementer report 2026-08-16); the overflow boundary table was RED-first in the P1/P2 fix round (2026-08-17).

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 790:2 — `6e4`, `1e19`, `15e-1`, `1.2345e4` volume | `TestCanonicalIntegerValueBoundaries`, `TestClosedBarIntegerRuleIsAboutTheValueNotTheSpelling` | yes (implementer report 2026-08-16: build-failure RED on the new symbols) | yes |
| B2 | if at 793:3 — non-integer exponent | not-applicable: unreachable on canonical input (defensive) | not-applicable | not-applicable |
| B3 | if at 796:3 — `15e-1`, decimal `close_minor` | `TestCanonicalIntegerValueBoundaries`, `TestClosedBarRejectsDecimalPriceNumber` | yes (implementer report 2026-08-16) | yes |
| B4 | if at 799:3 — `1e21` | `TestCanonicalIntegerValueBoundaries` | yes (implementer P1/P2 fix round 2026-08-17: overflow boundary table, RED-first) | yes |
| B5 | if at 804:2 — mantissa with `.` | not-applicable: unreachable on canonical input (defensive) | not-applicable | not-applicable |
| B6 | if at 808:2 — `-1`, `""`, `18446744073709551616`, negative `low_minor` | `TestCanonicalIntegerValueBoundaries`, `TestClosedBarRejectsDecimalPriceNumber` | yes (implementer report 2026-08-16; boundary table fix round 2026-08-17) | yes |
| B7 | for at 811:2 — exponent steps for `6e4` / `1e19` | `TestCanonicalIntegerValueBoundaries` | yes (implementer report 2026-08-16) | yes |
| B8 | if at 812:3 — `2e19` refused, `1e19` accepted, envelope `volume` `2e19` refused | `TestCanonicalIntegerValueBoundaries` | yes (implementer P1/P2 fix round 2026-08-17: overflow boundary table, RED-first) | yes |

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; bounds probed independently by both reviewers (review.md 2026-08-17).
