## ADDED Requirements

### Requirement: Offline regression composes fill protection and restart state
Regression verification SHALL exercise partial fill application, explicit scoped protection entry blocking, journal reopen and duplicate replay through existing implementation paths on the same persisted campaign lineage. It SHALL distinguish synthetic fixture verification from broker evidence and SHALL NOT replace live rollout approvals.

#### Scenario: Partial fill and protection survive replay
- **WHEN** an admitted synthetic order for eight shares has cumulative fill three, protection blocks new entry, and the journal is reopened before identical fill and protection requests replay
- **THEN** position and cumulative fill remain three, unsettled residual remains five until authoritative terminal evidence, protection and generation remain bound, and duplicate replay adds no duplicate exposure or protection event

#### Scenario: New entry refusal has a positive control
- **WHEN** a valid new next-leg command uses the latest version and correct sequence in the blocked campaign
- **THEN** it fails with the protection exposure-block reason without new exposure, while an independent valid unblocked campaign can admit its own next leg

### Requirement: Offline exit regression distinguishes evaluation from submission
The offline regression SHALL verify protective evaluation under entry blocking without treating every evaluation as permission for a new order. It SHALL preserve unsettled-entry and pending-exit restrictions using truthful fake transport state.

#### Scenario: Unsettled entry withholds a duplicate liquidation
- **WHEN** protective evaluation breaches the stop while the partially filled entry remains working and its cancellation explicitly fails or remains uncertain in the fake transport
- **THEN** evaluation continues, a new sell is withheld, and the existing delayed-alert contract is verified without hiding the working buy

#### Scenario: Terminal entry permits bounded protection
- **WHEN** explicit terminal entry evidence has been applied and a fresh observation satisfies existing protection guards
- **THEN** any fake exit is capped by confirmed sellable quantity, replay does not duplicate it, and an already pending sell is preserved without a second submission

### Requirement: Offline protection evidence includes meaningful negative controls
The regression SHALL include valid unblocked and non-breach controls, persisted-state assertions and behavioral mutation evidence. Tests SHALL use isolated temporary state and no live transport, operational change or historical evidence mutation.

#### Scenario: Safety invariant is disabled in a disposable mutation
- **WHEN** an entry-block guard or independent replay/exit-idempotency invariant is disabled in a disposable source variant
- **THEN** the corresponding regression fails behaviorally and the unchanged source passes; compilation errors alone do not count as detection
