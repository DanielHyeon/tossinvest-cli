## ADDED Requirements

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
