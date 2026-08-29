# Review

## Function Logic Map

Five bundles under `analysis/function-logic/`, all produced before the corresponding edit:

- `internal-strategymarket--aggregateclosedkrxfiveminute` — the edited production function. Written against the pre-edit revision `a6f17d0c…` and regenerated against the post-edit revision `aa6c93cd…`; 18 branches, 13 returns, 37 calls, unchanged by the edit. B7, B12 and the two instants in the final return are the edit surface.
- `internal-strategymarket--minutes`, `internal-strategycandle--officialpage`, `internal-strategyengine--verifiedbarfixtureat` — the three test fixtures whose labels shift by one minute so each keeps naming the same bucket.
- `internal-strategymarket--testnopositionproofrequiresfreshauthoritativezeropositionandorders` — carried at `"revision": "base"`. Its body is unchanged; it is required only because the three new tests were appended after it and the file revision moved. Recorded rather than silently exempted (the "logic map is always larger than the plan" rule).

## RED → GREEN

Three REDs were captured before any production line changed, and each failed for the reason predicted:

| RED | Observed failure before the fix |
|---|---|
| `TestTheBucketOpensOneMinuteBeforeItsFirstLabel` | `strategy market: incomplete_bucket: not aligned to KRX five-minute boundary` — the genuine first bucket of every KRX session was being refused |
| `TestABarLabelledAtTheOpenIsOutsideTheRegularSession` | `err=<nil>` — a bucket containing the pre-open minute `08:59`–`09:00` was accepted as regular |
| `TestTheLastRegularMinuteLabelIsAdmitted` | `strategy market: outside_regular_session: 2026-07-31T15:30:00+09:00` — the session's true last minute was being discarded |

All three pass after the fix. Full suite: `go test ./... -count=1` green, **9,425 tests across 102 packages, exit 0**; `go vet` clean; `$(go env GOROOT)/bin/gofmt -l internal/ tools/ cmd/` empty.

## Mutation evidence

The new tests were made to fail on purpose, with `go test -overlay` so the working tree was never written to: `m1_window` (session window reverted), `m2_align` (alignment anchored one minute off), `m3_openat` (open derivation reverted to the label). **All three KILLED.** Restoration was verified by counting the two edited symbols in `bars.go` (each exactly 1) and re-running the package suite (24 tests green) — not by `git checkout`, which would have erased the GREEN work along with the mutation.

## Blast radius

CodeGraph (live index 34,752 nodes): the aggregator's callers are `AdaptAndSealClosedKRXFiveMinute` (`strategycandle/adapter.go:32`) and `SealOfficialClosedKRXFiveMinute` (`bars.go:111`), whose own callers are all tests; the only consumer of the resulting `VerifiedBar` is `strategyengine.LaneInput.Bar`, and no file outside `internal/strategyengine` constructs that struct. The defect was therefore **latent** — it could only mis-label bars inside tests. Stated plainly because the opposite claim would have been an easy and false way to make this change look urgent.

Because each fixture keeps naming the same bucket, `OpenAt()` and `ClosedAt()` come out byte-identical to before for every existing lane assertion; that is why `internal/strategyengine`'s session-boundary, cutoff and provenance tests needed no edit at all.

## Direction of the behaviour change

Conservative where it can be: the pre-open minute that was admitted becomes a refusal. The session's genuine last minute becomes admissible — a widening, but only back to the `09:00`–`15:30` boundary the spec already froze, never past it. No refusal kind, threshold, stop, sizing or exposure value moves.

## Correction unit

The unit is the *value* — "the broker's minute timestamp is the bar's open" — not a single file:line. Three copies existed:

1. `internal/official/candle_reads.go:16` — corrected here, with the reason stated in place.
2. `internal/strategymarket.aggregateClosedKRXFiveMinute` — corrected here.
3. `docs/migration/openapi.latest.json` — the broker's own document, mirrored for reference. **Not edited**: correcting it would fork it silently from its source. `candle_reads.go` now records why this repository disagrees with it.

Two further copies were already correct and are left alone: `internal/official/strict_minute_candles.go:71` and `internal/officialbars/producer.go:409` both already say the document is wrong.

## Ownership

Edited: `internal/strategymarket/bars.go`, `internal/official/candle_reads.go`, and the three test files above. Not edited: any order, stop, sizing, Guardian, ledger, reconciliation, auth or fill path; any `cmd/`, engine wiring, router, scheduler, journal, toggle or container file; `internal/officialbars/**`; `docs/migration/openapi.latest.json`. No production caller is added or removed, and no toggle or container moves.
