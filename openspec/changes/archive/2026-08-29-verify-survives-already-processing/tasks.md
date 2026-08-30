# Tasks: verify-survives-already-processing

- [x] 1.1 [T] 정정의 일시적 거절 재시도 — 409 + 본문 `already-processing`에 한해 추가 2회,
  대기는 `retryAfterSeconds`(상한 5초, 주입 sleep seam), `order.amend.retries` 기록.
  RED: 409 1회 후 성공하는 fake에서 order-amend가 fail → GREEN(pass, 재시도 기록).
- [x] 1.2 [T] 접수·조건주문 생성은 여전히 재시도하지 않는다 — 전용 테스트로 고정.
- [x] 1.3 [T] 실패한 단계도 자기가 낸 주문을 취소한다 — 단계 본문이 오류로 끝나면 러너가
  그 단계의 미취소 산출물을 정리한다. `ErrOutsidePlan`·컨텍스트 취소에서는 하지 않는다.
  RED: 정정이 상한까지 실패했을 때 주문이 남고 다음 단계가 `ErrExposureCap` → GREEN.
- [x] 1.4 [T] sweep도 게이트를 지난다 — 계획에 취소 줄이 없으면 전송되지 않는다.
- [x] 2.1 Function Logic Map + `check_analysis.py`
- [x] 2.2 PM registry allowlist + fixture, `make sdd-sync && make sdd-check && make gate`
- [x] 2.3 measurements.md에 M24(정정에도 already-processing) 기록
