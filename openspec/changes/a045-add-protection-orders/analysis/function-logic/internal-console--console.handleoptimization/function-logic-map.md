# Function Logic Map: `Console.handleOptimization`
- Source: `internal/console/optimization.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Only GET/HEAD are accepted; all protection data comes from the server-owned commander and never from free-form browser input.
## Branches and early returns
- B1 rejects other methods; B2-B3 load exit policy safely; B4 enumerates fixed policies; B5-B7 expose protection status or a fail-closed load error.
## Calls and live bindings
- Calls configured read seams, `RegisteredCommonPolicies`, and the template renderer. Errors are displayed without activating anything.
## State mutations and fallbacks
- Builds a response view only. Nil protection wiring renders explicit OFF/UNWIRED state.
## Safety conclusion
- Read-only UI boundary; no order, activation, or toggle mutation is reachable.
