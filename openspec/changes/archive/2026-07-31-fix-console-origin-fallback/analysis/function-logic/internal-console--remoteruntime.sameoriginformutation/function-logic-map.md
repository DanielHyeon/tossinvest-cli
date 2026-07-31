# Function Logic Map: `remoteRuntime.sameOriginForMutation`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Origin header key | absent or present | browser request | present delegates to strict predicate |
| Referer header key | absent or present | browser request | present delegates to strict predicate |
| `r.TLS` | non-nil only for direct HTTPS | Go server connection | nil returns false in headerless fallback |
| `r.Host` | exact canonical host and port | request plus outer security middleware | mismatch returns false |
| `rr.origin` | exact configured HTTPS origin | validated remote public URL | mismatch returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | either explicit header key is present | none | returns strict `sameOrigin` result | explicit/empty/multiple/contradictory table |
| final | neither header key is present | none | true only for direct TLS and exact `"https://"+Host` | canonical TLS, non-TLS, wrong Host/port |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.CanonicalHeaderKey` | use the canonical request-map key | pure, no retry | Go AST |
| `remoteRuntime.sameOrigin` | preserve strict explicit-header precedence | false is final; no fallback | Go AST + tests |

## State mutations and fallbacks

- Pure predicate with no I/O, mutation, audit, config, engine, trading, or
  journal side effect.
- Does not read `X-Forwarded-Host`, `X-Forwarded-Proto`, path, query, or
  fragment.
- Direct TLS+Host is available only when both explicit header keys are absent.

## Safety conclusion

- Safe edit boundary: new leaf called only by `Console.mutating`; `/login`
  continues to call strict `sameOrigin`.
- High-risk impact: authentication boundary, covered by fail-closed evidence
  tables and CSRF-order integration tests; no trading-path impact.
