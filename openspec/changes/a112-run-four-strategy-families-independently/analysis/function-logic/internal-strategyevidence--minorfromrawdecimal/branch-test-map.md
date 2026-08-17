# Branch Test Map: `minorFromRawDecimal`

- Source: `internal/strategyevidence/breakout_bar.go`, SHA-256 `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b`; branch IDs follow `ast.json` (10 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: original delivery RED = build failure on the new symbols (implementer report 2026-08-16); leading-zero refusal and the overflow boundary table were RED-first in the P1/P2 fix round (2026-08-17); the 32-char bound removal (g5) has no RED statement in review.md.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 823:2 — raw `""` | `TestClosedBarRejectsSignedExponentOrPaddedRawDecimal` | yes (implementer report 2026-08-16: build-failure RED on the new symbols) | yes |
| B2 | if at 827:2 — raw with a fraction point | `TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars`, `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` | yes (implementer report 2026-08-16) | yes |
| B3 | if at 830:2 — sign / exponent / space / comma / trailing `.` / leading `.` / letter | `TestClosedBarRejectsSignedExponentOrPaddedRawDecimal` | yes (implementer report 2026-08-16) | yes |
| B4 | if at 835:2 — `00231.4350`, `01`, `00`, `0231.65`, `012345` | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow`, `TestClosedBarRefusesLeadingZeroRawDecimal` | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B5 | if at 838:2 — 5 dp at scale 4, 1 dp at scale 0, 30 dp long raw | `TestClosedBarRejectsOverPreciseRawForTheDeclaredScale`, `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` | yes (implementer report 2026-08-16; scale amendment follow-up mutants 4→3/4→5 killed) | yes |
| B6 | for at 842:2 — integer digits | every accepted raw (`TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` accept table) | yes (implementer report 2026-08-16) | yes |
| B7 | if at 844:3 — `18446744073709551616`, 28-digit integer part | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` | yes (implementer P1/P2 fix round 2026-08-17: overflow boundary table, RED-first) | yes |
| B8 | for at 849:2 — scale 4 padding / scale 0 no-op | `TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars`, `TestNewClosedBar1mEnvelopeDerivesScaleMinorsAndIdentityTogether` | yes (implementer report 2026-08-16) | yes |
| B9 | if at 851:3 — `0.5`→5000, `0.0001`→1, `231.1`→2311000 | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` | yes (implementer report 2026-08-16) | yes |
| B10 | if at 855:3 — `1844674407370955.1616`@4 refused, `.1615` accepted | `TestMinorFromRawDecimalRefusesLeadingZerosAndOverflow` | yes (implementer P1/P2 fix round 2026-08-17: overflow boundary table, RED-first) | yes |

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; overflow guards probed independently by both reviewers (review.md 2026-08-17).
