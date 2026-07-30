## Why

TossOS의 Codex·Claude MCP 설정은 각 세션마다 동일한 프로젝트 `GBRAIN_HOME`으로
`gbrain serve`를 실행한다. PGLite는 단일 프로세스만 열 수 있으므로 후발 서버가 기존
서버의 잠금을 기다리며 응답하지 않고, GBrain advisory 조회와 `make sdd-sync`까지
timeout시키는 장애가 재현된다.

## What Changes

- 프로젝트 GBrain wrapper가 커널 수명에 묶인 singleton lock으로 `serve` 중복 실행을
  PGLite 진입 전에 거절한다.
- 활성 서버가 있는 동안 CLI advisory 작업은 장시간 PGLite lock을 기다리지 않고 소유자
  진단과 함께 즉시 busy를 반환한다.
- `make sdd-sync`는 GBrain busy를 advisory warning으로 처리하고 CodeGraph hard-evidence
  동기화와 freshness 기록을 계속한다.
- stale lock 파일을 무조건 삭제하거나 살아 있는 다른 에이전트/부모 프로세스를 종료하지
  않으며, 운영 복구는 PID·명령·`GBRAIN_HOME` 검증 뒤 비소유 중복 `gbrain serve`만
  정상 종료한다.
- 동시 실행·프로세스 종료 후 자동 회수·busy 분류에 대한 회귀 테스트와 운영 문서를 추가한다.

## Capabilities

### New Capabilities

_없음._

### Modified Capabilities

- `sdd-workflow`: 프로젝트 GBrain을 단일 PGLite 소유자로 조립하고, advisory contention이
  hard-evidence SDD 동기화를 차단하지 않는 요구사항을 추가한다.

## Impact

- `tools/sdd/gbrain_project.py`
- `tools/sdd/sdd_sync.py`
- 위 도구의 Python 회귀 테스트
- `docs/WORKFLOW.md`
- `openspec/specs/sdd-workflow`

제품 런타임, 주문·Guardian·원장·reconciliation 경로와 설치된 `tossctl` 바이너리는
변경하지 않는다.
