# Branch Test Map: `cursorBefore`

- Source: `internal/officialbars/producer_test.go`, SHA-256 `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4`; branch IDs follow `ast.json` (0 branches, regenerated 2026-08-18 against the decision-30 sources, revision `current`).
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- The edit: `brokerTimestamp(t, oldest.Add(-time.Minute))` → `brokerTimestamp(t, oldestOpen)`. The measured rule is unchanged — cursor = the page's oldest bar's **label** minus one minute — but once the label was proved to be the close (`open + 1m`), that expression collapses to the bar's own open, so the explicit `−1m` was removed and the parameter renamed to name its clock space.
- Branchless helper: the AST reports no branches, so the single row below is the required happy-path row.
- Tests: `internal/officialbars/producer_test.go` — eight tests build their page cursors through it.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path at 219–222 — the oldest bar's open instant in, the measured `nextBefore` string out | the eight callers are `TestPollCrawlsPagesInTheMeasuredShape`, `TestPollRefusesACursorThatDoesNotMoveBackwards`, `TestPollStopConditions`, `TestPollStopsAtAnExplicitLowerBound`, `TestPollDefaultsToFourPages`, `TestPollKeepsThePageCountWhenTheReaderFails`, `TestPollRefusesAMergedListThatIsNotStrictlyDescending` and `TestPollEndToEndThroughTheRealOfficialClient`. The last is the strongest: its `httptest` handler dispatches on the exact `before` string the producer sends, so a cursor built here that did not round-trip byte for byte would fall into the handler's `default` arm and fail the test | n/a (test helper; the decision-30 RED was the 13 producer tests that failed one minute off, review.md 2026-08-18) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed, 0 skipped (2026-08-18) |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 — `ok internal/official 4.685s`, `ok internal/officialbars 1.548s`, 408 `=== RUN` / 408 PASS, 0 SKIP.
