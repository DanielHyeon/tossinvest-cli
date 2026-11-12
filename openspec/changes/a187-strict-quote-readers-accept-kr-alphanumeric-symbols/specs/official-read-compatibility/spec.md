## REMOVED Requirements

### Requirement: Minute symbol compatibility does not widen quote or order policy

**Reason**: The numeric-only rule for the two strict quote readers was a local restriction, not the published contract (REST snapshot 1.2.24 documents KRX symbols as numeric or alphanumeric six-character codes for the orderbook read). It silently removed alphanumeric KR symbols from consumers that price through the strict orderbook read (a184, 2026-11-12, symbol 0158S0).

**Migration**: Replaced by "KR alphanumeric symbol compatibility covers strict reads but not order policy" below. The order-policy half of the removed requirement is carried over unchanged, including the a087 published-wording scenario.

## ADDED Requirements

### Requirement: KR alphanumeric symbol compatibility covers strict reads but not order policy

The minute-candle reader and both strict quote readers SHALL accept KR symbols of exactly six ASCII uppercase alphanumeric characters through one shared lexical predicate. Other forms MUST be refused before any request and MUST NOT be normalized. Order intent validation and the a087 executable order probe policy MUST remain numeric-only for KR symbols. No strategy universe or order permission SHALL change.

#### Scenario: Same alphanumeric input reaches every strict reader
- **WHEN** KR 0101N0 is passed to the minute reader and each strict quote reader using offline fixtures
- **THEN** each reader sends exactly one request carrying symbol=0101N0 or symbols=0101N0 and returns that exact symbol

#### Scenario: Invalid KR forms are still refused with zero requests
- **WHEN** KR 0101n0, " 0101N0", 0101N, 0101N00 or 0101-0 is passed to each strict quote reader
- **THEN** each returns its symbol refusal and the server receives zero requests

#### Scenario: Numeric KR and US symbols are unchanged
- **WHEN** KR 005930 and US AAPL are passed to each strict quote reader
- **THEN** request query, returned symbol and currency are the same as before this change

#### Scenario: Order policy is not widened
- **WHEN** the order intent market inference and the a087 request-schema test are evaluated with KR 0101N0
- **THEN** their existing numeric-only behavior, request body and execution guards remain unchanged

#### Scenario: Updated published symbol wording
- **WHEN** the a087 request-schema test reads the refreshed snapshot
- **THEN** it recognizes the current published description while its existing numeric-only production inputs, request body and execution guards remain unchanged
