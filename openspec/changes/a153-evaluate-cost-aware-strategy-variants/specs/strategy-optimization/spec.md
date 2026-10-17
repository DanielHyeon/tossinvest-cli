## ADDED Requirements

### Requirement: Strategy comparisons are preregistered and cost aware
Evaluation SHALL freeze immutable data splits, variants and cost assumptions before holdout outcomes, report net performance and dependence-aware uncertainty, and mark reused holdouts as consumed. Exactly one primary candidate SHALL be fixed using train/validation before holdout access; selecting the best holdout variant SHALL NOT support promotion.

#### Scenario: Tuning after holdout
- **WHEN** parameters are revised after viewing test outcomes
- **THEN** the result is exploratory and cannot support promotion without new untouched evidence

### Requirement: Insufficient evidence never produces a promoted policy
Recommendation SHALL require complete provenance, at least twenty independent holdout sessions, positive market-separated equally weighted session mean net-R improvement lower bound of a 95% paired session bootstrap CI (2000 resamples, seed 153), nonworsening drawdown and positive net expectancy under doubled slippage; meeting those checks SHALL NOT activate trading.

#### Scenario: Small sample
- **WHEN** five sessions show positive mean returns
- **THEN** the report remains inconclusive and produces no approved runtime choice

### Requirement: Ablations preserve research boundaries
Comparisons SHALL isolate feature, stop and exit changes; unsupported volume-profile variants SHALL be excluded explicitly and r1/a128/a147 evidence SHALL remain immutable.

#### Scenario: Volume-profile unsupported
- **WHEN** a152 reports OHLCV-only data
- **THEN** the evaluation excludes VP and continues supported variants without fabricating profile values
