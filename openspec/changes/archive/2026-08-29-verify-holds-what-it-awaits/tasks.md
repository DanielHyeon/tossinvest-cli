# Tasks: verify-holds-what-it-awaits

> 이 change 단독으로는 **취소 대상 목록이 바뀌지 않는다**(리뷰 A1). 새로 취소되는 객체도,
> 새로 붙잡히는 객체도 없다 — 만드는 것은 붙잡음을 **표현할 수단**이다. 노출 상한·승인 창·
> 계획 인가·`ErrOutsidePlan` 레일은 전부 무변경이다.

- [x] 1.1 [T] `Artifact.HeldUntil StepID`·`Artifact.ChainID`(가산·`omitempty`).
  `FormatVersion` 무변경. RED: 필드 부재 기록의 round-trip과 정리 목록이 지금과 동일함을
  먼저 고정한다(회귀 기준선).
- [x] 1.2 [T] `heldAfter` — **`Outstanding`이 고른 그 줄**(gate를 지목한 줄)의 index와 gate
  단계의 마지막 줄 index를 비교한다(D2). 실패한 취소가 지목보다 뒤에 있으면 여전히 놓아준다는
  것을 함께 고정 — M22를 조건주문 쪽에 되살리지 않는다는 증거. RED/GREEN.
- [x] 1.3 [T] `cleanupFrom`을 단일 규칙으로 통합 — gate 부재 시 kind별 기본값
  (조건주문 → `conditional-cancel`, 주문 → gate 없음)으로 **기존 판정 보존**. RED: 기존
  기록 fixture로 목록이 바이트 단위로 같음 → GREEN.
- [x] 1.4 [T] **P1 — 붙잡힌 주문은 정리되지 않는다.** RED: `HeldUntil`이 있는 주문이
  지금은 무조건 대상 → GREEN: 대상에서 빠지고 승인 목록에도 안 오른다.
- [x] 1.5 [T] **해제** — gate 단계가 붙잡음 뒤에 terminal 판정을 기록하면 다시 대상이 된다.
  선언보다 앞선 판정은 놓아주지 않는다. RED/GREEN 양방향.
- [x] 1.6 [T] `markDeliberate`가 gate 단계와 사슬을 함께 찍는다(`Deliberate`는 유지 —
  화면·`undeliberate`·`redo.go:114`가 읽는다). 조건주문 3개 호출부가 `conditional-cancel`을
  gate로 선언해 **오늘과 같은 판정**을 명시적으로 표현한다. RED/GREEN.
- [x] 1.7 [T] 정정이 사슬을 잇는다 — 새 조건주문이 같은 `ChainID`를 싣고 붙잡음이 새
  객체로 이어진다. RED/GREEN.
- [x] 1.8 [T] **실기록 회귀** — `capability-verify.jsonl`·`capability-verify-us.jsonl`의
  실제 모양을 재현한 fixture에서 `PendingCleanup`·`Outstanding` 결과가 이 change 전후로
  동일함을 고정한다. 이 change가 실계좌 판정을 바꾸지 않았다는 증거.
- [x] 1.9 [T] 정적 가드 둘(AST) — ① 정리 판정 경로가 **시계를 읽지 않는다**(D3):
  `cleanupFrom`·`heldAfter`·`holdGate`가 `time.Now`·`r.now`·`CreatedAt`·`CancelledAt`을
  참조하지 않는다. ② production 코드가 쓰는 모든 `HeldUntil` 값이 **카탈로그에 실재하는
  단계**다(리뷰 A7: 오타난 StepID는 영원히 settled되지 않아 객체를 영구히 붙잡는다 —
  fail-closed가 조용한 교착이 되는 유일한 경로).
- [x] 2.1 Function Logic Map + Branch Test Map + risk-pattern-report,
  `python3 tools/logic-map/check_analysis.py --change verify-holds-what-it-awaits`
- [x] 2.2 Pre-Edit 선언을 review.md에 기록 (High-risk: 라이브 취소 대상 목록)
- [x] 2.3 적대적 Eng 관점을 포함한 gstack proposal-freeze 리뷰 → review.md
- [x] 2.4 `make sdd-sync && make sdd-check`, `go test ./... -count=1`, `make vet`,
  `make validate`, PM registry allowlist + `tools/pm/test_generate_master_tracker.py` fixture
- [x] 2.5 D5의 잔여 위험(붙잡힌 사슬을 끝내는 콘솔 조작 부재)을 issues.md에 이연 기록
- [x] 2.6 `make gate CHANGE=verify-holds-what-it-awaits`

## 이 change가 하지 않는 것

`conditional-trigger` 단계 구현, child 주문 관측, 발동 지연 시각 4종, `ProtectiveCapability`
산출은 후속 change다. 이 change는 그 단계가 **존재할 수 있는 조건**만 만든다.
