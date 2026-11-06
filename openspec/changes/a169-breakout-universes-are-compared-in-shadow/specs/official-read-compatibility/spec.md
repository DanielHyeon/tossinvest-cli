## ADDED Requirements

### Requirement: dormant shadow session storage preserves same-day interruption state

A runtime-unconnected shadow-session component SHALL persist `active` before granting a session handle and retain a process-exclusive owner lock until the handle is retired. It SHALL use one trusted root identity and validated civil day, validate all stored frames and transitions, and refuse ambiguous, corrupt, inaccessible or mismatched storage. It SHALL NOT automatically delete or repair records. File and directory synchronization SHALL precede a successful Begin result. The implementation boundary and supported storage SHALL follow `analysis/daily-session-plan.md`.

The permitted transitions SHALL be absent to active, paused to active, and active to paused, aborted or completed. A new Begin SHALL refuse existing active, aborted or completed. Only paused permits same-day collection resumption. Pause and Complete SHALL require the caller to stop new transmissions, drain prior work and retire that handle permanently. The storage component alone SHALL NOT claim to enforce HTTP draining, current trading-day identity or operational engine identity.

#### Scenario: a process exits while a session is active
- **WHEN** a process terminates with a durable active record and a new process begins the same root and day
- **THEN** acquisition SHALL fail even though the previous process lock has been released

#### Scenario: a quiescent session pauses and resumes normally
- **WHEN** the caller drains and retires its handle and a valid paused record is persisted
- **THEN** a later Begin MAY resume by persisting a new active record before granting its handle
- **AND** previously selected universes SHALL remain the responsibility of the existing caller-side restoration contract

#### Scenario: an ending operation returns an error after its record is committed
- **WHEN** Pause has recorded a valid paused state but a subsequent synchronization or cleanup step returns an error
- **THEN** the error SHALL NOT be reported as proof that the state was never recorded
- **AND** a later Begin SHALL decide from the validated durable record under the declared storage-preservation and caller-drain assumptions

#### Scenario: concurrent processes contend for the same day
- **WHEN** multiple processes begin against one trusted root and day
- **THEN** at most one active handle SHALL be granted at a time and competing acquisition SHALL fail without waiting

#### Scenario: an explicit abort or completion is followed by another collection attempt
- **WHEN** a valid aborted or completed record exists
- **THEN** a new Begin for that day SHALL fail without truncating or resetting the record

#### Scenario: the dormant storage component is published
- **WHEN** this internal component passes its scoped verification
- **THEN** no production measurement caller SHALL be connected by this lot and no live measurement or activation authority SHALL be inferred

### Requirement: an optional measurement observation fence starts from a stable current cursor

The shared observed budget SHALL support an optional immutable measurement fence bound to one verified Budget owner.
A successful baseline SHALL use a stable snapshot with no pending observation and a valid current state. Events and gaps
at or before that baseline, including historical ring overrun from an initial zero cursor, SHALL be excluded from new-session
incident detection without deleting or repairing them. Current pending, dirty state, identity failure and read failure SHALL
reject creation. A fence SHALL NOT silently reset its cursor after creation.

Each explicitly fenced actual wire, including token and 401 retries, SHALL check the same owner and shared observation state
after acquiring the wire lease but before its own Begin. New event/gap counters, counter regression, pending observation or
state failure SHALL reject that wire. The lease SHALL remain held through the existing transmission and response lifetime.
Observation failure or rate limiting of the fenced request SHALL latch that fence against later attempts with new collectors.
An already obtained HTTP result SHALL retain its original meaning. A request without a fence SHALL retain existing behavior.
These internal capabilities SHALL NOT imply that the measurement runtime, engine identity binding or durable daily abort
protocol is complete or activated.

#### Scenario: past gaps do not prevent a stable new baseline

- **GIVEN** a verified observed Budget has a valid state, historical gaps or events and no pending observation
- **WHEN** a new measurement fence captures the current cursor
- **THEN** earlier records remain unchanged and do not count as new incidents for that fence

#### Scenario: current unresolved state is not treated as historical

- **WHEN** baseline creation sees pending Busy or Orphan, dirty state, an identity mismatch or a read failure
- **THEN** it returns an error without creating a usable fence or repairing the state

#### Scenario: a shared incident precedes the next fenced transmission

- **GIVEN** another participant records a new event or gap after the baseline and releases the shared wire lease
- **WHEN** a fenced token or data request obtains that lease
- **THEN** it is rejected before its own Begin and first HTTP Write, with request-body and lease cleanup

#### Scenario: a new collector cannot clear an observation failure

- **GIVEN** a fenced attempt records a 429 or an observation integrity failure
- **WHEN** a later logical request uses a new collector with the same fence
- **THEN** it cannot start a new wire, and the previous HTTP result is not reclassified as not sent

#### Scenario: an absent fence preserves the existing transport

- **WHEN** an existing caller supplies no measurement fence
- **THEN** its existing admission, observation, response ownership and retry semantics remain unchanged

### Requirement: Stock basic info exposes the broker security type

The official stock basic-info read (`GET /api/v1/stocks`) SHALL carry the broker's `securityType` value verbatim into its
output as `security_type` for every returned symbol, because ETF, ETN and stock rows cannot otherwise be told apart by any
consumer. Producers that do not receive the value (WTS reads, `/prices`) SHALL leave the field empty and omit it from JSON, so
their output bytes are unchanged.

#### Scenario: an ETF and a stock are told apart

- **WHEN** `tossctl ops call stocks` reads two symbols whose broker responses carry `securityType` `ETF` and `STOCK`
- **THEN** the output rows carry `security_type` `ETF` and `STOCK` respectively, unchanged from the broker value

#### Scenario: producers without the value keep their bytes

- **WHEN** a quote is produced by a read that has no `securityType` (for example `/prices`)
- **THEN** its JSON has no `security_type` key and is byte-identical to the output before this change
