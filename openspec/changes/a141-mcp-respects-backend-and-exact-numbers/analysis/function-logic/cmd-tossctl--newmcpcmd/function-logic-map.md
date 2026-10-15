# Function Logic Map: `newMCPCmd`

- Source: `cmd/tossctl/mcp.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cfg.OpenAPI` and `opts.backend` | Existing config/flag contract | `resolveBackend` → `internal/app.ResolveBackend` | Invalid backend returns before server construction |
| Official credentials | present or absent | `official.LoadCredentials` | Load error returns; absent credentials may still permit authenticated WTS only |
| WTS session | present, absent, or unreadable | `session.FileStore.Load` | Unexpected load error returns; absent session is ordinary |
| Official capability | credentials AND enabled AND effective backend != `wts` | CLI-equivalent predicate in `internal/app.New` | Official client, trading service, auth status, and direct catalog capability must remain absent |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `loadConfig` fails | none | return config error | Existing config error coverage |
| B2 | OpenAPI path resolution fails | none | return path error | Existing resolver coverage |
| B3 | credential loading fails | none | return credential error | Existing credential coverage |
| B4 | eligibility permits credentials, enabled OpenAPI, and non-WTS preference | constructs direct official read/write capability | no error | guarded matrix: disabled/WTS deny; enabled auto/openapi retain |
| B5 | session loading fails other than `ErrNoSession` | none | return session error | Existing session coverage |
| B6 | session exists | builds WTS client | continue | WTS operation remains available |
| B7 | both effective capabilities absent | none | login guidance error | Offline no-credentials regression |
| B8 | no WTS session | builds sessionless router only | continue; catalog auth gate must refuse WTS operations | Existing WTS missing-session regression |
| B9 | backend resolution fails before credential paths and construction | none | return before any client construction | Invalid override regression |
| B10 | update-cache path resolves | bounded update probe | probe failure is silent | Existing behavior preserved |
| B11 | update exists | appends instruction text | continue | Existing behavior preserved |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `resolveBackend` | Obtain effective policy before capability construction | invalid value returns | AST `B9`, delegated resolver |
| `official.LoadCredentials` | Load capability material without exposing it | read failure returns | AST `B3` |
| `official.New` | Construct official direct client only under eligibility predicate | no network at construction | AST `B4`; `internal/app.New` parity |
| `trading.NewService` | Bind official-only mutation broker only with official client | config/preview/confirm gate retained | AST `B4`; MCP server contract |
| `hybrid.New` | Route allowed read capability | WTS-only when official capability absent | AST `B8-B9` |
| `mcp.NewServer` / `SetAuthStatus` | Publish only allowed dependencies and status | no secret values | AST calls 104/108 |

## State mutations and fallbacks

- Constructs clients and a local lineage service; starts the stdio server only at `Serve`.
- No order is placed by construction. The changed boundary controls whether the server receives any direct official-read or official-trading capability.

## Safety conclusion

- Safe edit boundary: construction ordering and eligibility predicate in `cmd/tossctl/mcp.go`; do not change engine wiring, catalog registry, WTS client, or confirmation flow.
- High-risk impact: yes. A direct official dependency can make financial reads or official mutation tooling reachable despite disabled/WTS policy.
