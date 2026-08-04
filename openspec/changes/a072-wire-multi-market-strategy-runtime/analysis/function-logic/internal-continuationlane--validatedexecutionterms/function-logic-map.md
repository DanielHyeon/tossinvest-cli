# Function Logic Map: `validatedExecutionTerms`
- Source: `internal/continuationlane/execution_terms.go`
- Requires exact sealed price provenance and stop < entry < target.
- Safety: no inferred price or target; returns false on any mismatch.
## Inputs and invariants
Sealed terms and exact stop provenance.
## Branches and early returns
Reject seal/provenance/order mismatch.
## Calls and live bindings
Pure validators only.
## State mutations and fallbacks
None.
## Safety conclusion
Fail closed.
