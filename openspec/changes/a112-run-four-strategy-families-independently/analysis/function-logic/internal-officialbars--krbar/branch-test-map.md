# Branch Test Map: `krBar`

- Source: `internal/officialbars/producer_test.go`, SHA-256 `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4`; branch IDs follow `ast.json` (0 branches, regenerated 2026-08-18 against the decision-30 sources, revision `current`).
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- The edit: at base it put the **open** instant on the wire (`Timestamp: brokerTimestamp(t, open)`); it now delegates to `krBarLabelled(t, open.Add(time.Minute), closePrice)` because the 2026-08-18 03:29 KST probe proved the broker's label is the bar's **close**. The argument is still an open instant, so the caller's expectations did not move.
- Branchless helper: the AST reports no branches, so the single row below is the required happy-path row.
- Tests: `internal/officialbars/producer_test.go`. Exactly one test builds fixtures through it.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path at 131–134 — an open instant in, a KRW `RawMinuteCandle` out whose `Timestamp` is that instant **plus one minute** | `TestPollUsesTheKoreanCalendarAndSessionPrefix` is the only caller: it asserts the `KRX:2026-08-14` session prefix and the KST-formatted request bound against bars whose expectations are written as opens, so the `+1m` is proved by the stored open matching the argument rather than the wire value. The KR boundary case is deliberately elsewhere — `TestPollTreatsTheKoreanClosingLabelAsTheLastRegularBar` uses `krBarLabelled` because there the label is the subject | n/a (test helper; the decision-30 RED was the 13 producer tests that failed one minute off, review.md 2026-08-18) | `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0, 408 tests passed, 0 skipped (2026-08-18) |

Verification (this bundle, 2026-08-18): `go test ./internal/official ./internal/officialbars -count=1 -race` exit 0 — `ok internal/official 4.685s`, `ok internal/officialbars 1.548s`, 408 `=== RUN` / 408 PASS, 0 SKIP.
