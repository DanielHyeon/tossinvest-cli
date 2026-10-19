## ADDED Requirements

### Requirement: Strategy comparisons are preregistered and cost aware
Offline exploratory evaluation SHALL declare fixed filter and cost hypotheses and bind immutable local input and source hashes. It SHALL compare baseline with candle-only, volume-only and MFI-only variants using identical a151 fixed-stop execution, independently for KR and US, and recompute the complete path under doubled entry and exit slippage. It SHALL NOT claim hashes establish preregistration or an untouched holdout; confirmatory preregistration and consumed-holdout enforcement are deferred and no confirmatory result is permitted.

#### Scenario: Independent filters
- **WHEN** valid a151-compatible candidates are evaluated
- **THEN** each filter acts alone and all selected variants share the fixed baseline execution/cost assumptions, with inherited a151 exclusions explicitly reported

#### Scenario: Tuning after outcomes
- **WHEN** an input or parameter is changed after outcomes were viewed
- **THEN** results remain exploratory and inconclusive, with no promotion or untouched-holdout assertion

### Requirement: Insufficient evidence never produces a promoted policy
All a153-v1 results SHALL remain inconclusive and unauthorized for production. Reports SHALL preserve full candidate/session denominators and missingness, distinguish known no-trade zero from unknown outcomes, use a disclosed common complete-session cohort, and report equally weighted session-mean simulated-risk-normalized contribution differences per market. Exploratory 95% paired session bootstrap bounds SHALL use 2000 resamples, seed 153 and linear percentile interpolation. Missing or fewer than twenty sessions SHALL be explicit; twenty or more SHALL NOT establish adequacy, population performance or policy eligibility.

#### Scenario: Suppression and missing outcomes
- **WHEN** a filter rejects a candidate with known closed baseline
- **THEN** its contribution is zero and its paired difference is minus baseline simulated R, while censored baseline or unknown filter is not assigned zero and prevents that session joining the common cohort

#### Scenario: Small sample
- **WHEN** five sessions show positive mean improvement
- **THEN** the report remains inconclusive and produces no approved runtime choice

### Requirement: Ablations preserve research boundaries
Comparisons SHALL use only confirmed signal-time candle/volume inputs and the existing a147 one-minute MFI definition, with receipt and precision checks before use. The tool SHALL preserve existing source/evidence/configuration and perform no collection or orders. Unsupported VP, SI, MA20, EMA exits, partial exits, portfolio drawdown and empirical promotion SHALL be explicitly deferred rather than synthesized.

#### Scenario: Future feature information
- **WHEN** a bar closes or arrives after signal confirmation, or exceeds accepted timestamp precision
- **THEN** it cannot establish a passing feature condition

#### Scenario: Volume-profile unsupported
- **WHEN** a152 reports OHLCV-only data
- **THEN** the evaluation excludes VP and continues independent candle, volume and MFI comparisons without fabricating profile values
