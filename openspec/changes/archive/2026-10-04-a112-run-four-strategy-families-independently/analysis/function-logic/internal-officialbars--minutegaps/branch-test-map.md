# Branch Test Map: `minuteGaps`

- Source: `internal/officialbars/producer.go`, SHA-256 `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772`; branch IDs follow `ast.json` (5 branches, regenerated 2026-08-18 against the decision-30 sources).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; the clamp edge cases were added RED-first in the ruling 29 fix round (gstack P2 "gap clamps untested").
- **Decision-30 correction (2026-08-18).** The broker's `timestamp` is the bar's close instant, not its open (US probe 03:29 KST, review.md). This function's behaviour is unchanged; only its documentation and its source line numbers moved. Branch count unchanged.
- Tests: `internal/officialbars/producer_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | for at 462:2 — consecutive observed pairs walked, on converted open instants | `TestPollReportsGapsOnConvertedOpenInstants` (broker labels 09:35/09:33/09:31 → holes opening at 09:33 and 09:31, 1 minute each), `TestPollGapsAreClampedToTheRegularWindow` (3 gaps), `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (388-minute gap) | 13 producer tests failed one minute off before the conversion landed (review.md, decision-30 correction 2026-08-18) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B2 | if at 465:3 — a hole starting before the regular open is clamped to the open | `TestPollGapsAreClampedToTheRegularWindow` (`Gaps[2]` from the open, 1 minute) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B3 | if at 468:3 — a hole reaching past the close is clamped to the close | `TestPollGapsAreClampedToTheRegularWindow` (`Gaps[0]` 386 minutes, `To` = close − 1 min) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B4 | if at 471:3 — adjacent minutes leave an empty window and report nothing | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollNeverAdmitsTheNewestObservedBar` (zero gaps) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B5 | if at 475:3 — a positive but sub-minute span | not-applicable: ruling 26 pins every bar instant to a minute boundary and the adapted calendar edges are whole minutes, so the span after B4 is always ≥ 1 minute; defensive | not-applicable | not-applicable |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed across the two packages.
