# Function Logic Map: `verifiedBarFixtureAt`

- Source: `internal/strategyengine/lane_test.go`
- Source SHA-256: `6b224a50c8db46a7c66b6fe0d0ef120bb103045b54a715f6a13a64f9b0c8fdf2` (current worktree; equal to `source_sha256` in `ast.json`, regenerated 2026-08-18 after the edit)
- Signature: `verifiedBarFixtureAt(t *testing.T, start time.Time, mutate func([]strategymarket.RawMinuteCandle)) strategymarket.VerifiedBar`
- Source range: `76:1-96:2`
- AST counts: branches 3, returns 1, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Test-fixture bundle: this function is test-only. a117 edits it because the corrected label convention changes what a fixture must emit to mean the same bucket.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `start` | the instant the five-minute bucket **opens** | every call site | `t.Fatal` on seal error |

`start` always meant the bucket's opening instant; the fixture previously emitted labels starting at that instant, which was correct only under the old reading. It now emits `start+1m` .. `start+5m`, so `OpenAt()` and `ClosedAt()` come out exactly as before and no lane assertion in this file moves.

## Branches and early returns

Exact AST return node: `95`. B1 (`range` 79) builds the five rows - the edited line is inside it; B2 (`if` 87) applies the caller's mutator; B3 (`if` 92) fails the test if the seal refuses.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategymarket.SealAdaptedOfficialMinutePage` / `SealOfficialClosedKRXFiveMinute` | mint the `VerifiedBar` the lane consumes | refusal ⇒ `t.Fatal` at B3 | `ast.json`; CodeGraph caller list |

## State mutations and fallbacks

None beyond test-local state.

## Safety conclusion

- Safe edit boundary: test fixture only; the production lane is untouched.
- High-risk impact: no.
- Untested branch: none - all three run across the file's sixteen call sites.
