# Function Logic Map: `ExecutionTermsPreimage.valid`
- Source: `internal/continuationlane/execution_terms.go`
- Validates private plan/evidence-bound seal; no fallback, mutation, or authority call.
- Safety: malformed or cross-evidence values fail closed.
## Inputs and invariants
Private sealed plan/evidence preimage only.
## Branches and early returns
Straight boolean validation.
## Calls and live bindings
Pure digest checks only.
## State mutations and fallbacks
None.
## Safety conclusion
Fail closed.
