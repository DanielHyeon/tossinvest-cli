# Proposal-freeze review: adopt-stockos-full-sdd

- 일시: 2026-07-27
- 범위: proposal.md, design.md, sdd-workflow delta
- 위험 등급: Normal workflow/tooling change, live-order/high-risk production path 영향 없음
- 보이스: Manager + gstack autoplan 관점(CEO, Eng, DX) + Security
- 사용자 전제: StockOS Full SDD 도구와 방법론을 TossOS에 전부 적용하되 재사용 가능성을 확인한다

## Evidence reviewed

- StockOS `.claude/CLAUDE.md`, `.codex/{agents.md,config.toml,hooks.json}`, `.mcp.json`, `.githooks/post-commit`
- StockOS `tools/{logic-map,sdd,sdd-history,pm}`, `scripts/memory-*`, `.sdd/control-graph`, PM registry
- TossOS `AGENTS.md`, `CLAUDE.md`, `docs/WORKFLOW.md`, `Makefile`, `.gitignore`, `openspec/specs/sdd-workflow/spec.md`
- 로컬 CLI 탐지: rtk, openspec, codegraph, codegraphcontext, create-context-graph, gbrain, docker 설치됨; ast-grep 누락
- 로컬 서비스 탐지: StockOS TypeDB·Neo4j 컨테이너 존재, 현재 중지 상태

## CEO review — scope and leverage

판정: **진행**. TossOS에는 OpenSpec과 gate는 있지만 현재 코드 사실을 함수 호출·분기 수준으로 고정하는 단계와 교차 세션 기억이 없다. 도구의 설치만 복제하면 규칙과 실제 산출물이 다시 갈라지므로 설정, hook, doctor, tests, 문서를 하나의 change로 묶는 현재 범위가 맞다.

기존 자산을 최대한 재사용한다. 전역 CLI와 gstack/Superpowers는 다시 vendoring하지 않고 프로젝트 등록만 수행한다. StockOS의 PM 항목, memory, graph index를 복제하는 것은 빠르지만 제품 사실을 오염시키므로 거절한다. Python AST 추출기를 Go 코드에 그대로 쓰는 것도 “도구가 있다”는 거짓 신뢰를 만들므로 거절한다.

### What already exists

| 하위 문제 | 기존 자산 | 결정 |
|---|---|---|
| 계약 | TossOS OpenSpec + strict gate | 유지·확장 |
| 호출 관계 | 전역 CodeGraph | TossOS 별도 index/MCP |
| 보조 문맥 | 전역 CodeGraphContext | TossOS 별도 index |
| 함수 내부 | StockOS Python AST/ast-grep | Go AST/Go rule로 교체 |
| TDD·리뷰 | 전역 Superpowers/gstack | 재사용 |
| 의미 기억 | 전역 GBrain | TossOS source pin |
| 관측 DB | 기존 TypeDB/Neo4j 컨테이너 | 서비스 공유, namespace 격리 |
| PM | StockOS generator/schema | generator 패턴만 재사용 |

### NOT in scope

- StockOS의 거래 전략·PM backlog·기억·인덱스 데이터 이관
- 실거래 엔진 코드, 주문 설정, 운영 토글 변경
- 외부 DB의 가용성을 CI hard gate로 만드는 일
- gstack 또는 Superpowers 자체 소스 vendoring

## Eng review — architecture, tests, failure modes

판정: **조건부 승인 후 조건 반영**. hard evidence와 advisory evidence 경계가 design에 명시되어 있고, 외부 그래프 장애가 개발을 막지 않는 구조다. 아래 조건을 구현에 반영한다.

