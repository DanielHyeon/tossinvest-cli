# Function Logic Map: `remoteRuntime.security`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `next` | non-nil console router handler | `Console.routes` | a nil handler would panic only after a request passes the wrapper; construction supplies a handler |
| `r.RemoteAddr` | parseable direct peer address | `remoteRuntime.peer` | B1 returns HTTP 403 |
| `r.URL.Path` | any console path; `/healthz` is special | `http.Request` | only loopback health bypasses allowed CIDR/Host checks |
| `r.Host` | configured public URL host for non-health traffic | `RemoteAccess.PublicURL` | B3 returns HTTP 403 |
| response security headers | fixed on every remote response | `remoteRuntime.security` | missing/wrong policy can break browser origin evidence or weaken browser isolation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `rr.peer(r)` cannot parse a direct peer | response headers already set; writes error body | HTTP 403, early return | malformed/outside peer remote tests |
| B2 | path is `/healthz` and peer is loopback | invokes `next.ServeHTTP` | returns after minimal health handler | `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal` |
| B3 | peer is outside allowed CIDRs OR Host differs from public URL | writes error body | HTTP 403, early return | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` and Host rejection tests |
| final | peer and Host are accepted | invokes `next.ServeHTTP` | downstream route result | trusted dashboard and console route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `w.Header().Set` | apply HSTS, CSP, frame, MIME, referrer, and cache policy before every branch | deterministic in-memory header mutation; no retry | AST calls at lines 278-283 |
| `rr.peer` | derive the actual direct peer without trusting forwarding headers | boolean failure, no retry | CodeGraph definition + B1 |
| `peer.IsLoopback` | restrict minimal health bypass | pure address check | AST + B2 |
| `rr.peerAllowed` | enforce configured CIDR allowlist | pure prefix membership check | CodeGraph + B3 |
| `strings.EqualFold` | enforce exact configured public Host | pure comparison | AST + B3 |
| `next.ServeHTTP` | continue only through health exception or accepted peer/Host | downstream owns its own method/origin/CSRF/audit gates | CodeGraph caller/callee trail |

## State mutations and fallbacks

- Mutates only the current response header map and response body/status on
  refusal paths.
- Does not mutate remote runtime, sessions, rate limits, settings, engine,
  orders, journal, or operating toggles.
- No fallback is added. Existing peer, health, CIDR, and Host branches remain in
  their current order.

## Safety conclusion

- Safe edit boundary in this function: replace only the `Referrer-Policy`
  literal from `no-referrer` to `same-origin`; `Console.render` and the two
  template meta declarations are separately mapped and tested policy surfaces.
- High-risk impact: yes, because this wrapper is part of the remote console
  security boundary. Mitigation is fail-closed origin logic left unchanged,
  exact header regression coverage, safe Chrome reproduction, race/full tests,
  and an independent security review.
