## ADDED Requirements

### Requirement: optimization 화면은 근거와 변경 수명주기를 표시한다
콘솔은 parameter registry, current/effective version, lane performance evidence, candidate diff, apply history와 rollback 동작을 표시해야 한다 (SHALL).

#### Scenario: 추천 불가
- **WHEN** 필수 성과 근거가 부족하다
- **THEN** 화면은 구체적인 누락 사유를 표시하고 apply control을 활성화하지 않는다

#### Scenario: 적용 승인
- **WHEN** 운영자가 preview와 version을 확인하고 apply한다
- **THEN** 결과 version과 restart 필요 여부를 표시하며 LIVE toggle은 별도 상태로 남는다

### Requirement: optimization 설정은 카테고리와 설명으로 탐색된다
콘솔은 `overview`, `exit-protection`, `position-management`, `candidate-filters`, `strategy-runtime`, `performance-history` 여섯 category를 고정 순서로 제공해야 하며 (SHALL), 모든 설정에 한국어 설명, parameter key, 단위, registry 기본값, desired/effective 값, 범위와 적용 시점을 표시해야 한다 (SHALL).

#### Scenario: 기본값과 현재값 구분
- **WHEN** 운영자가 설정 가능한 field를 연다
- **THEN** placeholder가 아닌 별도 label로 기본값·현재 desired·현재 effective와 적용 시점을 구분해 표시한다

#### Scenario: 모바일 category 탐색
- **WHEN** 360px viewport에서 optimization을 연다
- **THEN** 동일한 여섯 category와 deep link를 사용할 수 있고 페이지 전체의 수평 overflow 없이 설정과 설명을 읽을 수 있다

#### Scenario: category-scoped save
- **WHEN** 두 category에 미저장 draft가 있고 한 category에서 저장한다
- **THEN** 해당 category changed subset만 preview/apply하며 다른 draft와 LIVE 상태를 변경하지 않는다

### Requirement: 위험 설정은 before/after 확인을 요구한다
손절폭 확대, 보호 약화, lane 또는 LIVE 권한 변화는 일반 설정 저장과 구분해야 하며 (SHALL), before/after·적용 대상·restart 여부를 표시한 명시적 확인 없이는 적용해서는 안 된다 (MUST NOT).

콘솔은 StockOS lane-console의 화면 단위 navigation·partial save·effective mismatch 패턴을 따라야
하며 (SHALL), 운영자에게 자유 텍스트·숫자·symbol·확인 문구 입력을 요구해서는 안 된다 (MUST NOT).
모든 변경은 preset/radio/select/chip/toggle/discrete step과 server-defined reason code로 수행해야 한다
(SHALL).

#### Scenario: LIVE 보호 약화
- **WHEN** LIVE 상태에서 draft가 허용 손실폭이나 유예를 확대한다
- **THEN** 3초 대기, 확인 checkbox와 별도 위험 승인 전에는 저장 control이 활성화되지 않는다

#### Scenario: 입력 없는 위험 확인
- **WHEN** 위험 확대 candidate를 확인한다
- **THEN** 3초 대기, 영향 범위 확인 checkbox와 승인 button을 제공하되 typed phrase나 자유 reason 입력은 요구하지 않는다

#### Scenario: 자유 입력 control 회귀
- **WHEN** optimization HTML과 handler contract를 검사한다
- **THEN** text, textarea, number, contenteditable control이 0개이고 제출값은 registry option ID에 한정된다
