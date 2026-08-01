# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`
- Source: `internal/console/static_test.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- AST-derived route registrations are classified by verb and all state-changing paths must pass the common CSRF gate.
## Branches and early returns
- B1-B6 enumerate declarations, classify methods, and fail on any missing or bypassed protection route.
## Calls and live bindings
- Static parser and test assertions inspect source without executing mutations.
## State mutations and fallbacks
- Test-only bookkeeping; failures stop the gate.
## Safety conclusion
- Prevents protection preview/apply from escaping POST+CSRF policy during later edits.
