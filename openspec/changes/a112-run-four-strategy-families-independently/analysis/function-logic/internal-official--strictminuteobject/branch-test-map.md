# Branch Test Map: `strictMinuteObject`

- Source: `internal/official/strict_minute_candles.go`, SHA-256 `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d`; branch IDs follow `ast.json` (8 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; in the fix round the function was reduced to a pure extractor so that the deep walk became the single authority for duplicate keys and trailing values (the masked-by-sibling mutant pair the adversary reported).
- Tests: `internal/official/strict_minute_candles_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 435:2 — the opening token cannot be read | not-applicable: unreachable behind `strictMinuteCheckJSON` (an empty or truncated body is refused at `strictMinuteDecode` B3 first) | not-applicable | not-applicable |
| B2 | if at 438:2 — the whole body, `result`, or a candle given as a non-object | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `body is not an object`, `result is not an object`, `candle is not an object`) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B3 | for at 442:2 — envelope, `result` and candle members extracted | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`, `TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B4 | if at 444:3 — the key token cannot be read | not-applicable: unreachable behind the whole-body walk | not-applicable | not-applicable |
| B5 | if at 448:3 — a non-string object key | not-applicable: `encoding/json` yields only string keys inside an object | not-applicable | not-applicable |
| B6 | if at 452:3 — the member value cannot be decoded | not-applicable: unreachable behind the whole-body walk | not-applicable | not-applicable |
| B7 | if at 458:2 — the closing token cannot be read | not-applicable: unreachable behind the whole-body walk | not-applicable | not-applicable |
| B8 | if at 461:2 — the closing token is not `}` | not-applicable: the member loop ends only at the matching `}` | not-applicable | not-applicable |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
