## ADDED Requirements

### Requirement: Tracked order ownership is proven

The durable tracked-order source SHALL return a non-terminal fill snapshot only when journal evidence proves exactly one local engine owner through a confirmed mutation attempt or recorded replacement lineage in the same account, market-local trading day, symbol, and side. A broker-only observation without that evidence MUST remain external and MUST NOT be followed as a local engine order after it leaves the open list. Ambiguous same-scope ownership MUST fail closed as an identifier conflict without projecting a position, invoking fill hooks, or releasing a reservation.

#### Scenario: External open order disappears
- **WHEN** a broker open order with no local confirmed attempt or lineage is observed, stored, and later absent from the broker open list
- **THEN** it is not returned as a locally tracked order and no OrderByID follow-up is attributed to engine ownership

#### Scenario: Confirmed engine order leaves the open list
- **WHEN** a confirmed engine order or its recorded replacement leaves the broker open list before a terminal snapshot is stored
- **THEN** it remains in the tracked set and is read by identifier until a broker-derived terminal state is durably recorded

#### Scenario: Broker reuses an order identifier on a later trading day
- **WHEN** a previously terminal broker order identifier is observed again on a later market-local trading day in the same account
- **THEN** the later observation starts a new cumulative fill sequence and cannot inherit the prior order's terminal or filled quantity

#### Scenario: Canonical ownership is ambiguous
- **WHEN** more than one local intent claims the same account, market-local trading day, symbol, side, and broker order identifier
- **THEN** the journal records an identifier conflict atomically and performs no local fill projection, hook, or reservation release

#### Scenario: Replacement identifiers are reused outside the selected scope
- **WHEN** another account, market, or trading day has a confirmed amendment with the same parent or child order identifier
- **THEN** tracked orders, reconciliation local state, and live-order cancellation follow only the confirmed amendment in the selected canonical scope

#### Scenario: Legacy and scoped lineage disagree
- **WHEN** validated legacy lineage and schema-v16 scoped lineage name different successors in one canonical scope
- **THEN** resolution records an account-wide identifier conflict and refuses to choose either successor

#### Scenario: Legacy empty-scope evidence meets reused identifiers
- **WHEN** a schema-v15 fill snapshot has no account or trading-day scope and its order or lineage endpoint is reused on another trading day in the same account and market
- **THEN** the snapshot is not attributed to either reused tracked lineage scope, while a terminal snapshot in another market remains authoritative only for its matching market

#### Scenario: Two canonical orders share one opaque identifier
- **WHEN** two engine-owned orders in different canonical market, trading-day, symbol, or side scopes share the same broker order identifier
- **THEN** their cumulative snapshots coexist durably and the detector polls, derives lineage, and applies each scope independently without overwriting or skipping either order

#### Scenario: Reused identifier has fills and corrections in both scopes
- **WHEN** two canonical scopes sharing one opaque identifier each append fills or execution corrections
- **THEN** their event and correction streams, provenance, and trade outcomes remain separated by the complete canonical identity, and an order-id-only compatibility read fails closed as ambiguous

#### Scenario: New observation has only part of canonical scope
- **WHEN** a new observation supplies some but not all of account, market-local trading day, symbol, and side
- **THEN** persistence rejects it without overwriting an existing snapshot or re-emitting a cumulative fill delta

#### Scenario: External evidence predates a future exact owner
- **WHEN** a broker-only snapshot or event is stored before the first confirmed local intent and a later order reuses the same complete canonical identity
- **THEN** the earlier evidence remains external and cannot terminate, track, project, explain, or contribute P&L to the later order

#### Scenario: Two intents claim the same canonical identity
- **WHEN** two confirmed intents name the same account, market-local trading day, symbol, side, and opaque order identifier
- **THEN** legacy migration binds no event or correction, provenance/outcomes attribute no fill, and the existing identifier-conflict gate remains fail closed

#### Scenario: Legacy event has one preexisting confirmed owner
- **WHEN** an append-only blank-scope fill event predates schema v19 and exactly one owner's confirmed transition predates that event
- **THEN** migration leaves the original event unchanged, writes one additive canonical binding, and scoped reads, provenance, and outcomes consume the composed evidence

#### Scenario: Ownership and evidence share a released second
- **WHEN** legacy ownership and evidence have the same second-resolution timestamp
- **THEN** the evidence remains unbound because equality is not causal proof; new confirmations and fill commits serialize a strict durable order so owner-before-evidence remains attributable after restart

#### Scenario: Confirmed cancellation echoes the target identifier
- **WHEN** a confirmed CANCEL response repeats the broker identifier of the PLACE it cancelled
- **THEN** the cancellation is audit history but never a second order owner, and live-order lookup remains available to the emergency exit path

#### Scenario: External cumulative snapshot precedes a reused local order
- **WHEN** an external exact-scope snapshot exists before one local order-creating confirmation and a later observation reports that local order's cumulative fill
- **THEN** the new local sequence starts from zero and applies the full cumulative quantity rather than subtracting the external snapshot
