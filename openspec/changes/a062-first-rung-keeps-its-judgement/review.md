# a062 · Review

- 날짜: 2026-08-03
- 시점: proposal-freeze (첫 구현 task 착수 전)
- 대상: `proposal.md`, `design.md`, `specs/exit-policy/spec.md`
- 위험 등급: **High-risk** — 손절 판정 경로. WORKFLOW가 요구하는 **적대적 Eng 관점**을
  포함한다.
- 남은 절차: 구현 후 **별도 컨텍스트의 독립 검증**(WORKFLOW §9)은 아직 수행되지 않았다.

## Pre-Edit Gate

```text
- change id / task id: a062-first-rung-keeps-its-judgement / 5.1
- 대상 심볼: internal/exitpolicy.compareRecoveryStage
- 기존 동작 파악 근거:
    · 호출부는 SelectRecoverySnapshot 한 곳뿐 (repo 전역 grep, 현재 HEAD)
    · 호출 직전 정체성 검사: PositionID, PositionGeneration, sameRecoveryPolicy,
      EntryPrice, InitialStop, 양쪽 InputDigest 비존재 (recovery.go:101-107)
    · 기존 테스트: recovery_test.go 2건, recovery_validation_test.go.
      **rung을 활성화한 fixture가 하나도 없다** — 이 전이는 무검증 상태다
    · 운영 원장 재현: 저장 snapshot + recovery policy로 26,200 관측 시
      rung -1 → 0에서 bare ErrRecoveryIdentity 발생 확인
- upstream 상속 테스트 영향: no. 이 함수는 TossOS a042 계열 신설 코드이고
  upstream tossinvest-cli에는 존재하지 않는다. 650 baseline 회귀 없음을 make test로 확인한다
- 실패 테스트 선행 작성: yes (2.1~2.7, 3.1~3.2, 4.1)
- 안전 불변식 §0 위반 여부 검토: 통과
    · §0.3 손절 즉시성: 회복 방향. 현재는 격리로 판정이 **완전히 멈춘다**
    · §0.9 단방향 안전: 기준선을 낮추는 경로 없음. 후퇴는 saved_monotone 유지
    · §0.6 원장 스키마: 변경 없음
```

## 적대적 Eng — 제기된 반론과 처리

### A1. "가드를 지우는 것이다. 무엇을 잃는가?" — **잃는 것 없음, 논증**

이 분기가 막으려던 것은 ratchet snapshot과 ladder snapshot의 혼합 비교다. 그런데
`compareRecoveryStage`에 도달하려면 먼저 `sameRecoveryPolicy(saved.Policy,
recomputed.Policy)`를 통과해야 한다 — ID·version·digest 셋 다 일치해야 한다.

- ratchet 정책은 rung을 활성화하지 않는다 (`ActiveRung`은 항상 `NoRung`).
- ladder 정책은 `RatchetLevel`을 쓰지 않는다 (`NONE` 고정).

따라서 **정체성이 같은 쌍에서 한쪽만 rung을 갖는 경우는 하나뿐**이다: 같은 ladder가
아직 첫 rung을 밟지 않은 상태. 이것은 혼합이 아니라 전진이다.

정체성이 다른 쌍은 이 함수에 도달하지 못한다. 즉 이 분기는 **도달 가능한 모든
입력에서 오판**이었다.

### A2. "그러면 이 분기가 왜 있었나" — **추정, 그리고 그 추정을 테스트로 고정**

a042 설계 당시 "rung 축과 level 축을 섞지 말자"는 의도가 정체성 검사와 중복으로
들어간 것으로 보인다. 그 의도 자체는 옳고, a062는 그것을 **정체성 검사 쪽에 남긴다**.
2.4와 2.5가 그 의도를 테스트로 고정한다 — digest drift, 해석 불가 level 둘 다
`ErrRecoveryIdentity`를 유지한다.

### A3. "격리가 줄면 위험한 것을 놓치지 않나" — **줄어드는 것은 오탐뿐**

수정 후에도 격리되는 경우:

| 상황 | 결과 |
|---|---|
| policy identity drift | `ErrRecoveryIdentity` → 격리 |
| entry / initial stop 불일치 | `ErrRecoveryIdentity` → 격리 |
| 해석 불가 ratchet level | `ErrRecoveryIdentity` → 격리 |
| 같은 단계인데 파생선 다름 | `ErrRecoveryIdentity` → 격리 |
| protection·high·stage가 엇갈림 | `ErrRecoveryAmbiguous` → 격리 |
| snapshot 무결성 실패 | `stored_snapshot_corrupt` → 격리 (별도 경로, 무관) |

사라지는 것은 "같은 정책, 같은 포지션, 모든 축이 함께 오른 정상 전진" 하나다.

### A4. "보호선이 올라가는 변경이면 왜 §0.9를 통과하나" — **더 보수적이다**

지금: 격리 → **판정 없음** → 손절도 익절도 평가되지 않음. 25,700(본전)으로 올라갔어야 할
보호선이 24,929에 멈춰 있고, 그마저 아무도 확인하지 않는다.

수정 후: 판정 재개 → 보호선이 24,929에서 25,700(본전)으로 상승. 손절선이 **올라가는**
방향이고, 이는 §0.9가 요구하는 단방향 안전이다. 낮아지는 경로는
`RecoverySavedMonotone`이 계속 막는다.

### A5. "recomputed가 rung을 잃는 역방향은?" — **saved 유지, 2.2가 고정**

