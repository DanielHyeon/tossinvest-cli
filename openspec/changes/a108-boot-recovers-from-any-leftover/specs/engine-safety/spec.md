# engine-safety — a108 delta

> 2026-08-13 23:35 KST 호스트 재부팅이 남긴 반쪽 endpoint 잔재(`endpoint.json`만 남은
> `.strategy-runtime-read/`)가 엔진 기동을 영구 거부시켰다 — US 장중, 손절 감시 전면
> 정지. 회수 코드가 자기 종료 시퀀스가 만들 수 있는 상태를 "영원히 거부"로 처리했고,
> 조회 전용 export socket의 실패가 보호 루프 전체를 죽였다.

## ADDED Requirements

### Requirement: 런타임 endpoint 잔재 회수는 자기 수명주기가 만드는 모든 상태를 다룬다

엔진이 소유한 런타임 endpoint(control 디렉터리·descriptor·socket)의 기동 시 잔재 회수는 그 endpoint의 생성·종료·회수 시퀀스가 만들 수 있는 **모든** 부분 상태(빈 디렉터리, descriptor만, socket만, 둘 다)를 소유자 사망 검증 후 사람 개입 없이 회수해야 한다(SHALL). 소유자 생존 판정은 프로세스 ID 재사용에 오판되지 않는 수단이어야 하며(SHALL) — kill-0 단독 판정은 금지된다(SHALL NOT). 소유권·권한·symlink 검증과 낯선 엔트리의 거부는 유지되어야 한다(SHALL) — 회수의 확장은 검증의 완화가 아니다.

#### Scenario: 반쪽 잔재에서의 재기동 (2026-08-13 사고)

- **WHEN** control 디렉터리에 descriptor만 남은 상태(graceful shutdown이 socket을 unlink한
  뒤 프로세스가 죽음)에서 엔진이 기동하면
- **THEN** 잔재를 회수하고 기동을 계속한다 — 어떤 재시도 루프도 같은 상태에 영구히
  막히지 않는다

#### Scenario: 재사용된 PID는 주인이 아니다

- **WHEN** 잔재 descriptor의 PID 자리에 무관한 생존 프로세스가 있고 socket은 수락하지
  않으면
- **THEN** 소유자 사망으로 판정하고 회수한다

#### Scenario: 살아 있는 주인은 건드리지 않는다

- **WHEN** 잔재의 socket이 연결을 수락하면
- **THEN** 회수하지 않고 이번 기동 시도를 거부한다

### Requirement: 조회 전용 endpoint의 실패는 엔진을 죽이지 않는다

조회 전용 export endpoint(strategy projection 등)의 기동 실패는 엔진 기동을 중단시켜서는 안 되며(SHALL NOT), 엔진은 해당 endpoint 없이 보호·대사 루프를 계속하고 미해소 critical 알림으로 그 사실을 운영자에게 보여야 한다(SHALL). 엔진 싱글턴 보장은 journal flock이 단독으로 소유하며(SHALL), 조회 endpoint의 상태가 싱글턴 판정에 쓰여서는 안 된다(SHALL NOT).

#### Scenario: projection 기동 실패에서의 엔진 기동

- **WHEN** strategy projection endpoint 기동이 실패하면
- **THEN** 엔진은 projection 없이 루프를 시작하고 critical 알림을 발행하며, 손절·대사
  판정 경로는 영향받지 않는다

#### Scenario: 강등 기동의 싱글턴 불변

- **WHEN** projection 없이 강등 기동한 엔진이 살아 있는 동안 두 번째 엔진이 기동을
  시도하면
- **THEN** journal flock이 두 번째 기동을 거부한다
