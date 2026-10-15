## Why

MCP must honor the same backend authorization policy as CLI reads. Financial tool output must not round integer identifiers or decimal values when shortening large responses. This change implements the first correctness priority authorized on 2026-10-15; behavioral implementation claims require fresh AST and CodeGraph evidence before coding.

## What Changes

- Resolve backend policy before exposing direct official-read or official-trading dependencies to MCP.
- Preserve exact JSON numeric tokens during large-result truncation.
- Add offline routing and precision regression tests, independent adversarial review, and a further gstack review.
- Preserve existing journal, risk controls, shared token lifecycle, engine official-only policy, operating settings and live state.

## Capabilities

### New Capabilities

- `mcp-backend-and-number-integrity`: MCP respects resolved backend capability and maintains exact numbers in truncated JSON.

### Modified Capabilities

None.

## Impact

Expected code boundary: `cmd/tossctl/mcp.go`, `internal/hybrid`, `internal/mcp` and directly associated tests. No development-tool MCP configuration changes. No deployment, credential/token refresh against real services, live orders, toggles, restarts or operating configuration changes. Backend authorization is high risk and requires branch/logic maps and negative tests. Story: STORY-TOS-a141, feature: FEAT-TOS-014. Implementation: Terra; Manager: specification and review only.
