# CodeGraph baseline

- Base commit: `6765858af1beb472268cb699efa1964addb84d60`
- Capture: `base-commit.txt` created by `capture_change_base.py` before production edits.
- Index: `make sdd-sync` initialized CodeGraph in this worktree (2,798 files; 49,477 nodes; 183,327 edges).

## Queries and conclusions

| Query | Evidence | Current conclusion |
|---|---|---|
| `newMCPCmd` callers/callees/impact | Caller: `newRootCmd`; callees include config load, credential load, `official.New`, `trading.NewService`, `hybrid.New`, `mcp.NewServer`, and `SetAuthStatus` | Capability must be gated before all four direct publication/construction paths. |
| `toolResult` callers/callees/impact | Called by `handleToolsCall`; calls `shrinkToCap`; impact remains inside MCP server/test surface | Exact-number handling is local to truncation serialization. |
| `Client.Official` | Direct accessor; router decision lives in `route` | Do not edit accessor unless a post-change bypass remains. |
| `Catalog.Call` impact | MCP `handleToolsCall` and CLI ops caller; default backend accepts any non-nil `Deps.Client` | MCP construction must supply nil when policy forbids official capability. |

Direct source cross-check: `internal/app.New` constructs official only when credentials exist, OpenAPI is enabled, and resolved preference is not `wts`. `cmd/tossctl/mcp.go` currently does not apply that condition before `official.New`.

