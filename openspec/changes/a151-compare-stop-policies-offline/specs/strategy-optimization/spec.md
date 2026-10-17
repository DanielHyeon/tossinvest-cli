## ADDED Requirements

### Requirement: Stop-policy comparisons vary one decision at a time
Offline evaluation SHALL pair identical eligible entries across fixed, structural and ATR initial stops with the same exit and cost rules, recalculate quantity under identical caps, and freeze definitions before outcomes.

#### Scenario: One arm lacks warmup
- **WHEN** ATR history is insufficient for an entry
- **THEN** the paired comparison excludes that entry for all arms and reports the exclusion reason

### Requirement: EMA exits retain independent immediate protection in experiments
EMA9 close-based exit experiments SHALL retain the same initial protective stop and SHALL be evaluated separately from changes to initial stop distance.

#### Scenario: Intrabar stop and later EMA close
- **WHEN** a protective stop is touched before a closing EMA signal
- **THEN** protection takes precedence; the experiment does not wait for bar close

### Requirement: Ambiguous intrabar paths are disclosed
OHLC evaluation SHALL count ambiguous same-bar stop/target events and use conservative ordering with sensitivity bounds instead of claiming known execution order.

#### Scenario: Both levels touched
- **WHEN** one bar spans both target and stop
- **THEN** stop-first base results and ambiguity counts are reported


### Requirement: Stop comparisons freeze fill prices and target geometry
The baseline stop-only experiment SHALL use the next eligible one-minute open after signal confirmation, adverse tick/slippage rounding, gap-through-stop open pricing and the same absolute baseline 2.1 percent target for every arm. Missing executions SHALL be reported as unfilled or censored and SHALL NOT become realized fills.

#### Scenario: Gap below the protective stop
- **WHEN** the next eligible open is below the stop price
- **THEN** the simulated exit uses that worse open less adverse slippage and costs rather than the stop price
