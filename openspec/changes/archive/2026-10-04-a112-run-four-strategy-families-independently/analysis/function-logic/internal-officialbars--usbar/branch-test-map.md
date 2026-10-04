# Branch Test Map: `usBar`

- Source: `internal/officialbars/producer_test.go`, SHA-256 `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4`; branch IDs follow `ast.json` (0 branches, regenerated 2026-08-18 against the decision-30 sources, revision `current`).
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- The edit: at base the helper wrote the **open** instant onto the wire (`Timestamp: brokerTimestamp(t, open)`); it now delegates to `usBarLabelled(t, open.Add(time.Minute), closePrice)` because the 2026-08-18 03:29 KST probe proved the broker labels a bar with its **close**. The argument is still an open instant, so all ~30 call sites kept their expectations unchanged.
- Branchless helper: the AST reports no branches, so the single row below is the required happy-path row.
- Tests: `internal/officialbars/producer_test.go` — 30 tests in this file build their US fixtures through it.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path at 115–118 — an open instant in, a `RawMinuteCandle` out whose `Timestamp` is that instant **plus one minute** | exercised by every US fixture in the package; the ones that pin the `+1m` meaningfully are `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (stored opens must equal the arguments, not the wire values), `TestPollNeverAdmitsTheNewestObservedBar`, `TestPollCrawlsPagesInTheMeasuredShape` and `TestPollEndToEndThroughTheRealOfficialClient` (which serialises the helper's output as real JSON through `candleJSON` and lets the real `official.Client` parse it back) | n/a (test helper; the decision-30 RED was the 13 producer tests that failed one minute off, review.md 2026-08-18) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed, 0 skipped (2026-08-18) |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 — `ok internal/official 4.685s`, `ok internal/officialbars 1.548s`, 408 `=== RUN` / 408 PASS, 0 SKIP.
