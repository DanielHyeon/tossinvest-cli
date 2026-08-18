# Function Logic Map: `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders`

- Source: `internal/strategymarket/bars_test.go`
- Source SHA-256: `b94852735790b61add409d2cf155b54e1b7113bcfc389bc120d7236c722c8ed3` (the a117 comparison base `73a1e85c`; `ast.json` carries `"revision": "base"` because the body is unchanged and only the file revision moved)
- Signature: `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders(t *testing.T)`
- Source range: `153:1-177:2`
- AST counts: branches 4, returns 0, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Test-fixture bundle: this function is test-only. a117 edits it because the corrected label convention changes what a fixture must emit to mean the same bucket.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| stubbed position/order readings | fresh, authoritative, zero | the test's own stubs | `t.Fatalf` |

**Why this bundle exists:** a117 appended three new tests to the end of `bars_test.go`, and this function is the last one before that append. Its body is unchanged - the checker flags it because the file revision moved. Recorded rather than silently exempted.

## Branches and early returns

No return nodes (a test body). B1 (`if` 157), B2 (`if` 160), B3 (`range` 163) and B4 (`if` 173) are the fixture's own assertions and are untouched by a117.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NoPositionProof` constructors and `t.Fatalf` | assert the fail-closed proof rules | none | `ast.json` |

## State mutations and fallbacks

None beyond test-local state.

## Safety conclusion

- Safe edit boundary: not edited by a117; only its file position moved.
- High-risk impact: no.
- Untested branch: not-applicable - this is itself a test.
