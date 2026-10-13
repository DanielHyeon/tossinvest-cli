# a130 — 편입 뒤 수량이 늘어난 편입 포지션도 성과 행을 남긴다 (an-adopted-trade-keeps-its-outcome)

> 3판(2026-10-13). 1판 REJECT → 2판(사용자 결정 「거절만」, 2차 리뷰 APPROVE-WITH-FIXES) → 사용자가 2판 방침을 철회하고 **재계산**을 결정해 3판.

## Why

2026-10-07 손절 한 건(편입 3주 → 63초 뒤 대사 조정 3→6 → 손절 6주 전량 체결 → CLOSED)이 `trade_outcomes` 에 행을 남기지 못했고,
그 사실이 어디에도 기록되지 않았다. 운영 저널 사본(2026-10-12)에서 엔진 매도로 닫힌 편입 포지션 중 행이 없는 것이 12건이다.

원인(base `233d5d9c`, `tools/logic-map` AST — `analysis/function-logic/`):

- 성과 행의 유일한 생산 경로 `ApplyExitFill`(apply_hook.go:961) B14 안 `freezeTradeOutcomeTx`(trade_outcomes.go:129) 호출(:1023)이 반환값(거절 여부)을
  **버린다** — 거절은 로그 · 이벤트 없이 사라진다.
- 편입 포지션은 `adoptedRoundTripLegs`(:488) B5(:512)에서 엔진 매도 수량이 **편입 시점 수량**과 다르면 거절된다 — 편입 뒤 대사로 수량이 늘어난 포지션은
  늘어난 만큼 엔진이 팔아도 여기서 거절된다.
- 늘어난 수량의 매입가는 원장에 없다(대사 수렴은 평균가를 보내지 않고 조정 행은 편입 전 cost_basis 를 이어 적는다 — 리뷰 F-1).

## What Changes

1. **묶음별 편입가로 재계산**(사용자 결정 2026-10-13 2차): 편입 뒤 늘어난 묶음마다 매입가 = 그 조정 뒤 엔진의 첫 가격 관측(이 포지션의 exit 판정 기록),
   단 허용 간격 383초(운영 관측 간격 분포 p99 실측) 안에서만. 매수 수량 = 편입 수량 + 묶음 합, 엔진 매도 수량이 그와 같아야 하고, 실현 R = 비용 차감 손익
   ÷ (편입 시점 주당 위험 × 매도 수량) — 분자 · 분모가 같은 수량. 관측 없는 묶음 · 편입 뒤 감소(리뷰 F-3) · 재편입(I-2)은 거절.
2. **거절 사유 기록**: 판정이 고정 사유 코드 타입을 돌려주고, 생산 엔진은 거절된 종결마다 체결 트랜잭션의 **커밋 성공 뒤** `engine.log` 에
   `trade_outcome.not_frozen` 줄 하나(종목 · 시장 · 회차 · 사유)를 남긴다. 계좌 참조 · 포지션 id 없음. 쓰기는 별도 goroutine.
3. **backfill 도구**(`tools/a130-outcome-backfill`): 모집단 = 체결로 닫힌 편입 포지션 중 행 없는 것. 미리보기(읽기 전용, 쓰기 0, 확인 토큰) →
   사람 실행(토큰 일치 + 항목마다 짧은 쓰기 트랜잭션에서 재계산 = 미리 본 값일 때만 INSERT). 스키마가 빌드와 다르면 실행 거절. **mutating — 사람만 실행.**
4. **콘솔 안내문**: 완결 왕복 거래 표 아래 안내문이 실제 누락 사유와 거절 줄을 말한다.

## What Does NOT Change

- 편입 뒤 증가가 없는 편입 포지션과 엔진 진입 포지션의 산식(사본 동결 31행 재계산 31/31 일치). 동결 행의 불변(UPDATE 금지 트리거). 저널 스키마(35).
- 체결 반영 · 포지션 종결 · exit 상태 완료 · 주문 · 손절 판정. 성과 실패는 지금처럼 체결 트랜잭션을 되돌리지 않는다. 브로커 호출 0.
- converge 의 평균가 기록(a132, 별도 change).
- 거절 싱크가 nil 이면 거절 경로의 동작 · 읽기는 이 change 이전과 같다(편입 시기 · 관측 읽기는 편입 종결마다 돈다).

## Impact

- Affected specs: `trade-analytics`(「성과 원시 지표」 MODIFIED — 동결 조건과 거절 · 복구 경로 명시, 요구 3개 추가).
- Affected code: `internal/journal/trade_outcomes.go`(`freezeTradeOutcomeTx` · `computeTradeOutcome` · `adoptedRoundTripLegs` · `BackfillTradeOutcome`),
  `internal/journal/apply_hook.go`(`ApplyHooks` · `ApplyTx` 필드 · `runApplyHooks` 핸들 리터럴 · `ApplyExitFill`), `internal/journal/fills.go`(`RecordFill` 수집
  목록 · 커밋 뒤 보고), 새 파일 `internal/journal/trade_outcome_refusal.go` · `trade_outcome_backfill.go`, `internal/journal/adoption_static_test.go`(읽기
  허용 목록 두 줄), `internal/app/engine/gateway.go`(`bindApplyHooks` 선택 인자), `internal/app/engine/engine.go`(`NewContext` 바인딩 호출), 새 파일
  `internal/app/engine/trade_outcome_refusal_log.go`, `internal/console/templates_portfolio.go`(안내문), 새 도구 `tools/a130-outcome-backfill/`.
- High-risk: 원장(체결 반영 트랜잭션)과 엔진 기동 배선. 원장 쓰기 동작: 편입 뒤 증가 포지션 중 관측이 있는 것은 행이 생기고(새 산식), 편입 뒤 감소 · 재편입은
  행이 사라진다(보수). Pre-Edit + FLM 필수.
- 병행 a131(로그 계좌 제거) · a132(converge)와 파일이 겹치지 않는다.
- 배포: 사람이 KST 05:00~09:00 에 이미지 재빌드 · 교체(운영 저널 32→35 이관은 그 배포의 몫, 다음 창 10-14 05:00). backfill 은 그 뒤 사람이.
