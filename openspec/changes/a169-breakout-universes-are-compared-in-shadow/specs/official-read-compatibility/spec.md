## ADDED Requirements

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
