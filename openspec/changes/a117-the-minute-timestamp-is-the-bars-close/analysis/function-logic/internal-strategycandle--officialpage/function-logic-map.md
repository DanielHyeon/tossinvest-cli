# Function Logic Map: `officialPage`

- Source: `internal/strategycandle/adapter_test.go`
- Source SHA-256: `d0f227759181a0727ba87ed06ac6f5d163bafd048e5429a2df7778510bb717dc` (current worktree; equal to `source_sha256` in `ast.json`, regenerated 2026-08-18 after the edit)
- Signature: `officialPage(t *testing.T, adjusted bool) official.RawMinutePage`
- Source range: `15:1-40:2`
- AST counts: branches 5, returns 1, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Test-fixture bundle: this function is test-only. a117 edits it because the corrected label convention changes what a fixture must emit to mean the same bucket.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| served JSON body | five candle rows | the fixture's own literal | `t.Fatal` on reader error |

The five literal timestamps are shifted one minute later so the page still means the bucket that opens at `09:00`.

## Branches and early returns

Exact AST return node: `39`. B1 (`switch` 18) routes the fake server; B2/B3/B4 are its `case` arms (token, candles, not-found); B5 (`if` 36) fails the test on a reader error. a117 changes only the literal timestamps inside the B3 arm.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `official.Client.RawMinuteCandles` | drive the real KR reader over the fake server | error ⇒ `t.Fatal` at B5 | `ast.json` |

## State mutations and fallbacks

Registers a `t.Cleanup` server shutdown; no other state.

## Safety conclusion

- Safe edit boundary: test fixture only.
- High-risk impact: no.
- Untested branch: B4 (the not-found arm) is defensive and never taken.
