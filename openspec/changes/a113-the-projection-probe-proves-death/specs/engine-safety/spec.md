# engine-safety — a113 delta

> a109 가 형제 endpoint 에 세운 사망 증명(chmod-then-probe)을 strategy projection endpoint
> 요구 문면으로 명시해, a108 원형(`projectionSocketAccepts`)에 남은 owner-write 추정
> (a109 issues I1)을 계약 위반으로 만든다. freeze 리뷰 P1-1 에 따라 형제 요구(:356)가 아니라
> projection 자신의 요구를 고치고, 새 SHALL 은 **최종 이름의 socket** 으로 좁힌다 — staging
> 잔재는 a108 의례 그대로 journal flock 이 방어한다(design 선언된 생략).

## MODIFIED Requirements

### Requirement: strategy projection endpoint의 잔재 회수는 자기 수명주기가 만드는 모든 상태를 다룬다

엔진의 strategy projection endpoint(control 디렉터리·descriptor·socket)의 기동 시 잔재 회수는 그 생성·종료·회수 시퀀스가 만들 수 있는 모든 부분 상태(빈 디렉터리, descriptor만, socket만, 둘 다, 쓰다 만 산출물과 staging 잔재)를 소유자 사망 검증 후 사람 개입 없이 회수해야 하며(SHALL), 산출물 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 소유자 생존 판정은 프로세스 ID 재사용에 오판되지 않는 수단이어야 하며(SHALL) kill-0 단독 판정은 금지되고(SHALL NOT), 최종 이름 socket의 소유자 사망 판정은 관측된 권한 비트로 추정해서는 안 되며(SHALL NOT) 검증한 그 socket에 발행 계약 권한(0600)을 복원한 뒤의 connect probe로 증명해야 하고(SHALL), 그 권한 복원은 회수 경로에서만 일어나야 하며 조회 클라이언트는 endpoint의 권한을 바꿔서는 안 되고(SHALL NOT), 소유권·symlink 검증과 낯선 엔트리의 거부는 유지되어야 한다(SHALL).

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

#### Scenario: 쓰기 비트가 깎인 산 socket은 죽은 것이 아니다

- **WHEN** 수락 중인 최종 이름 socket의 권한에서 소유자 쓰기 비트가 외부 chmod로 깎인
  상태에서 엔진이 기동하면
- **THEN** 회수는 권한 비트로 사망을 추정하지 않고 0600 복원 후 probe로 생존을 확인해
  그 socket을 제거하지 않고 기동 시도를 거부한다
- **AND** 그 socket의 group/other 권한은 넓어지지 않는다

#### Scenario: 선임자의 늦은 정리가 후계자를 지우지 않는다

- **WHEN** 종료 중인 선임 프로세스의 지연된 정리와 후계 프로세스의 endpoint 발행이
  겹치면
- **THEN** 선임자는 자신이 발행한 경로만 제거할 수 있고 후계자의 socket은 사라지지
  않는다
