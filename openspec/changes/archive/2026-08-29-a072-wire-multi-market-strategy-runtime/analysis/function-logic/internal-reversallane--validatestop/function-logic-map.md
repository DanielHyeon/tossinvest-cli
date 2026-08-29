# Function Logic Map: `validateStop`
- Source: `internal/reversallane/evaluate.go`
- Validates private seal, plan/evidence scope, freshness, decimal, and non-retreat.
- Safety: stale or caller-forged provenance is typed-refused.
## Inputs and invariants
Plan/evidence-bound sealed candidate.
## Branches and early returns
Reject seal, decimal, or retreat.
## Calls and live bindings
Pure stop validation.
## State mutations and fallbacks
None.
## Safety conclusion
Fail closed.
