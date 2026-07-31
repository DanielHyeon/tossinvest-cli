# Function Logic Map: `remoteRuntime.sameOrigin`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rr.origin` | exact configured `https://host:port` | `newRemoteRuntime` / validated public URL | mismatch returns false |
| `Origin` header key/values | absent or exactly one non-empty origin | browser request | present invalid/mismatch returns false; never falls through |
| `Referer` header key/values | absent or exactly one parseable absolute URL | browser request | present invalid/mismatch returns false |

`r.TLS` and `r.Host` are deliberately not inputs to this strict predicate.
They are read only by the new mutation-only `sameOriginForMutation` leaf.

## Branches and early returns

Post-GREEN captured AST:

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | canonical Origin key is present | none | enters final Origin evaluation; Referer is not read | explicit canonical/cross-origin and contradiction table |
| B2 | present Origin has a value count other than one | none | false | empty slice and multiple Origin |
| Origin return | exactly one Origin value | none | non-empty trimmed value equals `rr.origin` | canonical/wrong/empty/space Origin |
| B3 | no Origin; Referer key absent or value count is not one | none | false | absent/multiple Referer and headerless login |
| B4 | single trimmed Referer value is empty | none | false | empty/space Referer |
| B5 | Referer parse errors, scheme is empty, or host is empty | none | false | malformed/relative Referer |
| final | Referer is an absolute URL | none | scheme+host equals `rr.origin` | same-origin paths and mismatch |

## New mutation-only leaf

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| SM-B1 | mutation wrapper sees either header key | none | delegates to strict predicate | contradictory evidence and invalid explicit values |
| SM-final | mutation wrapper sees neither key | none | direct TLS and exact `"https://"+Host` result | canonical/non-TLS/wrong Host/port |

`sameOriginForMutation` is a new pure leaf in this change, so it has no
pre-change AST. Its behavior is bound by the source, the integration tests, and
the unchanged `Console.mutating` gate order.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct canonical header map lookup | distinguish absence from explicit invalid values | no I/O, no retry | current Go AST |
| `strings.TrimSpace` | reject empty explicit evidence | no I/O, no retry | current Go AST |
| `url.Parse` | remove Referer path/query/fragment from origin identity | parse failure returns false | Go AST |
| `Console.mutating` caller | run before form/CSRF/handler | false becomes 403 with no mutation | CodeGraph callers + HEAD |
| `remoteRuntime.loginPost` caller | protect compatibility login POST | false becomes 403 before credentials/audit | CodeGraph callers + HEAD |

## State mutations and fallbacks

- Pure predicate: no mutation, I/O, audit, config, process, or account side effect.
- Strict predicate never uses TLS+Host fallback.
- New mutation-only leaf gives explicit header keys final precedence and lets only
  complete header-key absence reach direct TLS+Host.
- No forwarding header is read.
- Fallback is direct-TLS only; TLS-terminating proxies remain unsupported.

## Safety conclusion

- Implemented edit boundary: strengthened strict `sameOrigin`, added one pure
  mutation-only leaf, and changed only the predicate call in
  `Console.mutating`.
- High-risk impact: yes, authentication/session boundary. No trading-path impact.
