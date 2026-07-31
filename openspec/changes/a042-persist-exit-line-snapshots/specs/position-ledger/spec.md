## ADDED Requirements

### Requirement: exit snapshot 핵심 상태는 원자적으로 영속된다
journal은 policy ID/version/digest, snapshot/decision/observation ID, baseline, high-water, active rung, next target/protection, last observation source·time을 포지션 generation에 귀속해 저장해야 한다 (SHALL).

#### Scenario: 평가 commit 뒤 재시작
- **WHEN** 보호선 승격 transaction이 commit된 직후 프로세스가 종료된다
- **THEN** 재시작은 승격된 기준선과 동일한 policy/rung/high-water를 회수한다

#### Scenario: 평가 commit 전 crash
- **WHEN** snapshot transaction이 commit되기 전에 프로세스가 종료된다
- **THEN** journal은 부분 필드가 섞인 snapshot을 노출하지 않고 이전 완전 상태를 유지한다

### Requirement: migration은 additive이고 legacy를 추정하지 않는다
새 snapshot column은 nullable additive migration이어야 하며 (SHALL), legacy NULL을 0 또는 임의 정책값으로 backfill해서는 안 된다 (MUST NOT).

#### Scenario: legacy row 조회
- **WHEN** 새 column이 NULL인 기존 exit row를 읽는다
- **THEN** 결정적으로 복원 가능한 값만 파생하고 나머지는 unknown으로 표시한다

### Requirement: 화면용 snapshot은 저장값과 실효값의 근거를 구분한다
journal read model은 저장값, 현재 재계산값과 최종 실효값을 구분하고 각 값의 policy version, source, observed-at과 stale/unknown reason을 제공해야 한다 (SHALL).

#### Scenario: 저장값이 더 안전함
- **WHEN** 저장 protection이 현재 재계산 후보보다 높다
- **THEN** read model은 두 값을 모두 보존하고 effective source가 saved-monotone임을 표시한다

#### Scenario: 근거가 불완전함
- **WHEN** legacy row에서 다음 target을 결정적으로 복원할 수 없다
- **THEN** read model은 제품 기본값이나 0을 넣지 않고 unknown reason을 반환한다
