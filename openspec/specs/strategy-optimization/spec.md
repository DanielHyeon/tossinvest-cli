# strategy-optimization Specification

## Purpose
TBD - created by archiving change a050-add-strategy-optimization. Update Purpose after archive.
## Requirements
### Requirement: 최적화 설정은 versioned lifecycle을 가진다
시스템은 parameter registry, immutable settings snapshot, candidate, preview, apply, history와 rollback을 제공해야 한다 (SHALL).

#### Scenario: candidate preview
- **WHEN** 운영자가 base version과 변경 parameter를 제출한다
- **THEN** validation, 예상 영향, restart 범위와 a049 evidence digest를 포함한 read-only preview를 반환한다

#### Scenario: CAS apply
- **WHEN** 운영자가 current base version의 candidate를 승인한다
- **THEN** 새 settings version과 before/after/actor/reason audit를 원자적으로 기록한다

#### Scenario: rollback
- **WHEN** 운영자가 과거 snapshot으로 rollback을 승인한다
- **THEN** 과거 row를 수정하지 않고 동일 CAS 규칙의 새 version을 생성한다

### Requirement: 근거 부족은 추천 불가다
결정적 lineage, 최소 표본 또는 필수 metric이 부족하면 시스템은 최적 parameter를 생성해서는 안 된다 (MUST NOT).

#### Scenario: link_missing 포함
- **WHEN** 평가 표본의 필수 거래가 `link_missing`이다
- **THEN** 상태를 insufficient evidence로 표시하고 apply 후보를 자동 생성하지 않는다

### Requirement: 설정 적용은 LIVE 권한을 만들지 않는다
optimization apply 또는 rollback은 lane state, automation gate, kill switch 또는 기존 position snapshot을 변경해서는 안 된다 (MUST NOT).

#### Scenario: 설정 apply
- **WHEN** 새 strategy settings version을 적용한다
- **THEN** lane는 기존 OFF/ON 상태를 유지하고 활성 포지션 exit policy도 그대로다

### Requirement: parameter registry는 UI 계약을 완전하게 제공한다
각 설정 descriptor는 category, label, description, type, unit, default state/value, range/choices, owner change, apply timing과 safety direction을 제공해야 한다 (SHALL). 미승인·해당 없음·측정 불가를 숫자 0이나 빈 문자열로 대신해서는 안 된다 (MUST NOT).

변경 가능한 descriptor는 server-defined stable option ID의 유한 목록과 control kind를 제공해야 하며
(SHALL), UI는 자유 텍스트·숫자 직접 입력·contenteditable·typed confirmation을 제공해서는 안 된다
(MUST NOT). decimal/integer도 registry가 승인한 discrete choice 또는 bounded step option으로
표현해야 한다 (SHALL).

#### Scenario: 미승인 threshold descriptor
- **WHEN** a046 threshold가 승인되지 않았다
- **THEN** descriptor는 default state `unapproved`를 반환하고 UI가 임의 숫자 default를 생성하지 못하게 한다

#### Scenario: registry 불완전
- **WHEN** writable parameter에 description, default 또는 apply timing이 없다
- **THEN** 해당 field는 read-only configuration error로 표시되고 preview/apply 대상에서 제외된다

#### Scenario: 자유 입력 없는 descriptor
- **WHEN** 변경 가능한 decimal, integer, enum 또는 symbol-scoped descriptor를 렌더링한다
- **THEN** UI는 registry option ID, 현재 행 action 또는 bounded step choice만 제공하고 text/number 입력을 렌더링하지 않는다

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
