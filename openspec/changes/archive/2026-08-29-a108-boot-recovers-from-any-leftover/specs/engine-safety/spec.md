# engine-safety — a108 delta

> 2026-08-13 23:35 KST 호스트 재부팅이 남긴 반쪽 endpoint 잔재(`endpoint.json`만 남은
> `.strategy-runtime-read/`)가 엔진 기동을 영구 거부시켰다 — US 장중, 손절 감시 전면
> 정지. A1·A2 적대 리뷰(2026-08-14)가 요구를 두 방향으로 고쳤다: 회수만이 아니라
> **발행도** 전체적이어야 하고(쓰다 만 산출물이 최종 이름에 나타나면 안 된다), 강등
> **보고는 게이트에 연결되면 안 된다**(미전달 알림 행이 다음 부팅의 진입을 잠근다).
> 형제 endpoint 셋의 같은 병은 a109가 승계한다(design D5-2).

## ADDED Requirements

### Requirement: strategy projection endpoint의 잔재 회수는 자기 수명주기가 만드는 모든 상태를 다룬다

엔진의 strategy projection endpoint(control 디렉터리·descriptor·socket)의 기동 시 잔재 회수는 그 생성·종료·회수 시퀀스가 만들 수 있는 모든 부분 상태(빈 디렉터리, descriptor만, socket만, 둘 다, 쓰다 만 산출물과 staging 잔재)를 소유자 사망 검증 후 사람 개입 없이 회수해야 하며(SHALL), 산출물 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 소유자 생존 판정은 프로세스 ID 재사용에 오판되지 않는 수단이어야 하며(SHALL) kill-0 단독 판정은 금지되고(SHALL NOT), 소유권·symlink 검증과 낯선 엔트리의 거부는 유지되어야 한다(SHALL).

#### Scenario: 반쪽 잔재에서의 재기동 (2026-08-13 사고)

- **WHEN** control 디렉터리에 descriptor만 남은 상태(graceful shutdown이 socket을 unlink한
  뒤 프로세스가 죽음)에서 엔진이 기동하면
- **THEN** 잔재를 회수하고 기동을 계속한다 — 어떤 재시도 루프도 같은 상태에 영구히
  막히지 않는다

#### Scenario: 쓰다 만 잔재도 잔재다

- **WHEN** 0바이트·잘린 descriptor, 또는 chmod 전에 죽어 group/other 비트 없는 비-0600
  권한으로 남은 socket이 잔재로 남은 상태에서 엔진이 기동하면
- **THEN** 소유자 사망이 입증되는 한 회수하고 기동을 계속한다

#### Scenario: 재사용된 PID는 주인이 아니다

- **WHEN** 잔재 descriptor의 PID 자리에 무관한 생존 프로세스가 있고 socket은 수락하지
  않으면
- **THEN** 소유자 사망으로 판정하고 회수한다

#### Scenario: 살아 있는 주인은 건드리지 않는다

- **WHEN** 잔재의 socket이 연결을 수락하면
- **THEN** 회수하지 않고 이번 기동 시도를 거부한다

#### Scenario: 선임자의 늦은 정리가 후계자를 지우지 않는다

- **WHEN** 종료 중인 선임 프로세스의 지연된 정리와 후계 프로세스의 endpoint 발행이
  겹치면
- **THEN** 선임자는 자신이 발행한 경로만 제거할 수 있고 후계자의 socket은 사라지지
  않는다

### Requirement: 조회 전용 endpoint의 실패는 엔진을 죽이지 않는다

조회 전용 export endpoint(strategy projection 등)의 기동 실패는 엔진 기동을 중단시켜서는 안 되며(SHALL NOT), 엔진은 해당 endpoint 없이 보호·대사 루프를 계속하고 기동 경고와 관측 이벤트로 그 사실을 보고해야 하며(SHALL), 그 보고가 알림 outbox·알림 전달 상태·entry gate에 연결되어서는 안 되고(SHALL NOT — 미전달 행은 다음 부팅의 진입을 잠근다), 강등 기동 후 같은 프로세스 안에서 endpoint 재시도를 해서는 안 되며(SHALL NOT), 엔진 싱글턴 보장은 journal flock이 단독으로 소유한다(SHALL).

#### Scenario: projection 기동 실패에서의 엔진 기동

- **WHEN** strategy projection endpoint 기동이 실패하면
- **THEN** 엔진은 projection 없이 루프를 시작하고 기동 경고·관측 이벤트를 남기며,
  손절·대사 판정 경로는 영향받지 않는다

#### Scenario: 강등 기동은 진입 상태를 바꾸지 않는다

- **WHEN** 알림 전달이 불가능한 배포에서 강등 기동한 엔진을 재기동하면
- **THEN** 강등이 만든 미전달 알림 행은 존재하지 않고 entry gate는 그것 때문에 잠기지
  않는다

#### Scenario: 강등 기동의 싱글턴 불변

- **WHEN** projection 없이 강등 기동한 엔진이 살아 있는 동안 두 번째 엔진이 기동을
  시도하면
- **THEN** journal flock이 두 번째 기동을 거부한다
