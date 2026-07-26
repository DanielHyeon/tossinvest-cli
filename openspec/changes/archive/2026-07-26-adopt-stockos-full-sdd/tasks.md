# Tasks: adopt-stockos-full-sdd

## 0. Proposal freeze

- [x] 0.1 StockOS 도구 실소스와 TossOS 설치 상태를 대조하고 재사용/재구현 매트릭스를 proposal·design에 고정한다.
- [x] 0.2 `openspec validate adopt-stockos-full-sdd --strict`를 통과한다.
- [x] 0.3 gstack autoplan의 CEO·Eng·DX 관점으로 proposal/design/spec delta를 검토하고 `review.md`에 결정 근거를 기록한다.

## 1. Toolchain and project registration

- [x] 1.1 누락된 ast-grep을 설치하고 rtk/OpenSpec/CodeGraph/CodeGraphContext/Create Context Graph/GBrain/gstack/Superpowers/Docker doctor를 추가한다.
- [x] 1.2 TossOS 전용 CodeGraph MCP, CodeGraphContext index, GBrain source pin을 구성하고 smoke 조회한다.
- [x] 1.3 공유 TypeDB/Neo4j의 TossOS database/source 격리 정책과 비차단 연결을 구성한다.

## 2. Full SDD repository tooling

- [x] 2.1 Go AST 추출기, Go ast-grep 위험 패턴, Function Logic Map scaffold/report와 테스트를 구현한다.
- [x] 2.2 파일 기반 episodic memory와 retain/recall/rebuild 테스트를 이식한다.
- [x] 2.3 agent-save/commit event 마스킹·버퍼·best-effort graph sync와 훅을 구현한다.
- [x] 2.4 에이전트 설정 공유 블록 mirror/drift 검사와 테스트를 구현한다.
- [x] 2.5 TossOS PM registry/generator/check와 bootstrap story를 구현한다.

## 3. Instructions and gates

- [x] 3.1 `.claude/CLAUDE.md`, `.codex/agents.md`, 루트 `CLAUDE.md`, `AGENTS.md`를 Full SDD 계약으로 동기화한다.
- [x] 3.2 기존 사용자 편집을 보존하며 `docs/WORKFLOW.md`에 전체 실행 순서·Function Logic Map·기억·관측·PM·완료 증거를 통합한다.
- [x] 3.3 `.mcp.json`, Claude/Codex hooks, git hooks, `.gitignore`, Makefile의 `sdd-doctor/sdd-sync/sdd-check`를 구성하고 `make gate`에 연결한다.

## 4. Verification

- [x] 4.1 SDD 도구 단위 테스트와 hook offline smoke를 통과한다.
- [x] 4.2 Go 샘플 함수의 AST·Function Logic Map·위험 패턴 산출물을 생성해 end-to-end를 검증한다.
- [x] 4.3 `make sdd-check`, `make test`, `make vet`, `make validate`, `make gate CHANGE=adopt-stockos-full-sdd`를 통과한다.
- [x] 4.4 diff 독립 리뷰, High-risk 영향 없음, StockOS 데이터 미복제, 남은 외부 서비스 상태를 최종 보고한다.
