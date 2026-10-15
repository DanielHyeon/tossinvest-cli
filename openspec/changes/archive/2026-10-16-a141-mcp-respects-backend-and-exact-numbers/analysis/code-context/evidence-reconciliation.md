# Evidence reconciliation

- Baseline: `6765858af1beb472268cb699efa1964addb84d60`.
- CodeGraph and current source agree that `newMCPCmd` directly constructs and publishes official dependencies, while `Catalog.Call` treats a non-nil direct official client as authorization.
- Current source and AST agree that oversized `toolResult` decodes into `any`; Go's default decoder yields `float64`, so retained integers above `2^53` cannot remain exact. The downstream `trim` assertion only recognizes a `float64` omission count and therefore must be covered when exact-number decoding changes the dynamic type.
- CodeGraphContext names relevant symbols but is not used as authority for branch behavior.
- No mismatch blocks the proposed scope. Any change outside `cmd/tossctl/mcp.go`, `internal/mcp/server.go`, their tests, or this a141 analysis requires Manager approval.
