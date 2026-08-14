## ADDED Requirements

### Requirement: positions API는 flat 관측의 canonical freshness를 console과 공유한다
`GET /api/v1/positions`는 console과 같은 persisted effective snapshot 및 shared freshness 판정을 사용해야 한다 (SHALL): engine stopped가 확정되면 즉시 stale이고, running·unavailable·unwired에서는 integrity와 per-position 30초 age bound를 함께 적용한다. 성공한 flat observation refresh가 계속되는 managed position의 `exitLine`은 fresh/actionable이어야 하며 (SHALL), API adapter가 별도 시계나 raw seed 가격으로 line을 재계산해서는 안 된다 (MUST NOT).

#### Scenario: flat quote가 계속 관측되는 managed position
- **WHEN** position의 가격과 policy state는 변하지 않지만 마지막 성공 refresh가 freshness bound 안이다
- **THEN** API는 console과 같은 current protection, next target, next protection, freshness status와 evaluated-at을 반환한다

#### Scenario: seed에서 첫 flat 평가가 완료된다
- **WHEN** `SEED` position이 adoption t0와 같은 첫 공식 quote로 canonical evaluation을 영속한다
- **THEN** API는 `not_evaluated_yet`을 해제하고 완전한 actionable exitLine을 반환한다

#### Scenario: invalid quote만 도착한다
- **WHEN** 마지막 성공 evaluation 뒤 invalid 또는 stale quote만 도착해 persisted freshness가 age bound를 넘는다
- **THEN** API는 actionable 가격을 숨기고 console과 같은 typed stale/unknown reason을 반환한다

#### Scenario: console과 API의 30초 경계
- **WHEN** running·unavailable·unwired인 동일 snapshot을 29.999초, 정확히 30초와 30초 초과 시각에 두 adapter로 projection한다
- **THEN** 두 surface는 앞의 두 시각에 fresh, 초과 시각에만 stale이며 line visibility와 typed reason이 동일하다

#### Scenario: engine running과 stopped 판정
- **WHEN** 같은 snapshot과 runtime을 두 adapter가 running 또는 stopped로 확정한다
- **THEN** running도 30초 age bound를 적용하고 stopped는 둘 다 즉시 `engine_not_running`이다

#### Scenario: API blocking read 중 freshness 경계를 지난다
- **WHEN** 실제 positions 요청의 cache·journal·policy·runtime 또는 단일 engine-marker read가 진행되는 동안 snapshot age가 30초를 넘거나 marker가 stopped 경계를 지난다
- **THEN** API는 모든 blocking read 뒤의 한 response clock으로 판정해 즉시 stale/dash를 반환하며 broker cache를 다시 읽지 않는다

#### Scenario: stopped marker 뒤 wall clock이 rollback한다
- **WHEN** marker read는 engine을 stopped로 판정했지만 그 직후 response clock이 뒤로 움직여 marker refresh time이 다시 bound 안처럼 보인다
- **THEN** API는 pre-read stopped 판정을 running으로 승격하지 않고 `engine_not_running`과 dash를 유지한다

#### Scenario: invalid sibling은 running liveness를 빌리지 않는다
- **WHEN** running engine의 batch에서 다른 symbol은 유효하지만 이 position의 quote가 계속 invalid/missing이다
- **THEN** API와 console은 이 position의 마지막 성공 관측만 aging하여 30초 초과 뒤 같은 stale verdict로 숨긴다

### Requirement: positions API의 flat refresh는 read-only history 의미를 보존한다
API가 반환하는 최신 observed-at은 durable snapshot refresh 근거여야 하며 (SHALL), 의미가 동일한 refresh를 새 exit event, proposal 또는 order로 노출해서는 안 된다 (MUST NOT).

#### Scenario: 여러 flat refresh 뒤 API 조회
- **WHEN** 한 position이 동일 line으로 여러 번 refresh된 뒤 positions와 exit history를 조회한다
- **THEN** positions는 최신 observed-at을 반환하고 history는 첫 평가 이후 의미 없는 중복 event를 포함하지 않는다
