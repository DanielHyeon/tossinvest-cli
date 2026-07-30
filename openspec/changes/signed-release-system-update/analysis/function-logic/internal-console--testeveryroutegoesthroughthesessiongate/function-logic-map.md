# Function Logic Map: `TestEveryRouteGoesThroughTheSessionGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| extracted routes | complete route table | AST extractor | low count or ungated path fails |
| route floor | at least 23 after update route | `Console.routes` | stale extractor/list detected |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | iterate every route and assert session wrapper | none | test failure |
| B3 | route count below floor | none | test failure |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | parse actual registrations | static only | AST evidence |
| `len` | enforce extractor floor | deterministic | AST |

## State mutations and fallbacks

- The new update route cannot bypass the loopback session credential.

## Safety conclusion

- Safe edit boundary: raise floor to include the added route.
- High-risk impact: yes — unauthenticated install would replace executable code.
