# a078 · 설계

## D1 — 한쪽 `NoRung`은 비교 불가가 아니라 최하위 단계다

**결정**: `compareRecoveryStage`에서 "한쪽만 rung이 있으면 `ErrRecoveryIdentity`"
분기를 제거하고, rung이 하나라도 있으면 숫자 비교로 단계를 정한다.

```go
if a.ActiveRung != NoRung || b.ActiveRung != NoRung {
	switch {
	case a.ActiveRung < b.ActiveRung: return -1, nil
	case a.ActiveRung > b.ActiveRung: return 1, nil
	default:                          return 0, nil
	}
}
```

`NoRung == -1`이고 rung 인덱스는 `0` 이상이므로 순서가 이미 맞는다. 새 상수도, 새
매핑도 필요 없다.

**근거**: 이 함수는 `SelectRecoverySnapshot` 안에서만 호출되고, 호출 직전에

```go
if saved.PositionID != recomputed.PositionID ||
   saved.PositionGeneration != recomputed.PositionGeneration ||
   !sameRecoveryPolicy(saved.Policy, recomputed.Policy) ||
   saved.EntryPrice != recomputed.EntryPrice ||
   saved.InitialStop != recomputed.InitialStop || … { return …, ErrRecoveryIdentity }
```

가 통과했다. **두 후보는 같은 정책 정체성의 같은 포지션·같은 세대다.** ratchet 정책은
rung을 활성화하지 않고(`ActiveRung`는 항상 `NoRung`) ladder 정책은 `RatchetLevel`을
쓰지 않으므로, 정체성이 같은 쌍에서 한쪽만 rung을 갖는 상황은 **하나뿐**이다 —
ladder가 아직 첫 rung을 밟지 않았다.

제거되는 것은 안전 가드가 아니라 **이미 위에서 처리된 조건의 중복 오판**이다.

## D2 — 이 함수의 계약을 다시 적는다

`compareRecoveryStage(a, b)`는 "a가 b보다 몇 단계 앞서 있는가"를 답한다.

| 상황 | 값 | 호출자의 해석 |
|---|---|---|
| 어느 한쪽이라도 rung 활성 | rung 숫자 비교 | ladder 단계 |
| 양쪽 다 `NoRung` | `RatchetLevel` 순위 비교 | ratchet 단계, 또는 rung 이전의 ladder |
| 알 수 없는 `RatchetLevel` | `ErrRecoveryIdentity` | **그대로 유지** |

세 번째 줄은 남긴다. 순위표에 없는 level은 실제로 해석할 수 없는 값이고, 정체성
검사로는 걸러지지 않는다.

**양쪽 `NoRung`인 ladder 쌍**은 두 번째 줄로 간다. ladder snapshot의 `RatchetLevel`은
`NONE`이므로 rank 0으로 같고, 단계 0(동률)이 나온다 — 466100이 09:03:35까지 8번
정상 기록된 경로가 정확히 이것이다. 바뀌지 않는다.

## D3 — 호출자 쪽 결과가 어떻게 달라지는가

466100의 실제 값으로:

| | saved (09:03:35) | recomputed (09:03:40, 26,200) |
|---|---|---|
| `CurrentProtection` | 24,929 | **25,700** (rung 0의 StopPct 0 = 본전) |
| `NextProtection` | 25,700 | 26,008.4 |
| `HighWater` | 26,150 | 26,200 |
| `ActiveRung` | -1 | 0 |

수정 후: `protection = +1`, `high = +1`, `stage = +1` → 셋 다 `>= 0` →
`recomputed, RecoveryRecomputed`. **보호선이 24,929에서 25,700(본전)으로 올라간다.**
지금은 그 상승이 반영되지 않은 채 판정 자체가 멈춰 있다 — 손절선이 여전히 진입가보다
3% 아래에 있고 아무도 그것을 확인하지 않는다.

이 값은 원장의 실제 저장 snapshot으로 검증했다 (task 7.3).

역방향(재시작 후 recomputed가 rung을 잃음): `protection <= 0`, `high <= 0`,
`stage = -1` → 셋 다 `<= 0` → `saved, RecoverySavedMonotone`. **기준선은 낮아지지
않는다** — spec "복구된 기준선은 낮아질 수 없다" 그대로다.

축이 엇갈리는 경우(rung은 올랐는데 protection이 낮음): `stage = +1`, `protection = -1`
→ 두 조건 모두 불만족 → `ErrRecoveryAmbiguous` → 격리. **진짜 위험 케이스는 그대로
격리된다.**

## D4 — `stage == 0`의 파생선 검사

```go
if stage == 0 && (saved.NextTarget != recomputed.NextTarget ||
                  saved.NextProtection != recomputed.NextProtection) {
	return …, fmt.Errorf("%w: equal policy stage has different derived next lines", ErrRecoveryIdentity)
}
```

이 검사는 그대로 둔다. 수정 후 `NoRung ↔ rung 0`은 `stage != 0`이 되므로 이 분기에
들어오지 않는다 — 들어와서는 안 된다. 단계가 다르면 파생선이 다른 것이 **정상**이고,
지금 그것을 오류로 보던 것이 결함의 다른 얼굴이다.

## D5 — 기존 격리는 이 change가 풀지 않는다

수정은 앞으로 만들어질 격리를 막을 뿐이다. 466100의 활성 격리는 원장에 남아 있고,
`exitloop`는 계속 그 포지션을 건너뛴다.

해제 경로는 세 가지가 있고 전부 §0.7 사람 판정이다.

1. `Journal.ReleaseExitSnapshotQuarantine(HUMAN_REPAIR)` — **호출자가 없다.** CLI에도
   콘솔에도 배선돼 있지 않다.
2. 콘솔의 `자동관리 해제` → `새 generation 재편입`. 새 세대는 옛 격리와 세대가 달라
   막히지 않는다. 다만 재편입은 **현재가 기준으로 기준선을 다시 만든다** — 보호선이
   바뀌므로 그 자체가 운영 판단이다.
3. 그대로 두고 수동 관리.

a062는 셋 중 무엇도 하지 않는다. `issues.md` I1에 사실만 정리한다.

## D6 — 정지선

- 손절·익절 **계산식**을 바꾸지 않는다. 평가기(`EvaluateLadderSnapshot`)는 손대지 않는다.
- 격리 **생성 조건**을 넓히거나 좁히지 않는다. `record()`는 그대로다.
- 새 상태변경 경로, 새 CLI, 새 route 없음.
- 변경 범위는 `compareRecoveryStage` 한 함수의 분기 하나다.
