## ADDED Requirements

### Requirement: strategy engine은 canonical four-family matrix를 exact set으로 봉인한다
Strategy engine은 continuation, reversal, weekly-value와 breakout-retest의 KR/US descriptor 정확히 8개를 canonical matrix로 등록해야 하며 (SHALL), 각 descriptor는 market, family, horizon, lane ID/version, release/config/evidence schema와 `OFF/OFF/UNOBSERVED` 기본 상태를 가져야 한다 (SHALL). Descriptor validation, typed lane input, evaluator/proposal adapter와 production scope는 같은 exact matrix를 검증해야 하며 (SHALL), partial, duplicate, unknown 또는 market/version mismatch를 허용해서는 안 된다 (MUST NOT).

#### Scenario: exact eight descriptor registry
- **WHEN** production registry를 생성하고 검증한다
- **THEN** continuation/reversal/weekly-value/breakout-retest의 KR/US 8개만 존재하고 모두 OFF/OFF/UNOBSERVED다

#### Scenario: breakout adapter 일부만 배선
- **WHEN** descriptor에는 KR breakout이 있지만 typed input 또는 proposal adapter가 빠져 있다
- **THEN** production assembly가 fail-closed로 거부되고 기존 lane 또는 weekly adapter로 fallback하지 않는다

### Requirement: 모든 four-family lane 결정은 순수성과 lineage seal을 유지한다
각 canonical lane는 ApprovedCandidate와 exact typed immutable input을 받아 proposal 또는 stable refusal을 반환해야 하며 (SHALL), broker/journal/toggle mutation을 직접 수행해서는 안 된다 (MUST NOT). Accepted proposal은 family, market, lane/version, candidate/setup, evidence/config, owner scope, campaign/leg, execution terms와 arbitration provenance를 seal해야 한다 (SHALL).

#### Scenario: breakout input을 weekly input으로 대체
- **WHEN** route decision은 breakout lane인데 caller가 weekly-value tagged input을 제공한다
- **THEN** typed input mismatch로 거부되고 evaluator와 broker call은 0건이다

#### Scenario: complete four-family lineage
- **WHEN** 어느 family의 valid proposal이 shared dispatch 후보가 된다
- **THEN** family/lane/version과 candidate/setup/evidence/config/owner/arbitration identifier가 symbol/time 추정 없이 결정적으로 연결된다

### Requirement: production route manifest는 시장별 네 family를 원자 검증한다
Production route manifest는 KR/US 각 시장에 continuation, reversal, weekly-value와 breakout-retest descriptor 정확히 4개를 요구해야 하며 (SHALL), horizon/lane/version/family/config/evidence/scoring binding을 원자 검증해야 한다 (SHALL). Legacy three-family manifest 또는 한 family만 갱신된 partial manifest는 4-family 권위를 만들지 않고 runtime entry를 OFF로 유지해야 한다 (SHALL).

#### Scenario: legacy three-descriptor manifest
- **WHEN** production loader가 continuation/reversal/weekly-value만 가진 current-looking market manifest를 읽는다
- **THEN** four-family runtime activation으로 인정하지 않고 typed matrix refusal과 exposure-raising request 0건을 반환한다

#### Scenario: exact KR and US paired matrix
- **WHEN** 두 시장 모두 exact four-family descriptors와 valid digests를 가진다
- **THEN** registry/manifest 구조 검증은 통과하되 lane desired/effective 또는 LIVE approval은 자동으로 ON이 되지 않는다

### Requirement: production route authority는 평가 전 cross-family winner를 고르지 않는다
Production route authority는 시장의 eligible continuation, reversal, weekly-value와 breakout-retest candidate를 exact family binding과 함께 sealed `RouteSet`으로 모두 반환해야 하며 (SHALL), raw `Candidate.Score`, registry order 또는 market-wide single-proposal 규칙으로 pure evaluator 전에 한 family를 선택해서는 안 된다 (MUST NOT). Cross-family winner는 각 candidate가 exact typed evaluator/proposal path를 통과한 뒤 common calibrated score를 검증하는 market coordinator만 정할 수 있다 (SHALL).

#### Scenario: raw score가 가장 높은 candidate
- **WHEN** raw score가 가장 높은 continuation candidate와 더 높은 calibrated proposal을 만드는 breakout candidate가 같은 owner scope에 있다
- **THEN** RouteSet은 둘 다 평가 대상으로 보존하고 coordinator만 post-evaluation score로 winner를 정한다

#### Scenario: 두 symbol이 같은 market에 존재
- **WHEN** KR에 서로 다른 owner scope의 두 symbol candidate set이 동시에 존재한다
- **THEN** market-wide 하나를 사전 선택하지 않고 두 owner scope가 독립적으로 coordinator intake에 도달한다
