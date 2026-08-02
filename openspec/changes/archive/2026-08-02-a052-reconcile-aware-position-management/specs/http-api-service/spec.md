## ADDED Requirements

### Requirement: positions와 optimization API는 웹과 같은 adoption/reconcile 사실을 사용한다
`/api/v1/positions`와 `/api/v1/optimization`은 웹 position-management와 동일한 registry default, config desired, running-engine effective, candidate와 adoption-blocking tracker projector를 사용해야 한다 (SHALL). block DTO는 모든 journal cause가 아니라 adoption driver와 같은 runtime projection임을 source로 명시해야 한다 (SHALL). read API는 reconcile resolution capability 또는 mutation route를 노출해서는 안 된다 (MUST NOT).

positions item은 stable `adoptionStatus` enum(`UNKNOWN`, `MANAGED`, `EXCLUDED`, `RECONCILE_BLOCKED`, `ADOPTION_PENDING`, `UNMANAGED`), `statusKnown`, `adoptionLabel`, typed/sanitized `adoptionReason`, `included`, `excluded`, `candidate`, `designationKnown`과 nullable covering block(scope/market/symbol/reason/startedAt)을 반환해야 한다 (SHALL). optimization position-management는 desired/effective adoption blocks와 `effectiveKnown`을 반환해야 한다 (SHALL).

#### Scenario: 미국 include 보유분이 영구 차단으로 대기한다
- **WHEN** 미국 보유분이 include됐고 account-wide permanent quantity-mismatch block이 active다
- **THEN** positions API는 eligible false, candidate true, adoptionStatus `RECONCILE_BLOCKED` 및 sanitized block reason을 반환하고 optimization API는 actual desired/effective adoption 값을 반환한다

#### Scenario: managed와 exclude가 함께 있다
- **WHEN** managed position의 symbol이 exclude에도 있다
- **THEN** positions API의 adoptionStatus는 `MANAGED`이고 included/excluded 사실은 별도 boolean으로 보존된다

#### Scenario: released lifecycle은 raw adoption eligibility보다 우선한다
- **WHEN** journal position에는 adoption ID가 남아 있지만 authoritative lifecycle은 `RELEASED`다
- **THEN** positions API는 `UNMANAGED`, `OPERATOR_RELEASED`를 반환하고 `MANAGED`, `ADOPTION_PENDING`, `RECONCILE_BLOCKED`로 오표시하지 않는다

#### Scenario: API sidecar는 별도 network namespace에서 runtime을 읽는다
- **WHEN** console/engine과 HTTP API가 Compose의 서로 다른 network namespace에서 같은 private engine directory를 mount한다
- **THEN** HTTP API는 command loopback이 아니라 authenticated runtime-only Unix endpoint로 effective와 block projection을 읽고 Preview/Apply 권한을 얻지 않는다

#### Scenario: engine runtime을 읽지 못한다
- **WHEN** config desired는 읽히지만 engine control plane이 unavailable이다
- **THEN** optimization API는 effectiveKnown false를 반환하고 registry 기본값을 effective로 위장하지 않으며, non-managed positions item은 statusKnown/designationKnown false와 typed runtime-unavailable reason을 반환한다

#### Scenario: read API의 mutation 표면은 그대로다
- **WHEN** a052가 배포된다
- **THEN** HTTP API mutation allowlist에는 기존 optimization preview/application만 남고 reconcile 해제 endpoint는 없다

#### Scenario: raw exit evidence와 effective line
- **WHEN** legacy exit state에 raw t0/baseline은 있지만 canonical effective snapshot이 없다
- **THEN** positions API의 exitLine actionable 가격은 unknown을 유지하고 storedExitEvidence는 raw 값과 effectiveKnown false를 반환한다
