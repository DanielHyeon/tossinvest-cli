# Function Logic Map: `sealExecutionTerms`
- Source: `internal/strategyflow/types.go`
- Package-private sealing binds complete lineage to exact lane provenance and policy.
- Safety: invalid inputs return zero opaque terms.
## Inputs and invariants
Complete lineage and lane authority.
## Branches and early returns
Reject incomplete preimage.
## Calls and live bindings
Pure validation/hash.
## State mutations and fallbacks
Value construction only.
## Safety conclusion
Package-private seal.
