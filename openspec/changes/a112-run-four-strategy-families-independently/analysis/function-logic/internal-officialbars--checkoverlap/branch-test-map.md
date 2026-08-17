# Branch Test Map: `checkOverlap`

- Source: `internal/officialbars/producer.go`, SHA-256 `83960410b7b870ca60fad568002060de49ebc7271d72b94d46f665ff274b29b1`; branch IDs follow `ast.json` (3 branches, generated 2026-08-17 after GREEN).
- New function (lot L1b, not in the frozen base 016da624). RED per review.md: the first cut had no overlap check at all and the byte-identity test (then named TestPollRefusesAnOverlapThatIsNotByteIdentical, now `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical`) failed before GREEN; ruling 27 then rewrote the rule (forward-walk refusal, byte equality only for equal instants), kept the function as a declared interface guard and renamed both tests `TestPollInterfaceGuard…`.
- Tests: `internal/officialbars/producer_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 420:2 — well-formed successive pages fall through | `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B2 | case at 421:2 — page 2 starts newer than page 1's oldest bar | not-applicable through `official.Client` (ruling 27; comment at 413–418); interface guard `TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |
| B3 | case at 425:2 — the shared minute differs between the two pages | not-applicable through `official.Client` (ruling 27); interface guard `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical`; accepting side `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar` | build/behaviour RED captured by the implementer 2026-08-17 (review.md L1b implementation report) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 (2026-08-17) |

Verification (this bundle, 2026-08-17): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 404 tests passed across the two packages.
