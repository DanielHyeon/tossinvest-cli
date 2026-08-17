# Branch Test Map: `checkClosedBar1m`

- Source: `internal/strategyevidence/breakout_bar.go`, SHA-256 `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b`; branch IDs follow `ast.json` (28 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED evidence per review.md: the original delivery captured RED as a build failure on the new symbols against the unmodified file (2026-08-16 implementer report); the P1/P2 fix round (2026-08-17) was RED-first for the calendar-day, finality-minimum, successor, symbol and leading-zero rules. Where review.md carries no RED statement the row says so.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 466:2 — foreign/empty/case-variant `schema` | `TestClosedBarRejectsUnknownEnumValues` /schema | yes (implementer report 2026-08-16: build-failure RED on the new symbols) | yes |
| B2 | if at 470:2 — `market` JP/""/us/Us/USA/KR-in-US-fixture | `TestClosedBarRejectsUnknownEnumValues` /market | yes (implementer report 2026-08-16) | yes |
| B3 | if at 473:2 — `currency` EUR/""/usd/KRW/USDT under US | `TestClosedBarRejectsUnknownEnumValues` /currency | yes (implementer report 2026-08-16) | yes |
| B4 | if at 476:2 — `price_scale` 0/2/3/5 for USD | `TestClosedBarRejectsOverPreciseRawForTheDeclaredScale` | yes (implementer follow-up 2026-08-16: scale table amended, mutants 4→3/4→5 killed) | yes |
| B5 | if at 479:2 — symbol `AA:PL`, `:AAPL`, `AAPL:`, `aapl`, `Aapl`, ` AAPL`, "", NUL | `TestClosedBarRefusesSymbolWithRecordSeparatorOrLowerCase`, `TestClosedBarRejectsMalformedIdentityStrings` | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B6 | if at 483:2 — session `KRX:…` under US, bad date, no calendar | `TestClosedBarRejectsMalformedIdentityStrings`, `TestClosedBarRejectsUnknownEnumValues` /session_id | yes (implementer report 2026-08-16) | yes |
| B7 | if at 486:2 — empty `calendar_version` | `TestClosedBarRejectsMalformedIdentityStrings` /calendar version is empty | yes (implementer report 2026-08-16) | yes |
| B8 | if at 489:2 — interval 0/1000/30000/60001/300000 | `TestClosedBarRejectsIntervalOtherThanOneMinute` | yes (implementer report 2026-08-16) | yes |
| B9 | if at 492:2 — `bar_label` close_at/""/OPEN_AT/… | `TestClosedBarRejectsUnknownEnumValues` /bar_label | yes (implementer report 2026-08-16) | yes |
| B10 | if at 495:2 — open_at +1 ms / +59,999 ms / −30 s / 0 | `TestClosedBarRejectsOpenAtThatIsNotOnTheMinute` | yes (implementer report 2026-08-16) | yes |
| B11 | if at 498:2 — `finality` unknown/""/SUCCESSOR_OBSERVED/final | `TestClosedBarRejectsUnknownEnumValues` /finality (residual a3: a widened enum with a literal not in the table survives — recorded in review.md) | yes (implementer report 2026-08-16; mutant "finality enum widened" killed) | yes |
| B12 | if at 501:2 — `closed=false` | `TestClosedBarRejectsUnfinishedBar` | yes (implementer report 2026-08-16) | yes |
| B13 | if at 505:2 — observed = open, open−60 s, 0, open+59,999 ms; accept at open+60,000 | `TestClosedBarRequiresObservationAfterTheBarClosed`, `TestClosedBarRejectsBarFromTheFuture` | yes (implementer P1/P2 fix round 2026-08-17, RED-first; mutant "future-bar check deleted" killed earlier) | yes |
| B14 | if at 511:2 — `ParseMarket` error | not-applicable: unreachable after B2 (defensive; no test can reach it through the decoder) | not-applicable | not-applicable |
| B15 | if at 515:2 — `TradingDay` error | not-applicable: unreachable (embedded tzdata, market pinned by B2/B14; defensive) | not-applicable | not-applicable |
| B16 | if at 518:2 — header valid for 08-14, payload bar on 08-13/08-12/08-15; 20:00 ET post-market and 09:00 KST accepted | `TestClosedBarPayloadAloneRefusesABarFromAnotherCalendarDay`, `TestClosedBarAcceptsEveryBarOfTheMarketLocalSessionDay` (`TestClosedBarRequiresTheSessionCalendarDay` stops at the header helper and does not reach this check) | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B17 | range at 522:2 — the four price minors | every accepting test, e.g. `TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars` | yes (implementer report 2026-08-16) | yes |
| B18 | if at 526:3 — `low_minor` 0 | `TestClosedBarRejectsImpossiblePriceOrdering` /zero low | yes (implementer report 2026-08-16) | yes |
| B19 | if at 530:2 — low 231.9 above high 231.8 | `TestClosedBarRejectsImpossiblePriceOrdering` /low above high | yes (implementer report 2026-08-16) | yes |
| B20 | if at 533:2 — open 231.0 below low | `TestClosedBarRejectsImpossiblePriceOrdering` /open below low | yes (implementer report 2026-08-16) | yes |
| B21 | if at 536:2 — close 231.85 above high | `TestClosedBarRejectsImpossiblePriceOrdering` /close above high | yes (implementer report 2026-08-16) | yes |
| B22 | range at 539:2 — five raw↔minor pairs | every accepting test | yes (implementer report 2026-08-16) | yes |
| B23 | if at 550:3 — minor ≠ raw×scale, 5-dp USD raw, signed/exponent/padded raw, leading-zero raw | `TestClosedBarRejectsMinorThatDisagreesWithRawDecimal`, `TestClosedBarRejectsOverPreciseRawForTheDeclaredScale`, `TestClosedBarRejectsSignedExponentOrPaddedRawDecimal`, `TestClosedBarRefusesLeadingZeroRawDecimal` | yes (implementer report 2026-08-16; leading zero: fix round 2026-08-17) | yes |
| B24 | if at 554:2 — `revision` 0 | `TestClosedBarRejectsMalformedIdentityStrings` /revision is zero | yes (implementer report 2026-08-16) | yes |
| B25 | if at 558:2 — successor open+60,001 ms | `TestClosedBarRequiresSuccessorOpenAt` /not on the minute | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B26 | if at 561:2 — successor = open / open−60 s / 0 | `TestClosedBarRequiresSuccessorOpenAt` /at the bar itself, /before the bar, /zero | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B27 | if at 564:2 — successor after the observation instant; equal accepted | `TestClosedBarRequiresSuccessorOpenAt` /after the observation instant | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B28 | if at 567:2 — digest without prefix / upper-case / short | `TestClosedBarRejectsMalformedIdentityStrings` (digest cases) | yes (implementer report 2026-08-16) | yes |

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, four consumer packages, `go build ./...`, vet, gofmt (GOROOT) green; two independent reviewers reproduced (review.md 2026-08-17 L1a sections).
