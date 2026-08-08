## Context

GBrain의 TossOS 전용 데이터 홈은 `.sdd/gbrain-home`이고 저장 엔진은 PGLite다.
PGLite는 데이터 디렉터리당 단일 프로세스 연결만 허용한다. 반면 `.codex/config.toml`과
`.mcp.json`의 stdio MCP 설정은 각 Codex/Claude 세션마다
`tools/sdd/gbrain_project.py serve`를 별도 실행한다.

2026-07-30 실측에서 동일 `GBRAIN_HOME`을 가진 `gbrain serve` 두 개가 동시에 살아 있었고,
한 프로세스만 `.gbrain-lock/lock`의 heartbeat 소유자였다. 후발 프로세스와 직접 CLI는
잠금 획득을 기다리며 10초 timeout을 냈다. GBrain은 advisory이므로 이 contention이
CodeGraph hard-evidence 동기화나 개발 게이트를 멈추게 해서는 안 된다.

## Goals / Non-Goals

**Goals:**

- 동일 TossOS GBrain 홈에 진입하는 모든 wrapper 실행을 한 프로세스로 직렬화한다.
- 중복 MCP 시작과 CLI/MCP 경합을 PGLite 내부 timeout 전에 즉시 감지한다.
- 프로세스 crash·SIGKILL 뒤 별도 stale-lock 삭제 없이 자동 복구한다.
- 이미 wrapper 밖에서 시작된 legacy GBrain 소유자도 PID 생존 확인으로 감지한다.
- GBrain busy 상태에서도 CodeGraph hard-evidence 동기화는 완료한다.

**Non-Goals:**

- 여러 stdio MCP 클라이언트에 GBrain을 동시에 제공하는 HTTP broker 구축
- PGLite를 PostgreSQL/Supabase로 마이그레이션
- GBrain upstream의 잠금/IPC 구현 수정
- 살아 있는 잠금 소유자나 에이전트 부모 프로세스의 자동 종료

## Decisions

### D1. wrapper 입구에 POSIX advisory flock을 둔다

`gbrain_project.py`는 모든 하위 명령 전에
`.sdd/gbrain-home/.gbrain/tossos-process.lock`을 열고 `LOCK_EX|LOCK_NB`를 획득한다.
획득한 file descriptor는 `execve` 뒤에도 상속되게 해 실제 `gbrain` 프로세스 수명 동안
커널이 lock을 유지한다.

커널 flock은 프로세스 종료 시 자동 회수되므로 PID 파일만으로 singleton을 구현하는
방식과 달리 stale cleanup이나 PID 재사용 추측이 필요 없다. lock 파일에는 진단용
PID·명령·데이터 홈만 기록하고 권위는 파일 내용이 아니라 flock 소유권이다.

### D2. pre-wrapper legacy PGLite 소유자를 별도로 감지한다

새 flock을 획득했더라도 기존 `.gbrain-lock/lock`에 다른 살아 있는 PID와 steal grace
안의 `refreshed_at` heartbeat가 함께 기록되어 있으면 현재 프로세스는 GBrain을 실행하지
않고 temporary-failure로 종료한다. 이는 변경 배포 전에 시작된 `gbrain serve`가 새
flock을 보유하지 않는 전환 구간을 보호한다.

PID 생존은 `os.kill(pid, 0)`으로 확인하고 heartbeat freshness는 GBrain upstream과 같은
`GBRAIN_PGLITE_LOCK_STEAL_GRACE_SECONDS`(기본 600초) 계약을 사용한다. dead, malformed,
stale-heartbeat metadata는 삭제하지 않고 실제 GBrain 자체 stale-lock 복구에 맡긴다.
이 조건이 없으면 종료된 holder의 PID가 다른 프로세스에 재사용됐을 때 wrapper가 upstream
복구 진입을 영구 차단한다.

### D3. busy는 exit 75와 안정된 marker로 표현한다

중복 실행은 `EX_TEMPFAIL` 의미의 exit 75와 `[gbrain-project] busy:` stderr marker를
반환한다. 오류에는 확인 가능한 소유자 PID·명령·홈만 포함하며 토큰·환경 전체는 출력하지
않는다. wrapper는 기다리지 않으므로 에이전트 MCP 초기화가 잠금 timeout으로 고착되지 않는다.

### D4. sdd-sync는 wrapper를 사용하고 busy를 advisory로 처리한다

`sdd_sync.py`의 init/source/sync 명령은 raw `gbrain` 대신 project wrapper를 호출한다.
source probe가 exit 75 또는 busy marker를 반환하면 GBrain freshness는 갱신하지 않고
경고를 출력하되 failure 목록에는 넣지 않는다. CodeGraph와 CodeGraphContext 성공 상태는
그대로 기록한다.

다른 GBrain 오류·실제 timeout은 기존처럼 incomplete로 보고한다. advisory라는 이유로
모든 오류를 삼키지 않고, 정상적인 활성 소유자 contention만 좁게 허용한다.

### D5. 현재 복구는 비소유 중복 프로세스만 정상 종료한다

lock metadata의 PID, `/proc/<pid>/environ`의 `GBRAIN_HOME`, command line을 교차검증한다.
PGLite lock 소유자가 아닌 동일 홈의 중복 `gbrain serve`에만 SIGTERM을 보내고 종료를
확인한다. 잠금 소유자와 Codex/Claude 부모 프로세스는 건드리지 않는다.

## Risks / Trade-offs

- [두 번째 동시 에이전트는 GBrain MCP를 사용할 수 없음] → GBrain은 advisory이고
  hard evidence는 CodeGraph다. 동시 다중 클라이언트가 필요하면 별도 HTTP broker 또는
  PostgreSQL backend change로 해결한다.
- [활성 MCP 동안 GBrain 색인이 stale] → `sdd-check`는 기존 계약대로 warning만 내며,
  MCP 소유자가 종료된 뒤 `make sdd-sync`가 갱신한다.
- [PID metadata가 오래됨] → metadata는 진단 전용이며 busy 판정 권위는 flock과 실제 PID
  생존 검사다.
- [Windows에서 `fcntl` 부재] → TossOS의 지원 개발/운영 환경은 현재 Linux/WSL이며,
  미지원 플랫폼은 명시적 오류로 fail closed한다.

## Migration Plan

1. wrapper·sync 회귀 테스트를 배포한다.
2. 현재 PGLite lock 소유자와 같은 홈의 `gbrain serve` 목록을 재검증한다.
3. 비소유 중복만 SIGTERM으로 정상 종료한다.
4. 기존 소유자는 세션 종료까지 유지한다. 이후 시작되는 모든 프로세스는 새 flock을 사용한다.
5. `make sdd-sync`가 활성 소유자 아래에서 빠르게 advisory busy로 끝나고 hard index를
   갱신하는지 확인한다.

Rollback은 wrapper와 sync 변경을 되돌리는 것이다. 데이터 파일·PGLite lock 디렉터리는
수정하거나 삭제하지 않으므로 데이터 migration rollback은 없다.

## Open Questions

없음. 다중 동시 MCP 지원은 이 change의 fail-fast singleton 계약과 분리된 후속 인프라
change로 다룬다.
