## ADDED Requirements

### Requirement: General chart ordering is chronological

General official charts SHALL expose oldest-to-newest candle order from newest-first wire responses without changing candle timestamps or values.

#### Scenario: Multiple candles
- **WHEN** a general chart receives three descending-time candles
- **THEN** returned chart is ascending-time with latest candle last and unchanged timestamp/price/volume values

#### Scenario: Empty or singleton chart
- **WHEN** a valid general response has zero or one candle
- **THEN** adaptation succeeds without inventing or dropping data

### Requirement: Strict engine candle contracts are preserved

Strict candle readers SHALL preserve newest-first raw pages and close-time cursors. Existing officialbars close-to-open conversion SHALL remain unchanged. Domestic strict read lexical validation SHALL accept six ASCII uppercase alphanumeric characters while preserving market selection, exact request-symbol propagation and existing data validation. It MUST NOT invent response symbol identity verification absent from the wire contract. No strategy universe or order permission SHALL change.

#### Scenario: Domestic alphanumeric code
- **WHEN** a valid domestic read requests 0101N0 and receives matching valid data
- **THEN** lexical validation accepts it and existing content validations still run

#### Scenario: Invalid code or normalization attempt
- **WHEN** a request has invalid length, non-ASCII, punctuation, lowercase or surrounding whitespace
- **THEN** strict lexical refusal remains effective without silently normalizing the symbol

#### Scenario: Request symbol propagation
- **WHEN** a valid symbol is used for a strict read
- **THEN** its exact value is used in the request and returned Page.Symbol without claiming independent broker-response identity attestation

#### Scenario: Strict multi-page candles
- **WHEN** strict domestic or US pages contain newest-first close-time candles
- **THEN** raw order, cursor progression and officialbars open-time conversion remain unchanged

### Requirement: Public API baseline is reproducible

Repository documentation SHALL contain official REST and AsyncAPI snapshots with verified version, source URL, retrieval date and reproducible canonical hash metadata using exact jq -S . output bytes including trailing newline. Existing monitoring configuration SHALL remain unchanged.

#### Scenario: Public snapshot refresh
- **WHEN** current public contracts are retrieved
- **THEN** recorded versions, counts and hashes match saved JSON and the REST metadata remains compatible with its existing monitor

#### Scenario: Scope claims
- **WHEN** the compatibility receipt is produced
- **THEN** it distinguishes stored AsyncAPI evidence from unimplemented WebSocket clients or automated AsyncAPI monitoring
