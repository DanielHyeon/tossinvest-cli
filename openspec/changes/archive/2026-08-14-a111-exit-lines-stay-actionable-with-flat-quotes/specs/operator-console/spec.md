## ADDED Requirements

### Requirement: 지속 관측되는 flat managed position은 actionable exit line을 유지한다
콘솔의 `/positions`와 `/position-management`는 shared freshness 판정을 사용해야 한다 (SHALL): engine stopped가 확정되면 즉시 stale이며, running·unavailable·unwired에서는 canonical snapshot integrity와 마지막 성공 관측의 30초 age bound를 함께 적용한다. 유효한 flat 관측이 계속 영속되면 current protection, next target과 next protection을 계속 actionable하게 표시해야 하며 (SHALL), `SEED`, corrupt, generation mismatch 또는 실제 stale 증거를 raw 값으로 보충해서는 안 된다 (MUST NOT).

#### Scenario: unchanged first quote 뒤 관리 화면
- **WHEN** 새 관리 포지션이 t0와 같은 첫 공식 가격으로 `EVALUATED` snapshot을 얻는다
- **THEN** 두 console 화면은 `not_evaluated_yet` 대신 canonical current/next line과 evaluated-at을 표시한다

#### Scenario: 30초 이상 가격이 움직이지 않는다
- **WHEN** 가격과 policy state는 30초 이상 같지만 성공한 공식 관측 refresh가 age bound 안에서 계속 영속된다
- **THEN** 두 console 화면은 `오래된 평가`로 강등하지 않고 최신 canonical line을 표시한다

#### Scenario: 실제로 관측이 끊긴다
- **WHEN** engine liveness와 무관하게 한 position의 마지막 성공 snapshot 관측이 30초를 넘으며 그 뒤 성공한 refresh가 없다
- **THEN** 두 console 화면은 actionable 가격을 `—`로 숨기고 typed stale 사유를 표시한다

#### Scenario: age 경계와 engine liveness
- **WHEN** running·unavailable·unwired snapshot을 29.999초, 정확히 30초, 30초 초과에서 읽거나 engine을 stopped로 확정해 읽는다
- **THEN** 앞의 두 age는 fresh이고 초과만 stale이며, stopped는 age와 무관하게 즉시 `engine_not_running`이다

#### Scenario: console blocking read 중 freshness 경계를 지난다
- **WHEN** 실제 `/position-management` 요청의 journal·runtime·quarantine 또는 단일 engine-marker read가 진행되는 동안 snapshot age가 30초를 넘거나 marker가 stopped 경계를 지난다
- **THEN** console은 모든 blocking read 뒤의 한 response clock으로 판정해 즉시 stale 사유와 dash를 표시하고 추가 marker read를 만들지 않는다

#### Scenario: stopped marker 뒤 wall clock이 rollback한다
- **WHEN** marker read는 engine을 stopped로 판정했지만 그 직후 response clock이 뒤로 움직여 marker가 다시 fresh처럼 보인다
- **THEN** console은 stopped를 running으로 승격하지 않고 `engine_not_running`과 dash를 유지한다

#### Scenario: running engine에서 한 symbol만 invalid다
- **WHEN** engine은 running이고 같은 batch의 valid sibling은 계속 평가되지만 이 position의 quote만 invalid/missing이다
- **THEN** 이 position의 timestamp는 전진하지 않아 30초 초과 뒤 stale로 숨겨지며 sibling liveness가 대신 freshness를 만들지 않는다

#### Scenario: 저장 snapshot이 손상됐다
- **WHEN** observed-at, identity 또는 JSON/flattened tuple 무결성 검증이 실패한다
- **THEN** engine runtime이 running이어도 화면은 freshness를 추정하지 않고 unknown/corrupt로 fail-closed한다

### Requirement: flat refresh는 운영 event나 주문으로 표시되지 않는다
콘솔은 의미가 동일한 observation refresh를 새 exit transition, proposal, intent 또는 broker order로 표시해서는 안 된다 (MUST NOT). 기존 exit history는 실제 first evaluation, state/action transition과 arming decision만 유지해야 한다 (SHALL).

#### Scenario: 동일 가격 refresh가 여러 번 영속된다
- **WHEN** 한 evaluated position이 동일한 line으로 여러 번 refresh된다
- **THEN** 현재 line의 evaluated-at은 전진하지만 exit event와 order history 개수는 늘지 않는다
