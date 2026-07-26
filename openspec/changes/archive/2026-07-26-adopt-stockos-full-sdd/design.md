# Design: TossOS Full SDD toolchain

## 1. 설계 원칙

1. 계약과 증거를 섞지 않는다. OpenSpec은 의도, CodeGraph·현재 HEAD·Go AST는 현재 구현의 사실이다.
2. 보조 기억과 관측 그래프는 advisory다. stale·미접속이어도 OpenSpec, 테스트, 리뷰 게이트는 약화되지 않는다.
3. StockOS의 프로젝트 데이터를 복제하지 않는다. 도구 구현만 재사용하고 TossOS 인덱스·이벤트·PM 항목은 별도 생성한다.
4. hook은 관측 전용·비차단이다. 실패가 저장·커밋을 막지 않고 시크릿 원문을 저장하지 않는다.
5. Go 코드 내부 증거는 Python AST 도구로 대체하지 않는다.

## 2. Full SDD 실행 흐름

```text
memory recall / GBrain
          │
          ▼
OpenSpec change + strict validation
          │
          ▼
CodeGraph callers/callees/impact/context ── hard evidence
          │
          ├── CodeGraphContext ──────────── supporting evidence
          ▼
Go AST + ast-grep
          │
          ▼
Function Logic Map + Branch Test Map
          │
          ▼
RED → GREEN → REFACTOR → VERIFY
          │
          ▼
gstack review/security/QA + make gate
          │
          ▼
archive + PM sync + episodic retain/promotion
```

현재 HEAD·CodeGraph·CodeGraphContext가 충돌하면 현재 HEAD를 직접 읽고 CodeGraph를 다시 동기화한다. CodeGraphContext 단독 결과로 production 편집을 승인하지 않는다.

## 3. 구성 경계

### 3.1 에이전트 규칙

`.claude/CLAUDE.md`를 공유 블록의 정본으로 둔다. `.codex/agents.md`는 같은 공유 블록을 포함하며 `tools/sdd/check_agent_config_sync.py`가 byte 단위 drift를 검사한다. 루트 `CLAUDE.md`와 `AGENTS.md`는 실행 환경이 항상 읽는 진입점으로서 공유 계약과 `docs/WORKFLOW.md`를 명시적으로 가리킨다.

### 3.2 코드 증거

- `.mcp.json`과 `.codex/config.toml`은 저장소 기준 상대 경로의 CodeGraph MCP를 등록한다.
- `.codegraph/` 인덱스는 로컬 파생 데이터이므로 커밋하지 않는다.
- CodeGraphContext도 프로젝트 루트를 별도 인덱싱한다.
- GBrain 실행 파일은 전역 설치를 재사용하되 PGLite 단일-writer 충돌과 StockOS 기억 혼입을 막기 위해
  `.sdd/gbrain-home/`의 TossOS 전용 데이터 홈을 사용한다. `.gbrain-source`로 source를 pin하고
  MCP와 `sdd-sync`는 `tools/sdd/gbrain_project.py`를 통해 같은 홈을 사용한다. GBrain 결과는
  의미 검색·과거 결정 recall에만 쓰고 코드 사실의 최종 권위로 쓰지 않는다.

### 3.3 Function Logic Map

Go 표준 패키지 `go/parser`, `go/ast` 기반 CLI가 지정 함수의 signature, range, branch, early return, call, assignment를 JSON으로 출력한다. `ast-grep`은 goroutine, panic, float 변환, 파일/네트워크 side effect 등 검토 후보를 찾는다. 스캐너 발견은 결함 판정이 아니라 사람이 Function Logic Map에 분류할 증거다.

산출물 위치:

```text
openspec/changes/<change-id>/analysis/function-logic/<package>--<function>/
  ast.json
  function-logic-map.md
  branch-test-map.md
  risk-pattern-report.md
```

