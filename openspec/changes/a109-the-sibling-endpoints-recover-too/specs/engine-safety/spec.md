# engine-safety — a109 delta

> a108이 strategy projection에 세운 불변식의 일반형이다. a108 A1 리뷰가 좁힌 요구
> ("거짓 SHALL을 정본에 넣지 않는다")를, 세 형제 endpoint를 실제로 고치는 이 change가
> 일반 요구로 되돌린다.

## ADDED Requirements

### Requirement: 엔진이 소유한 모든 런타임 endpoint는 자기 잔재에서 회복한다

엔진이 소유한 모든 런타임 endpoint(position policy command·position policy runtime·alert control 포함)의 기동 시 잔재 회수는 자기 생성·종료·회수 시퀀스가 만들 수 있는 모든 부분 상태를 소유자 사망 검증 후 사람 개입 없이 회수해야 하며(SHALL), 산출물 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 살아 있는 주인의 endpoint 위에 두 번째 서버가 올라서서는 안 된다(SHALL NOT — 또는 journal flock이 그 유일한 방어임을 코드와 테스트가 명시해야 한다).

#### Scenario: pre-chmod socket 잔재에서의 재기동

- **WHEN** listen과 chmod 사이에 죽어 group/other 비트 없는 비-0600 socket이 남은
  상태에서 엔진이 기동하면
- **THEN** 세 endpoint 모두 잔재를 회수하고 기동을 계속한다

#### Scenario: 산 주인의 endpoint는 탈취되지 않는다

- **WHEN** 살아 있는 주인이 수락 중인 socket 위에서 두 번째 기동이 시도되면
- **THEN** 두 번째 기동은 그 endpoint를 넘겨받지 못한다