재시작 후 상태가 유실돼 재계산이 rung 미활성으로 돌아가면 `stage = -1`,
`protection <= 0`, `high <= 0` → `saved_monotone`. 저장된 rung n과 그 보호선이
유지된다. 이 경로는 **지금은 오류로 거부돼 격리되던** 것이므로, 수정 후 오히려
spec의 "복구된 기준선은 낮아질 수 없다"가 실제로 동작하기 시작한다.

### A6. "부분 체결·수량 변화가 rung 비교에 끼어들 수 있나" — **없음**

`compareRecoveryStage`는 `ActiveRung`과 `RatchetLevel`만 읽는다. 수량·체결·주문 상태를
보지 않는다. `ProjectedQuantity`는 `validateRecoverySnapshot`이 별도로 검사하고 a062는
그것을 건드리지 않는다.

### A7. "동시 관측자(race)는?" — **경로 불변**

`record()`는 이미 `decision_id` 유일 인덱스와 트랜잭션으로 중복 관측을 막는다. a062는
그 안에서 호출되는 순수 비교 함수만 바꾸므로 동시성 계약이 달라지지 않는다. 다만
**격리가 트랜잭션을 commit한 뒤 오류를 반환**한다는 기존 구조는 그대로이므로,
수정 후에는 그 commit이 일어나지 않는 경로가 늘어난다(= 정상 기록으로 이어진다).

### A8. "이 수정만으로 466100이 살아나나" — **아니다, 명시**

기존 격리 행은 남는다. `issues.md` I1에 해제 경로 세 가지와 각각의 대가를 정리했고,
어느 것도 a062가 자동으로 하지 않는다. **완료 보고에 반드시 포함한다.**

### A9. "다른 정책 종류(RUNNER 등)도 같은 결함인가" — **같다, 그리고 같이 고쳐진다**

`COMMON_LADDER_HYBRID_50`뿐 아니라 모든 ladder 정책이 같은 코드를 지난다. 등록된
common ladder 정책 전부가 첫 rung 전이에서 이 결함을 만난다. 한 곳을 고치면 전부
고쳐진다.

## Manager 셀프리뷰 — 스펙 품질

- ADDED만 사용했다. 기존 "복구된 기준선은 낮아질 수 없다"는 그대로 참이다.
- a062가 정의하는 것은 **무엇이 "안전한 후보 하나를 결정할 수 없다"인가**이고,
  기존 문장은 그 판정 뒤에 무엇을 하는가다. 충돌 없음.
- Scenario 6건이 판정 갈래를 전부 덮는다 — 전진·후퇴·엇갈림·정체성·해석불가·지속.

## 결론

**진행 승인.** 완료 보고에 다음을 반드시 포함한다.

1. 466100의 기존 격리는 풀리지 않는다 (`issues.md` I1). 해제는 운영자 결정.
2. 격리 생성 미로깅(I2)과 알림 미전달(I3)은 a062 범위 밖이며 별도 change가 필요하다.
3. 배포는 a061과 함께 한 번에 한다 (사용자 지시, 2026-08-03).

---

## 구현 후 검토 (2026-08-03, 별도 패스)

**한계 명시**: 구현을 만든 컨텍스트가 diff를 다시 읽은 것이며 WORKFLOW §9의 **별도
세션 독립 검증이 아니다**.

### 변이 검증 (tasks 7.1~7.2)

백업 파일로 복구했고, 복구 후 `diff`로 바이트 동일함을 확인했다.

| 변이 | 대상 | 결과 |
|---|---|---|
| 제거한 분기를 되돌림 | `exitpolicy` 2.1·2.2 | FAIL ✓ |
| 같은 변이 | `journal` 3.1 | FAIL ✓ |
| 같은 변이 | `engine` 4.1 | FAIL ✓ |
| 단계 비교 부호 반전 | `exitpolicy` 2.1·2.2 | FAIL ✓ |

세 계층이 모두 RED가 되는 것이 중요하다 — 비교 함수·원장 기록·관측 루프가 같은
결함의 서로 다른 면을 보고 있다는 증거다.

### RED 관찰 (수정 전)

- 2.1 첫 rung 활성화 → `exitpolicy: recovery candidate identity mismatch`
- 2.2 rung을 잃은 재계산 → 같은 오류 (저장 후보 유지가 **작동한 적이 없다**)
- 2.3 축이 엇갈리는 경우 → `ErrRecoveryAmbiguous`가 아니라 identity mismatch.
  **결함이 진짜 모호성 판정까지 가리고 있었다**는 뜻이고, 수정 후 비로소
  `ErrRecoveryAmbiguous`가 나온다.

### 실제 원장 값 검증 (task 7.3)

466100의 저장 snapshot과 recovery policy를 그대로 읽어 26,200 관측을 평가했다.

```
src=recomputed  err=<nil>
rung        -1 -> 0
protection  24929 -> 25700   (rung 0의 StopPct 0 = 본전)
high water  26150 -> 26200
```

수정 전에는 같은 입력이 bare `ErrRecoveryIdentity`였다. 원장 사본은 계좌 참조를
담고 있어 검증 후 즉시 삭제했다.

### 실행 결과

- `go test ./...` — 78 패키지 **6050건 green** (a061 8건 + a062 11건 포함).
- `go vet ./...` 무결함, `go fmt` 변경 없음.
- `openspec validate --all` 59/59.
- `check_analysis.py` — `compareRecoveryStage`(current)와 인접 함수
  `compareRecoveryDecimal`(base revision, 미편집) 증거 완성.

### 남은 것

- 466100의 기존 격리는 그대로다 (`issues.md` I1). 해제는 운영자 결정이며 현재 코드에
  "격리를 풀고 원래 기준선을 유지하는" 경로 자체가 없다.
- 격리 생성 미로깅(I2), 알림 publisher 미배선(I3)은 별도 change가 필요하다.
