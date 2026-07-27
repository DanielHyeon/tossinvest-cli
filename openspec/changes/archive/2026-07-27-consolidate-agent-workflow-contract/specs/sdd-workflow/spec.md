## MODIFIED Requirements

### Requirement: SDD 도구의 실재와 동기 검증
에이전트 규칙에 등재된 SDD 도구와 경로는 저장소의 `make sdd-check`로 검증 가능해야 한다(SHALL).
상세 개발 절차의 단일 정본은 `docs/WORKFLOW.md`여야 하며(SHALL), Claude·Codex 진입 파일은
동일한 최소 안전 부트스트랩과 `docs/WORKFLOW.md` 포인터를 포함해야 한다(SHALL).
설치되지 않은 필수 CLI, 존재하지 않는 산출물 경로, 최소 부트스트랩 drift 또는 정본 포인터
누락은 게이트를 실패시켜야 한다(SHALL).

#### Scenario: ast-grep 누락
- **WHEN** 개발 환경에서 `make sdd-check`를 실행했으나 ast-grep CLI가 없으면
- **THEN** 설치 명령을 포함한 오류로 실패한다

#### Scenario: Codex 최소 안전 부트스트랩 drift
- **WHEN** `.codex/agents.md`의 최소 안전 부트스트랩이 `.claude/CLAUDE.md`와 다르면
- **THEN** 설정 동기 검사가 실패하고 재생성 명령을 안내한다

#### Scenario: 상세 워크플로 정본 포인터 누락
- **WHEN** Claude 또는 Codex 진입 파일에서 `docs/WORKFLOW.md` 필수 참조가 누락되면
- **THEN** 설정 동기 검사가 실패한다

#### Scenario: CodeGraph hard-evidence index drift
- **WHEN** 마지막 `make sdd-sync` 이후 tracked 또는 untracked 소스가 변경되었다
- **THEN** `make sdd-check`는 stale CodeGraph fingerprint로 실패하고 재동기화를 요구한다
