# a130 — 편입 뒤 수량이 늘어난 편입 포지션도 성과 행을 남긴다 (an-adopted-trade-keeps-its-outcome)

## Why

2026-10-07 010170 손절(편입 3주 → 63초 뒤 대사 조정 3→6 → 손절 매도 6주 전량 체결 → CLOSED)이 `trade_outcomes` 에 행을 남기지
못했다. 운영 저널 사본(2026-10-12 스냅숏)에서 CLOSED 73건 중 이 결함으로 행이 빠진 것이 **11건**(2026-08-03 부터)이다.

원인(base `233d5d9c`, `tools/logic-map` AST 로 확인 — `analysis/function-logic/`):

- 성과 행의 유일한 생산 경로는 `ApplyExitFill`(internal/journal/apply_hook.go:961) 가 포지션 CLOSED 트랜잭션 안에서 부르는
  `freezeTradeOutcomeTx`(trade_outcomes.go:129) 다. 그 호출(apply_hook.go:1023)은 반환 bool 을 **버린다** — 거절은 로그·이벤트
  없이 사라진다.
- 편입 포지션은 `adoptedRoundTripLegs`(trade_outcomes.go:488) 의 분기 B5(:512) `!sell.priced || sell.quantity.Cmp(adopted) != 0`
  에서 거절된다. `adopted` 는 **편입 시점 수량**(`position_adoptions.quantity`)이라, 편입 뒤 대사 조정(`position_adjustments`)이
  수량을 늘리면 엔진이 늘어난 수량 전부를 팔아도 매도 수량이 편입 수량과 달라 거절된다.
- 같은 비교는 「사람이 일부를 밖에서 처분한 교차 경우」를 거절하려고 둔 것이다(round-5 결정). 그 의도는 유지해야 한다.

## What Changes

1. **편입 포지션의 매수 쪽 다리**(사용자 결정 2026-10-12 1항): 매수 수량 = 편입 수량 + 편입 **뒤** 양(+)의 대사 조정분,
   늘어난 수량의 매입가 = 그 조정의 브로커 평균가(`position_adjustments.new_avg_price`). 주당 최초 위험은 편입 시점 값
   (`exit_states.initial_risk`) 그대로이고, 실현 R 의 분자(손익)와 분모(위험 × 수량)는 **같은 수량**(= 엔진 매도 수량 = 매수 수량)
   으로 계산한다. 엔진 매도 수량이 그 매수 수량과 다르면 지금처럼 행을 남기지 않는다(교차 경우 거절 유지).
2. **조용한 거절 제거**(2항): 성과를 동결하지 못한 사유를 고정 사유 코드로 계산하고, 생산 엔진은 그 거절마다
   `engine.log` 에 `trade_outcome.not_frozen` 줄 하나(종목·시장·인스턴스 번호·사유 코드)를 남긴다. 계좌 참조와 포지션 id
   는 싣지 않는다. 줄은 체결 트랜잭션을 롤백하거나 붙잡지 않는다(쓰기는 별도 goroutine, 패닉은 거둠).
3. **backfill 도구**(3항): `tools/a130-outcome-backfill` — ① 미리보기: 읽기 전용 연결(`journal.OpenReadOnly`)로 성과 행이 없는
   CLOSED 포지션마다 계산 결과 또는 거절 사유를, 이미 있는 행마다 재계산 일치 여부를 출력하고 확인 토큰을 낸다(쓰기 0).
   ② 실행: 사람이 토큰을 붙여 다시 실행하면 계획을 다시 계산해 토큰이 같을 때만 기존 `BackfillTradeOutcome` 으로 채운다.
   **mutating** — 에이전트·시험은 운영 저널에 실행하지 않는다.
4. **콘솔 안내문**(4항): 완결 왕복 거래 표 아래 안내문이 실제 누락 사유(엔진 비관리 · 외부 매도/조정 종결 · 매도 수량 불일치 ·
   평균가 미관측)와 거절 줄 · backfill 도구를 말한다.

## What Does NOT Change

- 엔진 진입 포지션(진입 결정 경로 `roundTripLegs`)의 계산. 편입 뒤 조정이 없는 편입 포지션의 계산(기존 31행 재계산 불변).
- 매도 다리의 귀속(`engineSellOrders` · `sumSellFills`), 거절 시 행을 남기지 않는 동작, 동결 행의 불변(UPDATE 금지 트리거).
- 체결 반영 · 포지션 종결 · exit 상태 완료 · 주문 · 손절 판정. 성과 실패는 지금처럼 체결 트랜잭션을 되돌리지 않는다.
- 저널 스키마(버전 35 그대로 — 새 테이블 · 칼럼 없음). 브로커 호출 0.
- `OutcomeRefused` 싱크가 nil 이면(시험 · 배선 전) 체결 반영의 동작과 읽기는 이 change 이전과 같다.

## Impact

- Affected specs: `trade-analytics`(요구 3개 추가).
- Affected code: `internal/journal/trade_outcomes.go`(`freezeTradeOutcomeTx` · `computeTradeOutcome` · `adoptedRoundTripLegs` ·
  `BackfillTradeOutcome`), `internal/journal/apply_hook.go`(`ApplyHooks` 필드 · `ApplyTx` 필드 · `runApplyHooks` · `ApplyExitFill`),
  새 파일 `internal/journal/trade_outcome_refusal.go` · `trade_outcome_backfill.go`, `internal/app/engine/gateway.go`
  (`bindApplyHooks`), `internal/app/engine/engine.go`(`NewContext` 의 바인딩 호출 한 줄), 새 파일
  `internal/app/engine/trade_outcome_refusal_log.go`, `internal/console/templates_portfolio.go`(안내문), 새 도구
  `tools/a130-outcome-backfill/`.
- High-risk: 원장(체결 반영 트랜잭션 안의 성과 동결)과 엔진 기동 배선. 원장 쓰기 동작은 「거절이던 것이 행이 됨」 한 방향만
  바뀐다. Pre-Edit 선언 + Function Logic Map 필수.
- 병행 a131(로그 계좌 제거)과 파일이 겹치지 않는다(runtime.go · reconcileloop.go · exitloop.go · soak · verifylive · audit 무편집).
- 배포: 사람이 KST 05:00~09:00 에 이미지 재빌드 · 교체(스키마 불변이라 롤백은 이미지 핀 되돌림). backfill 실행은 배포 뒤 사람이.
