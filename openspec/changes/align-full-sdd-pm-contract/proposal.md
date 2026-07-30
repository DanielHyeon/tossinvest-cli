## Why

TossOS는 Full SDD를 표방하지만 실제 PM registry는 활성 OpenSpec 32개를 Story 대신
bootstrap allowlist로 우회하고, `docs/WORKFLOW.md`의 실행 순서도 StockOS 정본의
READY·evidence reconciliation·Pre-Edit·archive/PM 규칙을 완전하게 보존하지 않는다.
문구와 gate가 다른 상태를 끝내고 Story와 change의 1:1 계약을 기계적으로 강제해야 한다.

## What Changes

- **BREAKING** 활성 OpenSpec change의 bootstrap allowlist 예외를 제거한다.
- 모든 활성 TossOS change를 고유한 Delivery Story에 역등록하고
  Initiative → Epic → Feature → Story → OpenSpec change 양방향 연결을 완성한다.
- Story의 진행 상태를 수동 `status`가 아니라 proposal/tasks/archive 증거에서 파생한다.
- PM generator가 Story→change와 change→Story 양쪽 모두 정확히 1개인지 검사하게 한다.
- `docs/WORKFLOW.md`를 StockOS `.claude/CLAUDE.md`의 Full SDD 순서와 완료 계약에 맞추되,
  TossOS의 Go AST 도구, 공식 Open API/WTS 경계, upstream 회귀 계약, journal·Guardian
  불변식, 프로젝트 namespace와 배포 명령은 그대로 보존한다.
- READY, evidence reconciliation, Pre-Edit Gate, review/security, archive+PM sync,
  episodic→canonical 승격 조건을 명시적으로 복원한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `sdd-workflow`: Story↔OpenSpec 무예외 1:1, 파생 진행 상태, StockOS 기준 Full SDD
  단계와 완료 증거를 TossOS 개발 계약으로 강화한다.

## Impact

- 문서: `docs/WORKFLOW.md`, PM portfolio 원본과 generated tracker
- 도구: `tools/pm/generate_master_tracker.py`와 단위 테스트
- 계약: `openspec/specs/sdd-workflow`
- 에이전트 진입점: 기존 최소 bootstrap과 TossOS 고유 안전 불변식은 변경하지 않는다.
- production trading code, runtime config, 주문·위험·원장 동작: 영향 없음
