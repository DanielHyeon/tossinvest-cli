# Function Logic Map: `brokerTimestamp`

- Source: `internal/officialbars/producer_test.go`
- **Revision: `base`.** The AST and every coordinate below describe the file **at the frozen comparison base `d9d9f71f`**, not the working tree. Read the body with `git show d9d9f71f:internal/officialbars/producer_test.go`.
- Source SHA-256: `1d36e945513968599fdb553b18c1b8160436965aeb68ca345be054d25f12afde` (the base file; `git show d9d9f71f:internal/officialbars/producer_test.go | sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `brokerTimestamp(t *testing.T, instant time.Time) string`
- Source range: `106:1`–`109:2` at the base (ast.json `start`/`end`); the same body sits at `107:1`–`110:2` in the working tree
- AST evidence: `ast.json` regenerated 2026-08-18, revision `base`; branches 0, returns 1, calls 4, assignments 0, defers 0, go statements 0.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- Why this bundle is base-revision, measured rather than assumed: the correction inserted the new `usBar` doc comment at base line 110 — the blank line **immediately after** this helper's closing brace (`git diff --unified=0 d9d9f71f` reports `@@ -110,0 +112,3 @@`). `check_analysis`'s `intersects()` treats a zero-count hunk as touching `start..end+1`, so the base range 106–109 matches at 110 while the current range 107–110 does not reach the inserted lines 112–114. The helper is therefore required at its **base** revision, and its own four lines are byte-identical between `d9d9f71f` and the working tree.
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- Base-side body, verbatim: `t.Helper()` then `return instant.In(mustLocation(t, marketclock.MarketKR)).Format("2006-01-02T15:04:05.000-07:00")`.
- The single formatting authority for every wire instant in this test package. Its base doc comment states the contract: "브로커가 쓰는 그대로의 시각 표기다(KST 고정 오프셋, 밀리초 셋)" — the broker's own notation, a fixed KST offset with exactly three fractional digits.
- Two measured properties are encoded in that one layout string: the offset is **numeric** (`+09:00`, never `Z`) and the fraction is exactly **three digits**. Both are what `official.strictMinuteInstant` refuses when they are wrong, so a fixture built here is guaranteed to be accepted by the reader for the right reason.
- **Asia/Seoul for both markets.** A US bar's wire instant is rendered in KST too, because that is the measured request and response shape — the same reason `PollClosedBars` formats its `before` bound through `marketclock.MarketKR.Location()`.
- What it does **not** do, at base or now: it attaches no meaning to the instant it formats. It is the layer below the label-vs-open question, which is exactly why decision 30 left it untouched while rewriting its callers.

## Branches and early returns

The AST reports **0 branches** at the base revision: `t.Helper()` then one unconditional `return` (AST return node `108`). Nothing to enumerate, so the table carries a single `—` row.

| Branch | AST kind | Source location (base) | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| — | (none) | 106–109 | branchless: `t.Helper()` then one unconditional return | happy path only; see `branch-test-map.md` B1 |

## Calls and live bindings

| Callee expression | Source location (base) | Evidence |
|---|---|---|
| `t.Helper()` | 107 | failures attribute to the calling test, not to the helper |
| `mustLocation(t, marketclock.MarketKR)` | 108 | loads `Asia/Seoul`; fails the test outright if the zone is unavailable, so no caller can receive a wrongly-offset string |
| `instant.In(…)` | 108 | re-expresses the instant in KST without changing it |
| `.Format("2006-01-02T15:04:05.000-07:00")` | 108 | the measured wire layout: numeric offset, exactly three fractional digits |

## State mutations and fallbacks

- No assignments (AST `assignments` 0), no package state, no I/O, no goroutines, no defers, no clock read. Pure function of its two arguments.
- No fallback: the only failure path is `mustLocation`, and it fails the test rather than returning a string with a wrong or missing offset.

## Safety conclusion

- Safe edit boundary: a test formatting helper. It is not linked into production and touches no order, stop-loss, sizing, guardian, journal or toggle surface.
- High-risk impact: no production branch behind it. Its value is that it is the **one** place the wire layout is written down, so the whole fixture corpus moves together if the measured layout is ever re-measured.
- Decision-30 relationship, stated precisely: this helper was **not** semantically changed. It formats whatever instant it is handed; what the correction changed is which instant its callers hand it (`usBar`/`krBar` now pass `open + 1m`, `cursorBefore` now passes the oldest bar's open). The gate requires this bundle because of an adjacency in the diff, not because its logic moved — and recording that distinction is the honest answer rather than inventing a change it did not have.
