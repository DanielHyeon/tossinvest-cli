# a083 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `Tracker.Observe`의 Function Logic Map과 Branch Test Map을 **편집 전에**
      작성한다. High-risk 경로이므로 면제 없다.
- [x] 1.2 `Converger.ConvergeQuantities`의 Function Logic Map과 Branch Test Map을
      작성한다.
- [x] 1.3 Pre-Edit 선언을 `review.md`에 기록한다.
- [x] 1.4 CodeGraph로 `AdjustmentApplied`, `Tracker.Observe`, `ConvergeQuantities`의
      정의·호출부·impact를 확인하고 프로덕션 호출자가 `Converger` 하나뿐임을 기록한다.
- [x] 1.5 기존 해제 테스트 4개가 드라이버의 실제 순서를 재현하지 않는다는 것을
      확인해 기록한다.

## 2. RED — 드라이버 (정본)

- [x] 2.1 주기 1에서 수렴 + 같은 diff 관측 → 차단 유지·`Released = 0`이고 credit이
      보존된다. (현재 코드에서도 통과 — 보존을 확인하는 것은 2.2다)
- [x] 2.2 주기 2의 재조회가 일치 → `ADJUSTMENT_APPLIED`로 해제되고 `Released = 1`,
      `reconcile_states.released_at`과 `release_cause`가 기록되며 entry gate에서 그
      심볼의 차단이 사라진다. **현재 코드에서 실패한다.**
- [x] 2.3 주기 2의 재조회가 여전히 불일치 → 해제되지 않고 credit이 소멸한다.

## 3. RED — Tracker

- [x] 3.1 같은 as-of의 관측은 credit을 쓰지도 버리지도 않는다. **현재 코드에서 실패한다.**
- [x] 3.2 더 나중 as-of의 일치 관측이 credit을 쓰고 해제한다.
- [x] 3.3 더 나중 as-of의 불일치 관측이 credit을 소멸시킨다 (기존 규칙 유지).
- [x] 3.4 관측 diff에 as-of가 없으면 일치해도 해제하지 않는다.
- [x] 3.5 credit의 as-of가 관측보다 나중이면 사용하지 않는다.
- [x] 3.6 회귀: 다른 심볼의 credit은 해제하지 않는다.
- [x] 3.7 회귀: 조정 없는 일치는 `AwaitingAdjustment`로 남는다.
- [x] 3.8 회귀: 영구 불일치는 credit으로 풀리지 않는다.
- [x] 3.9 회귀: 재시작(`Restore`)은 credit을 복원하지 않는다.
- [x] 3.10 회귀: persist 실패 시 해제가 커밋되지 않은 credit은 남고 미사용 credit은
      그대로 보존된다.
- [x] 3.11 무관한 심볼이 불일치해도 일치한 심볼의 credit이 소멸하지 않는다 (D2b).
      **현재 코드에서 실패한다.**
- [x] 3.12 3.11 이후 전체가 일치하는 첫 관측에서 그 심볼이 해제된다 (D2b).

## 4. RED — Converger

- [x] 4.1 `ConvergeQuantities`가 diff의 as-of를 credit에 전달한다.
- [x] 4.2 회귀: 재적용된 조정(`Applied = false`)도 credit된다.
- [x] 4.3 회귀: as-of 없는 diff는 수렴 자체를 거부한다(기존 동작).

## 5. GREEN

- [x] 5.1 `AdjustmentCrediter`와 `Tracker.AdjustmentApplied`에 비교 as-of 인자를 추가한다.
- [x] 5.2 `Tracker.adjusted`를 심볼 → 비교 as-of 맵으로 바꾼다.
- [x] 5.3 `Observe`가 사용 가능한 credit만 사용·소멸시키고 나머지를 보존한다 (D2, D4).
- [x] 5.4 `ConvergeQuantities`가 `asOf`를 전달한다.
- [x] 5.5 2·3·4장 전부 GREEN.

## 6. REFACTOR

- [x] 6.1 `mismatch.go` 상단 주석의 해제 규칙 설명에 "어느 재조회인가"를 적는다.
- [x] 6.2 `converge.go`의 "spent by the *next* observation" 주석을 실제 계약으로 고친다.
- [x] 6.3 `reconcileloop.go`의 단계 순서 주석에 credit 수명이 순서에 의존하지 않음을 적는다.

## 7. VERIFY

- [x] 7.1 변이 검증: 사용 조건을 `>=`로 되돌리면 3.1이 RED가 되는지 확인하고 되돌린다.
- [x] 7.2 변이 검증: 미사용 credit도 삭제하도록 되돌리면 2.2가 RED가 되는지 확인하고 되돌린다.
- [x] 7.3 운영 원장의 실제 값(TSLA 0.000154 → 1.000154, as-of 2026-08-05T05:22:25Z)으로
      두 주기를 재생해 해제가 붙는지 확인한다 (읽기 전용 fixture).
- [x] 7.4 upstream 상속 테스트 회귀 없음 (650 green).
- [x] 7.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 7.6 `make gate CHANGE=a083-a-block-releases-itself`.

## 8. 리뷰와 기록

- [x] 8.1 적대적 Eng 리뷰를 받고 `review.md`에 기록한다.
- [x] 8.2 발견 사항을 `issues.md`에 남긴다 — missing-order 차단은 credit 발행자가
      없어 여전히 운영자 전용이라는 점, 차단 evidence 문자열이 첫 관측에 고정되어
      화면이 오래된 숫자를 보여주는 점.

## 10. 개정 2 — 독립 리뷰가 연 blocking (2026-08-05)

- [x] 10.1 D10 RED: 커밋된 수렴이 오류 반환 경로에서 credit되지 않음을 고정한다.
      `a083b_partial_convergence_test.go`.
- [x] 10.2 D10 GREEN: `ConvergeQuantities`의 crediter 호출을 `defer`로 옮겨 앞으로
      생길 반환 경로까지 덮는다.
- [x] 10.3 D8 RED: 재분류된 불일치(ExternalPos)가 차단을 푼다는 것을 고정한다.
- [x] 10.4 D7 RED: credit보다 나중에 생긴 차단을 그 credit이 푼다는 것을 고정한다.
- [x] 10.5 D9 RED: 답할 차단이 없는 credit이 살아남아 미래의 차단을 푼다는 것을 고정한다.
- [x] 10.6 D7·D8·D9 GREEN: `creditAnswers`, `symbolsInDispute`의 ExternalPos,
      `answerableBlockFor`.
- [x] 10.7 회귀: 개정 1의 D2b(무관한 심볼이 credit을 버리지 않는다)가 유지되는지
      확인한다.
- [x] 10.8 `issues.md`에 B1·B2와 I6~I9를 기록한다.
- [ ] 10.9 gstack 독립 리뷰 재실행 + `make gate`.
