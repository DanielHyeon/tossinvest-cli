## ADDED Requirements

### Requirement: 승인 후보는 독립 lane의 순수 결정으로 변환된다
strategy lane는 ApprovedCandidate와 versioned market inputs를 받아 EntryDecision 또는 명시적 refusal을 반환해야 하며 broker·journal·운영 토글을 직접 변경해서는 안 된다 (MUST NOT).

#### Scenario: 진입 결정
- **WHEN** 활성 lane가 유효 ApprovedCandidate를 평가해 진입 조건을 충족한다
- **THEN** candidate ID, lane ID/version, stop/target와 RiskIntent 입력을 가진 결정이 생성된다

#### Scenario: lane OFF
- **WHEN** lane desired/effective state가 OFF다
- **THEN** 신규 EntryDecision과 buy mutation은 0건이고 기존 exit loop는 계속된다

### Requirement: strategy entry는 공식 LIVE 경로만 사용한다
승인된 strategy entry는 Guardian, durable journal과 official Open API gateway를 순서대로 통과해야 하며 paper/shadow/canary order path를 가져서는 안 된다 (MUST NOT).

#### Scenario: 운영자 LIVE 승인
- **WHEN** 전체 gate가 통과하고 운영자가 lane와 automation을 명시적으로 승인한다
- **THEN** 다음 유효 결정은 공식 LIVE gateway를 사용한다

#### Scenario: Guardian refusal
- **WHEN** Guardian이 첫 실패 단계에서 거부한다
- **THEN** broker request는 0건이고 refusal과 provenance가 journal에 기록된다
