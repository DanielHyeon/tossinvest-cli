## ADDED Requirements

### Requirement: Session features have explicit reset seed and time semantics
Feature calculation SHALL use session-prefix unadjusted closed one-minute bars for typical-price VWAP and session-reset closed five-minute EMA9 seeded by the first nine closes. It SHALL return method and timing provenance and SHALL NOT replace frozen v1 runtime inputs.

#### Scenario: Session boundary
- **WHEN** the first bar of a new supported session arrives
- **THEN** VWAP accumulators and EMA seed reset and EMA remains not_ready until nine complete five-minute bars exist

### Requirement: Incomplete session inputs cannot become ready indicators
The calculator SHALL reject future, duplicate, malformed or mixed-session bars and SHALL mark incomplete prefixes, zero total volume and insufficient warm-up not_ready. Unsupported calendars SHALL NOT be inferred.

#### Scenario: Mid-session start
- **WHEN** only the last twenty minutes of a session are available
- **THEN** no session VWAP or ready EMA is fabricated

### Requirement: VWAP slope and EMA period are reproducible
The calculator SHALL identify EMA periods as five-minute closed bars and compute VWAP percent slope between consecutive five-minute boundaries using an explicit version.

#### Scenario: Equivalent replay
- **WHEN** the same canonical prefix is replayed
- **THEN** values, readiness and input digest are identical with no network or broker calls
