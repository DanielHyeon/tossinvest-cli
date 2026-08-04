# a078 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `compareRecoveryStage`의 Function Logic Map과 Branch Test Map을 **편집 전에**
      작성한다. High-risk 경로이므로 면제 없다.
- [x] 1.2 Pre-Edit 선언을 `review.md`에 기록한다.
- [x] 1.3 원장에서 결함을 재현해 기록한다 — 저장 snapshot과 recovery policy로
      26,200 관측 시 rung -1 → 0에서 bare `ErrRecoveryIdentity`가 나오는 것 (읽기 전용).
- [x] 1.4 `compareRecoveryStage`의 호출부가 `SelectRecoverySnapshot` 하나뿐이고
      그 직전에 정체성 검사가 있음을 현재 HEAD에서 확인한다 (D1의 전제).
- [x] 1.5 기존 테스트가 이 전이를 덮지 않는다는 것을 확인해 기록한다.

## 2. RED — exitpolicy

- [x] 2.1 saved rung 미활성 + recomputed rung 0(protection·high 동반 상승) →
      `recomputed` 선택, 오류 없음. 현재 코드에서 **실패**한다.
- [x] 2.2 saved rung n + recomputed rung 미활성(protection·high 하락) →
      `saved_monotone`. 현재 코드에서 **실패**한다.
- [x] 2.3 회귀: rung은 올랐는데 protection이 낮으면 `ErrRecoveryAmbiguous`.
- [x] 2.4 회귀: policy digest·entry·initial stop 불일치는 `ErrRecoveryIdentity`.
- [x] 2.5 회귀: 양쪽 rung 미활성 + 알 수 없는 ratchet level은 `ErrRecoveryIdentity`.
- [x] 2.6 회귀: 양쪽 rung 미활성 + 정상 ratchet level 순위 비교는 지금과 동일하다.
- [x] 2.7 회귀: 같은 단계에서 파생선이 다르면 지금과 같이 거부한다.

## 3. RED — journal (record 경로)

- [x] 3.1 canonical snapshot이 rung 미활성인 상태에서 rung 0 판정을 기록하면
      격리되지 않고 `effective_source = recomputed`로 저장된다. 현재 코드에서 **실패**한다.
- [x] 3.2 회귀: 축이 엇갈리는 판정은 지금과 같이 `ambiguous_recovery`로 격리된다.

## 4. RED — engine (판정 지속)

- [x] 4.1 ladder 포지션이 첫 익절선을 넘은 다음 관측 주기에서도 working set에 남아
      판정된다. 현재 코드에서 **실패**한다.

## 5. GREEN

- [x] 5.1 `compareRecoveryStage`의 한쪽-`NoRung` 오류 분기를 제거한다 (D1).
- [x] 5.2 2·3·4장 전부 GREEN.

## 6. REFACTOR

- [x] 6.1 함수 주석에 계약을 적는다 — 정체성은 호출자가 이미 확인했고, 여기서 답하는
      것은 단계 순서뿐이라는 것 (D2).

## 7. VERIFY

- [x] 7.1 변이 검증: 제거한 분기를 되돌리면 2.1이 RED가 되는지 확인하고 되돌린다.
- [x] 7.2 변이 검증: 단계 비교 부호를 뒤집으면 2.2가 RED가 되는지 확인하고 되돌린다.
- [x] 7.3 466100의 실제 저장값으로 수정 후 결과가 `recomputed`(보호선 24,929 → 25,700)인지 확인한다.
- [x] 7.4 upstream 상속 테스트 회귀 없음.
- [x] 7.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 7.6 `make gate CHANGE=a078-first-rung-keeps-its-judgement`.

## 8. 리뷰와 기록

- [x] 8.1 적대적 Eng 리뷰를 받고 `review.md`에 기록한다.
- [x] 8.2 발견 사항을 `issues.md`에 남긴다 — 466100 격리 해제 경로 부재, 격리 생성
      미로깅, 알림 publisher 미배선.
- [x] 8.3 PM story/tracker 동기화.
