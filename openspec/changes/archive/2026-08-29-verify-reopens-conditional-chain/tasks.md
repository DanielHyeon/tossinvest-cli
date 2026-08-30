# Tasks: verify-reopens-conditional-chain

> 선행: `verify-clears-leftovers` 완료(됨) — 이 change는 그것이 도입한 정리 가드의 결함을
> 고친다. 후행: 이 change가 `verify-execution-capability` task 2.5의 KR 측정 경로를 다시
> 연다. **2c `add-protection-orders`는 여전히 2.5~2.9 완료 전에 작성하지 않는다.**

- [x] 1.1 [T] 판정과 객체의 선후 — `cleanupFrom`의 조건주문 가드가 `conditional-cancel`의
  최신 항목 색인과 그 조건주문을 만든 항목 색인을 비교한다. 기준은 증거 기록의 순서다.
  RED: `conditional-cancel`이 `skipped`(조건주문 등록 **이전** 기록)인 기록에서 등록된
  조건주문이 정리 대상으로 나옴 → GREEN: 대상이 아니다.
- [x] 1.2 [T] 나중의 취소 실패는 가드를 연다 — 조건주문 등록 **이후**에 기록된
  `conditional-cancel` fail은 그 조건주문을 정리 대상으로 만든다. 잔여물이 영원히 남지
  않는다는 것을 고정한다. RED/GREEN.
- [x] 1.3 [T] fail-closed — 생성 항목이나 취소 단계 항목을 찾지 못하면 정리하지 않는다.
  RED/GREEN.
- [x] 1.4 [T] 2026-07-28 실측 회귀 — `capability-verify.jsonl`의 실제 순서
  (07-26 cancel skipped → 07-28 register pass → cleanup)를 fixture로 고정하고,
  cleanup이 `grLKqi…`를 대상으로 삼지 않음을 증명한다. 이 테스트가 없으면 같은 교착이
  다시 자란다.
- [x] 1.5 [T] 대상이 사라진 전제 단계의 재개통 — `RedoSet`이 세 조건(통과 + deliberate
  조건주문 생성 / 살아 있는 조건주문 없음 / 비-deferred 의존 단계 중 미통과 존재)이 모두
  참일 때만 통과한 단계를 포함한다. RED: KR 기록에서 `conditional-register`가 집합에 없음
  → GREEN: 있다.
- [x] 1.6 [T] 재개통 규칙의 좁음 — 정상 종료한 US 기록(register·persist·modify·cancel 전부
  pass, trigger만 deferred)에서 `conditional-register`가 집합에 **없음**을 고정한다.
  조건주문이 아직 살아 있는 기록에서도 없음을 고정한다. RED/GREEN.
- [x] 1.7 [T] 표와 집합이 같은 답을 한다 — 콘솔 재측정 표의 행 선택이 `RedoSet`과 동일한
  근거를 쓴다. RED: `conditional-register`가 집합에 있는데 표에 없음 → GREEN.
- [x] 1.8 [T] 승인 모델 불변 — 재개통된 단계도 계획에 열거되고 배치 승인 없이는 아무것도
  전송되지 않는다. `Plan.Authorises`가 목록 밖 요청을 계속 거절한다. RED/GREEN.
- [x] 2.1 Function Logic Map + Branch Test Map + risk-pattern-report —
  `internal/verifylive.cleanupFrom`, `internal/verifylive.RedoSet`,
  `internal/console.newFuncMap`(또는 `redoable` 배선 함수).
  `python3 tools/logic-map/check_analysis.py --change verify-reopens-conditional-chain`
- [x] 2.2 Pre-Edit 선언을 review.md에 기록 (High-risk: 라이브 취소·조건주문 등록 경로)
- [x] 2.3 적대적 Eng 관점을 포함한 gstack proposal-freeze 리뷰 → review.md
- [x] 2.4 `make sdd-sync && make sdd-check`, `go test ./... -count=1`, `make vet`,
  `make validate`, PM registry 동기화
- [x] 2.5 `make gate CHANGE=verify-reopens-conditional-chain`

> **인계**: 실제 KR 재측정(2026-07-30 09:00~15:30 KST 장중 창, 콘솔 `[재측정]`, 사람 승인)은
> 이 change의 task가 아니라 `verify-execution-capability` task 2.5다. 이 change의 산출물은
> **그 창을 다시 쓸 수 있게 만드는 것**까지이며, 창을 쓰는 것은 그 change의 몫이다.
> 운영 전제 두 가지는 review.md A6·A7에 있다: KR 보유가 있어야 하고, 승인이 두 번 필요하다.
