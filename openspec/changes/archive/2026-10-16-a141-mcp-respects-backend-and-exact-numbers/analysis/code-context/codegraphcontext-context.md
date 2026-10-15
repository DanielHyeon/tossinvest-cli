# CodeGraphContext supporting evidence

`codegraph context` for `newMCPCmd` lists the CLI app constructor and the same config, credential, WTS-session, official-client, trading-service, and MCP-server dependencies used by this command. The adjacent `internal/app.New` source is the authoritative parity reference.

`codegraph context` for `toolResult` identifies `handleToolsCall`, `shrinkToCap`, `maxResultBytes`, and existing MCP result tests. `trim` is the downstream marker-count consumer and must accept the exact-number decoder representation.

This context is advisory. The implementation boundary and branch claims are verified against current source and the AST bundles under `analysis/function-logic/`.

