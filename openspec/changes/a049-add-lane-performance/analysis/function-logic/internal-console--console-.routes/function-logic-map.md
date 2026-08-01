# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.remote` | nil, trusted-network remote, or token-authenticated remote | validated `Console` construction | nil returns the session-gated local mux; non-nil wraps the mux in remote security |
| registered route graph | literal unique paths | this function plus the three typed `register*` helpers | Go `ServeMux` panics on duplicate patterns during construction; static route tests pin the complete table |
| `/performance-history` | session-gated display route | `handlePerformanceHistory` | handler accepts GET/HEAD only and returns 405 without reaching its reader for other methods |
| state-changing routes | exact allowlisted paths behind `session0(mutating(...))` | `static_test.go` route analysis | tests fail if a mutating route loses session/CSRF gating or a read route gains mutation vocabulary |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.remote != nil && !c.remote.trustedNetwork` | registers token login/logout routes; logout remains session+mutation gated | continues building the same mux | `TestRemoteLoginIssuesADistinctBoundSecureSession`; `TestTrustedNetworkConsoleNeedsNoApplicationSession` |
| B2 | `c.remote != nil` after every route is registered | wraps the completed mux in peer/origin/session/CSP remote security | early return of secured handler | `TestRemotePeerHostOriginAndCSRFAreIndependentGates`; `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal` |
| fallthrough | `c.remote == nil` | no additional side effect | returns local mux | `TestEveryRouteRefusesARequestWithoutTheSessionToken`; `TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewServeMux` / `mux.HandleFunc` | builds a closed literal route table | duplicate patterns panic immediately; no network or retry | AST call list; `registeredRoutes` static scan |
| `c.session0` | applies the application session boundary to all non-health/token-login routes | rejects missing/invalid session before leaf handler | `TestEveryRouteRefusesARequestWithoutTheSessionToken` |
| `c.mutating` / `c.startExclusive` / `c.credentialHTTPS` | preserves existing CSRF, exclusivity and credential transport gates for acting routes | fail closed in wrappers; no retry here | console static/security suites |
| `c.handlePerformanceHistory` | serves the new fixed-query performance page | reader error is rendered as an explicit unchanged/read failure; POST is 405 | `performance_history_test.go` |
| `c.registerOverview`, `c.registerOrders`, `c.registerSignals` | composes the established read surfaces | their own registration files are included by static AST route scan | `registeredRoutes` and dashboard route tests |
| `c.remote.security` | adds trusted peer/origin/session headers to the whole mux | fail closed on invalid remote request | remote console tests |

## State mutations and fallbacks

- The function mutates only a newly allocated in-memory mux during construction; it does not write config, journal, performance data, broker state, lane state, or LIVE approval state.
- `/performance-history` is deliberately bound through `session0` only. Its leaf handler owns GET/HEAD enforcement and is tested to reject POST before the injected read interface is called.
- A missing performance reader is an explicit `seam 미배선` view, not a synthetic zero and not a fallback to broker polling.
- Existing operating routes and wrapper order are byte-for-byte unchanged apart from inserting the new read route beside `/optimization`.

## Safety conclusion

- Safe edit boundary: one authenticated, read-only route registration and no new mutating wrapper or capability.
- High-risk impact: the host route table is security-sensitive, so the full static route guard and both remote/local branches remain required even though the added leaf has no trading authority.
