# Branch Test Map: `strictMinuteCursor`

- Source: `internal/official/strict_minute_candles.go`, SHA-256 `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d`; branch IDs follow `ast.json` (6 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: the first cut treated an absent `nextBefore` as terminal and the `cursor absent` subtest failed before GREEN; the recheck's "cursor sub-minute acceptance untested" P2 was closed by an explicit accepting test.
- Tests: `internal/official/strict_minute_candles_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 385:2 — a two-key `result` whose second key is `hasMore`, so the key-count rule passes and only this rule refuses | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `cursor replaced by another key`) | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B2 | if at 390:2 — `null` cursor ⇒ terminal page, empty `NextBefore` | `TestStrictMinuteCandlesAcceptsTheKoreanMarket`, `TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys`, `TestStrictMinuteCandlesAcceptsNestingUpToTheBound` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B3 | if at 393:2 — number, object, array and boolean cursors | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `cursor is a number`, `cursor is an object`, `cursor is an array`, `cursor is a boolean`) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B4 | if at 397:2 — an empty-string cursor | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `cursor is an empty string`) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B5 | if at 401:2 — a cursor with a `Z` offset | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `cursor breaks the timestamp grammar`) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B6 | if at 404:2 — cursor equal to and newer than the oldest bar refused; empty page accepted; sub-minute cursor accepted | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `cursor is not older than the oldest bar`, `cursor is newer than the oldest bar`), `TestStrictMinuteCandlesAcceptsAnEmptyPageThatStillCarriesACursor`, `TestStrictMinuteCandlesAcceptsASubMinuteCursor` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
