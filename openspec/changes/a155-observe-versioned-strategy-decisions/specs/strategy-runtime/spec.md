## ADDED Requirements

### Requirement: Observation has no execution authority
The observation component SHALL emit append-only diagnostic records using versioned evidence without access to broker mutations, reservation writes, position changes or order-capable coordinator submission.

#### Scenario: All candidate conditions pass
- **WHEN** an observation fixture meets every proposed entry condition
- **THEN** a diagnostic record is produced and orders, reservations and trading state mutations remain zero

### Requirement: Observation failure cannot delay protection
Observation SHALL use nonblocking producer enqueue, a maximum 256-record queue, maximum 64KiB encoded records and drop-newest counters, with no observation disk I/O or retry on trading/exit threads. Append durability SHALL precede checkpoint advancement and replay SHALL deduplicate stable record IDs without truncating evidence. It SHALL report failed records while preserving independent protective exits.

#### Scenario: Storage failure
- **WHEN** the observation destination is full or unavailable
- **THEN** failure is visible and existing exit handling continues without waiting for observation recovery

### Requirement: Actual observation requires a separate registered session approval
Replay validation SHALL NOT authorize actual capture; actual sessions SHALL require an explicitly approved bounded manifest and SHALL preserve existing evidence paths and schedules.

#### Scenario: No session approval
- **WHEN** offline replay and code review pass
- **THEN** no collector, schedule, restart or API measurement is started
