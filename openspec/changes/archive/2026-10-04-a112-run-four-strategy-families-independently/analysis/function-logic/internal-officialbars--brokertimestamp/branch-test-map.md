# Branch Test Map: `brokerTimestamp`

- Source: `internal/officialbars/producer_test.go` **at the frozen base `d9d9f71f`** (revision `base`), SHA-256 `1d36e945513968599fdb553b18c1b8160436965aeb68ca345be054d25f12afde`; branch IDs follow `ast.json` (0 branches, regenerated 2026-08-18). Coordinates below are base coordinates — the same body sits at 107–110 in the working tree.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- Measured qualification: the helper's own four lines are **byte-identical** between `d9d9f71f` and the working tree (base 106–109, current 107–110, +1 line). The correction inserted the new `usBar` doc comment at base line 110 — the blank line right after its closing brace (`@@ -110,0 +112,3 @@`) — and `check_analysis` counts a zero-count hunk as touching `start..end+1`, which is why this helper is required at its base revision rather than its current one.
- Branchless helper: the AST reports no branches, so the single row below is the required happy-path row.
- Tests: `internal/officialbars/producer_test.go`.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path at base 106–109 — any instant in, the broker's wire notation out (`Asia/Seoul` fixed offset, exactly three fractional digits) | at base every fixture reached it through `usBar`, `krBar` and `cursorBefore`, and six tests called it directly: `TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical`, `TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded`, `TestPollInterfaceGuardDropsTheEqualInstantOverlapBar`, `TestPollInterfaceGuardRefusesAnAdjacentDuplicateInstant`, `TestPollTruncatesTheSubSecondPollInstant` and `TestPollEndToEndThroughTheRealOfficialClient`. The last two are the ones that can actually falsify the layout: `TestPollTruncatesTheSubSecondPollInstant` compares the producer's own `before` literal against a string built here, and the end-to-end test round-trips it through real JSON and the real `official.Client` parser, which refuses a `Z` offset or a four-digit fraction | n/a (test helper; the decision-30 RED was the 13 producer tests that failed one minute off, review.md 2026-08-18) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed, 0 skipped (2026-08-18) |

Note on the GREEN column: the command runs the **working tree**, not the base. That is honest here precisely because the base and current bodies are byte-identical — the run exercises the same four lines the base AST describes.

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 — `ok internal/official 4.685s`, `ok internal/officialbars 1.548s`, 408 `=== RUN` / 408 PASS, 0 SKIP.
