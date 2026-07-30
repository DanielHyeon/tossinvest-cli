## ADDED Requirements

### Requirement: 프로젝트 GBrain 단일 프로세스 소유권
The TossOS project-local GBrain wrapper SHALL serialize MCP and CLI processes
entering the same `GBRAIN_HOME` with a kernel-lifetime singleton lock. 살아 있는 소유자가
있을 때 후발 프로세스는 PGLite 내부 timeout을 기다리거나 잠금 파일을 삭제해서는 안 되며,
소유자 진단을 포함한 temporary-failure로 즉시 종료해야 한다.

#### Scenario: 두 에이전트가 동시에 GBrain MCP를 시작한다
- **WHEN** 첫 번째 `gbrain serve`가 project singleton lock을 보유한 동안 두 번째 세션이 같은 wrapper로 `serve`를 시작하면
- **THEN** 두 번째 실행은 실제 GBrain/PGLite를 시작하지 않고 exit 75와 busy 진단을 반환한다

#### Scenario: 소유 프로세스가 비정상 종료한다
- **WHEN** singleton lock 소유 프로세스가 정상 cleanup 없이 종료하면
- **THEN** 커널은 소유권을 자동 회수하고 다음 wrapper 실행은 stale 파일 삭제 없이 GBrain을 시작할 수 있다

#### Scenario: 변경 전 GBrain 소유자가 남아 있다
- **WHEN** singleton flock을 사용하지 않는 기존 프로세스가 같은 홈의 PGLite lock에 살아 있는 PID와 steal grace 안의 heartbeat로 기록되어 있으면
- **THEN** 새 wrapper는 그 lock을 삭제하거나 기다리지 않고 legacy 소유자를 busy로 보고한다

#### Scenario: legacy PID는 살아 있지만 heartbeat가 stale이다
- **WHEN** PGLite lock의 PID가 존재하더라도 `refreshed_at`이 GBrain steal grace보다 오래되었으면
- **THEN** wrapper는 lock을 삭제하지 않고 실제 GBrain 실행에 넘겨 upstream stale-lock recovery가 소유권을 판정하게 한다

### Requirement: GBrain contention의 advisory 격리
SDD synchronization SHALL execute GBrain commands through the project wrapper and
treat verified active-owner contention only as a GBrain freshness warning. 해당 contention은
CodeGraph hard-evidence 동기화와 성공 상태 기록을 실패시키거나 지연시켜서는 안 된다.
contention 이외의 GBrain 오류는 기존 incomplete 진단을 유지해야 한다.

#### Scenario: 활성 MCP 중 make sdd-sync를 실행한다
- **WHEN** project GBrain MCP가 singleton을 보유한 상태에서 SDD sync의 source probe가 실행되면
- **THEN** probe는 빠르게 busy로 분류되고 CodeGraph hard-evidence 동기화 결과는 정상 기록된다

#### Scenario: GBrain 자체 오류가 발생한다
- **WHEN** singleton contention이 아닌 GBrain init, source registration 또는 sync 오류가 발생하면
- **THEN** SDD sync는 해당 GBrain 오류를 incomplete failure로 보고하고 GBrain freshness를 갱신하지 않는다

### Requirement: GBrain 중복 복구의 소유자 보존
Operational recovery SHALL cross-check command line, `GBRAIN_HOME`, and the PGLite
lock owner PID.
자동 복구는 동일 홈의 비소유 중복 `gbrain serve`에만 정상 종료 신호를 보낼 수 있으며,
활성 잠금 소유자·에이전트 부모 프로세스·잠금 데이터 디렉터리를 종료 또는 삭제해서는 안 된다.

#### Scenario: 비소유 중복 프로세스를 복구한다
- **WHEN** 동일 홈의 두 `gbrain serve` 중 한 PID만 PGLite lock owner로 확인되면
- **THEN** 복구 절차는 비소유 GBrain 자식에만 SIGTERM을 보내고 소유자와 부모 에이전트를 유지한다
