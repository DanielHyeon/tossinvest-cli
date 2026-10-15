# Function Logic Map: `mcpOfficialEligible`

- Source: `cmd/tossctl/mcp.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Inputs are optional credentials, OpenAPI config and resolved backend preference. Return true exactly when credentials != nil AND OpenAPI.Enabled AND preference != "wts". Caller resolves preference first.

## Branches and early returns

AST branches is null: there is no B1. A single return expression uses Go short-circuit conjunction of the three eligibility conditions. This is a capability predicate, not a credential parser or network authorization request.

## Calls and live bindings

No function calls. newMCPCmd uses the result before official/trading dependency construction.

## State mutations and fallbacks

No mutation, persistence or fallback. False leaves officialClient/tradingSvc absent in the caller.

## Safety conclusion

High-risk capability boundary. TestMCPOfficialEligibilityMatchesCLI covers each condition; guarded command tests verify its actual use. Enabled/WTS gate removal mutants are detected. Journal, token storage and trading confirmations remain unchanged.
