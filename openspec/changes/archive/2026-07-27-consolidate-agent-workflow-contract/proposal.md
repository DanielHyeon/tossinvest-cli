## Why

TossOS의 상세 개발 절차가 `docs/WORKFLOW.md`와 Claude/Codex 공유 블록에 중복되어 있어
같은 규칙이 서로 다른 표현으로 유지되고 drift·충돌 위험이 있다. 상세 절차는 하나의 정본으로
통합하고, 에이전트별 파일은 세션 시작에 필요한 최소 안전 계약과 정본 포인터만 제공해야 한다.

## What Changes

- `docs/WORKFLOW.md`를 TossOS 개발 절차와 완료 조건의 단일 상세 정본으로 명시한다.
- `.claude/CLAUDE.md`와 `.codex/agents.md`의 공유 블록을 최소 안전 부트스트랩으로 축소한다.
- 부트스트랩은 LIVE 주문 금지, 안전 불변식, 권위 경계, 필수 진입 순서와
  `docs/WORKFLOW.md` 강제 참조를 유지한다.
- agent config 동기화 검사를 전체 워크플로 복제 검사에서 최소 부트스트랩 동기화 검사로 변경한다.
- 관련 테스트와 `sdd-workflow` 정본 요구사항을 새 구조에 맞게 갱신한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `sdd-workflow`: Claude/Codex 파일이 전체 SDD 계약을 복제하는 대신 동일한 최소 안전
  부트스트랩을 포함하고 상세 절차의 단일 정본인 `docs/WORKFLOW.md`로 라우팅하도록 변경한다.

## Impact

- `.claude/CLAUDE.md`
- `.codex/agents.md`
- `docs/WORKFLOW.md`
- `tools/sdd/check_agent_config_sync.py`
- `tools/sdd/test_agent_config_sync.py`
- `openspec/specs/sdd-workflow/spec.md`
- `make sdd-check`의 agent config sync 단계
