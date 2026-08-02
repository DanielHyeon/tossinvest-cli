## ADDED Requirements

### Requirement: positions는 모든 시장에서 기준선 근거 상태를 직접 설명한다
`/positions`는 시장과 무관하게 실효 snapshot, 저장 원장 근거, 미생성 상태를 서로 다른 라벨로 표시해야 한다 (SHALL). 저장 원장 근거는 표의 주 정보로 읽을 수 있어야 하지만 actionable `ExitLine` 가격으로 복사되어서는 안 된다 (MUST NOT).

#### Scenario: KR legacy managed position
- **WHEN** KR 관리 포지션에 baseline과 initial stop은 있으나 canonical effective snapshot이 없다
- **THEN** 표는 `원장 기준선 · 실효 미확인`, 저장 baseline과 최초 손절을 접지 않은 상태에서 표시하고 current effective/next target을 만들지 않는다

#### Scenario: US legacy managed position
- **WHEN** US 관리 포지션에 같은 legacy 원장 근거가 있다
- **THEN** KR과 동일한 증거 상태와 필드를 표시하며 시장별로 숨기거나 다른 가격을 계산하지 않는다

#### Scenario: stale canonical snapshot
- **WHEN** canonical snapshot이 stale이다
- **THEN** 기존 `오래된 평가`와 사유를 유지하고 actionable 및 raw 가격을 표의 주 기준선으로 표시하지 않는다

#### Scenario: 다른 lifecycle generation의 저장 근거
- **WHEN** 현재 position-policy generation과 exit state 또는 snapshot generation이 다르다
- **THEN** 과거 가격과 snapshot identity를 모두 숨기고 세대 불일치 사유를 표시한다

#### Scenario: 손상되었거나 검증되지 않은 저장 근거
- **WHEN** snapshot 상태가 partial/invalid/corrupt이거나 시도된 lifecycle lookup이 현재 generation을 확인하지 못한다
- **THEN** 저장 가격을 숨기고 손상 또는 관리 세대 확인 불가 사유를 표시한다

### Requirement: 미편입 후보는 기준선이 아직 없는 이유와 정책 폭을 표시한다
KR 또는 US 포지션이 `ADOPTION_PENDING` 또는 `RECONCILE_BLOCKED`이고 exit state가 없으면 `/positions`는 `기준선 미생성`과 편입 후 생성된다는 설명을 표시해야 한다 (SHALL). running effective adoption 설정을 아는 경우 최초 손절폭을 percentage로 표시해야 하며 (SHALL), broker average/current price에서 보호 가격을 계산해서는 안 된다 (MUST NOT).

#### Scenario: reconcile-blocked US candidate
- **WHEN** include된 US 보유분은 effective initial stop 3%를 사용하지만 account reconcile block 때문에 아직 편입되지 않았다
- **THEN** 행은 `기준선 미생성`, `편입 후 확정`, `effective 최초 손절폭 3%`를 표시하고 숫자 가격선은 만들지 않는다

#### Scenario: pending KR candidate
- **WHEN** KR 보유분이 편입 후보이고 runtime effective 설정을 읽을 수 있다
- **THEN** US와 동일한 미생성 설명과 effective percentage를 표시한다

#### Scenario: runtime unavailable
- **WHEN** desired include에 지정된 KR 또는 US 보유분이 있지만 candidate 여부나 effective stop percentage를 증명할 runtime commander를 읽지 못한다
- **THEN** 화면은 편입 요청 저장과 실행 상태 미확인을 구분하고, desired/default 값을 대신 사용하지 않으며 기준선/percentage를 `알 수 없음`으로 유지한다

#### Scenario: managed 또는 released 상태와 desired include가 충돌한다
- **WHEN** 이미 엔진이 관리 중이거나 운영자가 해제한 종목이 desired include 목록에도 남아 있다
- **THEN** 현재 lifecycle 상태가 편입 예약보다 우선하며 해당 행을 `편입 예약됨` 또는 runtime-unknown candidate로 표시하지 않는다

### Requirement: 기준선 복원 화면은 입력과 mutation을 추가하지 않는다
`/positions`는 form, visible input, button, contenteditable 또는 reconcile/order mutation route를 추가해서는 안 된다 (MUST NOT).

#### Scenario: read-only responsive view
- **WHEN** 375px viewport와 정적 route 계약을 검사한다
- **THEN** 세 증거 상태는 읽을 수 있고 visible input이나 POST capability는 없다
