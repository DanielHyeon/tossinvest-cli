# position-exit-policy-management — a069 delta

> **ADDED만 쓴다.** 기존 "release와 re-adopt는 lifecycle을 분리한다"는 그대로 참이고,
> a063이 정의하는 판정 격리 해제는 **lifecycle을 아예 건드리지 않는** 세 번째 동작이다.
> 기존 두 요구사항의 SHALL과 충돌하지 않는다.

## ADDED Requirements

### Requirement: 판정 격리 해제는 lifecycle을 바꾸지 않는다

시스템은 exit snapshot 격리의 해제를 운영자 승인으로만 수행해야 하며 (SHALL), 그 해제는 adoption generation·진입가·초기 손절·high-water·rung을 바꿔서는 안 된다 (MUST NOT).
해제는 운영자에게 제시된 바로 그 quarantine version에 대한 compare-and-swap이어야 하고 (SHALL), version이 어긋나면 아무것도 쓰지 않고 거부해야 한다 (SHALL). 해제 기록에는 주체·사유·증거가 남아야 하며 (SHALL), 그 증거 문자열은 서버가 조립해야 한다 (SHALL — 운영자에게 확인 문구를 타이핑시켜서는 안 된다(SHALL NOT)).

해제는 판정을 **다시 시도하게 할 뿐** 판정을 느슨하게 해서는 안 된다 (MUST NOT). 해제 뒤 다음 관측에서 같은 격리 조건이 다시 성립하면 시스템은 그 포지션을 다시 격리해야 한다 (SHALL). 사람이 아닌 어떤 경로도 격리를 자동으로 해제해서는 안 된다 (MUST NOT).

#### Scenario: 정상 해제
- **WHEN** 운영자가 격리된 포지션의 현재 quarantine version을 담은 preview를 3초 지연 뒤 확인과 함께 승인한다
- **THEN** 해당 격리 행만 released 표시되고, 주체·사유·서버가 조립한 증거가 함께 기록되며, 저장된 exit snapshot과 adoption generation은 그대로다

#### Scenario: 다음 관측에서 판정 재개
- **WHEN** 격리가 해제된 포지션이 다음 exit 관측 주기에 들어온다
- **THEN** 그 포지션은 판정 대상 집합에 돌아오고 저장된 기준선으로 손절 평가가 재개된다

#### Scenario: 원인이 남아 있는 해제
- **WHEN** 해제된 포지션의 다음 판정이 여전히 하나의 검증된 후보를 고를 수 없다
- **THEN** 새 quarantine version으로 다시 격리되고 판정은 다시 거부된다

#### Scenario: stale quarantine version
- **WHEN** 운영자가 본 뒤 다른 write가 있었고 오래된 quarantine version으로 해제를 적용한다
- **THEN** 거부하고 격리 상태와 저장된 snapshot을 바꾸지 않는다

#### Scenario: 확인 없는 적용
- **WHEN** 위험 확인 없이 또는 3초 지연 전에 해제를 적용한다
- **THEN** 거부하고 아무것도 변경하지 않으며 승인 권한을 소비하지 않는다

#### Scenario: 격리가 없는 포지션
- **WHEN** 활성 격리가 없는 포지션에 해제를 요청한다
- **THEN** 거부하고 아무것도 쓰지 않는다

#### Scenario: 기준선 보존
- **WHEN** 격리 해제가 적용된다
- **THEN** exit state의 진입가·초기 손절·high-water·활성 rung과 adoption generation이 해제 전과 동일하다
