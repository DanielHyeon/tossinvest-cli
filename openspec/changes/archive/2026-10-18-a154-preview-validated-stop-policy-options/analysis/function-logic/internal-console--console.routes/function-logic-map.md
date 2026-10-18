# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`; AST: `ast.json`; scan: `risk-pattern-report.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Console options and remote mode | configured console | Console construction | session and CSRF wrappers gate protected routes |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remote exists and not trustedNetwork | register login/logout | continue | remote auth tests |
| B2 | remote exists | none | return remote security wrapper else mux | remote and local route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mux.HandleFunc, session0, mutating, register helpers | construct guarded route table | no network or state writes at registration | AST and static route tests |
| remote.security | apply remote security wrapper | selected only when configured | AST B2 and remote tests |

## State mutations and fallbacks

Constructs a local mux; no DB, config, broker, or journal mutation. Unmatched root delegates to redirectRoot.

## Safety conclusion

- Safe edit boundary: add one compute-only route protected by existing session0/mutating and outer no-store; preserve all other registrations.
- High-risk impact: route security boundary; existing policy and trading behavior unchanged.
