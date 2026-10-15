# Function Logic Map: `TestToolResultKeepsSmallPayload`

- Source: `internal/mcp/server_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Inputs retain existing JSON and backend contracts; no network or persistent state is changed.

## Branches and early returns

B1 existing error assertion; unchanged function is included because the same source file gained regressions.

## Calls and live bindings

Uses only standard numeric/JSON helpers; no transport, credential, or session call is introduced.

## State mutations and fallbacks

Returns a derived value only. This test compares the small response text with its expected JSON; it does not exercise omission-marker handling.

## Safety conclusion

- Safe edit boundary: local validation only.
- High-risk impact: none in the test function itself; it protects unchanged small-response serialization.
