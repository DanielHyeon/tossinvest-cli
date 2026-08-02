## ADDED Requirements

### Requirement: 포지션 관리는 실제 adoption desired/effective를 구분한다
`/position-management`는 registry 기본값, config file의 desired adoption 값, running engine이 시작 때 로드한 effective 값을 별도 label로 표시해야 한다 (SHALL). engine runtime을 읽지 못한 경우 effective를 기본값이나 desired로 대체해서는 안 된다 (MUST NOT).

#### Scenario: 저장 설정과 실행 설정이 다르다
- **WHEN** config desired는 ON·3%이고 running engine effective는 OFF·5%다
- **THEN** 화면은 두 값을 각각 표시하고 어느 하나를 다른 값으로 덮지 않는다

#### Scenario: engine control plane을 읽지 못한다
- **WHEN** config desired는 읽히지만 running engine runtime은 unavailable이다
- **THEN** desired는 표시하고 effective는 `알 수 없음`으로 표시한다

### Requirement: 편입 보조 상태는 candidate와 reconcile 차단을 함께 설명한다
`/positions`와 `/position-management`는 기존 관리 판정 라벨 옆에 stable adoption status를 표시해야 한다 (SHALL). projector는 running effective settings가 known일 때 `candidate=(globalEnabled || included) && !excluded`를 계산하되 `journal unknown > already managed > operator released > excluded > candidate and covering reconcile block > candidate pending > unselected` 순서로 평가해야 한다 (SHALL). runtime unavailable인 non-managed 행은 desired를 effective로 위장하지 않고 `UNKNOWN`과 runtime unavailable 이유를 표시해야 한다 (SHALL). 미국 시장이라는 이유만으로 `편입 불가`라고 단정해서는 안 된다 (MUST NOT).

#### Scenario: include된 미국 보유분이 account-wide 차단을 만난다
- **WHEN** 미국 보유분에 adoption include가 있고 entry/adoption record는 아직 없으며 account-wide quantity-mismatch block이 active다
- **THEN** 두 화면은 기존 `관리 편입` 판정과 `대사 차단으로 대기` 보조 상태, block 사유를 표시하고 미국 시장 미지원으로 설명하지 않는다

#### Scenario: managed와 exclude가 함께 있다
- **WHEN** 이미 entry/adoption 근거로 managed인 symbol이 장래 candidate exclude에도 있다
- **THEN** 현재 보호 상태는 `MANAGED`로 유지되고 exclude가 기존 관리 lifecycle을 해제한 것으로 표시되지 않는다

#### Scenario: 미지정 행과 account block이 함께 있다
- **WHEN** global adoption이 OFF이고 include되지 않은 미편입 symbol에 account-wide block이 존재한다
- **THEN** 행은 `UNMANAGED`이며 candidate가 아니므로 `RECONCILE_BLOCKED`로 표시되지 않는다

#### Scenario: include와 exclude가 함께 있다
- **WHEN** 미편입 symbol이 include와 exclude에 모두 있다
- **THEN** candidate는 false이고 `EXCLUDED`가 표시된다

#### Scenario: runtime은 unavailable이고 desired만 include다
- **WHEN** journal은 readable이고 미편입 symbol이 desired config include에 있으나 running engine runtime을 읽지 못한다
- **THEN** desired include 사실은 설정 요약에 보존되지만 행의 effective status는 `UNKNOWN`이며 `ADOPTION_PENDING`으로 승격되지 않는다

#### Scenario: 운영자가 external lifecycle을 release했다
- **WHEN** adoption ID는 남아 있지만 authoritative position-policy lifecycle이 `RELEASED`다
- **THEN** 두 화면은 `UNMANAGED`, `OPERATOR_RELEASED`, `관리 외(운영자 해제)`를 표시하고 candidate 또는 account block보다 release를 우선한다

### Requirement: 저장 exit 근거와 실효 보호선을 구분한다
canonical persisted effective snapshot이 없는 exit state는 현재 실효 보호선이나 다음 익절 가격을 만들어 내서는 안 된다 (MUST NOT). 다만 journal에 저장된 t0 entry, initial stop, baseline과 high-water는 `원장 기록 · 실효 미확인` 증거로 별도 표시해야 하며 (SHALL), actionable effective line과 동일한 필드/라벨로 표시해서는 안 된다 (MUST NOT).

#### Scenario: legacy seed-only exit state
- **WHEN** exit state에 entry/initial-stop/baseline은 있으나 canonical effective snapshot이 없다
- **THEN** 화면은 current protection과 next target을 `—`로 유지하고 저장된 가격들을 `원장 기록 · 실효 미확인` 상세로 표시한다

#### Scenario: canonical effective snapshot이 있다
- **WHEN** exit state에 유효한 canonical effective snapshot이 있다
- **THEN** 기존 operatorview freshness와 effective-source 판정을 그대로 사용해 실효 보호선과 다음 익절을 표시한다

### Requirement: a052 운영 상태 표면은 읽기 전용이다
a052는 reconcile preview/apply route, capability, free-text field 또는 journal mutation을 console에 추가해서는 안 된다 (MUST NOT). 기존 position-policy lifecycle mutation surface와 a052 runtime read endpoint는 별도 권한으로 유지해야 한다 (SHALL). Compose sidecar용 shared Unix endpoint는 authenticated GET runtime만 제공하고 lifecycle Preview/Apply를 제공해서는 안 된다 (MUST NOT).

#### Scenario: 정적 route와 HTML 검사
- **WHEN** a052 route table과 `/positions`, `/position-management` HTML을 검사한다
- **THEN** reconcile resolution POST route와 text/textarea/number/contenteditable 입력이 없고 shared runtime surface에는 authenticated GET 외의 command가 없다
