## ADDED Requirements

### Requirement: 신규 Story와 OpenSpec은 같은 번호를 사용한다
TossOS는 `a040` 이후 신규 번호형 change를 `aNNN-kebab-intent`로 명명하고 정확히 하나의 `STORY-TOS-aNNN`과 연결해야 한다 (SHALL). 기존 비번호형 change와 `STORY-TOS-001~039`는 historical legacy로 허용해야 한다 (SHALL).

#### Scenario: 정상 신규 change
- **WHEN** `a041-complete-exit-line-contract`와 `STORY-TOS-a041`이 서로를 가리킨다
- **THEN** PM 검증은 번호·slug·1:1 mapping을 승인한다

#### Scenario: 번호 불일치
- **WHEN** `STORY-TOS-a041`이 `a042-*` change를 가리킨다
- **THEN** PM 검증은 번호 불일치로 실패한다

#### Scenario: legacy 보존
- **WHEN** 기존 `STORY-TOS-039`가 기존 무번호 archived change를 가리킨다
- **THEN** PM 검증은 migration을 강요하지 않고 기존 mapping을 승인한다

### Requirement: 번호와 intent 형식은 기계적으로 검증된다
PM 검증기는 신규 change의 3자리 번호 중복, 빈 intent, 대문자·underscore·비-kebab slug를 거부해야 한다 (MUST).

#### Scenario: 중복 번호
- **WHEN** 서로 다른 두 active change가 같은 `a047` 번호를 사용한다
- **THEN** 검증은 두 경로를 모두 지목하며 실패한다
