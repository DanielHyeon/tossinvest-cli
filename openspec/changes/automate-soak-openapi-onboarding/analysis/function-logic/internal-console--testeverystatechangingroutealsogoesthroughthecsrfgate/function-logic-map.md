# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The literal state-changing route set and AST-extracted route table must match.
Every listed route is CSRF-gated and every other route is not.

## Branches and early returns

| Branch | Meaning | Evidence |
|---|---|---|
| B1-B4 | iterate routes and reject missing/excess gates | static test |
| B5-B6 | iterate required paths and reject missing routes | static test |

## Calls and live bindings

`registeredRoutes` reads the current package AST. Failures are test-only.

## State mutations and fallbacks

None.

## Safety conclusion

- Safe edit boundary: enumerate the new write-only credential save route.
- High-risk impact: the guard strengthens authentication coverage.
