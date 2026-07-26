# TossOS — Full SDD agent contract (Codex)

이 파일은 `.claude/CLAUDE.md`의 공유 블록 mirror다.

<!-- SDD_SHARED_START -->
## 0. 최상위 안전 불변식

TossOS는 실제 돈을 다루는 자동매매 제품이다. 이 절은 모든 방법론과 도구보다 우선한다.

1. 개발·테스트 중 사람 승인 없는 LIVE 주문 side effect를 만들거나 실행하지 않는다.
2. 대화형 에이전트는 `mutating: true` 명령을 자동 실행하지 않는다.
3. 토글 OFF는 upstream 동작과 동일해야 한다.
4. 손절·비상 청산의 즉시성을 약화하거나 지연하지 않는다.
5. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 경로는 High-risk다.
6. 손절·익절·사이징 변경은 명확한 근거가 있는 보수 방향만 허용한다.
7. 운영 토글 flip과 live 검증은 사람이 직접 승인한다.
8. 시크릿·세션·계좌 개인정보·검증되지 않은 수익성 결론을 기억/그래프/로그에 저장하지 않는다.

## 1. Full SDD 책임 계층

| 계층 | 도구 | 책임 |
|---|---|---|
| 계약(OpenSpec) | OpenSpec | 의도, SHALL, 수용 기준, change scope |
| 증거(CodeGraph hard evidence) | CodeGraph + 현재 HEAD | 정의, callers/callees, impact, 현재 코드 사실 |
| 증거 보조(CodeGraphContext) | CodeGraphContext | 관련 문맥 후보, 교차검증 |
| 함수 내부 증거 | Go AST + ast-grep + Function Logic Map | branch, early return, mutation, fallback, side effect |
| 실행(Superpowers TDD) | Superpowers | RED → GREEN → REFACTOR → VERIFY |
| 게이트(gstack) | gstack | proposal review, code review, security, QA, ship/handoff |
| 기억 | 파일기반 episodic + GBrain | 검증된 과거 학습과 의미 검색 |
| 관측 | SDD Control Graph + Create Context Graph | change·test·review·agent event 관계의 비차단 관측 |
| PM 조정 계층 | TossOS portfolio generator | initiative→epic→feature→story→change 역추적 |

권위 충돌 시 순서는 다음과 같다.

1. 안전 불변식
2. 승인된 OpenSpec
3. 현재 HEAD + 실행 테스트 + CodeGraph
4. 공식 API fixture/사람 승인 실측
5. CodeGraphContext, GBrain, memory, 관측 그래프

보조 문맥과 기억은 advisory다. production 편집의 단독 근거가 될 수 없다.

## 2. 표준 Full SDD 사이클

### Step 0 — 기억 회고

```bash
scripts/memory-recall.sh "<키워드>"
python3 tools/sdd/gbrain_project.py search "<키워드>"
```

기억은 가설로만 취급하고 현재 OpenSpec과 코드로 재검증한다.

### Step 1 — 계약(OpenSpec)

신규 기능·동작 변경·안전 경로 수정은 먼저
`openspec/changes/<change-id>/`의 proposal, design, spec delta, tasks로 정의한다.

```bash
python3 tools/sdd/capture_change_base.py --change <change-id>
openspec validate <change-id> --strict --no-interactive
```

change 생성 직후 구현 전 `base-commit.txt`를 한 번만 고정한다. 첫 구현 task 전에 gstack
proposal-freeze review를 수행하고 `review.md`에 남긴다.
Requirement 수준 변경 시 영향 관점을 다시 검토한다.

### Step 2 — 증거(CodeGraph + 현재 HEAD)

정확한 문자열은 `rg`, 의미/심볼 관계는 CodeGraph를 우선한다.

```bash
codegraph sync .
codegraph query "<symbol-or-question>"
codegraph context <symbol>
codegraph callers <symbol>
codegraph callees <symbol>
codegraph impact <symbol>
```

대상 정의, 직접 caller/callee, 테스트, config/live binding, 영향 파일을 기록한다.
CodeGraph가 현재 HEAD와 다르면 sync 후 파일을 직접 읽는다.

### Step 2.1 — CodeGraphContext 보조 문맥

