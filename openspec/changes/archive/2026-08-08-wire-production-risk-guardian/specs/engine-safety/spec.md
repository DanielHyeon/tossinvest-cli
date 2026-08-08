## ADDED Requirements

### Requirement: Production startup constructs one Guardian after durable scope exists
The engine profile SHALL, when `engine.automation_gate.enabled` is true and no
explicit test Guardian is injected, resolve the official account and open the
writable engine journal before constructing exactly one production
`RiskGuardian`. It SHALL construct no Guardian and no loop set when the gate is
off. A Guardian construction failure SHALL close the journal, record/refuse
startup, and start no loop.

#### Scenario: Gate-on production assembly
- **WHEN** the real CLI assembler loads a valid gate, resolves an account, and opens the engine journal
- **THEN** it constructs exactly one `RiskGuardian` scoped to that account and journal before running the interlock

#### Scenario: Gate-off production assembly
- **WHEN** the real CLI assembler loads an automation gate that is off
- **THEN** it does not construct a Guardian and the command starts no engine loop

#### Scenario: Guardian construction fails
- **WHEN** the configured gate cannot produce a valid production Guardian after the journal is open
- **THEN** startup is refused, the journal is closed, and no loop or order side effect begins

### Requirement: Interlock and runtime share the Guardian identity
The engine SHALL pass the same Guardian instance to the startup interlock and,
after verification, publish that instance on `Context.Guardian`. The exit
observer SHALL obtain its `ReductionIssuer` from that field and SHALL NOT
construct, substitute, or bypass another Guardian.

#### Scenario: Verified context constructs exit observation
- **WHEN** the startup interlock verifies a production Guardian
- **THEN** the context publishes that exact instance and the exit observer uses it for reduce-only issuance

#### Scenario: Guardian cannot issue reductions
- **WHEN** the verified context carries a Guardian that does not implement `ReductionIssuer`
- **THEN** exit-observer construction fails before any observation loop starts

### Requirement: Command regression exercises production assembly
The command package SHALL have a regression test that invokes the actual
production assembly helper with an isolated config directory, a real SQLite
journal on an allowlisted test filesystem, and an `httptest` official broker.
The test SHALL inject no Guardian, SHALL contact no live endpoint, and SHALL
prove verified Guardian and exit-observer wiring while protection remains the
shipped `UNWIRED` value. The durability test override SHALL exist only under a
dedicated Go test build tag and SHALL have no flag, environment variable,
config key, or ordinary production symbol.

#### Scenario: Isolated USD CLI assembly
- **WHEN** the command regression supplies valid USD limits, credentials, attestation, account response, and no Guardian override
- **THEN** the actual CLI assembly constructs exactly one real `RiskGuardian`, returns the configured USD snapshot, records an isolated reduce-only decision through the context journal, and constructs an exit observer from that same Guardian
