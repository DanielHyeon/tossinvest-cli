# Change: adopt-stockos-full-sdd

## Why

TossOS는 OpenSpec·gstack·테스트 게이트의 일부만 사용하고 있고, StockOS에서 실제 운용 중인 코드 증거·함수 내부 분석·두 계층 기억·관측 그래프·에이전트 설정 동기화가 빠져 있다. 특히 기존 함수를 수정하기 전 호출 관계와 함수 내부 분기를 함께 고정하는 장치가 없어, 문서상 SDD와 실제 편집 절차가 어긋날 수 있다.

이 change는 StockOS의 Full SDD를 TossOS의 Go 코드베이스와 안전 불변식에 맞게 이식한다. 도구 이름만 문서에 적는 것이 아니라 설치 상태, 저장소 설정, 자동 훅, 검증 명령과 생성 산출물까지 연결한다.

## What Changes

- Full SDD 권위 계층을 `OpenSpec → CodeGraph hard evidence → CodeGraphContext 보조 문맥 → Go AST/ast-grep Function Logic Map → Superpowers TDD → gstack gate → 기억·관측`으로 고정한다.
- StockOS의 설치 도구를 재사용 가능성에 따라 아래처럼 적용한다.

| 분류 | 도구 | TossOS 적용 |
|---|---|---|
| 전역 그대로 재사용 | rtk, OpenSpec, CodeGraph, CodeGraphContext, Create Context Graph, GBrain, gstack, Superpowers, Docker | 중복 설치 없이 프로젝트 등록·설정·doctor 검증 |
| 신규 설치 | ast-grep CLI | 전역 설치 후 Go 위험 패턴 스캔에 사용 |
| 경로 조정 후 재사용 | agent save/commit history hook, 파일 기억, 설정 동기 검사, PM tracker generator | TossOS 경로·식별자·검증 명령으로 이식 |
| 구현 교체 | Python AST 추출기·Python ast-grep 규칙 | Go 표준 AST 추출기·Go 위험 패턴으로 교체 |
| 서비스만 공유 | StockOS TypeDB·Neo4j | 같은 로컬 서비스 사용 가능, database/source는 TossOS 전용으로 격리 |
| 재사용 금지 | StockOS PM 항목·기억 데이터·CodeGraph 인덱스 | 제품별 사실 오염을 막기 위해 TossOS 데이터로 새로 생성 |

- Claude, Codex, 범용 에이전트가 동일한 공유 SDD 블록을 읽도록 `.claude/CLAUDE.md`, `.codex/agents.md`, 루트 `CLAUDE.md`, `AGENTS.md`를 연결하고 drift 검사를 추가한다.
- `.mcp.json`, `.codex/config.toml`, Claude/Codex save hook, git post-commit hook을 TossOS 루트에 맞게 구성한다.
- CodeGraph·CodeGraphContext·GBrain 프로젝트 인덱스와 파일 기반 episodic memory를 초기화한다.
- TypeDB SDD Control Graph와 Neo4j Create Context Graph는 관측 전용·비차단으로 연결한다.
- `make sdd-doctor`, `make sdd-sync`, `make sdd-check`를 추가하고 기존 `make gate`가 Full SDD 구성 검사를 포함하게 한다.

## Impact

- 영향 스펙: `sdd-workflow`
- 영향 파일: 에이전트 규칙, 개발 워크플로, Makefile, `.mcp.json`, `.claude/`, `.codex/`, `.githooks/`, `.sdd/`, `tools/{logic-map,sdd,sdd-history,pm}`, `scripts/memory-*`, `docs/{frameworks,memory,pm,templates}`
- 라이브 주문·위험·원장 경로 영향: 없음
- 외부 서비스: 기존 로컬 TypeDB/Neo4j를 선택적으로 공유하되 데이터 네임스페이스를 분리한다. 서비스 장애는 개발·커밋을 차단하지 않는다.

## Success Criteria

1. 열거된 전역 도구가 모두 doctor에서 탐지되고 `ast-grep`이 실제 실행된다.
2. CodeGraph, CodeGraphContext, GBrain이 TossOS 루트에 대해 조회 가능한 상태다.
3. Go 함수 하나에 대해 AST 추출, Function Logic Map scaffold, 위험 패턴 보고서를 생성할 수 있다.
4. Claude/Codex/AGENTS 공유 규칙 drift가 자동 검사로 실패한다.
5. save/commit hook은 이벤트를 마스킹된 로컬 버퍼에 남기며 외부 그래프가 꺼져 있어도 성공한다.
6. `openspec validate adopt-stockos-full-sdd --strict`, SDD 도구 테스트, `make gate CHANGE=adopt-stockos-full-sdd`가 통과한다.
