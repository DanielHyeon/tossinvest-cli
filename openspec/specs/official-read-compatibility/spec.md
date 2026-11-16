# official-read-compatibility Specification

## Purpose
TBD - created by archiving change a143-official-read-compatibility. Update Purpose after archive.

## Requirements

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

### Requirement: General price reads preserve the broker data timestamp

The general official price adapter SHALL preserve the `timestamp` of each `/api/v1/prices` row verbatim, as a separate raw broker data time string on the domain quote, without decoding it and without adding any branch that can drop a row or return an error because of that field. The adapter MUST NOT replace, shift or derive the existing received time (`FetchedAt`) from the broker data time. A null or absent `timestamp` SHALL leave the broker data time empty and MUST NOT be filled from a local clock; a malformed `timestamp` SHALL be preserved as received. The domain quote's JSON form SHALL omit the broker data time when it is empty, so that quotes from producers that do not set it serialise byte-for-byte as before. No additional request, retry or rate-budget cost SHALL be introduced.

#### Scenario: Both markets carry a KST offset
- **WHEN** a KR row carries `2026-08-27T19:59:59.000+09:00` and a US row carries `2026-08-28T05:28:30.000+09:00`
- **THEN** each quote's broker data time is that string unchanged and its received time is unchanged

#### Scenario: Missing or malformed timestamp
- **WHEN** a row has a null, absent, empty or malformed `timestamp`
- **THEN** adaptation succeeds with the row kept, the broker data time is empty (null, absent, empty) or the received text (malformed), and the received time is still set

#### Scenario: Quotes without a broker data time serialise as before
- **WHEN** a quote produced by the WTS quote reader or the official stock-metadata adapter is written as JSON
- **THEN** the output contains no broker data time key and is byte-identical to the output before this change

#### Scenario: No new broker call
- **WHEN** a batch of symbols is read
- **THEN** exactly one `/api/v1/prices` request is made, as before

### Requirement: KR alphanumeric symbol compatibility covers strict reads but not order policy

The minute-candle reader and both strict quote readers SHALL accept KR symbols of exactly six ASCII uppercase alphanumeric characters through one shared lexical predicate. Other forms MUST be refused before any request and MUST NOT be normalized. Order intent normalization, including its numeric-only KR market inference, and the a087 executable order probe policy MUST remain unchanged. No order path, strategy universe or order permission SHALL change.

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
- **WHEN** order intent market inference is evaluated with KR 0101N0 and the a087 probe tests run with their existing numeric inputs
- **THEN** market inference still returns us for 0101N0 and kr only for six digits, and the a087 symbol policy, request body and execution guards are untouched by this change

#### Scenario: Updated published symbol wording
- **WHEN** the a087 request-schema test reads the refreshed snapshot
- **THEN** it recognizes the current published description while its existing numeric-only production inputs, request body and execution guards remain unchanged
