# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.remote` | nil, trusted, or authenticated remote runtime | `Console.New` | only remote login/logout route set changes |
| `c.opts.ReleaseDownloader` / `ReleaseCandidateStager` | nil or both wired | `Console.Options` | download route is absent unless both are wired |
| handler registrations | unique exact paths | `http.ServeMux` | duplicate registration panics during construction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | untrusted remote access configured | registers login/logout in addition to session-gated routes | none | existing remote route/session tests |
| B2 | release download and staging both wired | registers the mutating download route | none | existing signed release/static route tests |
| B3 | unconditional route registrations | builds only HTTP handler graph; performs no account/config mutation | returns mux | route capability/static tests, a048 read-only route test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewServeMux` / `HandleFunc` | build fixed route graph | duplicate pattern panics at construction | AST + Go `net/http` contract |
| `c.session0` | require local/remote authenticated session | refuses or redirects before handler | CodeGraph context + static tests |
| `c.mutating` | require POST/origin/CSRF on action routes | fail-closed response; no handler call | CodeGraph context + CSRF tests |
| `registerOverview/registerOrders/registerSignals` | register modular read screens | each owns a fixed route | CodeGraph related symbols |

## State mutations and fallbacks

- Mutates only a fresh `ServeMux` during `Console.New`; it does not change
  account, config, journal or process state.
- a048 adds one unconditional session-gated GET registration. It must not be
  wrapped in `mutating` and must reject non-GET/HEAD in its handler.

## Safety conclusion

- Safe edit boundary: add only the exact `/strategy-runtime/market-schedule`
  session-gated registration beside the existing optimization route.
- High-risk impact: no direct trading impact; authentication/route capability
  tests remain mandatory.
