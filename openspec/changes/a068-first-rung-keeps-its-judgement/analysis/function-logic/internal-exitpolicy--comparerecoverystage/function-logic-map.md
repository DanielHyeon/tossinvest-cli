# Function Logic Map: `compareRecoveryStage`

- Source: `internal/exitpolicy/recovery.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a068-first-rung-keeps-its-judgement/base-commit.txt`
- 위험 등급: **High-risk** (손절 판정 경로). Pre-Edit 선언은 `review.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a ExitLineSnapshot` | 재계산 후보 (`recomputed`) | `SelectRecoverySnapshot` | 호출 전에 `validateRecoverySnapshot` 통과 |
| `b ExitLineSnapshot` | 저장 후보 (`saved`) | 같음 | 같음 |
| `a.ActiveRung`, `b.ActiveRung` | `NoRung(-1)` 이상 | ladder 평가기 / 원장 | `< NoRung`은 `validateRecoverySnapshot`이 이미 거부 |
| `a.RatchetLevel`, `b.RatchetLevel` | 알려진 5단계 | ratchet 평가기 / 원장 | 순위표에 없으면 `ErrRecoveryIdentity` |

**호출자가 이미 보장하는 것 (recovery.go:101-107, 이 함수 진입 전)**:
`PositionID`, `PositionGeneration`, policy ID/version/digest, `EntryPrice`,
`InitialStop`이 두 후보에서 **모두 일치**하고 양쪽 `InputDigest`가 비어 있지 않다.
따라서 이 함수는 **같은 정책의 같은 포지션·같은 세대**만 본다. 이것이 a068 변경의
안전 근거다.

**호출부**: `SelectRecoverySnapshot` 한 곳뿐 (repo 전역 grep, 현재 HEAD).

## Branches and early returns

번호는 수정 **후**의 AST 기준이다. 수정 전에는 B1과 B2 사이에
"정확히 한쪽만 `NoRung`이면 `0, ErrRecoveryIdentity`" 분기가 하나 더 있었고, a062가
제거한 것이 그것이다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 한쪽이라도 `ActiveRung != NoRung` | 없음 | ladder 단계 비교로 진입 | 2.1, 2.2, 2.3 |
| B2 | rung 숫자 switch | 없음 | 아래 세 갈래 | 2.1, 2.2 |
| B3 | `a.ActiveRung < b.ActiveRung` | 없음 | `-1, nil` (저장 후보가 앞섬) | 2.2 |
| B4 | `a.ActiveRung > b.ActiveRung` | 없음 | `1, nil` (재계산이 앞섬) | 2.1 |
| B5 | 같은 rung | 없음 | `0, nil` | 2.7 |
| B6 | `rank` 클로저의 level 순회 | 없음 | 순위 또는 미발견 | 2.5, 2.6 |
| B7 | 순회 중 level 일치 | 없음 | `i, true` | 2.6 |
| B8 | 한쪽이라도 rank 실패 | 없음 | `0, ErrRecoveryIdentity` — **유지** | 2.5 |
| B9 | ratchet 순위 switch | 없음 | 아래 세 갈래 | 2.6 |
| B10 | `ar < br` | 없음 | `-1, nil` | 2.6 |
| B11 | `ar > br` | 없음 | `1, nil` | 2.6 |
| B12 | 같은 순위 | 없음 | `0, nil` | 2.6 |

**a062가 바꾸는 것은 제거된 분기 하나다.** B1의 참 분기에서 곧바로 B2로 간다.
`NoRung == -1`이고 rung 인덱스는 `0` 이상이므로 B3/B4/B5의 숫자 비교가 이미 올바른
순서를 준다 — 새 상수도 새 매핑도 없다.

**B8은 남긴다.** 순위표에 없는 `RatchetLevel`은 실제로 해석 불가한 값이고 정체성
검사로 걸러지지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rank` (지역 클로저) | ratchet level을 순위 정수로 | 순수, 오류 없음; 미발견은 `false` | AST |

I/O 없음. 시계 없음. 전역 상태 없음. 순수 비교 함수다.

## State mutations and fallbacks

- 아무것도 변경하지 않는다. 두 값을 읽고 `(int, error)`를 반환한다.
- fallback 방향: 판정할 수 없으면 오류를 반환하고 호출자가 격리한다. a062는 그
  방향을 유지하되 **판정할 수 있는 입력을 판정 불가로 오판하던 분기**만 없앤다.

## Safety conclusion

- Safe edit boundary: 한 함수의 한 분기. 평가기(`EvaluateLadderSnapshot`)와
  격리 생성(`record`)은 손대지 않는다.
- High-risk impact: **yes** — 이 함수의 결과가 어떤 보호선이 실효가 되는지를 정한다.
  그래서 Pre-Edit 선언과 적대적 Eng 리뷰를 거쳤다.
- §0.3 손절 즉시성: **회복 방향**. 현재는 이 분기 때문에 포지션이 격리되어 판정이
  완전히 멈춘다. 수정 후 판정이 재개된다.
- §0.9 단방향 안전: 기준선이 낮아지는 경로를 만들지 않는다. 후퇴는 B4를 거쳐
  `saved_monotone`으로 이어지고, 축이 엇갈리면 여전히 `ErrRecoveryAmbiguous`다.
