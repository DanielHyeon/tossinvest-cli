## ADDED Requirements

### Requirement: Parker RR diagnostics distinguish space target and realized results
Diagnostics SHALL identify each RR numerator, denominator, units, price basis and policy version without altering frozen v1 decisions or interpreting legacy RR labels as a stronger predicate.

#### Scenario: Legacy floor fixture
- **WHEN** entry=100, stop=99.3, forward space=1.2 and target=102.1 are supplied
- **THEN** space RR is 12/7, target RR is 3, realized RR is unknown without fills/costs, and the frozen v1 decision remains identical

### Requirement: Parker RR diagnostics reject undefined risk
RR diagnostics SHALL return a typed invalid or unknown state for nonpositive risk or unavailable realized evidence.

#### Scenario: Invalid denominator
- **WHEN** entry equals stop
- **THEN** no finite RR or approved risk decision is fabricated
