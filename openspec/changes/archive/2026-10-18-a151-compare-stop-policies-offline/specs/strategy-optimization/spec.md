## ADDED Requirements

### Requirement: Stop-policy comparisons vary one decision at a time
Offline evaluation SHALL pair identical eligible entries across fixed, structural and ATR initial stops with the same exit and cost rules, recalculate quantity under identical caps, and freeze definitions before outcomes.

#### Scenario: One arm lacks warmup
- **WHEN** ATR history is insufficient for an entry
- **THEN** the paired comparison excludes that entry for all arms and reports the exclusion reason

### Requirement: Stop comparison excludes strategic exit changes
The initial-stop tool SHALL support only the three initial stop arms with a common target and protective touch exits. EMA9 strategic exits SHALL be rejected as unsupported and reserved for a separately specified future experiment.

#### Scenario: EMA comparison requested
- **WHEN** an input asks this tool to execute an EMA9 exit comparison
- **THEN** the tool rejects the unsupported mode and does not silently change the exit rules

### Requirement: Ambiguous intrabar paths are disclosed
OHLC evaluation SHALL count ambiguous same-bar stop/target events and use conservative ordering with sensitivity bounds instead of claiming known execution order.

#### Scenario: Both levels touched
- **WHEN** one bar spans both target and stop
- **THEN** stop-first base results and ambiguity counts are reported


### Requirement: Stop comparisons freeze fill prices and target geometry
The baseline stop-only experiment SHALL use the next eligible one-minute open strictly after signal confirmation, adverse tick/slippage rounding, gap-through-stop open pricing and the same absolute baseline 2.1 percent target for every arm. Missing executions SHALL be reported as unfilled or censored and SHALL NOT become realized fills.

#### Scenario: Gap below the protective stop
- **WHEN** the next eligible open is below the stop price
- **THEN** the simulated exit uses that worse open less adverse slippage and costs rather than the stop price

### Requirement: Offline comparisons support KR and US with explicit units
The tool SHALL support KR/KRW and US/USD regular sessions with explicit caller-supplied calendar, IANA timezone, tick-grid, cost and same-currency cap manifests as specified in design. It SHALL reject missing, inconsistent or unsupported units and tick ranges, disclose unverified provenance and SHALL NOT infer a live exchange calendar or aggregate account returns across currencies.

#### Scenario: US daylight saving and early close
- **WHEN** a US fixture supplies a valid America/New_York session with its applicable offset and an explicit early-close calendar manifest
- **THEN** the tool evaluates only aligned bars within that session and applies its explicit USD tick and cost model

#### Scenario: US currency or timezone mismatch
- **WHEN** USD bars are paired with KRW caps or an offset inconsistent with America/New_York
- **THEN** the candidate is rejected with a reason rather than silently converted

### Requirement: Comparison inputs enforce temporal and numeric validity
The tool SHALL enforce the design's bounded input grammar, unique IDs and JSON keys, OHLC consistency, contiguous aligned bar paths and distinct close/received times. ATR14 SHALL be seeded from fourteen true ranges following the first close anchor in fifteen same-session five-minute bars, then use Wilder smoothing. Signal confirmation SHALL freeze the structural low, ATR and all model/budget/cap snapshots. Only the next scheduled minute open MAY subsequently determine simulated entry, absolute fixed/ATR stops, common target and quantity; its high/low/close and later data SHALL NOT influence those values. Full entry/outcome bars MAY arrive after their close because the open execution is explicitly hypothetical and pre-scheduled, not a retrospective live order.

#### Scenario: Late receipt or insufficient ATR warmup
- **WHEN** a required signal input arrived after confirmation or fewer than fifteen contiguous five-minute bars are available
- **THEN** the candidate is excluded from all three arms with its reason

#### Scenario: Missing next entry bar
- **WHEN** the first scheduled minute open strictly after confirmation is missing
- **THEN** the observation is unfilled, with no jump to a later open

### Requirement: Quantity and costs use exact declared simulation rules
The tool SHALL calculate stop geometry, tick rounding, integer quantity, caps, side fees and tax with exact arithmetic under the frozen design. Each arm SHALL use the same per-candidate risk budget and caps in quote currency, without portfolio compounding. A zero cap or invalid stop SHALL exclude the candidate from all arms. It SHALL label returns simulated R rather than actual fill evidence.

#### Scenario: Zero cap or invalid structural stop
- **WHEN** one arm has zero quantity or its rounded stop is at or above entry
- **THEN** all three arms are excluded from the pre-entry comparison and the reason is reported

### Requirement: Paired comparisons disclose complete cohort and limitations
The tool SHALL report all candidates, exclusions, eligible and complete three-arm cohort counts separately by market. Main paired differences SHALL use identical IDs for which all three arms closed. Missing paths and session-end open positions SHALL remain censored, with mark-to-market separate from realized simulation. Empty denominators SHALL remain null/not_ready. Arm-specific sensitivity endpoints SHALL NOT be represented as jointly achievable paired bounds. Profitability conclusions SHALL remain inconclusive in this delivery.

#### Scenario: One arm remains open at session end
- **WHEN** two arms close but the third remains open
- **THEN** the observation is excluded from the complete three-arm realized comparison while its arm outcomes and censoring remain visible

### Requirement: Offline evaluation preserves source artifacts and runtime isolation
The tool SHALL read only explicit local research inputs, preserve their bytes, and refuse output overwrites including symlink targets. It SHALL NOT fetch data, invoke order or live runtime paths, modify operating configuration or collect new sessions.

#### Scenario: Existing output or protected input selected
- **WHEN** output would replace an existing file, input or symlink target
- **THEN** the command refuses without modifying those artifacts

#### Scenario: Adverse entry slippage crosses structural protection
- **WHEN** the rounded simulated entry is above a valid structural stop but the entry bar open is already at or below that stop
- **THEN** the same open-gap precedence applies immediately after simulated entry and uses the open less exit slippage, not the more favorable stop price
