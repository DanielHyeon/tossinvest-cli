# Branch Test Map: `strictMinuteFinalAttempt`

- Source: `internal/official/strict_minute_candles.go`, SHA-256 `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced`; branch IDs follow `ast.json` (2 branches, regenerated 2026-08-18 against the decision-30 sources).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; ruling 28 then changed the selector from "the last successful attempt" to "the last attempt, which must be 2xx" (mutant N19 equivalent by construction, both reviewers concurring).
- **Decision-30 correction (2026-08-18).** The broker's `timestamp` is the bar's close instant, not its open (US probe 03:29 KST, review.md). This function's behaviour is unchanged; only its documentation and its source line numbers moved. Branch count unchanged.
- Tests: `internal/official/strict_minute_candles_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 240:2 — an empty attempt list | not-applicable: `c.get` returning nil implies `doRequest` emitted at least one trace (comment at 241; D6 declared untested) | not-applicable | not-applicable |
| B2 | if at 245:2 — the last attempt is an error or a non-2xx status | not-applicable: `c.get` already returned an error in that case (comment at 246); untaken side (401 then 200 ⇒ second attempt's `BodyDigest` and `ReadAt`) pinned by `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` | not-applicable | not-applicable |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed across the two packages.
