# Function Logic Map: `TestStopNeverRetreatsOnScaleIn`
- Source: `internal/reversallane/structural_evaluate_test.go`
## Inputs and invariants
Fresh sealed stop candidates.
## Branches and early returns
Tests retreat, equal boundary and tighter stop.
## Calls and live bindings
Pure evaluator.
## State mutations and fallbacks
Test values only.
## Safety conclusion
Non-retreat preserved.
