# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| console remote mode | trusted network or authenticated remote login | `c.remote` | login routes are exposed only when application auth is required |
| route registry | fixed literal paths and wrapper chains | this function | unmatched paths fall through authenticated root redirect/404 handling |
| history route (planned) | GET/HEAD only | `readOnly` wrapper | other methods return 405 before any metadata read |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remote application login required | register login/logout routes | continue building mux | remote route tests |
| B2 | release download/staging capability present | register release fetch route | otherwise route absent | system update static/runtime tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | authenticate protected screens | refuses missing/invalid session | route/session tests |
| `readOnly` | enforce GET/HEAD before read handlers | returns 405 with Allow header | orders/positions/history method tests |
| `mutating` | enforce POST+CSRF for write handlers | refusal performs no mutation | static and CSRF tests |

## State mutations and fallbacks

- This function only constructs an in-memory mux.
- The planned history wrapper changes accepted methods, not route data or authority.

## Safety conclusion

- Safe edit boundary: wrap the existing `/history` handler with the already-proven `readOnly` middleware.
- High-risk impact: no account mutation; safety increases because authenticated POST cannot spend read budget.