```bash
codegraphcontext update . --quiet
codegraphcontext report .
```

관련 문맥 후보와 누락 가능성을 찾는 데만 사용한다. CodeGraph hard evidence를 대체하지 않는다.

### Step 2.2 — 증거 조정

- HEAD·CodeGraph·CodeGraphContext 일치: Function Logic Map 입력으로 사용
- HEAD와 인덱스 불일치: 인덱스 갱신 후 재조회
- OpenSpec과 코드 불일치: 스펙 결함 등급을 판단하고 blocking이면 구현 중단
- 기억과 현재 증거 불일치: 기억을 폐기하거나 새 evidence로 갱신

### Step 2.5 — Function Logic Map

기존 함수 내부 로직, High-risk 함수, 복잡한 분기 수정 전에 다음 산출물을 만든다.

```text
openspec/changes/<change-id>/analysis/function-logic/<package>--<function>/
  ast.json
  function-logic-map.md
  branch-test-map.md
  risk-pattern-report.md
```

```bash
OUT=$(python3 tools/logic-map/scaffold_analysis.py \
  --change <change-id> --file <file.go> --func <Receiver.Method>)
go run ./tools/logic-map --file <file.go> --func <Receiver.Method> > "$OUT/ast.json"
python3 tools/logic-map/risk_pattern_report.py <file.go> \
  --output "$OUT/risk-pattern-report.md"
python3 tools/logic-map/check_analysis.py --change <change-id>
```

Map에는 입력/불변식, 모든 분기와 early return, mutation, side effect, timeout/retry,
fallback, live binding, branch별 테스트를 기록한다. 구현 후 다시 추출해 최신화한다.
검사는 persisted `base-commit.txt`(CI의 동일 commit `SDD_BASE_REF`) 대비 수정된 기존 Go 함수를 계산한다. AST의 source
SHA-256·파일·함수·분기와 Markdown map이 일치하지 않거나 대상 함수가 하나라도 빠지면 실패한다.

### Step 3 — 실행(Superpowers TDD)

1. RED: 요구사항·분기와 연결된 실패 테스트를 먼저 확인한다.
2. GREEN: 테스트를 통과시키는 최소 구현을 한다.
3. REFACTOR: 동작과 안전 불변식을 보존하며 정리한다.
4. VERIFY: 대상 테스트, 패키지 테스트, 전체 테스트·vet·race/복구 테스트를 위험도에 맞춰 실행한다.

테스트가 불가능한 문서·설정 작업은 재현 가능한 정적 검사와 smoke 명령을 먼저 실패시키고 고친다.

### Step 4 — 게이트(gstack + repository)

```bash
make sdd-sync
make sdd-check
make test
make vet
make validate
make gate CHANGE=<change-id>
```

High-risk 변경은 Pre-Edit 선언과 적대적 Eng/security review가 추가로 필요하다.
UI는 실제 브라우저 QA, CLI/DX는 처음 사용하는 개발자 흐름과 오류 문구를 검증한다.

### Step 5 — archive + PM sync

모든 task, review, test가 green이고 독립 검증이 끝난 뒤에만 change를 완료한다.

```bash
python3 tools/pm/generate_master_tracker.py --check
openspec archive <change-id>
```

archive 후 story 상태와 생성 tracker를 동기화한다.

### Step 6 — 파일기반 episodic 기억

검증된 재사용 학습만 episode로 작성한다.

```bash
scripts/memory-retain.sh docs/memory/episodes/<episode>.md
python3 scripts/memory_index.py promote <id> --evidence "<test/review evidence>"
```

retain은 항상 episodic이다. canonical 승격은 별도 증거가 있을 때만 명시적으로 한다.
ledger 변경은 lock 아래 원자 처리하며 canonical ID 재사용·강등과 canonical 경로 symlink를 거절한다.
pre-commit과 `make sdd-check`는 모든 memory 원문과 ledger를 검증한다.
GBrain은 advisory 의미 검색이며 OpenSpec·CodeGraph·테스트·gstack을 대체하지 않는다.

## 3. High-risk Pre-Edit Gate

기존 High-risk 함수를 수정하기 직전에 아래를 구현 보고와 Function Logic Map에 기록한다.

