# Branch Test Map: `Store.SealBarSeries`

- Source: `internal/strategyevidence/breakout_series.go`, SHA-256 `ece27d1ede03408e1f819f5d65f42fca9a252dd3e693b4cbadf834bee4e9abc5`; branch IDs follow `ast.json` (11 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: original delivery RED = build failure on the new symbols against the unmodified file (implementer report 2026-08-16); read-side binding and the NOT EXISTS removal were RED-first in the P1/P2 fix round (2026-08-17); the duplicate-minute guard was deleted post-recheck (g2) — no branch remains for it.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 61:2 — JP/"" market, blank symbol, empty/foreign-calendar session, 30 s interval, MaxBars 513/−1, zero cutoffs | `TestSealBarSeriesRefusesInvalidQuery` | yes (implementer report 2026-08-16: build-failure RED on the new symbols) | yes |
| B2 | if at 66:2 — prefix upper bound impossible | not-applicable: unreachable (prefix ends in `:`; defensive) | not-applicable | not-applicable |
| B3 | if at 79:2 — `QueryContext` error | not-applicable: driver failure, not injectable here | not-applicable | not-applicable |
| B4 | for at 83:2 — three bars inserted out of order | `TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest` | yes (implementer report 2026-08-16) | yes |
| B5 | if at 85:3 — a drifted/mismatched row aborts the read with zero bars | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`, `TestSealBarSeriesRefusesAPayloadThatDisagreesWithItsHeader`, `TestSealBarSeriesRefusesRatherThanHidingADriftedMinute`, `TestSealBarSeriesRefusesACorrectionThatMovesTheBar` | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B6 | if at 90:3 — taken side (`continue`): lower revision scanned after a higher one | `TestSealBarSeriesPicksTheHighestRevisionNumericallyNotTextually` (r1..r11 appended; asserts the SQL scan order is textual `r1,r10,r11,r2,…`; full cutoff → r11 with the `continue` side taken 8 times; cutoff at r10 → r10, proving numeric not textual comparison). Untaken side: `TestSealBarSeriesPrefersTheLatestVisibleRevision`, `TestSealBarSeriesCorrectionIsAppendOnlyAndReplayStable` | yes (compiling mutant `&& false` on the guard → winner degrades to r9; 2026-08-17) | yes |
| B7 | if at 95:2 — `rows.Close` error | not-applicable: driver failure | not-applicable | not-applicable |
| B8 | if at 98:2 — `rows.Err` error | not-applicable: driver failure | not-applicable | not-applicable |
| B9 | range at 102:2 — winners → slice | `TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest`, `TestSealBarSeriesRoundTripsEnvelopesBuiltByTheConstructor` | yes (implementer report 2026-08-16) | yes |
| B10 | if at 103:3 — `RegularSessionOnly` drops the extended-hours bar and changes the digest | `TestSealBarSeriesRegularSessionFilter` | yes (implementer report 2026-08-16) | yes |
| B11 | if at 109:2 — 513 bars refused (zero bars, empty digest), 512 accepted | `TestSealBarSeriesRefusesMoreThanFiveHundredTwelveBars` | yes (implementer report 2026-08-16; mutant 512→513 killed) | yes |

Scope/cutoff/digest properties (SQL predicates and the digest are calls, not branches, but the review cites them): `TestSealBarSeriesIgnoresOtherKindsSymbolsAndSessions`, `TestSealBarSeriesScopeExcludesLaterSessionsOtherKindsAndOtherSymbols`, `TestSealBarSeriesDualCutoffIndependence`, `TestSealBarSeriesDigestGoldenVector`, `TestSealBarSeriesDigestChangesWithEveryInput`, `TestSealBarSeriesDigestIsDomainSeparatedFromSnapshotDigest`, `TestSealBarSeriesIsStructurallySelectOnly`, `TestClosedBarAppendQuarantinesSameRevisionWithADifferentDigest` (quarantined payload never reaches the series).

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; reproduced by both reviewers (review.md 2026-08-17).
