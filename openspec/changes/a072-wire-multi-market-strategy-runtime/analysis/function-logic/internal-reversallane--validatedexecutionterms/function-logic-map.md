# Function Logic Map: `validatedExecutionTerms`
- Source: `internal/reversallane/execution_terms.go`
- Requires sealed terms and sealed fresh stop authority, then strict ordering.
- Safety: no inferred execution values.
## Inputs and invariants
Sealed terms and fresh sealed stop.
## Branches and early returns
Reject authority or order mismatch.
## Calls and live bindings
Pure validation.
## State mutations and fallbacks
None.
## Safety conclusion
Fail closed.
