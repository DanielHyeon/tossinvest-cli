# Branch Test Map: `minuteGaps`

- Source: `internal/officialbars/producer.go`, SHA-256 `83960410b7b870ca60fad568002060de49ebc7271d72b94d46f665ff274b29b1`; branch IDs follow `ast.json` (5 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; the clamp edge cases were added RED-first in the ruling 29 fix round (gstack P2 "gap clamps untested").
- Tests: `internal/officialbars/producer_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | for at 437:2 — consecutive observed pairs walked | `TestPollGapsAreClampedToTheRegularWindow` (3 gaps), `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (388-minute gap) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B2 | if at 440:3 — a hole starting before the regular open is clamped to the open | `TestPollGapsAreClampedToTheRegularWindow` (`Gaps[2]` from the open, 1 minute) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B3 | if at 443:3 — a hole reaching past the close is clamped to the close | `TestPollGapsAreClampedToTheRegularWindow` (`Gaps[0]` 386 minutes, `To` = close − 1 min) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B4 | if at 446:3 — adjacent minutes leave an empty window and report nothing | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollNeverAdmitsTheNewestObservedBar` (zero gaps) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B5 | if at 450:3 — a positive but sub-minute span | not-applicable: ruling 26 pins every bar instant to a minute boundary and the adapted calendar edges are whole minutes, so the span after B4 is always ≥ 1 minute; defensive | not-applicable | not-applicable |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
