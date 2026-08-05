# a083 · 대사 차단이 스스로 풀린다

- **Feature**: `FEAT-TOS-005` — Position adoption and common exit policy
- **Story**: `STORY-TOS-a083`
- **Spec**: `reconciliation`
- **위험 등급**: **High-risk** (reconciliation, 진입 차단 상태기계). 적대적 Eng 리뷰와 Pre-Edit 선언 필수.

## Why

**수량 불일치로 걸린 차단은 자동으로 풀리지 않는다. 운영 원장에 8건이 최대 30시간째 걸려 있고, 자동 해제는 배포 이래 한 번도 일어난 적이 없다.**

`reconciliation` 스펙은 이미 이렇게 요구한다.

> 비영구 차단의 자동 해제는 **조정 이벤트가 반영된 뒤의 재조회 일치**에만 근거하며
> 신규 release cause(ADJUSTMENT_APPLIED 계열)와 원인 기록을 남긴다.

그리고 시나리오까지 있다 — "조정 반영 후 자동 해제: 조정 이벤트가 반영되고 재조회가
일치하면 비영구 차단이 ADJUSTMENT_APPLIED 원인 기록과 함께 자동 해제된다".

이 시나리오는 프로덕션에서 **도달 불가능**하다.

### 원장 증거 (2026-08-05, 읽기 전용 조사)

```
reconcile_states 활성 8건 — 전부 QUANTITY_MISMATCH, release_cause 전부 NULL
  466100  2026-08-03T23:27:50Z   1일 6시간
  475150  2026-08-04T04:34:50Z   1일 0시간
  IONQ    2026-08-04T07:48:55Z   21시간
  042660  2026-08-04T08:06:38Z   21시간
  NNE     2026-08-04T23:24:06Z    6시간
  103590  2026-08-04T23:29:19Z    6시간
  032820  2026-08-05T00:43:44Z    4시간
  TSLA    2026-08-05T05:22:25Z    0시간

release_cause 히스토그램 (전체 이력):  NULL 8,  OPERATOR 6,  ADJUSTMENT_APPLIED 0
```

`ADJUSTMENT_APPLIED` 해제는 **0건**이다. 지금까지 풀린 6건은 전부 사람이 푼 것이다.

수렴은 정상 동작한다. `position_adjustments`에 차단된 심볼마다 조정이 있다.

```
TSLA    0.000154 → 1.000154   2026-08-05T05:22:25Z
032820  17 → 9                2026-08-05T00:43:44Z
103590  2 → 4                 2026-08-05T00:02:40Z
466100  19 → 24 → 12 → 0      2026-08-05T00:00:35Z ~ 00:52:07Z
```

그리고 최근 사이클 로그는 비교가 **이미 일치**한다고 말한다.

```
{"stable":true,"folded":0,"converged":0,"blocked":8,"released":0, ...}   × 매 60초
```

투영은 계좌와 같아졌고, 조정은 기록됐고, 재조회는 일치하는데, 차단은 남아 있다.

### 왜 도달 불가능한가

`ReconcileDriver.RunOnce`의 한 사이클은 이렇게 돈다.

```
1. 스냅샷 수집 → 안정화 → local state 읽기
2. diff = Compare(snapshot, local)                    ← 불일치 있음
3. ConvergeQuantities(diff)  → 조정 기록 → AdjustmentApplied(심볼)   ← credit 부여
4. Tracker.Refresh
5. Tracker.Observe(diff)     ← **3번과 같은 diff**. 여전히 불일치
```

5번의 `diff`는 3번의 조정 **이전**에 계산된 것이다. 그래서 `diff.BlocksEntry()`가
참이고, 관측 끝에서 `mismatch.go`가 credit을 통째로 버린다.

```go
} else {
    // A completed observation spends its credits. If it still disagreed, the
    // adjustment did not settle it; if releases committed, their work is done.
    t.adjusted = nil
}
```

다음 주기에는 비교가 일치하지만 credit이 없다. `Observe`는 그 차단을
`AwaitingAdjustment`로 분류하고 유지한다. `ConvergeQuantities`는 불일치가 없으니
아무것도 credit하지 않는다. **credit은 만들어진 그 사이클 안에서, 자신을 만든 비교를
관측하는 호출에 의해, 쓰이지도 못한 채 소멸한다.**

`Tracker.persist`가 유일한 자동 해제 경로이므로 결과는 하나다 — 자동 해제 없음.