1. Go AST extractor는 표준 라이브러리만 사용하고 JSON 출력에 파일·함수 range·branch·return·call·assignment를 포함한다.
2. hook은 모든 예외를 흡수해 exit 0을 반환하되, 로컬 JSONL write 실패는 stderr 경고를 남긴다.
3. 이벤트 수집은 토큰·비밀번호·쿠키·이메일 등 민감 패턴을 저장 전에 마스킹한다.
4. `make sdd-check`는 재현 가능한 repo-local 검사만 차단 조건으로 삼는다. TypeDB·Neo4j 실행 여부는 doctor 정보로만 보고한다.
5. `.claude/CLAUDE.md`와 `.codex/agents.md`의 공유 블록은 byte-identical 검사와 생성 명령을 제공한다.
6. PM check는 이 change 도입 이전 활성 change를 bootstrap allowlist로 명시해 기존 병행 작업을 깨뜨리지 않는다.

### Architecture

```text
agent instructions
      │
      ├── OpenSpec ─────────────── make validate/gate
      ├── CodeGraph MCP ────────── hard evidence
      ├── CodeGraphContext/GBrain  advisory
      ├── Go AST + ast-grep ────── Function Logic Map
      └── save/commit hooks
                │
                ├── local masked JSONL (always)
                ├── TypeDB tossos_sdd (best effort)
                └── Neo4j tossos-* sources (best effort)
```

### Failure modes

| 실패 | 영향 | 구조적 대응 | 차단 여부 |
|---|---|---|---|
| CLI 누락 | 절차 실행 불가 | doctor가 설치 명령 출력 | 차단 |
| stale CodeGraph | 잘못된 호출 근거 | sync 후 현재 HEAD 교차검증 | 차단 |
| CodeGraphContext/GBrain stale | 보조 문맥 부정확 | advisory 표기 | 비차단 |
| TypeDB/Neo4j 중단 | 중앙 관측 누락 | 로컬 JSONL 보존 | 비차단 |
| hook 예외 | 편집/커밋 지연 | catch-all + exit 0 | 비차단 |
| agent config drift | 도구별 규칙 불일치 | byte drift 검사 | 차단 |
| StockOS 데이터 혼입 | 잘못된 제품 사실 | TOS namespace와 빈 registry | 차단 |

### Test map

| 경로 | 검증 |
|---|---|
| Go source → AST JSON | Go fixture test |
| source paths → logic-map scaffold | Python unittest + smoke |
| event → masking → JSONL | Python unittest |
| Claude source → Codex mirror | Python unittest + drift negative test |
| PM YAML → generated tracker | Python unittest/check mode |
| tools/config → developer command | `make sdd-doctor`, `make sdd-check` |
| 전체 저장소 | test, vet, validate, gate |

## DX review — installation and daily use

판정: **승인**. 여러 CLI를 개별 기억하게 하지 않고 `make sdd-doctor`, `make sdd-sync`, `make sdd-check` 세 명령으로 노출한다. doctor 오류는 “무엇이 없음 / 왜 필요함 / 설치 또는 복구 명령”을 함께 출력해야 한다.

개발자 여정은 clone → 전역 CLI 확인 → project sync → SDD check → change 작업 순서다. 이미 설치된 CLI를 무조건 재설치하지 않아 환경을 망가뜨리지 않고, 신규 clone에서는 doctor가 정확히 부족한 항목을 알려준다. 생성 산출물 위치와 Function Logic Map 적용/면제 조건은 WORKFLOW 한 곳에서 찾을 수 있게 한다.

## Security review

판정: **승인**. 저장소에 Neo4j/TypeDB 자격증명을 넣지 않는다. `.env*`, 로컬 DB, 이벤트 버퍼, graph index는 ignore한다. hook 입력은 allowlisted metadata만 취하고 민감 문자열을 fail-closed 마스킹한다. 이 change는 `mutating: true` tossctl 명령이나 라이브 주문을 실행하지 않는다.

## Decision audit

