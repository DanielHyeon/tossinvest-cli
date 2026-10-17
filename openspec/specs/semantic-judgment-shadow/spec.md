# semantic-judgment-shadow Specification

## Purpose
Define reproducible, offline evaluation of semantic-judgment input quality and separately versioned J2 comparisons without changing shadow evidence or trading behavior.
## Requirements
### Requirement: Input quality is evaluated from an immutable snapshot
The offline evaluation SHALL record an immutable source manifest and hashes, validate judgment-to-state and judgment-to-label links, and derive quality classifications in a separate sidecar without modifying source evidence. It SHALL distinguish unavailable feed, successfully read feed with no matching symbol headlines, market versus symbol headlines, observation-time article freshness, unsupported fields, missing values and unknowns; no absence SHALL be inferred from unknown provenance. Only input already present in the historical snapshot is eligible; publication time alone does not establish availability at observation, and later-collected material without as-of availability proof is excluded or marked unknown. Derived outputs MUST NOT disclose account identifiers or credentials.

#### Scenario: News feed was read but symbol list is empty
- **WHEN** a snapshot confirms a successful news feed read with no matching symbol headlines
- **THEN** the sidecar records an empty symbol-news list separately from feed unavailability and does not assert that no relevant news existed

#### Scenario: Source data cannot be validated
- **WHEN** source hashes or required identity links fail validation
- **THEN** the affected evidence is reported invalid and no paired comparison treats it as validated

### Requirement: Existing selection is compared within a complete cycle
The first stage SHALL report separately by lot, market and local session; compare observed selection, J1-only top K, J2-survivor uniform expectation and full-pool uniform expectation using the same cycle and K, with the declared existing tie rule. K is the actual r1 selection count even when less than ten; for K=0 the cycle SHALL be reported and excluded from selection-return and uniform-K comparisons. Labels SHALL join by judgment ID; the paired return comparison SHALL exclude incomplete-label cycles and report their count. Each state SHALL be counted once, results SHALL be described as cost-free horizon price changes, and overlapping observations SHALL NOT be treated as independent account trades. Return comparisons SHALL weight eligible candidates equally within each cycle and cycles equally within each market/session. For each quality stratum, the report SHALL include state count, J2 exclusion rate, selection rate, 60-minute mean and missing-label count; any row-weighted stratum mean SHALL be distinguished from the cycle-weighted comparison and repeated observations disclosed; unknown SHALL remain distinct and overlapping strata SHALL NOT be summed. Insufficient data SHALL be inconclusive; thresholds SHALL NOT be optimized after seeing outcomes.

#### Scenario: One candidate lacks a label
- **WHEN** a candidate in a cycle lacks a valid horizon label
- **THEN** that cycle is excluded from paired return comparisons and the reason and count are reported

#### Scenario: No r2 response exists
- **WHEN** the immutable snapshot supports stage one but no compatible r2 response is saved
- **THEN** stage one may complete independently and the empirical r2 comparison remains blocked

### Requirement: J2 revision is paired without replacing original J1 scores
The r2 comparison SHALL use a separate lot and match frozen state, observation identity, model and model version, unchanged J1 question and all other request inputs; only J2 wording changes. The isolated selection replay SHALL keep the original r1 J1 scores and substitute only matched r2 J2 scores. A paired replay SHALL require compatible responses for every eligible candidate in a cycle; otherwise the entire cycle SHALL be excluded with reason and count. K SHALL be the original r1 cycle selection count; if r2 has fewer than K survivors, defer that cycle from the equal-K comparison and report actual counts separately without filling or altering r1. This fixed-J1 replay SHALL be identified as rule sensitivity, not joint r2 response performance, and SHALL disclose that J1 response effects are unmeasured. The r2 question SHALL retain the existing binary answer structure, distinguish missing information from an explicit adverse signal, and SHALL NOT imply that lack of visible adverse evidence establishes safety or sufficient information.

#### Scenario: Compatible response is missing or provenance mismatches
- **WHEN** saved r2 responses are absent or lack matching state, model version or question provenance
- **THEN** empirical comparison is blocked for those pairs without inventing answers or claiming a pass

#### Scenario: Some candidates lack compatible r2 responses
- **WHEN** even one eligible candidate in a cycle lacks a compatible r2 response
- **THEN** exclude the entire cycle from paired replay and report the reason and count

#### Scenario: Fewer r2 survivors than original K
- **WHEN** the r2 survivor count is less than the original r1 cycle selection count K
- **THEN** defer that cycle from equal-K comparison and report actual selection counts without filling missing places

#### Scenario: Synthetic response is supplied for testing
- **WHEN** a synthetic r2 response is used in an offline fixture
- **THEN** it tests comparison plumbing only and is not reported as an empirical model result

### Requirement: Future labels are excluded from model inputs
Any separately authorized future response collection SHALL use only data available at the original observation time, exclude future labels and future-dated input, and record the exact input and model provenance. If the frozen state contains future information, the pair SHALL be ineligible; filtering only the r2 input SHALL NOT be treated as an exact-input match. Later model training knowledge or version changes SHALL be disclosed as limitations; a historical comparison SHALL NOT be presented as a prospective causal result.

#### Scenario: A candidate input contains a future label
- **WHEN** an input would expose an outcome later than its observation time
- **THEN** that input is rejected before any separately authorized model request