`converge.go`는 전제를 이렇게 적어 두었다.

> The credit is spent by the *next* observation, whatever that observation finds.

그 "next observation"이 실제로는 같은 사이클의, 같은 비교에 대한 관측이다.
`reconcileloop.go`도 같은 전제를 적어 두었다 — "what the fold wrote shows up in the
*next* cycle's re-read". 두 파일이 같은 것을 전제하고, 어느 쪽도 그것을 강제하지 않는다.

### 왜 테스트가 못 잡았나

`mismatch_test.go`의 해제 테스트 4개는 전부 `tracker.AdjustmentApplied("AAPL")`를
**사이클 밖에서 직접** 부른 뒤 곧바로 *일치하는* diff를 관측한다. 드라이버의 실제
순서 — 불일치 diff로 수렴하고 **같은 불일치 diff를 관측** — 를 재현하는 테스트가
하나도 없다.

`TestAnAdjustmentIsSpentByTheRecheckItAnswers`는 오히려 결함을 사양으로 굳혀 두었다.
"조정 후의 재조회가 여전히 불일치하면 credit은 소멸한다"는 규칙 자체는 옳지만, 실제
루프에서 credit 직후의 관측은 **언제나** 그 credit을 만든 불일치 diff이므로, 이 규칙
아래에서 credit은 항상 헛되이 소멸한다.

### 무엇이 걸려 있나

차단은 **편입**을 막는다. 화면에 "관리 편입 · 대사 차단으로 대기 · 기준선 미생성 ·
엔진 보호 미적용"으로 나오는 상태가 그것이다. 042660과 103590은 21시간·6시간째
손절·익절 기준선 없이 보유 중이다. 계좌를 손으로 거래하면 불일치는 정상적으로
발생하고, 그때마다 그 종목은 사람이 풀어줄 때까지 영구히 보호 밖에 남는다.

## What Changes

**credit에 그것이 계산된 비교의 as-of를 각인하고, 같은 비교를 관측하는 호출은 credit을
쓰지도 버리지도 않게 한다.**

스펙 문장("조정 이벤트가 반영된 뒤의 **재조회**")이 말하는 "재조회"는 조정보다 나중에
수집된 비교다. 지금 코드에는 그 선후를 판정할 수단이 없어서 "다음에 오는 호출"로
근사했고, 드라이버에서 그 근사가 깨진다. `Diff.AsOf`가 이미 있으므로 근사를 사실로
바꿀 수 있다.

- `AdjustmentCrediter.AdjustmentApplied`가 credit의 비교 as-of를 함께 받는다.
- `Tracker.adjusted`가 심볼 → 비교 as-of 맵이 된다.
- `Observe`는 **자신의 diff보다 엄격히 앞선** as-of의 credit만 사용·소멸시킨다.
  같은 as-of의 credit은 보존되고, 다음 주기의 재조회가 쓴다.
- 선후를 판정할 수 없으면(as-of 부재·파싱 불가) credit을 쓰지 않는다 — 차단 유지가
  보수 방향이다.

## Impact

- **Specs**: `reconciliation` (ADDED 1 — 기존 해제 규칙을 좁히지도 넓히지도 않고,
  "어느 재조회인가"만 확정한다)
- **Code**: `internal/reconcile/mismatch.go`, `internal/reconcile/converge.go`
- **Tests**: `internal/reconcile/mismatch_test.go`, `internal/reconcile/converge_test.go`,
  `internal/reconcile/restore_test.go`, `internal/app/engine/reconcileloop_test.go`
- **Schema**: 없음. credit은 의도적으로 in-memory이고 재시작 시 복원되지 않는다는
  기존 계약을 그대로 유지한다.
- **§0 영향**: 청산 경로는 건드리지 않는다. 이 변경은 **진입 차단을 푸는** 방향이므로
  §0.9의 "보수 방향" 판단이 필요하다 — design D3에서 다룬다.

## Non-goals

- 영구 불일치(`reconciliation_mismatch_permanent`)의 자동 해제. 운영자 확인뿐이라는
  기존 SHALL을 그대로 둔다.
- 차단 evidence 문자열의 갱신. 화면이 첫 관측의 숫자를 계속 보여주는 문제는 별건이며
  `issues.md`에 기록한다.
- 재시작 후 credit 복원. 기존 계약(복원하지 않는다)을 유지한다.
