# Function Logic Map: `TestScaleInStopNeverRetreatsAndInvalidStopRefuses`
- Source: `internal/continuationlane/allocation_risk_test.go`
## Inputs and invariants
Sealed stop provenance fixtures.
## Branches and early returns
Checks tighter saved stop and invalid candidate.
## Calls and live bindings
Pure evaluator only.
## State mutations and fallbacks
Test values only.
## Safety conclusion
Stop authority stays fail closed.
