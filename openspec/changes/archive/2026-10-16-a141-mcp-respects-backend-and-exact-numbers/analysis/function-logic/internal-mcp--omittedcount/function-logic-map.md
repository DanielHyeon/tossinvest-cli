# Function Logic Map: `omittedCount`

- Source: `internal/mcp/server.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Input is an omission marker value. Accept only nonnegative bounded counts from json.Number, integral finite in-range float64, or int. This function has no backend or credential inputs.

## Branches and early returns

B1 dispatches type; B2 parses json.Number; B3/B4 reject invalid float before conversion; B5 accepts non-negative int.

## Calls and live bindings

Uses only standard numeric/JSON helpers; no transport, credential, or session call is introduced.

## State mutations and fallbacks

Returns a derived value only. Invalid omission markers fail closed to a fresh count.

## Safety conclusion

- Safe edit boundary: local validation only.
- High-risk impact: numeric truncation metadata; covered by focused offline tests.
