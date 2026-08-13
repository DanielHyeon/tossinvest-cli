# protection-orders — a107 delta

> a100 design D1이 프로덕션 보호 core를 `protectionlifecycle`로 확정했다. 이 delta는 그
> 결정의 잔여물을 치운다: 자체 SQLite를 요구하는 두 번째 core(`protection.Controller` +
> `Repository`)는 봉인된 채 남아 있을 이유가 없다. 봉인은 유지 비용을 내고, 죽은 core는
> 배선 사고의 상시 후보다.

## ADDED Requirements

### Requirement: 프로덕션 보호 core는 하나다

저장소는 브로커 상주 보호의 상태 전이 core를 하나만 가져야 하며(SHALL), 그 상태 영속은
trading journal 하나여야 한다(SHALL). 자체 영속 저장소를 요구하는 대체 보호 core를 저장소에
보유해서는 안 된다(SHALL NOT).

#### Scenario: 보호 경로 조립

- **WHEN** 보호 경로가 프로덕션에 조립되면
- **THEN** 상태 전이는 `protectionlifecycle`이고 영속은 trading journal이며, 보호를 위한
  두 번째 데이터베이스가 기동하지 않는다

#### Scenario: 두 번째 core의 부활 시도

- **WHEN** 자체 영속 저장소를 갖는 보호 상태 core가 저장소에 추가되면
- **THEN** 정적 가드가 이를 거부한다
