## Context

현재 `.claude/CLAUDE.md`의 `SDD_SHARED` 블록이 공통 계약의 정본이고
`.codex/agents.md`는 이를 byte-identical하게 복제한다. 그러나 동일한 상세 절차가
`docs/WORKFLOW.md`에도 더 구체적인 형태로 존재한다. 이 때문에 에이전트 진입 파일과 상세
워크플로 문서가 서로 독립적으로 진화할 수 있고, 동기화 검사는 Claude/Codex 간 drift만 막을 뿐
`WORKFLOW.md`와의 의미 drift는 막지 못한다.

## Goals / Non-Goals

**Goals:**

- 상세 개발 규칙의 단일 정본을 `docs/WORKFLOW.md`로 고정한다.
- 자동 발견되는 에이전트 파일에는 안전한 작업 시작에 필요한 최소 계약만 둔다.
- Claude/Codex 최소 계약의 byte-identical 동기화와 필수 정본 포인터를 기계적으로 검증한다.
- 기존 LIVE 주문, 토글 OFF, 손절·비상 청산, 운영 승인 불변식을 그대로 보존한다.

**Non-Goals:**

- TossOS 런타임 운용용 `AGENTS.md` recipe 변경
- SDD 실행 순서, 위험도 분류 또는 완료 게이트 완화
- `.codex/agents.md` 삭제
- 다른 작업의 코드나 진행 중 OpenSpec change 수정

## Decisions

1. `docs/WORKFLOW.md`를 상세 절차의 단일 정본으로 사용한다.
   에이전트별 파일에 상세 명령과 도구 설명을 반복하지 않아 문서 간 의미 drift 표면을 줄인다.

2. `SDD_SHARED` 블록은 최소 안전 부트스트랩으로 유지한다.
   자동 로딩 여부가 환경마다 다르므로 파일을 삭제하지 않고, 안전 불변식·권위 순서·필수 읽기
   순서·완료 전 게이트를 짧게 보장한다.

3. `.claude/CLAUDE.md`를 부트스트랩 생성 source로 유지한다.
   기존 생성 명령과 운영 습관을 보존하고 `.codex/agents.md`는 계속 자동 생성 가능한 mirror로 둔다.

4. 동기화 검사는 상세 명령·경로의 존재를 공유 블록에서 검사하지 않는다.
   그 항목들은 `WORKFLOW.md`와 별도 doctor/gate 검사의 책임으로 옮긴다. 대신 공유 블록의 최소
   안전 문구와 두 진입 파일의 `docs/WORKFLOW.md` 포인터를 검사한다.

## Risks / Trade-offs

- [에이전트가 포인터 문서를 읽지 않을 위험] → 공유 블록에 필수 진입 순서와 완료 보고 금지
  조건을 유지하고 루트 `AGENTS.md`도 같은 경로를 가리키게 한다.
- [상세 도구 경로 검사가 약해질 위험] → doctor와 기존 저장소 게이트가 실제 도구 존재를
  검증하도록 유지하고, sync 검사는 문서 복제 여부에만 집중한다.
- [사용자 편집 충돌] → 이미 수정된 다른 파일과 change는 건드리지 않고 대상 문서의 현재 내용을
  기준으로 최소 범위 패치를 적용한다.

## Migration Plan

1. delta spec과 review 기록을 준비하고 strict validation을 통과시킨다.
2. `WORKFLOW.md`에 단일 정본·부트스트랩 역할을 명시한다.
3. `.claude/CLAUDE.md` 공유 블록을 축소하고 생성기로 `.codex/agents.md`를 동기화한다.
4. sync 검사와 단위 테스트를 새 계약에 맞게 변경한다.
5. 정본 spec을 갱신하고 `make sdd-sync`, `make sdd-check`, 관련 테스트를 실행한다.

Rollback은 대상 파일의 이 change diff를 되돌리고 기존 전체 공유 블록 생성 방식을 복원하는 것이다.

## Open Questions

없음.
