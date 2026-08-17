# Branch Test Map: `adoptPage`

- Source: `internal/officialbars/producer.go`, SHA-256 `83960410b7b870ca60fad568002060de49ebc7271d72b94d46f665ff274b29b1`; branch IDs follow `ast.json` (2 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; the provenance rule this function carries (`ObservedAt` from the response, not from `PollAt`) was pinned in the ruling 29 fix round.
- Tests: `internal/officialbars/producer_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 400:2 — a page's candles adopted in order with the response's `ReadAt` and `BodyDigest` | `TestPollTakesObservedAtFromTheResponseNotThePollInstant`, `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollNeverAdmitsTheNewestObservedBar` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B2 | if at 402:3 — a candle timestamp `time.Parse` cannot read | not-applicable: the reader refuses such a candle before the producer sees it (grammar, parse and ruling 26 minute alignment); defence for a foreign `CandleReader`, undriven by any interface-guard test | not-applicable | not-applicable |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
