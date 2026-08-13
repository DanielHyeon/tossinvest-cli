# Branch Test Map: `buildGateway`

Source: `internal/app/engine/gateway.go` (**234-355**, 편집 후). AST 기준 branches **5** / returns **7**.

커버리지는 주장이 아니라 측정값이다 —
`go test ./internal/app/engine/ -count=1 -coverprofile`
(**편집 전 exit 0** · **편집 후 exit 0 · 439건 통과**)의 블록 카운트를 함수 범위로 잘라 읽었다.

## 분기 — 편집 전후를 나란히

| Branch | 편집 후 위치 | Condition | 조건 평가 | **오류 팔 실행** | 근거 블록 (편집 후) |
|---|---|---|---|---|---|
| B1 | `:239` | `checkProjectionWired` 오류 | yes | **no** | `239.57-241.3` count=0 |
| B2 | `:261` | `tracker.Restore` 오류 | yes | **no** | `261.45-263.3` count=0 |
| B5 | `:269` — **a098 이 신설한 갈래** | `restoreAlertEntryLatch` 오류 | yes | **no** | `269.71-271.3` count=0 |
| B3 | `:293` | `NewPairedReadinessAdapter` 오류 | yes | **no** | `293.16-295.3` count=0 |
| B4 | `:317` | `execgw.New` 오류 | yes | **no** | `317.16-319.3` count=0 |

> **다섯 분기의 오류 팔이 **전부** 안 덮여 있다 — 신설한 것도 포함해서.**
> 즉 **새 갈래가 이웃보다 나쁘지 않다**. 그것이 이 표가 말할 수 있는 전부이고,
> *"덮여 있다"*는 아니다. 다섯을 덮으려면 조립 단계 협력자를 실패시키는 하네스가
> 필요하고, 그것은 a098 의 범위가 아니다 — **「안 하는 것」에 이름을 붙일 후보다.**

정상 경로(`:243-261`·`:269`·`:276-293`·`:296-317`·`:323-354`)는 전부 count≥1 이다 —
a098 의 R8 테스트 넷이 이 함수를 **프로덕션 조립 그대로** 통과시킨다.

## 신설 leaf `restoreAlertEntryLatch` (`:153-168`)

**새 함수이므로 FLM 은 `not-applicable`** 이지만, 커버리지는 잰다.

| 블록 | 무엇 | count |
|---|---|---:|
| `153.100-155.16` | 진입 ~ `UndeliveredCount` 오류 조건 | 1 |
| **`155.16-157.3`** | **원장 읽기 실패 → 조립 거절** | **0** |
| `158.2-158.22` | `undelivered <= 0` 조건 | 1 |
| `158.22-160.3` | 미전달 0 → **아무것도 안 한다** | 1 |
| `164.2-167.12` | **`gate.Block(...)`** | 1 |

**다섯 중 넷이 실행된다.** 안 덮인 하나는 원장 읽기 실패 경로이고, 위 다섯 분기의
오류 팔과 같은 종류다.

> ⚠ **이 표를 처음 쓸 때 범위를 틀렸다 — 적어 둔다.** `restoreAlertEntryLatch` 를
> `:215-233` 으로 잡고 읽었는데 그 범위는 **`checkProjectionWired`** 였고,
> 그래서 *"모든 블록 count≥1 — 오류 팔까지 덮였다"*는 결과가 나왔다.
> **그럴듯했기 때문에 위험했다.** 함수 범위를 `awk` 로 다시 재서 잡았다.
> **커버리지 숫자는 범위가 맞을 때만 측정값이고, 아니면 남의 함수 이야기다.**

## R8 이 이 함수에서 관측하는 것

| # | 테스트 | 종류 | 어느 자리 |
|---|---|---|---|
| 1 | `TestARestartWithAnUndeliveredCriticalAlertKeepsEntryLatched` | **RED 오늘** | 조립 직후 `Blocks()` — B5 통과 후의 상태 |
| 2 | `TestAnEmptyOutboxLeavesTheGateUnlatchedOnRestart` | 회귀 핀 | leaf 의 `undelivered <= 0` 팔 |
| 3 | `TestTheRestoreNeverClearsALatchItDidNotSet` | 회귀 핀 | **leaf 를 직접** 부른다 — 아래 |
| 4 | `TestTheRestoredLatchDetailCarriesNoAlertContent` | **RED 오늘** | leaf 의 `gate.Block` 인자 (불변식 8) |

> **3번이 조립을 안 거치는 이유.** 첫 판은 `buildGateway` 로 조립한 **뒤에** `Block` 을
> 걸었는데, 복원은 그보다 먼저 끝나 있다 — 그래서 `Clear` 를 부르는 구현을 두고
> 돌려도 **넷이 전부 통과했다.** 관측 순서가 코드 순서와 반대였다.
> 게이트를 **미리 잠근 채** 복원을 돌리는 것이 "안 푼다"를 볼 수 있는 유일한 순서이고,
> 그러려면 leaf 를 직접 불러야 한다.
>
> | derive 뮤테이션 (`undelivered<=0` 에서 `Clear`) | 첫 판 | 고친 판 |
> |---|---|---|
> | 결과 | **통과 (이빨 없음)** | **FAIL** |

## 산출물 근거

- 분기·이탈 열거: `ast.json` — `go run ./tools/logic-map`, 편집 전후 각 1회
- 커버리지: `go test ./internal/app/engine/ -coverprofile`, 편집 전후 각 1회 (둘 다 exit 0)
- 함수 범위: `awk` 로 `^func …` ~ `^}` 를 세어서 확정 (위 ⚠ 의 이유)
- 뮤테이션: `restoreAlertEntryLatch` 의 `undelivered <= 0` 팔에 `gate.Clear` 삽입 → 원복 후
  `grep -c 'gate.Clear'` = **0** 확인
