# Function Logic Map: `minutes`

- Source: `internal/strategymarket/bars_test.go`
- Source SHA-256: `ffeec4a89d78fe4855e5ce30829c88c97f959a21d242602cf3b9b1221f81558f` (current worktree; equal to `source_sha256` in `ast.json`, regenerated 2026-08-18 after the edit)
- Signature: `minutes() []RawMinuteCandle`
- Source range: `16:1-23:2`
- AST counts: branches 1, returns 1, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Test-fixture bundle: this function is test-only. a117 edits it because the corrected label convention changes what a fixture must emit to mean the same bucket.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| none | - | the fixture's own literals | none; it builds a five-row page |

The fixture names the KRX bucket that opens at `09:00`. Under the corrected convention its five minute labels are `09:01`-`09:05`, not `09:00`-`09:04`; the bucket it means is unchanged, so every derived value in the tests that consume it (`OpenAt`, `ClosedAt`) is byte-identical to before.

## Branches and early returns

Exact AST return node: `22`. B1 (`range` at 18) fills five rows with consecutive labels starting one minute after the bucket opens. No early exit.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` / `Format` | build the RFC3339 label | none | `ast.json` |

## State mutations and fallbacks

None. Pure construction of a slice.

## Safety conclusion

- Safe edit boundary: test fixture only; no production symbol reads it.
- High-risk impact: no.
- Untested branch: none - B1 runs in every test that uses the fixture.
