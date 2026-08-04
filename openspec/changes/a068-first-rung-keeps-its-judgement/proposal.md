# a068 · 첫 rung을 밟아도 판정이 계속된다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a068`
- **Spec**: `exit-policy`
- **위험 등급**: **High-risk** (손절 판정 경로). 적대적 Eng 리뷰와 Pre-Edit 선언 필수.

## Why

**ladder 포지션이 첫 익절선을 넘는 순간 격리되고, 그때부터 손절 판정 대상에서 빠진다.**

2026-08-03 09:03:40, 466100 클로봇에서 실제로 일어났다. 09:03:35까지 8번 정상
판정됐고(전부 `active_rung = -1`), 가격이 첫 익절선 26,162.6을 넘어 rung 0을
활성화하려는 순간 `record()`가 거부하고 포지션을 격리했다. 그 뒤로 엔진은 이 포지션을
**판정하지 않는다**.

```
2026-08-03T09:03:45Z WARN exit.judgement_refused severity=critical symbol=466100
  detail: the stored protection state or the observed price is not usable, so this
          position is not being judged at all: journal: exit snapshot generation is
          quarantined (version 1): ambiguous_recovery
```

## 원인 — 같은 정책 안의 정상 전진을 "비교 불가"로 단정한다

`internal/exitpolicy/recovery.go`:

```go
func compareRecoveryStage(a, b ExitLineSnapshot) (int, error) {
	if a.ActiveRung != NoRung || b.ActiveRung != NoRung {
		if a.ActiveRung == NoRung || b.ActiveRung == NoRung {
			return 0, ErrRecoveryIdentity      // ← 한쪽만 rung이 있으면 오류
		}
```

이 오류가 `SelectRecoverySnapshot`을 통해 그대로 올라가고
`internal/journal/exit_state.go`가 `ambiguous_recovery`로 격리한다. 원장의 evidence는
정확히 그 sentinel의 **맨몸 문자열**(48바이트, 접미사 없음)이라 다른 경로는 배제된다 —
다른 반환점은 전부 `%w: …`로 감싼다.

**이 가드의 의도는 ratchet snapshot과 ladder snapshot을 섞지 않는 것이다.** 그런데
`compareRecoveryStage`에 도달하기 전에 `SelectRecoverySnapshot`이 이미 policy
ID·version·digest, 진입가, 최초 손절, 세대를 전부 일치 확인했다. 그 뒤에 남은 한쪽
`NoRung`은 **"아직 rung을 밟지 않았다"** 는 뜻일 뿐이고, `NoRung == -1 < 0`이므로
비교 가능한 최하위 단계다.

## 재현 (결정적)

저장된 snapshot과 recovery policy를 원장에서 그대로 꺼내 in-process로 평가했다.

| 관측가 | rung 전이 | `SelectRecoverySnapshot` |
|---|---|---|
| 26,150 / 26,100 / 25,900 / 24,900 | -1 → -1 | `recomputed` ✓ |
| **26,200** (첫 익절선 26,162.6 초과) | **-1 → 0** | **`exitpolicy: recovery candidate identity mismatch`** |

정체성 필드는 전부 동일했다 — 같은 포지션·세대, 같은 정책
(`COMMON_LADDER_HYBRID_50` / `1.0.0` / 같은 digest), 같은 진입가 25,700, 같은 최초 손절
24,929, 양쪽 input digest 존재. `ValidateRecoveryDerivation`도 양쪽 다 통과한다.
**오직 rung 전진 하나 때문에 거부됐다.**

## 지금 살아 있는 위험

보유 4종목이 전부 같은 상태다 — ladder 정책, canonical snapshot 있음, `active_rung = -1`.

| 종목 | 다음 익절선 | 고점 |
|---|---|---|
| 439960 | 18,334.18 | 18,060 |
| IONQ | 38.0223 | 37.41 |
| TSLA | 320.67 | 315.19 |
| 042660 | (SEED — 첫 snapshot 기록 시 합류) | — |

**각각 첫 익절선을 넘는 순간 466100과 똑같이 격리되고 손절 판정이 멈춘다.** 수익
구간에 들어선 바로 그 순간 보호가 사라진다는 뜻이다.

042660이 8/2에 rung 0을 정상 활성화한 적이 있는데, 그것은 **첫 ladder 판정이 곧바로
rung 0이어서 비교할 saved가 없었기** 때문이다(`saved_snapshot_json` NULL →
`SelectRecoverySnapshot(nil, …)`은 비교 없이 recomputed를 반환). 익절선 아래에서
한동안 머물다 넘는 종목만 걸린다 — 즉 정상적인 보유 대부분이다.

## What Changes

`compareRecoveryStage`에서 한쪽만 `NoRung`인 경우를 오류가 아니라 **단계 비교**로
처리한다. `NoRung == -1`이고 rung은 `0` 이상이므로 기존 숫자 비교가 이미 올바른 순서를
준다 — 특수 케이스를 지우면 된다.

바뀌지 않는 것:

- `SelectRecoverySnapshot`의 정체성 검사(정책·세대·진입가·최초 손절·digest). 이것이
  ratchet/ladder 혼합을 막는 진짜 가드다.
- 축이 엇갈리는 경우(`protection`은 saved가 높은데 rung은 recomputed가 높은 등)는
  그대로 `ErrRecoveryAmbiguous` → 격리.
- 후퇴(recomputed가 rung을 잃음)는 그대로 `saved_monotone` 선택. 기준선은 낮아지지
  않는다.
- `record()`의 격리 조건·fail-closed 규칙·손상 snapshot 처리 전부 그대로.

## 안전 방향 (§0.9)

**더 보수적인 방향이다.** 지금 상태는 "판정을 아예 하지 않음"이고, 수정 후는 "판정이
계속됨"이다. 손절 즉시성을 회복하는 변경이지 약화하는 변경이 아니다. 표시할 값을
새로 계산하거나 기준선을 낮추는 부분은 없다.

## Non-Goals

- **466100의 기존 격리를 자동 해제하지 않는다.** 수정은 앞으로의 격리를 막을 뿐,
  이미 걸린 것은 그대로 남는다. 해제는 §0.7 사람 판정이며 `issues.md` I1에 선택지를
  정리했다.
- 격리 생성이 로그에 남지 않는 문제(`issues.md` I2)를 고치지 않는다.
- 알림 publisher 미배선(`issues.md` I3)을 고치지 않는다.
- ratchet 경로의 `RatchetLevel` 순위 비교는 건드리지 않는다.

## Impact

- `internal/exitpolicy/recovery.go` — `compareRecoveryStage` 한 분기 제거.
- `internal/exitpolicy/recovery_test.go` — 이 전이를 덮는 테스트가 **하나도 없었다**.
  결함이 무테스트 상태로 출하됐다.
- `exit-policy` spec — ADDED 1건. 기존 "안전한 후보 하나를 결정할 수 없으면 격리"
  문장은 그대로 참이고, a062는 **무엇이 결정 가능한가**를 명시한다.
