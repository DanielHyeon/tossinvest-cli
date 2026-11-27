# a209 — 콘솔 주문 화면의 청산 근거 연결은 무장 행만 센다

## 왜

콘솔 주문 화면(`/orders`)이 체결된 청산 주문을 **전부** `ambiguous_exit_evidence`(모호한 청산 근거)로 보여 준다.
운영자는 어떤 판정이 그 주문을 냈는지 화면에서 볼 수 없다.

원인은 리더 하나다. 엔진이 낸 청산 발의(intent)는 설계상 `exit_events` 행을 **두 개** 가진다.

| 행 | action | 쓰는 곳 | 증거 열 |
|---|---|---|---|
| 무장 행 | 판정 action(운영에서는 `STOP_LOSS_LADDER` 15,572 · `LADDER_PARTIAL` 10,604) | `internal/journal/exit_state.go` `appendExitEventTx` | 20 열 채움 |
| 종결 훅 행 | `PROPOSAL_FILLED` · `PROPOSAL_REFUSED` · `PROPOSAL_CANCELLED` | `internal/journal/apply_hook.go`(체결 훅 · 해제) | NULL, 같은 `proposed_intent_id` |

`ReadOnly.BrokerOrderExitLinks`(`internal/journal/account_views.go`)는 intent 당 행 수를 `action` 구분 없이 센다.
두 행이 세어져 `event_count = 2` 가 되고, 단일 행 LEFT JOIN(`HAVING count(*)=1`)은 NULL 을 돌려주며, 집계 갈래가 그 링크를 모호로 표시한다
(분기 열거: `analysis/function-logic/internal-journal--readonly.brokerorderexitlinks/` AST).

운영 저널 읽기 전용 집계(2026-11-28, `mode=ro`, 행 수만 셈):

| intent 당 (무장 행, 훅 행) | intent 수 |
|---|---|
| (1, 1) | 26,171 |
| (1, 0) | 5 |
| 그 밖((0, n) · (≥2, n)) | 0 |

훅 행 분포: `PROPOSAL_REFUSED` 13,041 · `PROPOSAL_CANCELLED` 12,996 · `PROPOSAL_FILLED` 134. 체결된 청산 주문 134 건이 전부 (1, 1) 모양이므로
화면의 모든 체결 청산 주문이 모호로 보인다. 시험이 못 잡은 이유는 픽스처가 무장 행만 심었기 때문이다
(`account_exit_orders_test.go` `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous` · `a207_compact_evidence_test.go` `TestA207ReadersReadACompactRowAsARepeat`).

## 무엇을 바꾸나

- `ReadOnly.BrokerOrderExitLinks` 의 SQL 두 곳 — 행 수(`event_count`)와 단일 행 LEFT JOIN — 이 **무장 행만** 센다.
  훅 action 목록은 이미 있는 `a207HookActions`(`internal/journal/a207_compact_evidence.go`) 하나를 쓴다. 목록을 베끼지 않는다.
- 무장 행이 둘 이상이면 지금처럼 `ambiguous_exit_evidence` 다(fail-closed 유지).

## 범위 밖(비목표)

- 원장 쓰기 · 스키마 · 주문 · 손절 · 대사 판정은 바꾸지 않는다. 이 리더의 유일한 호출자는 콘솔 표시(`internal/console/orders.go` `ledgerOrderEvidence`)다.
- 압축 무장 행(a207, `effective_source='recomputed_repeat'`)에 연결된 체결 주문의 표시 사유는 바꾸지 않는다 — 지금처럼
  `repeat_of_refused_proposal`. D8 앵커를 따라가 사유를 보여 줄지는 열린 질문 1(design §5).
- `a207HookActions` 의 이름 · 위치는 바꾸지 않는다(a207 함수 본문을 건드리지 않기 위함).
- 같은 `proposed_intent_id` 로 묶는 다른 리더(trade_outcomes · a087 · provenance)는 고치지 않는다 — 읽기 전용 점검 결과 같은 결함이 없다(design §4).

## 위험 등급

Normal(표시 리더). High-risk 경로(주문 · 손절 · 원장 쓰기 · 대사)를 바꾸지 않는다. 비례 원칙: 변이 원장 · 다중 적대 리뷰는 `not-applicable`,
경량 gstack 리뷰 1회(최대 2회). 기존 Go 함수 하나(`ReadOnly.BrokerOrderExitLinks`)의 본문을 바꾸므로 gate ⑤ FLM 번들은 만든다.

## 영향

- 화면: 체결 · 거절 · 취소로 끝난 청산 주문이 자기 무장 행의 근거를 보여 준다.
- 롤백: 커밋 되돌리기 + 이미지 재빌드. 데이터 변경 0.
- 배포 뒤 확인: 콘솔 `/orders` 에서 체결된 청산 주문이 더는 `ambiguous_exit_evidence` 가 아님을 화면으로 확인(tasks 3.x — 사람 승인 뒤).
