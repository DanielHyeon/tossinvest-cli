# Function Logic Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| commands | non-nil narrow service including runtime GET | engine assembly | refuse startup |
| engineDir | real owner-safe directory | private FS validator | refuse startup and clean created leaf |
| listener/token | loopback plus 256-bit random bearer | OS/rand | close/cleanup on failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil service or empty directory | none | error | existing startup tests |
| B2 | unsafe/create/validate control dir | may create then remove private dir | error | symlink/mode tests |
| B3 | listen or RNG failure | closes listener/dir | error | injected/contract tests |
| B4 | health/positions/runtime GET wrong method | none | 405 typed error | transport method table |
| B5 | authenticated runtime GET succeeds/fails | journal read only | JSON or typed error | runtime RPC tests |
| B6 | descriptor publish fails | closes listener and cleanup | error | descriptor tests |
| B7 | success | publishes 0600 descriptor, serves loopback | server | end-to-end client test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| private FS validators | descriptor capability safety | fail closed | CodeGraph + tests |
| `commands.Runtime` | startup effective and active states | propagate typed server error | new runtime GET contract |
| HTTP server | bounded local transport | existing time/header limits | AST |

## State mutations and fallbacks

- Filesystem mutation is limited to private endpoint descriptor/control directory lifecycle.
- Runtime endpoint is GET-only and carries no mutation capability/token.

## Safety conclusion

- Safe edit boundary: add one authenticated GET handler beside existing reads.
- High-risk impact: yes; control-plane method and bearer checks are security boundaries.
