# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| registered routes | statically extracted fixed paths/wrappers | package Go AST | test failure |
| state-changing allowlist | exact complete path set | operator-console spec | missing/extra gate failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | listed route lacks CSRF | none | test error | self |
| B2 | unlisted route has CSRF | none | test error | self |
| B3 | listed route absent | none | test error | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | parse every registration | fails on opaque registration | CodeGraph + AST |

## State mutations and fallbacks

- Add the exact signed-download POST path to both static state-change lists.

## Safety conclusion

- Safe edit boundary: enumeration only.
- High-risk impact: yes; the guard proves session/CSRF coverage.
