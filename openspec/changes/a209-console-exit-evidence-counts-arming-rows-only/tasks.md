# a209 tasks

> 사용자 목표: 콘솔 주문 화면이 체결 · 거절 · 취소로 끝난 청산 주문을 `ambiguous_exit_evidence` 로 보이지 않게 함(사용자 착수 승인 2026-11-28).
> 위험 등급 Normal(표시 리더). 비례 원칙 — 변이 원장 · 다중 적대 리뷰 not-applicable, 경량 gstack 리뷰 최대 2회.

## 1. 계약 · 증거

- [x] 1.1 기억 회고: ai-memory(세션 기록뿐 — 결정 페이지 0) · Hindsight `tossos-dev-episodes`(직접 선례 0)
- [x] 1.2 Story `STORY-TOS-a209` → FEAT-TOS-004, proposal · design · operator-console spec 델타, `openspec validate --strict`
- [x] 1.3 CodeGraph 1.6.0: `codegraph callers BrokerOrderExitLinks`(생산 호출자 `internal/console/orders.go:716` 하나 · 나머지 시험) ·
      `codegraph affected internal/journal/account_views.go --filter '*_test.go'`(1,267 — journal 패키지 전체가 퍼짐, 심볼 단위 callers 로 좁힘) ·
      a166 census `shapes.json` 에 account_views.go 없음(rg 0)
- [x] 1.4 운영 저널 읽기 전용(`mode=ro`) 행 수 집계 — design F7
- [x] 1.5 `base-commit.txt` = baeddfe31 고정, 편집 전 AST(`analysis/ast-evidence/`)
- [x] 1.6 다른 리더 셋 읽기 전용 점검 — design §4(결함 없음)

## 2. 구현

- [x] 2.1 RED: `a209_exit_links_hook_rows_test.go` 세 시험(`analysis/tdd/red-journal.txt` — FILLED · REFUSED · CANCELLED 가 `ambiguous_exit_evidence` 로 실패, 무장 둘은 통과)
- [x] 2.2 GREEN: `ReadOnly.BrokerOrderExitLinks` 서브쿼리 둘에 `a207HookActions` 제외(`analysis/tdd/green-journal.txt`)
- [x] 2.3 FLM 번들 `ReadOnly.BrokerOrderExitLinks`(revision current) · `check_analysis.py`
- [x] 2.4 선택 시험 `go test ./internal/journal/ ./internal/console/ -count=1` · `make lint` — review.md 「시험 기록」
- [x] 2.5 경량 gstack 리뷰 r1(review.md — testing 1 수리, 적대 5 중 코드 결함 0, 2회차 불필요)

## 3. 사람 승인 뒤(이 change 밖에서 사람이 시작)

- [ ] 3.1 main 병합 · `make image CHANGE=a209-console-exit-evidence-counts-arming-rows-only` · 콘솔 교체
- [ ] 3.2 화면 확인(규칙 13): 콘솔 `/orders` 를 브라우저로 띄워 체결된 청산 주문이 `ambiguous_exit_evidence` 가 아니고 근거를 보이는지,
      콘솔 에러 · 실패 요청 0 인지 확인. 주문 · 토글 · 승인 버튼은 누르지 않음
- [ ] 3.3 design §5 열린 질문 1 · 2 결정(2 는 a207 archive 전에)
- [ ] 3.4 `make gate CHANGE=a209-console-exit-evidence-counts-arming-rows-only` → archive
