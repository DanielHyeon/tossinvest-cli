# Function Logic Map: `executionTermsIdentity`
- Source: `internal/strategyflow/types.go`
- Hashes every opaque scope, provenance, policy and lineage field with length framing.
- Safety: package-private only; callers cannot construct or reseal fields.
## Inputs and invariants
Opaque terms fields.
## Branches and early returns
Iterates exactly three price terms.
## Calls and live bindings
Length-framed SHA-256.
## State mutations and fallbacks
Local hash only.
## Safety conclusion
All fields bound.
