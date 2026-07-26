# sdd-workflow Specification (delta)

## ADDED Requirements

### Requirement: Full SDD 권위 계층
모든 신규 기능·동작 변경은 OpenSpec 계약, 현재 HEAD와 CodeGraph hard evidence, CodeGraphContext 보조 문맥, 기존 함수 내부 변경 시 Go AST·ast-grep Function Logic Map, Superpowers TDD, gstack 게이트 순서로 수행되어야 한다(SHALL). CodeGraphContext·GBrain·기억·관측 그래프는 advisory이며 OpenSpec, 현재 HEAD, 테스트, gstack을 대체해서는 안 된다(SHALL NOT).

#### Scenario: 보조 문맥과 현재 HEAD 충돌
- **WHEN** CodeGraphContext 또는 기억 결과가 현재 HEAD와 다르면
- **THEN** 현재 HEAD와 CodeGraph를 다시 확인·동기화한 뒤 그 결과를 구현 근거로 사용한다

#### Scenario: 기존 함수 내부 변경
- **WHEN** 기존 함수의 분기·early return·mutation·side effect를 변경하면
- **THEN** 구현 전에 Function Logic Map과 Branch Test Map을 만들고 변경 후 source hash·함수·분기와 묶인 증거로 최신화한다

#### Scenario: Function Logic Map 면제 시도
- **WHEN** `not-applicable` 면제를 기록했지만 비교 기준 대비 기존 Go 함수가 수정되었다
- **THEN** gate는 수정 함수를 diff에서 계산하고 해당 함수의 완전한 증거 묶음이 없으면 실패한다

### Requirement: SDD 도구의 실재와 동기 검증
에이전트 규칙에 등재된 SDD 도구와 경로는 저장소의 `make sdd-check`로 검증 가능해야 하며(SHALL), Claude·Codex 공유 SDD 블록은 drift 없이 동기화되어야 한다(SHALL). 설치되지 않은 필수 CLI, 존재하지 않는 산출물 경로, 공유 블록 drift는 게이트를 실패시켜야 한다(SHALL).

#### Scenario: ast-grep 누락
- **WHEN** 개발 환경에서 `make sdd-check`를 실행했으나 ast-grep CLI가 없으면
- **THEN** 설치 명령을 포함한 오류로 실패한다

#### Scenario: Codex 규칙 drift
- **WHEN** `.codex/agents.md`의 공유 블록이 `.claude/CLAUDE.md`와 다르면
- **THEN** 설정 동기 검사가 실패하고 재생성 명령을 안내한다

#### Scenario: CodeGraph hard-evidence index drift
- **WHEN** 마지막 `make sdd-sync` 이후 tracked 또는 untracked 소스가 변경되었다
- **THEN** `make sdd-check`는 stale CodeGraph fingerprint로 실패하고 재동기화를 요구한다

### Requirement: 두 계층 기억
파일 기반 episodic memory를 primary 정본으로 사용하고 GBrain을 의미 검색 보조 계층으로 사용해야 한다(SHALL). 기억은 검증 전 자동으로 canonical 승격되어서는 안 되며(SHALL NOT), 시크릿·개인정보·검증되지 않은 실거래 수익 결론을 저장해서는 안 된다(SHALL NOT).

#### Scenario: 검증되지 않은 작업 학습 저장
- **WHEN** 완료 전 학습을 retain하면
- **THEN** episodic 상태로만 기록되고 canonical 승격은 별도 검증 근거를 요구한다

#### Scenario: 동시 retain과 canonical ID 재사용
- **WHEN** 여러 에이전트가 동시에 memory를 retain하거나 기존 canonical ID를 episodic으로 다시 retain한다
- **THEN** ledger transaction은 직렬화되고 canonical ID의 교체·강등은 거절된다

### Requirement: 비차단 SDD 관측
에이전트 저장과 git commit은 마스킹된 로컬 이벤트로 관측되어야 하며(SHALL), TypeDB SDD Control Graph와 Neo4j Create Context Graph로 best-effort 동기화될 수 있다. 관측 서비스·CLI 오류는 저장, 커밋, 테스트를 차단해서는 안 된다(SHALL NOT). StockOS와 서비스를 공유할 때 database와 source namespace는 TossOS 전용으로 분리되어야 한다(SHALL).

#### Scenario: TypeDB 중단 상태의 커밋
- **WHEN** TypeDB에 접속할 수 없는 상태에서 post-commit hook이 실행되면
- **THEN** 로컬 마스킹 이벤트는 보존되고 hook은 성공 종료하여 커밋을 방해하지 않는다

#### Scenario: 공유 Neo4j 사용
- **WHEN** TossOS 이벤트를 공유 Neo4j에 ingest하면
- **THEN** TossOS source 이름으로 기록되고 StockOS source를 덮어쓰지 않는다

### Requirement: PM 계층과 OpenSpec 역추적
활성 OpenSpec change는 PM story와 1:1로 연결되거나 명시적 bootstrap 예외를 가져야 하며(SHALL), generator는 initiative→epic→feature→story→change의 양방향 참조와 고아 항목을 검사해야 한다(SHALL). StockOS의 PM 데이터·기억·인덱스를 TossOS 사실로 복제해서는 안 된다(SHALL NOT).

#### Scenario: 고아 OpenSpec change
- **WHEN** 활성 change가 어떤 TossOS story에도 연결되지 않고 bootstrap 예외에도 없으면
- **THEN** PM check가 실패하고 누락 change id를 출력한다
