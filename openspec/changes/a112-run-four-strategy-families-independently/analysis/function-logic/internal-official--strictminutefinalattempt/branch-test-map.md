# Branch Test Map: `strictMinuteFinalAttempt`

- Source: `internal/official/strict_minute_candles.go`, SHA-256 `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d`; branch IDs follow `ast.json` (2 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: build-failure RED on the new symbols before GREEN; ruling 28 then changed the selector from "the last successful attempt" to "the last attempt, which must be 2xx" (mutant N19 equivalent by construction, both reviewers concurring).
- Tests: `internal/official/strict_minute_candles_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 229:2 — an empty attempt list | not-applicable: `c.get` returning nil implies `doRequest` emitted at least one trace (comment at 230; D6 declared untested) | not-applicable | not-applicable |
| B2 | if at 234:2 — the last attempt is an error or a non-2xx status | not-applicable: `c.get` already returned an error in that case (comment at 235); untaken side (401 then 200 ⇒ second attempt's `BodyDigest` and `ReadAt`) pinned by `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` | not-applicable | not-applicable |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
