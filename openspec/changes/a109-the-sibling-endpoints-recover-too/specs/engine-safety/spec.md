# engine-safety — a109 delta

> a108이 strategy projection에 세운 불변식의 일반형이다. a108 A1 리뷰가 좁힌 요구
> ("거짓 SHALL을 정본에 넣지 않는다")를, 세 형제 endpoint를 실제로 고치는 이 change가
> 일반 요구로 되돌린다.

## ADDED Requirements

### Requirement: 엔진이 소유한 모든 런타임 endpoint는 자기 잔재에서 회복한다

엔진이 소유한 모든 런타임 endpoint(position policy command·position policy runtime·alert control 포함)의 기동 시 잔재 회수는 자기 생성·종료·회수 시퀀스가 만들 수 있는 모든 부분 상태(발행 전 staging 잔재와 구버전 staging 잔재 포함)를 소유자 사망 검증(connect probe — PID 불사용) 후 사람 개입 없이 회수해야 하며(SHALL), socket 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 수락 중인 socket 위에 두 번째 서버가 올라서서는 안 된다(SHALL NOT).

#### Scenario: pre-chmod socket 잔재에서의 재기동

- **WHEN** listen과 chmod 사이에 죽어 group/other 비트 없는 비-0600 socket이 남은
  상태에서 엔진이 기동하면
- **THEN** 두 socket endpoint 모두 잔재를 회수하고 기동을 계속한다

#### Scenario: 산 주인의 endpoint는 탈취되지 않는다

- **WHEN** 살아 있는 주인이 수락 중인 socket 위에서 두 번째 기동이 시도되면
- **THEN** 두 번째 기동은 그 socket을 unlink하지 않고 거부된다

#### Scenario: staging 잔재는 우리 잔재다

- **WHEN** 발행 전 임시 이름(신규 staging 또는 구버전 CreateTemp 이름)의 정규 파일·
  socket만 남은 상태에서 엔진이 기동하면
- **THEN** 잔재를 회수하고 기동을 계속하며 회수 후 control 디렉터리에 잔재가 없다

#### Scenario: 낯선 엔트리는 건드리지 않는다

- **WHEN** control 디렉터리에 이 endpoint가 만들 수 없는 이름 또는 모양의 엔트리가
  있는 상태에서 기동하면
- **THEN** 회수는 아무것도 제거하지 않고 그 endpoint의 기동을 거부한다

### Requirement: 엔진 기동은 자기 endpoint 표면의 실패로 죽지 않는다

position policy command·position policy runtime·alert control endpoint의 기동 실패는 엔진 부팅을 실패시키지 않고 해당 표면 없이 계속해야 하며(SHALL), 그 강등은 stderr 안내와 obs 이벤트로 보고하되(SHALL) 그 보고가 critical 등급·obs 등급표 등재·원장 outbox 적재 중 어느 것도 사용해서는 안 된다(SHALL NOT — 미전달 outbox 행은 다음 부팅의 진입 게이트를 잠근다, a108 D3-2).

#### Scenario: endpoint 하나의 실패가 보호 루프를 세우지 않는다

- **WHEN** 세 endpoint 중 어느 하나의 Start가 어떤 이유로든 실패한 채 엔진이 기동하면
- **THEN** 엔진은 그 표면 없이 부팅을 완료하고 손절·청산 루프는 정상 가동한다

#### Scenario: 강등 보고는 진입 게이트를 잠그지 않는다

- **WHEN** 강등 보고가 발행된 뒤 엔진이 재시작하면
- **THEN** 그 보고로 인해 잠긴 진입 게이트가 없다
