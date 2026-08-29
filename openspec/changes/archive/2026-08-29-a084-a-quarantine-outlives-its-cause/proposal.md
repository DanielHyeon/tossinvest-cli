# a084 · 격리가 자신의 원인보다 오래 살아남지 않는다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a084`
- **Spec**: `exit-policy`
- **위험 등급**: **High-risk** (손절 판정 경로, 원장 스키마). 적대적 Eng 리뷰와 Pre-Edit 선언 필수.

## Why

**격리를 만든 결함은 고쳤는데, 그 결함이 만든 행은 스스로 낫지 않는다. PLTR은 지금
16시간째 손절이 평가되지 않는 상태로 남아 있다.**

a078은 ladder 포지션이 첫 rung을 활성화하는 순간 격리되던 결함을 고쳤다. 그 change의
`issues.md` I1이 남긴 문장이 지금의 문제다.

> a062는 **앞으로의** 격리를 막는다. 이미 원장에 있는 행은 그대로이고 `exitloop`는
> 계속 이 포지션을 건너뛴다. (…) **즉 현재 코드에는 "격리를 풀고 원래 기준선을
> 유지하는" 경로가 없다.** 이것 자체가 설계 공백이고, 후속 change 후보다.

a079가 콘솔 해제 버튼을 만들어 사람이 풀 수 있게 했다. 남은 것은 **기계가 스스로
알아볼 수 있는 경우까지 사람을 기다리는가**다.

### 원장 증거 (2026-08-05, 읽기 전용 조사)

활성 격리 3건. 전부 `ambiguous_recovery`, evidence는 전부 접미사 없는
`exitpolicy: recovery candidate identity mismatch`.

저장된 snapshot과 recovery policy를 **현재 소스로** 재생했다. 세 건 모두 격리 시점이
자신의 첫 rung 교차 바로 뒤이고, 현재 비교기로는 가격축 전 구간에서 한 번도 격리되지
않는다.

```
포지션    저장 high water   첫 rung 교차가   현재 비교기 재생 결과
PLTR      148.55            148.73          전 구간 0건 격리 (rung -1→0, protection 141.717→146.1)
032820    11340             11350.70        전 구간 0건 격리 (rung -1→0, protection 10815.5→11150)
NNE       18.45             18.53           전 구간 0건 격리 (rung -1→0, protection 17.654→18.2)
```

즉 **원인은 이미 사라졌고 결과만 남아 있다.** a078이 예측한 그대로다.

> a078이 배포되면 (…) 세 축이 모두 `>= 0` → `RecoveryRecomputed`가 선택된다. 즉
> **해제가 붙는다.** 그러나 배포만으로는 풀리지 않는다. 격리는 판정 경로 **앞단**에서
> `ErrExitSnapshotQuarantined`로 끊으므로 수정된 비교기에 도달하지 못한다.

끊는 지점은 `ExitObserver.workingSet`이다. 활성 격리가 있으면 그 포지션은
`refused`로 빠지고 판정 함수를 아예 보지 못한다.

### 무엇이 걸려 있나

격리된 포지션은 **판정 대상이 아니다 — 손절 포함**. 화면이 그렇게 적어 두었고
알림도 그렇게 말한다.

```
exit.judgement_refused  severity=critical  symbol=PLTR
  the stored protection state or the observed price is not usable, so this
  position is not being judged at all
```

PLTR은 2026-08-04T13:31:05Z부터 보호선 141.717을 붙든 채 아무도 평가하지 않는다.
가격은 그 사이 148.55 → 159.62로 움직였다. 손절이 걸려야 할 상황이 왔더라도 엔진은
그것을 보지 못했을 것이다.

## What Changes

**격리 행에 그것을 만든 복구 선택기의 개정 번호를 각인하고, 선택기가 바뀐 빌드는
그 격리를 한 번 다시 판정한다.**

격리는 "두 후보 중 하나를 안전하게 고를 수 없다"는 **판정**이다. 판정을 바꿀 수 있는
것은 입력이거나 판정자다. 격리된 포지션은 새 snapshot을 쓰지 않으므로 입력은 바뀌지
않는다 — 그것이 격리의 정의다. 그러므로 답을 바꿀 수 있는 것은 **판정자뿐**이고,
판정자가 바뀐 시점이 유일하게 정당한 재시도 시점이다.

- `exitpolicy`에 복구 선택기의 개정 번호 상수를 둔다. `SelectRecoverySnapshot`의
  의미가 바뀔 때 손으로 올린다.
- `exit_snapshot_quarantines`에 `selector_revision` 열을 additive-nullable로 추가한다.
  NULL은 "이 change 이전에 기록되어 어느 개정에서 나왔는지 모른다"이며 재시도 대상이다.
- `workingSet`은 활성 격리의 각인이 현재 개정과 다르면 그 포지션을 **거르지 않고**
  판정으로 보낸다. 판정은 지금과 똑같이 `SelectRecoverySnapshot`을 다시 돌린다.
- 재판정이 성공하면 격리를 `SELECTOR_REVISED`로 해제하고 판정을 기록한다.
- 재판정이 여전히 거부하면 옛 행을 `SELECTOR_REVISED`로 닫고 **현재 개정으로 각인한
  새 행**을 연다. 같은 개정에서는 다시 재시도하지 않으므로 순환하지 않는다.

이것은 a079가 콘솔 해제에 대해 이미 문서화한 것과 **같은 동작**이다.

> The release does not repair anything and does not weaken any rule: it fills in
> one column, and the next observation runs the very same recovery selection that
> quarantined the position. If the cause still holds, the position is quarantined
> again immediately, under a new version.

사람이 눌렀을 때 안전한 동작이라면, 판정자가 바뀌었다는 기계적으로 확인 가능한 근거가
있을 때도 안전하다. 근거는 오히려 더 강하다 — 사람은 "고쳤을 것"이라 믿고 누르지만,
각인 비교는 실제로 바뀌었다는 사실을 안다.

## Impact

- **Specs**: `exit-policy` (ADDED 1)
- **Schema**: `SchemaVersion` 29 → 30. `exit_snapshot_quarantines.selector_revision`
  additive-nullable 1열. 기존 행 backfill 없음 (§0.6).
- **Code**: `internal/exitpolicy/recovery.go` (개정 상수),
  `internal/journal/schema.go`, `internal/journal/exit_snapshot.go`,
  `internal/journal/exit_state.go`, `internal/app/engine/exitloop.go`
- **Tests**: `internal/exitpolicy`, `internal/journal`, `internal/app/engine`
- **§0 영향**: §0.3 청산 즉시성을 **회복**하는 방향이다. 지금 격리된 포지션은 손절이
  평가되지 않으며, 재판정은 그 평가를 재개한다. §0.6 additive-nullable 준수.

## Non-goals

- 시간 기반 재시도. design D2에서 기각한다.
- 콘솔 해제 버튼의 대체. a079 경로는 그대로 남고, 개정이 같은 격리는 여전히 사람만
  풀 수 있다.
- 격리 사유 문자열의 개선. `ambiguous_recovery`의 evidence가 접미사 없는 sentinel이라
  어느 분기에서 나왔는지 사후에 알 수 없는 문제는 `issues.md`에 기록하고 별건으로 둔다.
- 이미 CLOSED된 포지션(032820, NNE)의 격리 행 정리. 판정 대상이 아니므로 손대지 않는다.
