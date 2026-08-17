# Branch Test Map: `strictMinuteObject`

- Source: `internal/official/strict_minute_candles.go`, SHA-256 `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced`; branch IDs follow `ast.json` (8 branches, regenerated 2026-08-18 against the decision-30 sources).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; in the fix round the function was reduced to a pure extractor so that the deep walk became the single authority for duplicate keys and trailing values (the masked-by-sibling mutant pair the adversary reported).
- **Decision-30 correction (2026-08-18).** The broker's `timestamp` is the bar's close instant, not its open (US probe 03:29 KST, review.md). This function's behaviour is unchanged; only its documentation and its source line numbers moved. Branch count unchanged.
- Tests: `internal/official/strict_minute_candles_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 447:2 — the opening token cannot be read | not-applicable: unreachable behind `strictMinuteCheckJSON` (an empty or truncated body is refused at `strictMinuteDecode` B3 first) | not-applicable | not-applicable |
| B2 | if at 450:2 — the whole body, `result`, or a candle given as a non-object | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `body is not an object`, `result is not an object`, `candle is not an object`) | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B3 | for at 454:2 — envelope, `result` and candle members extracted | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`, `TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys` | n/a (new code; pinned by mutation, see review.md) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B4 | if at 456:3 — the key token cannot be read | not-applicable: unreachable behind the whole-body walk | not-applicable | not-applicable |
| B5 | if at 460:3 — a non-string object key | not-applicable: `encoding/json` yields only string keys inside an object | not-applicable | not-applicable |
| B6 | if at 464:3 — the member value cannot be decoded | not-applicable: unreachable behind the whole-body walk | not-applicable | not-applicable |
| B7 | if at 470:2 — the closing token cannot be read | not-applicable: unreachable behind the whole-body walk | not-applicable | not-applicable |
| B8 | if at 473:2 — the closing token is not `}` | not-applicable: the member loop ends only at the matching `}` | not-applicable | not-applicable |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed across the two packages.
