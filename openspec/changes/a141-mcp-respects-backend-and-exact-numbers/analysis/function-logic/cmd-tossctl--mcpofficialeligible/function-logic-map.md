# Function Logic Map: `mcpOfficialEligible`

- Source: `cmd/tossctl/mcp.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Inputs retain existing JSON and backend contracts; no network or persistent state is changed.

## Branches and early returns

B1 returns true only for non-nil credentials, enabled OpenAPI, and non-WTS backend.

## Calls and live bindings

Uses only standard numeric/JSON helpers; no transport, credential, or session call is introduced.

## State mutations and fallbacks

Returns a derived value only. Invalid omission markers fail closed to a fresh count.

## Safety conclusion

- Safe edit boundary: local validation only.
- High-risk impact: numeric truncation metadata; covered by focused offline tests.
