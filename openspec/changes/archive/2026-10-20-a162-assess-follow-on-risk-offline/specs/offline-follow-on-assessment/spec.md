## ADDED Requirements

### Requirement: Reuse native follow-on proposals without creating authority
The system SHALL assess KR and US continuation, reversal and weekly follow-on requests using their existing native proposal APIs without changing source state, allocation, stop rules or refusal/invalidation semantics. It SHALL use existing same-call one-way capture to read private plan/risk data, reject capture failure and MUST NOT accept detached evidence as an alternate native input or mint execution authority.

#### Scenario: Existing proposal refuses
- **WHEN** the native proposal rejects an incomplete, stale, latched or otherwise invalid follow-on request
- **THEN** assessment preserves native kind, code and invalidation code and returns no hypothetical quantity or authority

#### Scenario: Weekly reservation evidence is absent
- **WHEN** a weekly request lacks the native reservation evidence required by its proposal
- **THEN** assessment refuses without constructing a reservation or substituting a161 inspection data

### Requirement: Assess remaining budget only under explicit risk assumptions
The system SHALL require a positive account-currency per-unit risk upper-bound assumption bound to the family, native risk binding digest and market. The digest SHALL preserve each family's meaning: continuation RiskBudgetDigest versus reversal/weekly PlanDigest. After a successful proposal it SHALL calculate remaining budget as initial budget minus filled and held risk, and hypothetical quantity as the minimum of proposal quantity and the floor of remaining budget divided by the assumption. Quantity SHALL represent additional risk not already in held risk, not replacement or retry of an existing reservation. Weekly slot evidence MUST NOT be added as a second monetary reservation. It SHALL validate at most 78 canonical ASCII decimal digits before parsing and at most 256 bits after parsing and for intermediate arithmetic. It SHALL NOT infer missing amounts, FX or costs.

#### Scenario: Budget reduces the candidate
- **WHEN** budget is 1000, filled risk is 200, held risk is 300, per-unit bound is 100 and candidate quantity is 8
- **THEN** the hypothetical quantity is 5 and planned risk is 500 without modifying the candidate or reserving that risk

#### Scenario: Missing or mismatched assumptions
- **WHEN** the bound is absent, zero, malformed, too large, or bound to a different market, currency or plan digest
- **THEN** assessment refuses without a usable partial quantity

#### Scenario: Exhausted or overdrawn budget
- **WHEN** filled and held risk exhaust or exceed budget
- **THEN** assessment reports no budget or invalid overdrawn state respectively and returns no successful quantity

#### Scenario: Large input or quotient
- **WHEN** a monetary string exceeds 78 digits or 256 bits, or an intermediate risk sum exceeds 256 bits
- **THEN** assessment refuses, while a valid quotient larger than uint64 is limited by candidate quantity before integer conversion

### Requirement: Results remain hypothetical inspection data
Every result SHALL mark inspection-only status, unverified ledger currentness, caller-assumed risk bounds and absence of execution authority. The result MUST NOT be an admission credential or contain execution caps, native plans, leases or dispatchable intents. The module SHALL NOT access brokers, journals, operating configuration or network services.

#### Scenario: Positive offline quantity
- **WHEN** an assessment produces a positive hypothetical quantity
- **THEN** the result still cannot reserve risk, authorize a trade, calculate authoritative qFinal or promote a policy

#### Scenario: No side effects
- **WHEN** assessment succeeds or refuses
- **THEN** all supplied native state remains unchanged and no persistent or operational state is modified

#### Scenario: Offline timestamps do not attest account currentness
- **WHEN** synthetic native evidence passes its own freshness rules
- **THEN** assessment still reports unverified ledger currentness and no execution authority, even when it returns a quantity
- **AND** stale native evidence preserves the native refusal and exposes no quantity

#### Scenario: Later weekly legs preserve native history
- **WHEN** KR or US weekly legs 3 through 7 are assessed in tagged integration tests
- **THEN** their fixtures advance all preceding legs using native reservation and positive-fill transitions, preserve filled and held risk accounting, and assessment leaves that state unchanged
- **AND** a missing required reservation is rejected by the native proposal without an offline quantity

#### Scenario: Non-nil foreign exchange evidence remains immutable
- **WHEN** tagged integration requests contain valid native non-nil FX evidence
- **THEN** the before and after native capture encodings are identical for assessment
- **AND** an independent FX mutation control is detected by the snapshot comparison, and invalid FX input preserves the native refusal without creating new authority
