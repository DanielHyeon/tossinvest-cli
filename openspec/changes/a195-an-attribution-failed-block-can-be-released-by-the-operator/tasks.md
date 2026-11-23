# a195 Tasks

> **High-risk(대사 해제).** 실계좌 주문을 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호·계좌 금지). 브로커는 가짜 수집기·httptest 만.
> 전체 시험·전체 게이트는 종결 단계에서만. 원장 스키마·설정·토글 변경 없음. 리뷰는 최대 2회차, 리뷰가 낸 새 가드·새 정책은 Manager 결정 전 구현 금지.

## 1. 계약과 증거 (production Go·시험 편집 0)

- [ ] 1.1 Story `STORY-TOS-a195`(FEAT-TOS-005) 등록 확인, `base-commit.txt` 고정(`tools/sdd/capture_change_base.py`).
- [ ] 1.2 CodeGraph hard evidence(버전 명기): `buildReconcileResolveRuntime`·`runEngineReconcileResolve`·`Tracker.Resolve`·`Tracker.Blocks` 의 definition/callers — `analysis/code-context/`.
- [ ] 1.3 `tools/logic-map` AST: proposal·design 이 인용한 함수(`buildReconcileResolveRuntime`, `runEngineReconcileResolve`, `Tracker.Resolve`, `Tracker.Observe`, `blockFromReconcileState`)를 `analysis/cited-ast/` 에 현재 HEAD 로 뽑고, proposal 표의 「조사 인용」 네 줄을 그 AST 와 대조해 틀린 줄은 문서를 고친다(틀리면 Manager 에게 보고).
- [ ] 1.4 편집 표면 세기: `buildReconcileResolveRuntime` 이 a166 census(`internal/accountbudgetruntime` 의 shapes.json·sites.json)나 다른 AST 구조 고정 시험에 있는지, `ATTRIBUTION_FAILED` 거절을 단언하는 기존 시험이 몇 개인지(`cmd/tossctl/engine_reconcile_test.go` 등). census 에 있으면 **멈추고 Manager 에게 돌아온다**.
- [ ] 1.5 편집 대상 기존 함수의 FLM/BTM 번들(pre-edit 판) — `analysis/function-logic/`.
- [ ] 1.6 `openspec validate a195-… --strict` 통과, proposal-freeze 적대 리뷰(작성자와 다른 계열, 최대 2회) — `review.md`.
- [ ] 1.7 Manager 1단계 검증·freeze.

## 2. Pre-Edit

- [ ] 2.1 Pre-Edit 선언을 `review.md` 에 확정(WORKFLOW 양식).

## 3. RED → GREEN

- [ ] 3.1 RED: 새 시험 파일 `cmd/tossctl/a195_reconcile_resolve_attribution_test.go` — design §3 의 1~7.
- [ ] 3.2 GREEN: 새 파일의 원인 판정 함수 + validate 클로저가 그것을 부르게 편집. 기존 시험 중 뒤집히는 단언이 있으면 그 함수의 번들과 함께 고친다.
- [ ] 3.3 편집 뒤 번들의 AST·FLM·BTM 재생성(옛/새 ast 정렬로 재번호).

## 4. 검증

- [ ] 4.1 변이 원장(`analysis/mutation/`) — design §3 의 변이 4개, 변이마다 닿음(양성 대조군) 먼저, 사본 하네스는 무변이 대조군 GREEN·시작 sha 단언 후.
- [ ] 4.2 선택 시험(design §3) 원문, `make lint`, `openspec validate --strict`, `python3 tools/logic-map/check_analysis.py --change a195-…`. 미실행 범위 기록.
- [ ] 4.3 구현 적대 리뷰(다른 컨텍스트) → gstack review 1회 — `review.md`.

## 5. 종결

- [ ] 5.1 Manager 검증(diff·시험 원문 재실행).
- [ ] 5.2 완료 게이트(격리 worktree) → archive → Story 경로 갱신 → push.
- [ ] 5.3 (사람 승인) 운영 반영과 해제 명령 실행 — design §4. 결과는 `review.md` 에 시각·상태만.
