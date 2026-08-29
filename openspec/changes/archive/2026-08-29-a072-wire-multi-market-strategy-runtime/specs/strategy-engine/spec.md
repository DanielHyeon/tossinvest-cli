## ADDED Requirements

### Requirement: production strategy dispatch는 router와 campaign lineage를 보존한다
Strategy engine은 production dispatch에서 router와 campaign lineage를 보존해야 한다 (SHALL). ApprovedCandidate와 immutable evidence를 market/horizon router에 전달하고
선택된 lane의 순수 EntryDecision 또는 typed refusal을 production runtime에 반환해야 한다
(SHALL). 결정은 market, candidate/evidence digest, router/version, lane/version과 campaign/leg
identifier를 포함해야 하며 (SHALL), lane과 router는 Guardian, broker, journal mutation 또는
운영 토글 writer를 직접 호출해서는 안 된다 (MUST NOT).

#### Scenario: US reversal lane 선택
- **WHEN** current US ApprovedCandidate가 router를 거쳐 reversal lane에서 수락된다
- **THEN** EntryDecision은 US market, router/lane version과 campaign/leg lineage를 포함하고 broker call은 아직 0건이다

#### Scenario: router refusal
- **WHEN** candidate horizon과 활성 lane binding이 일치하지 않는다
- **THEN** typed router refusal을 반환하고 lane evaluation, Guardian과 broker request는 0건이다

#### Scenario: 순수 lane의 mutation 의존성
- **WHEN** strategy lane 또는 router dependency closure를 검사한다
- **THEN** broker mutator, journal writer와 operating-setting writer가 존재하지 않는다