| # | 결정 | 수용/거절 | 근거 |
|---|---|---|---|
| 1 | 전역 도구 재사용 | 수용 | 중복 설치·버전 drift 방지 |
| 2 | ast-grep 신규 설치 | 수용 | 실제 누락, Function Logic Map 필수 증거 |
| 3 | Python AST 그대로 복사 | 거절 | Go 분기를 분석하지 못함 |
| 4 | StockOS graph/PM/memory 데이터 복제 | 거절 | 제품별 사실 오염 |
| 5 | TypeDB/Neo4j 서비스 공유 | 수용 | database/source 격리 시 비용 절감 |
| 6 | 외부 graph를 gate 필수 조건화 | 거절 | 신규 clone·offline 개발 비재현 |
| 7 | hook 실패로 편집/커밋 차단 | 거절 | 관측 계층이 계약/실행을 지배하면 안 됨 |
| 8 | Full SDD 규칙을 Claude/Codex/AGENTS에 동기화 | 수용 | 실행 환경별 drift 방지 |

## Review result

**APPROVED**. 위 Eng 조건을 tasks 1~3 구현과 테스트에 반영한다. Requirement 수준 변경이 생기면 이 review를 재실행한다.

Function Logic Map: not-applicable — 이 change는 기존 production 함수 내부 로직을 수정하지 않고
개발 workflow·도구와 신규 leaf 유틸리티만 추가한다. 신규 Go AST leaf는 전용 단위 테스트와
end-to-end smoke로 검증한다.

## Implementation and requirement-change review

- 일시: 2026-07-27
- 독립 검토: gstack performance specialist, adversarial specialist, red-team specialist,
  별도 Codex CLI read-only review
- 추가 Requirement: Function Logic Map diff binding, CodeGraph worktree freshness,
  memory transaction 직렬화
- High-risk production 영향: 없음. 주문·손절·익절·Guardian·원장·인증·체결 코드는 수정하지 않음

초기 검토에서 hook process storm, 전체 history 재처리, memory 동시 기록 유실·승격 secret/symlink,
허위 Function Logic Map 통과, broad Claude auto-approval, stale CodeGraph gate 통과,
CCG credential/output 잔존, TypeDB 실패 event offset 유실, PM duplicate/forward-link 누락을 확인했다.
모두 구현과 negative/regression test로 수정했다.

수정 후 정책은 다음과 같다.

1. save/commit은 로컬 event를 먼저 보존하고 PID/start-identity lock worker가 요청을 병합한다.
2. graph ingest는 byte-offset·100건 batch이며 실패한 TypeDB batch는 checkpoint를 전진하지 않는다.
3. CCG는 실제 credential 없이 private ephemeral workspace에서 실행하고 output 전체를 삭제한다.
4. memory ledger는 file lock과 atomic index replace를 사용하며 모든 원문/ledger를 gate에서 검증한다.
5. Function Logic Map은 diff에서 수정된 기존 함수를 계산하고 source hash·함수·branch coverage를 대조한다.
6. `make sdd-check`는 CodeGraph hard-evidence fingerprint를 차단 조건으로, advisory index와
   외부 TypeDB/Neo4j 가용성은 경고 조건으로 둔다.
7. Claude auto-approval은 exact read-only OpenSpec/CodeGraph/GBrain search 명령으로 축소했다.

후속 적대 검토에서는 change base override, base-file read fail-open, 실제 CodeGraph DB 부재,
legacy event 재마스킹 누락, CCG retry checkpoint, worker lock ownership, memory rollback의 각
fault-injection 경로를 추가로 확인해 수정했다. 마지막 red-team과 adversarial 재검토 결과는
모두 **CLEAN**이다.

현재 검증: Full SDD Python 63 tests와 Go logic-map tests, `make test`, `make vet`,
`make validate` 통과. 공유 TypeDB·Neo4j는 running이며 namespace는 TossOS 전용이다.
`make lint`의 유일한 실패는 이 change가 수정하지 않은 기존
`internal/verifylive/plan.go` gofmt drift다. `make gate CHANGE=adopt-stockos-full-sdd`의
8단계는 모두 통과했고 이후 OpenSpec archive와 PM 완료 동기화를 수행했다.