기존 함수 내부 편집, high-risk 코드, 복잡한 분기 수정에서는 이 묶음이 필수다. gate는 diff에서
수정된 기존 함수를 계산하고 source hash·함수·분기와 산출물을 대조한다. 새 leaf 함수나
문서·설정만의 변경은 `not-applicable` 근거를 review에 기록할 수 있다.
비교 기준은 proposal freeze 전에 기록한 change별 `base-commit.txt`다. 누락·invalid ref·diff
실패는 빈 변경으로 간주하지 않고 gate를 실패시킨다.

### 3.4 기억

파일 기반 memory가 primary다. `docs/memory/ledger.jsonl`과 원문 Markdown이 정본이고 JSONL/SQLite 검색 인덱스는 파생이다. ledger transaction은 file lock으로 직렬화하며 canonical ID의 재사용·강등과 symlink 탈출을 거절한다. GBrain은 저장소 의미 검색과 교차 세션 recall을 제공한다. 어떤 기억도 검증 전 canonical로 자동 승격하지 않는다.

### 3.5 관측 그래프

- TypeDB: 기존 `stockos-sdd-typedb` 서비스 재사용 가능, TossOS database 이름은 `tossos_sdd`.
- Neo4j: 기존 `infra-neo4j-1` 서비스 재사용 가능, Create Context Graph source 이름은 `tossos-agent-saves`, `tossos-commits`.
- event 파일은 `.sdd/history/events/` 아래 project-local JSONL로 분리한다.
- save/commit hook은 로컬 기록 후 즉시 종료하고 lock-protected worker가 요청을 병합한다.
- TypeDB/Neo4j ingest는 source별 byte-offset checkpoint와 100건 배치로 증분 처리한다.
- 자격증명은 저장소에 넣지 않고 환경변수로만 읽는다. CCG 호출은 credential-free 0700
  임시 workspace에서 수행하고 정상 종료 시 전체 삭제한다.

외부 DB가 없으면 hook은 로컬 버퍼까지만 기록하고 exit 0이다.

### 3.6 PM 계층

StockOS의 STK 항목은 가져오지 않는다. 동일한 initiative → epic → feature → story → OpenSpec change 1:1 연결 원칙과 generator만 TossOS ID(`TOS`)로 적용한다. 생성 문서는 파생이며 registry와 개별 YAML이 정본이다.

## 4. 자동화와 실패 정책

| 경로 | 실패 정책 |
|---|---|
| `make sdd-check` 설정·도구·memory·CodeGraph freshness 계약 | 차단 |
| OpenSpec strict, tests, vet | 차단 |
| Function Logic Map 필수 대상 누락 | 리뷰에서 차단 |
| save/post-commit history capture | 경고 후 비차단 |
| TypeDB/Neo4j ingest | 비차단 |
| CodeGraphContext/GBrain freshness | stale 표시, 비차단 |

`make gate`는 `make sdd-check`를 포함한다. 단, 외부 DB의 실행 여부는 CI/신규 clone에서 재현되지 않으므로 gate 조건이 아니다.

## 5. 검증 전략

- Python stdlib `unittest`: memory schema/검색, event 마스킹, config mirror, PM graph.
- Go test: AST extractor fixture와 branch/call 추출.
- CLI smoke: 각 전역 도구 version/help, CodeGraph/Context/GBrain project status, ast-grep sample scan.
- hook smoke: 외부 DB가 꺼진 상태에서 agent-save·commit event 스크립트가 exit 0이고 JSONL을 생성하는지 확인.
- repository gate: `make sdd-check`, `make test`, `make vet`, `make validate`.

## 6. Rollback

`git config --unset core.hooksPath`로 git hook을 끌 수 있다. Claude/Codex hook 파일과 MCP 설정을
제거하면 자동 관측이 멈춘다. `.codegraph/`, `.sdd/gbrain-home/`, 파생 memory DB와 관측
JSONL은 삭제 가능한 캐시다. 공유 TypeDB/Neo4j의 StockOS database/source는 건드리지 않는다.
