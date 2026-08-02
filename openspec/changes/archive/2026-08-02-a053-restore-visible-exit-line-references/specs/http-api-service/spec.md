## ADDED Requirements

### Requirement: positions API는 비실효 기준선 참조를 별도 계약으로 제공한다
`GET /api/v1/positions`는 actionable `exitLine`과 별도로 nullable `exitLineReference`를 제공해야 한다 (SHALL). reference는 `LEGACY_RAW`, `ADOPTION_PLAN` 또는 generation/runtime/lifecycle unknown 상태를 typed kind로 표시하고 항상 `effectiveKnown=false`여야 한다 (SHALL). lifecycle generation이 다르거나 lifecycle을 검증할 수 없으면 이전 가격이나 identity를 반환해서는 안 된다 (MUST NOT).

#### Scenario: legacy raw evidence
- **WHEN** KR 또는 US 포지션에 same-generation raw exit state만 있다
- **THEN** `exitLine`의 current/next 가격은 `—`이고 `exitLineReference`와 호환 `storedExitEvidence`에 non-effective 원장 근거가 반환된다

#### Scenario: US adoption plan
- **WHEN** US candidate가 pending 또는 reconcile-blocked이고 running effective stop percentage가 알려져 있다
- **THEN** `exitLineReference.kind=ADOPTION_PLAN`, stop percentage와 가격 미확정 설명을 반환하며 계산된 가격은 반환하지 않는다

#### Scenario: generation mismatch
- **WHEN** current lifecycle generation이 stored exit generation과 다르다
- **THEN** `exitLine`, `storedExitEvidence`, `exitLineReference` 어디에도 과거 가격이나 snapshot identity가 없다

#### Scenario: corrupt 또는 lifecycle-unverified evidence
- **WHEN** raw exit tuple의 snapshot 상태가 partial/invalid/corrupt이거나 요구된 lifecycle lookup이 현재 generation을 증명하지 못한다
- **THEN** API는 raw 가격을 반환하지 않고 typed unknown reason만 제공한다
