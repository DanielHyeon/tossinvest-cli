## ADDED Requirements

### Requirement: Validated initial-stop choices reuse settings lifecycle
The registry SHALL expose strategy-and-market-scoped finite initial-stop choices separately from existing common exit policies, with evidence and approval state, new-campaign-only timing and unchanged legacy defaults.

#### Scenario: Unvalidated policy
- **WHEN** a candidate lacks accepted holdout evidence or operator approval
- **THEN** the UI displays unavailable with reasons and cannot apply the candidate

### Requirement: Preview makes risk consequences visible without activation
Preview SHALL show planned stop distance, quantity, risk basis, costs, provenance and policy versions and SHALL NOT mutate trading settings or existing positions.

#### Scenario: Existing position during preview
- **WHEN** an operator previews a wider candidate stop
- **THEN** existing protection and operating toggles remain unchanged

### Requirement: Approved selection cannot widen existing positions
Any later approved CAS apply SHALL preserve existing position snapshots and lane/automation states and SHALL reject stale base versions.

#### Scenario: Stale version
- **WHEN** a selection was previewed against an old configuration version
- **THEN** apply is refused without partial changes
