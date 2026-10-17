## ADDED Requirements

### Requirement: Volume profile feasibility separates executions from proxies
The feasibility report SHALL classify actual trades, documented volume bins, OHLCV proxies and unavailable inputs, preserving source, time semantics, coverage and correction handling.

#### Scenario: Only OHLCV exists
- **WHEN** only five-minute candles and orderbook snapshots are provided
- **THEN** true volume profile is unsupported and neither candles nor book size are relabeled as executed price-bin volume

### Requirement: Unsupported volume profile does not block non-profile research
Feasibility SHALL complete with an unsupported result when inputs are unavailable and SHALL supply an explicit exclusion reason to downstream experiments.

#### Scenario: No licensed trade data
- **WHEN** no suitable existing input is available
- **THEN** no new collection starts and the report completes without inventing LVN/HVN evidence
