# Function Logic Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`
- Source: `internal/console/orders_static_test.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Every console option seam and route needs an explicit read/mutation classification and rationale.
## Branches and early returns
- B1-B17 enumerate fields/routes and fail for duplicate, missing, unclassified, or unexplained entries.
## Calls and live bindings
- Uses Go AST/source inspection and test assertions only.
## State mutations and fallbacks
- Test-local maps only; no runtime state.
## Safety conclusion
- Makes the new `Protections` capability and both routes visible to static governance.
