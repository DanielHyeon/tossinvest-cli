# Function Logic Map: `ExecutionTermsPreimage.valid`
- Source: `internal/reversallane/execution_terms.go`
- Validates private plan/evidence-bound seal with no fallback.
- Safety: external callers cannot mint accepted authority.
## Inputs and invariants
Private sealed plan/evidence preimage.
## Branches and early returns
Straight boolean validation.
## Calls and live bindings
Pure digest checks.
## State mutations and fallbacks
None.
## Safety conclusion
Fail closed.
