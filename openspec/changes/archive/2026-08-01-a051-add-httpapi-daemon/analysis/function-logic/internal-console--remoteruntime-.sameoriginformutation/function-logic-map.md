# Function Logic Map: `remoteRuntime.sameOriginForMutation`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Origin/Referer header keys | absent or present | browser request | any presence delegates to strict predicate; false is final |
| `r.TLS` and `r.Host` | direct TLS and canonical host/port | Go HTTP server | plaintext or mismatch returns false |
| `rr.origin` | validated configured HTTPS origin | `newRemoteRuntime` | mismatch returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | either Origin or Referer key present | none | return strict `sameOrigin` result | explicit evidence table |
| final | both keys absent | none | parse configured origin, then direct TLS canonical transport only; parse/transport failure is false | TLS/host/forwarded tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `remoteRuntime.sameOrigin` | preserve strict explicit-header precedence | false final; no fallback | AST + console regression tests |
| `networkboundary.OriginMatches` (planned extraction) | canonical direct transport fallback and forwarding-header refusal | pure; parse/transport failure false | networkboundary tests |

## State mutations and fallbacks

- Pure predicate with no mutation, audit, session, config, engine, journal, or broker side effect.
- Existing console has no trusted-proxy configuration, so forwarding headers remain refused.
- Direct TLS+Host fallback remains available only when both explicit keys are absent.

## Safety conclusion

- Safe edit boundary: retained B1 and delegated only canonical comparisons/fallback to shared networkboundary leaves.
- High-risk impact: yes (authentication/session boundary); no LIVE/trading path impact.
