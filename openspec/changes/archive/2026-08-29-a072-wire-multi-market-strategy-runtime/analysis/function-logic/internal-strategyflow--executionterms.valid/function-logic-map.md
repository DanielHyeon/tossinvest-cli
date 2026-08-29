# Function Logic Map: `ExecutionTerms.Valid`
- Source: `internal/strategyflow/types.go`
- Recomputes the private seal over opaque lineage, quantity, provenance and policy values.
- Safety: public getters return copies; no public reseal path exists.
## Inputs and invariants
Opaque complete terms.
## Branches and early returns
Reject zero or malformed seal.
## Calls and live bindings
Pure validation and hash.
## State mutations and fallbacks
Copy-only identity clearing.
## Safety conclusion
Tamper evident.
