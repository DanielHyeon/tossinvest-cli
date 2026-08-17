# Branch Test Map: `strictMinuteCandles`

- Source: `internal/official/strict_minute_candles.go`, SHA-256 `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d`; branch IDs follow `ast.json` (7 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols, then behaviour RED on the candle key-count rule (`unknown candle key`) before GREEN.
- Tests: `internal/official/strict_minute_candles_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 292:2 — `candles` given as an object and as `null` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `candles is not an array`, `candles is null`) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B2 | if at 296:2 — the element split fails | not-applicable: `strictMinuteCheckJSON` already proved the body well-formed and B1 proved the value is an array; defensive | not-applicable | not-applicable |
| B3 | if at 299:2 — three candles for a count of two refused, for a count of three accepted | `TestStrictMinuteCandlesRefusesMoreCandlesThanRequested` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B4 | range at 305:2 — one- and two-candle pages decoded in order | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`, `TestStrictMinuteCandlesAcceptsTheKoreanMarket` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B5 | if at 307:3 — fourteen per-candle refusals propagate unchanged | `TestStrictMinuteCandlesRefusesMalformedBodies` (candle subtests incl. `unknown candle key`, `missing candle key`, `bare number price`, `foreign currency`, `instant with non-zero seconds`) | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B6 | if at 310:3 — ascending pair and duplicated instant | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `ascending instants`, `duplicate instants`) | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B7 | if at 314:3 — a bar one minute past `before` refused; a bar exactly at `before` accepted | `TestStrictMinuteCandlesRefusesInstantsNewerThanBefore` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
