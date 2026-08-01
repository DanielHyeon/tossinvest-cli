# Function Logic Map: `remoteRuntime.sameOrigin`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rr.origin` | validated configured `https://host:port` | `newRemoteRuntime` | mismatch returns false |
| `Origin` | absent or exactly one canonical value | browser | explicit invalid/mismatch is final false |
| `Referer` | absent or exactly one absolute URL | browser | explicit invalid/mismatch returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | persisted configured origin cannot be parsed (defensive; startup validation normally prevents it) | none | false | invalid configuration fails closed |
| final | configured origin parsed | none | returns `present && matches` from the shared strict explicit predicate | precedence/opaque/repeated/path/port table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `networkboundary.ExplicitOriginMatches` (planned extraction) | centralize canonical scheme/host/effective-port and strict precedence | pure; parse failure false; no retry | current AST + networkboundary RED/GREEN tests |
| `Console.mutating` / `remoteRuntime.loginPost` | authentication boundary callers | false becomes 403 before mutation/audit | CodeGraph callers/impact |

## State mutations and fallbacks

- Pure predicate: no mutation, I/O, audit, session, engine, journal, or broker side effect.
- `/login` remains explicit-header-only; it must not gain direct TLS fallback.
- The extraction will not configure a proxy for the existing console. Forwarded evidence remains refused there.

## Safety conclusion

- Safe edit boundary: local URL parsing/equality was replaced only with the shared pure explicit-origin predicate; callers and gate order are unchanged.
- High-risk impact: yes (authentication/session boundary), no trading-path or LIVE capability impact.
