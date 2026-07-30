## Context

StockOS의 현재 SDD 정본은 안전 → OpenSpec 계약 → CodeGraph hard evidence →
CodeGraphContext 보조 문맥 → evidence reconciliation → Function Logic Map →
Pre-Edit Gate → TDD → gstack/검증 → archive/PM sync → memory 승격 순서를 고정한다.
또한 활성 change마다 Story가 반드시 하나 존재하며 진행 상태는 파일 증거에서 파생된다.

TossOS는 Go·토스 Open API·upstream 상속이라는 별도 특성이 있어 도구 명령과 안전
불변식은 다를 수밖에 없다. 문제는 이러한 정당한 차이가 아니라, 현재 PM generator가
32개 활성 change를 bootstrap allowlist로 우회하고 `WORKFLOW.md`가 StockOS의 일부
필수 단계와 완료 순서를 축약한 점이다.

## Goals / Non-Goals

**Goals:**

- 활성 OpenSpec change와 TossOS Delivery Story를 예외 없이 1:1로 연결한다.
- Initiative/Epic/Feature/Story의 양방향 참조를 generator가 검증한다.
- Story 진행 상태를 proposal/tasks/archive 파일에서 결정적으로 파생한다.
- StockOS Full SDD 단계와 증거 영수증을 TossOS workflow에 복원한다.
- TossOS의 Go Function Logic Map, 공식 Open API/WTS 경계, upstream 회귀,
  journal·Guardian·배포 특성을 그대로 유지한다.

**Non-Goals:**

- StockOS의 STK portfolio 항목이나 Python/KIS 명령을 TossOS로 복사하지 않는다.
- 이미 archive된 역사 change 전부에 소급 Story를 만들지 않는다. StockOS와 동일하게
  활성 change의 역방향 1:1부터 강제하고 archive history는 별도 파생 기록으로 다룬다.
- production trading code, runtime toggle, 배포 컨테이너를 변경하지 않는다.

## Decisions

### 1. Story가 change보다 먼저 존재한다

새 작업은 먼저 portfolio에 Delivery Story를 만들고 고유 `change_id`와 예정 경로를
기록한 다음 같은 ID의 OpenSpec change를 생성한다. 대안인 change-first+allowlist는
현재와 같은 영구 우회 통로를 만들므로 폐기한다.

### 2. bootstrap allowlist를 제거한다

generator는 활성 change 집합과 Story change 집합이 정확히 같아야 통과한다. Story 하나가
여러 change를 가리키거나 change 하나가 여러 Story에 연결되는 경우도 모두 실패한다.
현재 32개 활성 change는 제품 영역별 TossOS Feature에 Story로 backfill한다.

### 3. 상태는 증거에서 파생한다

portfolio는 `intent`와 계층·계약 연결만 보관한다. 진행 상태는 proposal 존재,
tasks 체크박스, active/archive 위치에서 파생한다. 수동 `status=done`과 실제 active
change가 충돌하는 현재 가능성을 제거한다.

### 4. 방법론과 프로젝트 특성을 분리한다

`docs/WORKFLOW.md`의 단계·READY·증거·완료 계약은 StockOS 정본을 따른다. 하지만 다음
TossOS 항목은 그대로 유지한다.

- Go AST extractor와 `go test`/race/vet
- 토스 공식 Open API 주문 경로와 WTS 조회 전용 경계
- upstream 650 테스트와 OFF 동작 보존
- journal single-writer, Guardian, reconciliation, 세션 안전
- TossOS 전용 GBrain/TypeDB/Neo4j namespace
- Docker Compose와 `make gate` 진입점

### 5. 기존 archive는 소급 생성하지 않는다

StockOS generator와 같은 정책으로 active change의 역방향 1:1만 필수화한다. Story가 이미
archive change를 가리키면 유효성을 검사하지만, Story 없이 끝난 과거 archive는
`archived-change-map` 성격의 역사로 남긴다.

## Risks / Trade-offs

- [32개 Story backfill의 분류 오류] → change proposal의 목적을 기준으로 Feature를 나누고
  generator 양방향 검증과 generated tracker를 사람이 검토한다.
- [한 번에 allowlist를 제거하면 중간 상태가 깨짐] → Story/계층 backfill을 먼저 완성한 뒤
  마지막에 allowlist와 우회 코드를 제거한다.
- [StockOS 문서를 그대로 복사해 TossOS 특성이 손실됨] → 보존 목록을 design과 workflow에
  명시하고 diff 리뷰에서 관련 문단 삭제를 금지한다.
- [수동 status 제거로 UI 출력 변화] → 기존 status 의미를 derived status로 대체하고
  generator 테스트로 proposal/tasks/archive 상태를 고정한다.

## Migration Plan

1. 이번 Story와 OpenSpec change를 먼저 1:1로 만든다.
2. TossOS 제품 Initiative/Epic/Feature를 추가하고 활성 change 32개 Story를 등록한다.
3. generator와 테스트를 RED→GREEN으로 변경해 무예외 1:1과 파생 상태를 강제한다.
4. registry allowlist를 제거하고 generated tracker를 재생성한다.
5. `docs/WORKFLOW.md`를 StockOS 단계와 대조해 수정한다.
6. PM test, OpenSpec strict, SDD sync/check, 대상 gate를 통과한다.

Rollback은 generator·portfolio·workflow·change delta를 한 묶음으로 되돌리는 것이다.
production runtime에는 rollback 대상이 없다.

## Open Questions

없음. 사용자 지시와 StockOS 정본이 Story 1:1 및 기준 절차를 확정했다.
