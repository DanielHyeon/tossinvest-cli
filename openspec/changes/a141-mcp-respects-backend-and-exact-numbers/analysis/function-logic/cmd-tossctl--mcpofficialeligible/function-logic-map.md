# Function Logic Map: `cmd-tossctl--mcpofficialeligible`

- Source: `cmd/tossctl/mcp.go`
- AST evidence: `ast.json`

## Invariant

credentials AND enabled AND preference not wts.

## Tests

`TestMCPOfficialEligibilityMatchesCLI; TestMCPCmdWTSDoesNotWireOfficialDependencies`.
