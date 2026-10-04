# Branch Test Map: `NewClosedBar1mEnvelope`

- Source: `internal/strategyevidence/breakout_bar.go`, SHA-256 `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b`; branch IDs follow `ast.json` (6 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: the combined constructor was added in the implementer follow-up (2026-08-16, "add the combined constructor + tests", VERIFY re-run green); the successor field was RED-first in the P1/P2 fix round (2026-08-17). review.md does not state a RED observation for the constructor's own tests, so those rows say not-applicable.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 302:2 — header refusals through the constructor (KRW under US, `KRX:` session, market `jp`, revision 0, open_at +1 s, observed at/before open, `" "` symbol, empty calendar version, off-day bar) | `TestNewClosedBar1mEnvelopeRefusesInconsistentInput`, `TestClosedBarRequiresTheSessionCalendarDay` | not-applicable (new function; behaviour pinned by the named tests — off-day case RED-first in the 2026-08-17 fix round) | yes |
| B2 | if at 306:2 — `marketAndCode` error after the header accepted the market | not-applicable: unreachable (defensive) | not-applicable | not-applicable |
| B3 | if at 309:2 — zero `SuccessorOpenAt` | `TestNewClosedBar1mEnvelopeCarriesTheObservedSuccessor` /zero | yes (implementer P1/P2 fix round 2026-08-17, RED-first: `successor_open_at_ms` added) | yes |
| B4 | if at 314:2 — non-zero pre-1970 successor | not-applicable: no fixture (defensive); other successor violations reach `checkClosedBar1m` via `NewEnvelope` — `TestNewClosedBar1mEnvelopeCarriesTheObservedSuccessor` | not-applicable | not-applicable |
| B5 | if at 325:2 — raw close 5 dp, raw volume `12345.5`, raw open `+231.4350` | `TestNewClosedBar1mEnvelopeRefusesInconsistentInput` | not-applicable (new function; behaviour pinned by the named test and the scale mutants 4→3/4→5 killed in the follow-up) | yes |
| B6 | if at 356:2 — `json.Marshal` error | not-applicable: unreachable (defensive) | not-applicable | not-applicable |

Happy path: `TestNewClosedBar1mEnvelopeDerivesScaleMinorsAndIdentityTogether` (scale 4/USD derived, minors 2314350/2318000/2311000/2316500/12345 recomputed, raw preserved, header == helper header, clocks bound), `TestNewClosedBar1mEnvelopeCarriesTheObservedSuccessor`, `TestClosedBarAcceptsEveryBarOfTheMarketLocalSessionDay` (20:00 ET post-market bar), `TestSealBarSeriesRoundTripsEnvelopesBuiltByTheConstructor` (store round-trip).

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; reproduced by both reviewers (review.md 2026-08-17).
