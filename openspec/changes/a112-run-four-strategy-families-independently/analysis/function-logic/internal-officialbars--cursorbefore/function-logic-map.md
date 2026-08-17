# Function Logic Map: `cursorBefore`

- Source: `internal/officialbars/producer_test.go`
- Source SHA-256: `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `cursorBefore(t *testing.T, oldestOpen time.Time) string`
- Source range: `219:1`–`222:2` (ast.json `start`/`end`), revision `current`
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 0, returns 1, calls 2, assignments 0, defers 0, go statements 0.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- Builds the `nextBefore` cursor a page would carry, from the open instant of that page's **oldest** bar, so a fixture never has to spell the measured cursor rule out by hand.
- **What decision 30 changed — the arithmetic, not the value.** The measured contract is "cursor = the page's oldest bar's **label** minus one minute". At base, when the wire carried the *open* instant, that was written `brokerTimestamp(t, oldest.Add(-time.Minute))`. Once the label was proved to be the **close** (`open + 1m`), the same rule reads `label − 1m = (open + 1m) − 1m = open`, so the body is now `brokerTimestamp(t, oldestOpen)` and the `-time.Minute` is gone. The parameter was renamed `oldest` → `oldestOpen` to say which space it is in.
- This is a *simplification that preserves the emitted string*: for the same bar, base and current produce the same cursor text. What changed is which instant the fixtures pass in, because the bars themselves moved.
- The rule it encodes is the reader's, not the producer's: `strictMinuteCursor` requires the cursor to be strictly older than the page's oldest bar, and the producer's own D4 guard (`PollClosedBars` B27) requires it to be strictly older than the bound the request carried. Both live in **label space** and are never converted.
- `t.Helper()` first, so a formatting failure is reported at the calling test's line.

## Branches and early returns

The AST reports **0 branches**: `t.Helper()` then one unconditional `return` (AST return node `221`). Nothing to enumerate, so the table carries a single `—` row.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| — | (none) | 219–222 | branchless: `t.Helper()` then one unconditional return | happy path only; see `branch-test-map.md` B1 |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `t.Helper()` | 220 | failures attribute to the calling test |
| `brokerTimestamp(t, oldestOpen)` | 221 | formats the instant into the broker's wire layout (fixed KST offset, three fractional digits). The `.Add(-time.Minute)` that stood here at base is **deliberately absent** — under decision 30 the subtraction is already inside the label→open relation |

## State mutations and fallbacks

- No assignments (AST `assignments` 0), no package state, no I/O, no goroutines, no defers, no clock read. Pure function of its two arguments.
- No fallback: the only failure path is `brokerTimestamp` → `mustLocation`, which fails the test rather than returning a malformed cursor.

## Safety conclusion

- Safe edit boundary: a test fixture builder; no production linkage, no order, stop-loss, sizing, guardian, journal or toggle surface.
- High-risk impact: no production branch behind it. But it is the fixture that makes the crawl tests honest: if it emitted a cursor that was **not** strictly older than the page's oldest bar, the reader's cursor rule and the producer's `CURSOR_LOOP` guard would fire and the pagination tests would fail loudly rather than pass falsely — the failure mode is visible, not silent.
- Recorded residual (review.md 2026-08-17), unchanged by this correction: the KR pagination contract is documented but unmeasured — M-B measured US only, so the first KR full-session crawl under L5 is a human-observed run. This helper encodes the US-measured cursor rule for both markets.
