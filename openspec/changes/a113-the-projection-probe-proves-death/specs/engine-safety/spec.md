# engine-safety — a113 delta

> a109가 형제 endpoint에 세운 사망 증명(chmod-then-probe)을 요구 문면으로 명시해,
> a108 원형(`projectionSocketAccepts`)에 남은 owner-write 추정(a109 issues I1)을
> 계약 위반으로 만든다.

## MODIFIED Requirements

### Requirement: 엔진이 소유한 모든 런타임 endpoint는 자기 잔재에서 회복한다

엔진이 소유한 모든 런타임 endpoint(position policy command·position policy runtime·alert control 포함)의 기동 시 잔재 회수는 자기 생성·종료·회수 시퀀스가 control 디렉터리 안 파일에 만들 수 있는 모든 부분 상태(descriptor·socket·현행 및 구버전 staging 잔재)를 소유자 사망 검증(connect probe — PID 불사용) 후 사람 개입 없이 회수해야 하며(SHALL), 그 사망 검증은 관측된 권한 비트로 사망을 추정해서는 안 되고(SHALL NOT) 접속 가능 권한 복원 후 probe로 증명해야 하며(SHALL), socket 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 수락 중인 socket 위에 두 번째 서버가 올라서서는 안 된다(SHALL NOT).

#### Scenario: pre-chmod socket 잔재에서의 재기동

- **WHEN** listen과 chmod 사이에 죽어 group/other 비트 없는 비-0600 socket이 남은
  상태에서 엔진이 기동하면
- **THEN** 두 socket endpoint 모두 잔재를 회수하고 기동을 계속한다

#### Scenario: 산 주인의 endpoint는 탈취되지 않는다

- **WHEN** 살아 있는 주인이 수락 중인 socket 위에서 두 번째 기동이 시도되면
- **THEN** 두 번째 기동은 그 socket을 unlink하지 않고 거부된다

#### Scenario: 쓰기 비트가 깎인 산 socket은 죽은 것이 아니다

- **WHEN** 수락 중인 socket의 권한에서 소유자 쓰기 비트가 외부 chmod로 깎인 상태에서
  같은 endpoint의 회수가 실행되면
- **THEN** 회수는 권한 비트로 사망을 추정하지 않고 probe로 생존을 확인해 그 socket을
  제거하지 않는다

#### Scenario: staging 잔재는 우리 잔재다

- **WHEN** 발행 전 임시 이름(신규 `.s-` staging 또는 현행·구버전 공통 CreateTemp
  이름)의 정규 파일·socket만 남은 상태에서 엔진이 기동하면
- **THEN** 잔재를 회수하고 기동을 계속하며 회수 후 control 디렉터리에 잔재가 없다

#### Scenario: 낯선 엔트리는 건드리지 않는다

- **WHEN** socket을 발행하는 endpoint의 control 디렉터리에 그 endpoint가 만들 수 없는
  이름 또는 모양의 엔트리가 있는 상태에서 기동하면
- **THEN** 회수는 아무것도 제거하지 않고 그 endpoint의 기동을 거부한다
