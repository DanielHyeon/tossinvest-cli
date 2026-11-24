# a200 2.1 — Pre-Edit 선언 (High-risk)

- 날짜: 2026-11-24, base `28680af3d`(재고정), 작성 Opus 팀메이트. Manager freeze 2026-11-24(`review.md`).
- 대상 symbol(기존 함수 편집 1): `internal/journal/position_projection.go:ProjectPosition`(65-105) — B5(if@101) 안, `enterReconcileInTx`@102 직전 한 자리. 편집 전 번들 `../function-logic/internal-journal--projectposition/`(source sha `5b62d380b860`, merge 뒤에도 같음).
- 새 파일(새 함수만): `internal/journal/a200_adjustment_close_fill_recognition.go` — 판정에 쓸 사실 읽기, 판정 순수 함수, 표지 쓰기. 쓰는 기존 함수는 호출만: `ApplyTx.PendingState` · `ApplyTx.Query` · `appendExitEventFromHook` · `riskcalc.CompareDecimal` · `position.Classify`.
- 호출부: `ProjectPosition` 의 생산 실행 자리 1(`runApplyHooks` Project@284 ← RecordFill fills.go:551; strategy 되메우기 :1348 은 BUY 만이라 매도 거절에 안 닿음).
- 기존 시험(영향): position_projection_test.go 거절 → RECONCILE 군(`TestAnOversellFreezesTheProjectionAndBlocksTheSymbol` · `TestASellWithNothingHeldIsAttributionFailure` 등), exit_state_test.go fill 경로, trade_outcomes_test.go, adjustment_close_test.go, apply_hook_test.go 구조 시험, a130 · a066 시험, reconcile 의 a192 · a195 · converge · mismatch 시험.
- 불변식: ① 인정은 투영 행을 쓰지 않는다(수량 0 그대로) ② 인정은 `reconcile_states` 를 쓰지 않는다 ③ 인정이 아니면 오늘의 `enterReconcileInTx` 경로 그대로(문구 · 원인 무변) ④ 전이표 · `ApplyExitFill` · 수렴 · 대사 해제 · 스키마 무편집 ⑤ 판정은 `resolveFillOrigin` 을 다시 부르지 않고 `origin.IntentID` 를 받는다 ⑥ 새 파일에 `pending_intent_id` 철자 0.
- 실패 시험(RED 계획): design §3-1(본 사고 모양 — 차단 0 · 표지 1), 편입 포지션 같은 모양, 조정 뒤 부분 체결 둘(둘 다 인정), 부분 익절 수량(주문 < 감소분), 실제 Tracker W4 · W5(그 주기 AF 0 · 다음 일치에서 진입 열림). 「계속 거절」 시험은 편집 전에도 통과(현재 동작 고정).
- 오류 계약: 사실 읽기의 DB 오류는 tx 오류(체결 롤백 — 오늘의 차단 쓰기 오류와 같은 방향). 못 읽는 decimal 은 「인정 아님」(오늘의 AF).
- rollback: 새 파일 삭제 + `ProjectPosition` 한 자리 되돌림. 원장 변화는 표지 행(새 action 문자열)뿐 — 스키마 무변이라 옛 이진이 그대로 읽음(옛 이진에서 표지는 콘솔에 원문으로 보일 뿐).
- 설정 영향: 없음(토글 없음 — 안전 불변식 3 「토글 OFF = upstream」 은 해당 없는 버그 수리; 인정은 차단을 덜 거는 방향이지만 수량 · 주문 경로는 무변).
