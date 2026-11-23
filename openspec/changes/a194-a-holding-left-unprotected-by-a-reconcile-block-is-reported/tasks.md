# a194 Tasks

> **High-risk(대사·편입·알림).** 실계좌 주문을 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호·계좌 금지).
> 전체 시험·전체 게이트는 종결 단계에서만. 원장 스키마·설정·토글 변경 없음. `internal/obs` 타입 열거 무편집.
> 리뷰는 최대 2회차, 리뷰가 낸 새 가드·새 정책은 Manager 결정 전 구현 금지.

## 1. 계약과 증거 (production Go·시험 편집 0)

- [x] 1.1 Story `STORY-TOS-a194`(FEAT-TOS-005) 등록 확인, `base-commit.txt` 고정.
- [x] 1.2 CodeGraph hard evidence(버전 명기): `judgeHoldings`·`blocked`·`alertUnmanaged`·`alertAdoptionFailed`·`reconcileUnmanagedKey`·`RunOnce`·`Tracker.Blocks`·`Tracker.EntryAllowed`·`Tracker.Permanent` — `analysis/code-context/`.
- [x] 1.3 `tools/logic-map` AST 를 `analysis/cited-ast/` 에 현재 HEAD 로 뽑고 proposal 표의 「조사 인용」 다섯 줄과 design D1·D3·§2 의 주장(`EntryAllowed` 사유로 운영자 전용을 가를 수 있는가, `alertAdoptionFailed` 의 key·등급·기록 입구, `RunOnce` 가 `Observe` 실패 시 `judgeHoldings` 를 건너뛰는가)을 대조한다. 틀리면 문서를 고치고 Manager 에게 보고.
- [x] 1.4 편집 표면 세기: `judgeHoldings` 를 고정하는 구조 시험·census 유무, 알림 수를 단언하는 기존 시험 전수(`reconcileloop_test.go` 등), `exit.position_adoption_failed` 의 사실 값을 열거하는 시험·골든 유무.
- [x] 1.5 `judgeHoldings` 의 FLM/BTM 번들(pre-edit 판).
- [x] 1.6 `openspec validate a194-… --strict` 통과, proposal-freeze 적대 리뷰(다른 계열, 최대 2회) — `review.md`.
- [x] 1.7 Manager 1단계 검증·freeze — Manager freeze 2026-11-23(커밋 `ba41f0059`), 남은 두 건(R2-2 주기 상한·R2-3(a) E5) 반영.

## 2. Pre-Edit

- [x] 2.1 Pre-Edit 선언을 `review.md` 에 확정.

## 3. RED → GREEN

- [x] 3.1 RED: `internal/app/engine/a194_reconcile_block_report_test.go` — design §4 의 1~13(「한 번」「창」은 `Notify` 호출 수·전달 시도 수로, 보유 보고의 중복 방지는 실 Notifier + 실 원장 outbox 하네스로).
- [x] 3.2 GREEN: 새 파일 `internal/app/engine/reconcile_block_report.go` + `judgeHoldings` 편집(차단 분기에서 보유를 모으고 순회 뒤 보고 호출) + 드라이버 구조체 새 필드(유예 시계·차단 보고 래치, 맵은 새 함수 안에서 생성). 편집하는 기존 함수 본문은 `judgeHoldings` 하나 — `NewReconcileDriver`·`Context.ReconcileDriver`·`RunOnce`·`alertUnmanaged` 무편집.
- [x] 3.3 design §4-14 의 기존 시험 무편집 통과 확인. 편집 뒤 번들 AST·FLM·BTM 재생성.

## 4. 검증

- [x] 4.1 변이 원장(`analysis/mutation/`) — design §4 의 변이 9개, 닿음 확인·무변이 대조군 GREEN 뒤.
- [x] 4.2 선택 시험 원문(design §4), `make lint`, `openspec validate --strict`, `check_analysis.py --change a194-…`. 미실행 범위 기록.
- [ ] 4.3 구현 적대 리뷰(다른 컨텍스트) → gstack review 1회 — `review.md`.

## 5. 종결

- [ ] 5.1 Manager 검증(diff·시험 원문 재실행).
- [ ] 5.2 완료 게이트(격리 worktree) → archive → Story 경로 갱신 → push.
- [ ] 5.3 (사람 승인) 운영 반영. 반영 뒤 첫 대사 주기에 활성 운영자 전용 차단의 알림이 나가는지 `alert_outbox` 로 확인(시각·타입만) — `review.md`.
