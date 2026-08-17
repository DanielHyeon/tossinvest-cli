# Function Logic Map: `krBar`

- Source: `internal/officialbars/producer_test.go`
- Source SHA-256: `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `krBar(t *testing.T, open time.Time, closePrice string) official.RawMinuteCandle`
- Source range: `131:1`–`134:2` (ast.json `start`/`end`), revision `current`
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 0, returns 1, calls 3, assignments 0, defers 0, go statements 0.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- The KRX twin of `usBar`: given the instant a bar **opens**, return one `official.RawMinuteCandle` shaped as the broker sends it, with KRW fields (`Open` 71000, `High` 71100, `Low` 70900, `Volume` 12, `Currency` KRW) and a caller-chosen close.
- **What decision 30 changed.** At base the struct was built inline with `Timestamp: brokerTimestamp(t, open)` — the open instant on the wire. It now delegates to `krBarLabelled(t, open.Add(time.Minute), closePrice)`, because the 2026-08-18 03:29 KST US probe proved the label is the bar's **close**. The argument stays an open instant, so the caller's expectation did not move.
- Honest scope limit, carried from review.md: KR's own labelling convention is **formally unmeasured** — the US probe measured US, and the KR 2026-08-17 09:10 observation (a bar labelled 20:00 KST carrying the 150,860-volume NXT closing print) only corroborates it. This helper applies the same convention to KR because the producer does; the KR probe stands for 2026-08-18 09:00–15:30 KST.
- `t.Helper()` first, so a fixture failure is reported at the calling test's line.

## Branches and early returns

The AST reports **0 branches**: `t.Helper()` then one unconditional `return` (AST return node `133`). Nothing to enumerate, so the table carries a single `—` row rather than an invented branch.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| — | (none) | 131–134 | branchless: `t.Helper()` then one unconditional return | happy path only; see `branch-test-map.md` B1 |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `t.Helper()` | 132 | failures attribute to the calling test |
| `open.Add(time.Minute)` | 133 | **the decision-30 conversion, in the test direction**: open instant → broker close label |
| `krBarLabelled(t, …, closePrice)` | 133 | the label-space builder added by the correction (`producer_test.go:136`); it owns the KRW literals and the `brokerTimestamp` formatting |

## State mutations and fallbacks

- No assignments (AST `assignments` 0), no package state, no I/O, no goroutines, no defers, no clock read. Pure function of its three arguments.
- No fallback: any failure comes from `krBarLabelled` → `brokerTimestamp` → `mustLocation` and fails the test outright rather than returning a partial candle.

## Safety conclusion

- Safe edit boundary: a test fixture builder, never linked into production, touching no order, stop-loss, sizing, guardian, journal or toggle surface.
- High-risk impact: no production branch behind it. Its one duty is that KR fixtures speak the same two clock spaces as US fixtures, so that `TestPollUsesTheKoreanCalendarAndSessionPrefix` keeps proving KR session identity against **open** instants while the wire carries labels.
- The KR boundary case that decision 30 actually turns on — the bar labelled 15:30 KST opening at 15:29 — is deliberately **not** built through this helper: `TestPollTreatsTheKoreanClosingLabelAsTheLastRegularBar` uses `krBarLabelled` directly, because there the label itself is the subject.