```text
Pre-Edit Gate:
- change id / task id:
- 대상 심볼:
- 현재 동작 근거(HEAD, CodeGraph, caller/callee, 테스트):
- CodeGraphContext 보조 문맥 확인: yes/no/not-applicable
- Function Logic Map/Branch Test Map 최신: yes/no/not-applicable
- upstream 상속 영향과 회귀 방지:
- RED 테스트 관측:
- 안전 불변식 위반 여부: 통과/차단
```

확신이 없으면 의존 있음으로 판단하고 호출부·테스트를 더 확인한다.

## 4. Function Logic Map 면제

새 leaf 함수, 문서, 단순 설정, 생성 파일만 바꾸고 기존 함수 내부 로직을 건드리지 않으면
`not-applicable`로 면제할 수 있다. `review.md`에 `Function Logic Map: not-applicable`과
근거를 남긴다.
High-risk 기존 함수는 면제할 수 없다.

## 5. SDD Control Graph와 Create Context Graph

save/commit hook은 `.sdd/history/events/`에 마스킹 JSONL을 쓴다.

- TypeDB endpoint: `localhost:1729`, database: `tossos_sdd`
- Neo4j source: `tossos-agent-saves`, `tossos-commits`
- TypeDB/Neo4j 서비스는 StockOS와 공유 가능하지만 database/source는 공유하지 않는다.
- hook과 graph ingest는 관측 전용·비차단이다.
- 자격증명은 환경변수로만 읽는다. CCG는 credential-free 0700 임시 workspace에서 실행 후 삭제하고,
  Neo4j persistence는 TossOS namespace의 HTTP ingest가 수행한다.
- 관측 그래프는 OpenSpec, CodeGraph, Function Logic Map, 테스트, gstack의 대체물이 아니다.

자동 graph sync는 명시적으로 활성화한다.

```bash
export TOSSOS_TYPEDB_SYNC=1
export TOSSOS_CONTEXT_GRAPH_SYNC=1
```

## 6. PM 조정 계층

TossOS ID만 사용한다: `INIT-TOS` → `EPIC-TOS` → `FEAT-TOS` → `STORY-TOS` → change.
StockOS의 STK backlog를 TossOS 사실로 복사하지 않는다. 활성 change는 story와 1:1로 연결하거나
registry의 명시적 bootstrap 예외여야 한다.

```bash
python3 tools/pm/generate_master_tracker.py
python3 tools/pm/generate_master_tracker.py --check
```

## 7. Enterprise Scale Addendum

- 코드·스펙·테스트·PM·기억의 source of truth와 파생 산출물을 구분한다.
- generated index/tracker/graph는 재생성 가능해야 한다.
- 외부 서비스 장애가 핵심 개발 게이트를 우회하거나 불필요하게 차단하지 않게 한다.
- 에이전트별 설정 공유 블록 drift는 `tools/sdd/check_agent_config_sync.py`로 차단한다.
- 새로운 SDD 도구를 문서에 추가할 때 doctor와 실재 경로 검사를 함께 추가한다.

## 8. 완료 보고 금지 조건

다음이 하나라도 없으면 완료라고 보고하지 않는다.

- 실행 명령과 실제 결과
- 변경 파일과 diff stat
- change/task DoD
- Function Logic Map 적용 또는 면제 근거
- High-risk 영향 및 Pre-Edit Gate 여부
- upstream 테스트 회귀 여부
- PM/check 및 agent config sync 결과
- 남은 위험·외부 서비스 상태

## 9. Skill routing

사용자 요청이 설치된 skill과 맞으면 해당 skill 지침을 먼저 읽고 사용한다.

- 제품 아이디어/범위 → office-hours, plan-ceo-review
- architecture/계획 → plan-eng-review, autoplan
- 버그/원인 조사 → investigate
- QA → qa 또는 qa-only
- 코드 리뷰 → review
- 보안 → cso
- ship/deploy/PR → ship 또는 land-and-deploy
- GBrain 실행 파일 설치/진단 → setup-gbrain, TossOS 데이터 동기화 → `make sdd-sync`

skill, 기억, graph는 안전 불변식과 승인 범위를 확장하지 않는다.
<!-- SDD_SHARED_END -->
