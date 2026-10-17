## ADDED Requirements

### Requirement: Prospective MFI shadow evaluation is preregistered

The evaluation SHALL create an immutable prospective plan before its first observation, recording all KR and US session dates or an objective consecutive-session selection rule, regular-session completeness rules, candidate and cycle selection, data provenance and timezone, 60-minute label timing, maximum latest-bar age and maximum label delay, missingness limits, interruption, interim reporting and stop rules, and analysis version. Any later amendment SHALL be separately versioned with its reason and distinguish pre- and post-amendment observations. Safety or operational stops remain permitted; outcome-driven early stopping SHALL NOT be represented as confirmatory success. Five new full regular sessions per market SHALL be the exploratory target, not a claim of statistical sufficiency. The evaluation SHALL retain and report all attempted partial or failed sessions and SHALL NOT silently replace sessions according to results. A missing maximum latest-bar age or label delay SHALL prevent the run. New collection or model/API calls require separate authorization.

#### Scenario: Session is interrupted or incomplete
- **WHEN** a registered session ends early or misses the preregistered completeness rules
- **THEN** it is retained with its reason and is not counted as a full session; no unregistered replacement is selected based on its result.

#### Scenario: Collection is not separately authorized
- **WHEN** only the present design is approved
- **THEN** no collection, model/API call, schedule or runtime change is performed.

### Requirement: MFI evidence preserves as-of identity and r1

The evaluation SHALL append a separately identified sidecar with source and offset-bearing original closed-bar timestamps, OHLCV and observation provenance, linked by market, local session, cycle, symbol and evidence digest. The bars SHALL be linked to the original judgment input or identical evidence demonstrably received before judgment. Later retrieval of past bars SHALL NOT establish as-of availability or identity without independent proof; a close time and offset alone are insufficient. The r1 JSON bytes, prompts, model, selection, decisions and historical evidence SHALL remain unchanged. Labels and later bars SHALL NOT be supplied to or used in an earlier judgment.

#### Scenario: Source time cannot be proved
- **WHEN** the source offset, close time, identity or availability at observation cannot be established
- **THEN** the candidate is excluded with a counted reason; the evaluator SHALL NOT infer an offset from a convenient outcome or rewrite historical evidence.

### Requirement: MFI trend and outcomes are calculated consistently

The evaluation SHALL calculate one-minute MFI14 and its change from three closed bars earlier using at least 18 valid consecutive closed bars. For positive and negative typical-price-times-volume flow sums P and N, MFI is `100 - 100/(1 + P/N)` when N > 0; N=0 and P>0 gives 100, P=0 and N>0 gives 0, and both zero is undefined. A three-bar delta within 1e-9 MFI points of zero is unchanged. It SHALL count and exclude future, missing, duplicate, nonconsecutive, invalid or insufficient bars, bars older than the preregistered age, undefined directional flow, and missing, invalid or excessively delayed 60-minute labels. It SHALL join each candidate's paired questions once by state/symbol/cycle and its outcome by judgment ID, and report selected and nonselected results separately for each market and local session.

#### Scenario: Closed bars or label are insufficient
- **WHEN** the 18-bar sequence is incomplete or the 60-minute label is missing or invalid
- **THEN** the evaluator reports the reason and excludes that candidate from the relevant calculation or outcome denominator rather than inventing a value.

#### Scenario: A bar closes after the judgment
- **WHEN** a bar close timestamp is later than the candidate observation time
- **THEN** that bar is not used and the exclusion is reported.

### Requirement: Observational findings remain bounded

The evaluation SHALL report counts, missingness, cost-free 60-minute mean price change and strictly-positive outcome proportion by MFI rising/falling/unchanged and selected/nonselected groups, plus same-cycle rising-minus-falling means with equal cycle weight and matched-cycle count. It SHALL preserve denominators from all observed candidates through MFI-valid/excluded reasons and label-valid/missing counts, broken down by market, session and selected/nonselected/selection-unknown; unknown selection or ambiguous identity SHALL NOT be classified as nonselected. It SHALL disclose repeated symbols, overlapping horizons and incomplete sessions, and SHALL NOT present the observational contrast as a same-K strategy comparison, account return, proof of causality or evidence that changing Jev inputs improves performance.

#### Scenario: Evidence is too small or ambiguous
- **WHEN** the prespecified evidence is inadequate or provenance is ambiguous
- **THEN** the result is inconclusive with denominators and limitations, not a performance PASS.
