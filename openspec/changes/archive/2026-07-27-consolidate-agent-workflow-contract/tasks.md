## 1. 계약과 문서 구조

- [x] 1.1 OpenSpec proposal, design, delta spec을 strict validation한다.
- [x] 1.2 `docs/WORKFLOW.md`를 상세 개발 절차의 단일 정본으로 명시한다.
- [x] 1.3 Claude/Codex 공유 블록을 동일한 최소 안전 부트스트랩으로 축소한다.
- [x] 1.4 활성 change를 PM bootstrap 예외에 등록한다.

## 2. 동기화 게이트

- [x] 2.1 agent config sync 검사를 최소 부트스트랩과 WORKFLOW 포인터 기준으로 변경한다.
- [x] 2.2 sync 검사 단위 테스트를 새 계약에 맞게 보강한다.
- [x] 2.3 `sdd-workflow` 정본 spec을 delta와 일치시킨다.

## 3. 검증

- [x] 3.1 agent config sync 단위 테스트와 OpenSpec strict validation을 통과시킨다.
- [x] 3.2 격리된 clean worktree에서 `make sdd-sync`, `make sdd-check` 및
  `make gate CHANGE=consolidate-agent-workflow-contract`를 실행한다.
- [x] 3.3 diff와 사용자 소유 변경 비간섭 여부를 독립 검토하고 review 기록을 갱신한다.
