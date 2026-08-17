# Branch Test Map: `checkOverlap`

- Source: `internal/officialbars/producer.go`, SHA-256 `8d45ca93b090cfe9e10a93e5a658991ed3376820b56dfa05e49b809171c16772`; branch IDs follow `ast.json` (3 branches, regenerated 2026-08-18 against the decision-30 sources).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: the first cut had no overlap check at all and the byte-identity test (then named TestPollRefusesAnOverlapThatIsNotByteIdentical, now `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical`) failed before GREEN; ruling 27 then rewrote the rule (forward-walk refusal, byte equality only for equal instants), kept the function as a declared interface guard and renamed both tests `TestPollInterfaceGuard…`.
- **Decision-30 correction (2026-08-18).** The broker's `timestamp` is the bar's close instant, not its open (US probe 03:29 KST, review.md). This function's behaviour is unchanged; only its documentation and its source line numbers moved. Branch count unchanged.
- Tests: `internal/officialbars/producer_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 445:2 — well-formed successive pages fall through | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B2 | case at 446:2 — page 2 starts newer than page 1's oldest bar | not-applicable through `official.Client` (ruling 27; comment at 438–443); interface guard `TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |
| B3 | case at 450:2 — the shared minute differs between the two pages | not-applicable through `official.Client` (ruling 27); interface guard `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical`; accepting side `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-18) |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed across the two packages.
