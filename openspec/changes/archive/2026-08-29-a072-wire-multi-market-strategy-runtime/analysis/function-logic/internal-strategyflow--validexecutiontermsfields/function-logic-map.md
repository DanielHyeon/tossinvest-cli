# Function Logic Map: `validExecutionTermsFields`
- Source: `internal/strategyflow/types.go`
- Validates complete scope, policy identity, three provenance tuples and strict ordering.
- Safety: branchless boolean validation, no normalization or fallback.
## Inputs and invariants
Complete provenance and policy identity.
## Branches and early returns
Boolean conjunction only.
## Calls and live bindings
Pure decimal/provenance checks.
## State mutations and fallbacks
None.
## Safety conclusion
Strict validation.
