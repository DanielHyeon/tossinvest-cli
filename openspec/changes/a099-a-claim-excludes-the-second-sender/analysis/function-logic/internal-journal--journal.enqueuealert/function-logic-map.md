# Function Logic Map: `Journal.EnqueueAlert`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 번들이 존재하는 이유는 design D13이다.** D13은 *"기록만 하는 호출자는 임차를
> 잡으면 안 된다"*를 주장하고, 그 주장은 **이 함수가 안에서 무엇을 하는지**를 근거로 쓴다.
> **함수 내부를 근거로 쓰는 문서는 AST 산출물이 먼저다.**
>
> 1라운드 A-P5가 이 경로를 지목했고, 1판 설계는 이 경로를 다루지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | 호출자 | 위임한 함수의 트랜잭션 오류로 나온다 |
| `a Alert` | `event_key` 비어 있지 않음 | 호출자 | 위임한 함수가 검사한다 |
| `remindAfter` | **인자가 아니다 — `0` 리터럴** (`:120`) | **이 함수 자신** | — |

**불변식 — a099가 지켜야 하는 것**: *"기록만 하는 호출자는 발송 권한을 얻지 않는다."*

**오늘 이 불변식은 무의미하다.** 발송 권한이라는 것이 원장에 없기 때문이다.
a099가 그것을 만드는 순간 이 불변식이 실질을 갖는다.

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — **분기 0** · 이탈 1 · 호출 1 · 대입 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | **분기가 없다** | — | — | — |
| 이탈 `:121` | 무조건 | 위임한 함수가 행을 쓴다 | `id, err` — **`owed`를 버린다** | 기존 + **a099 R16** |

### ⛔ 분기가 없다는 것이 이 번들의 발견이다

`ast.json`의 `"branches": null`이다. 이 함수에는 **조건이 하나도 없다.**
본문은 대입 하나와 반환 하나뿐이고, 유일한 호출이 `j.ClaimAlertForDelivery`다(`:120`).

**그러므로 이 함수 안에는 임차를 거절할 자리가 없다.**

| 결론 | 근거 |
|---|---|
| 「기록 전용은 임차를 안 잡는다」를 **이 함수 안에서 구현할 수 없다** | 분기가 0이다 |
| 구현 자리는 **`ClaimAlertForDelivery` 안**이다 | 유일한 호출 `:120` |
| 이 함수가 전달할 수 있는 것은 **인자뿐**이다 | 대입 1, 이탈 1 |

D13이 *"임차를 안 잡는 모드로 위임한다"*라고 적은 것은 이 형태와 맞는다.
**다만 "그 사실이 함수 이름과 doc comment에 드러난다"는 부분은 이 함수의 변경이 아니라
`ClaimAlertForDelivery`의 계약 변경이다** — Pre-Edit 대상이 이 함수가 아니라 그쪽이라는 뜻이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.ClaimAlertForDelivery` `:120` | 행을 쓰고 id를 받는다 | 트랜잭션 하나. `_txlock=immediate`로 write lock을 이미 잡는다 | `ast.json` calls |

**live bindings — 프로덕션 호출자 하나뿐이다**:

```
internal/execgw/replay.go:551   _, _ = g.journal.EnqueueAlert(ctx, journal.Alert{
```

**반환 둘을 다 버린다** (`_, _ =`). 이 호출자는 id도 오류도 안 본다.
**그러므로 이 호출자는 발송을 몰고 갈 수 없고, 몰고 갈 의도도 없다.**

테스트 호출자: `internal/journal/outbox_test.go` ·
`internal/journal/a096_claim_for_delivery_test.go` ·
`internal/journal/a097_rearm_is_a_new_episode_test.go`.

## State mutations and fallbacks

- **이 함수 자신은 상태를 안 바꾼다.** 모든 mutation은 `ClaimAlertForDelivery` 안에 있다.
- **`remindAfter = 0`을 넘긴다** (`:120`). doc comment `:116-119`가 이유를 적는다 —
  이 호출자는 발송하지 않으므로 알림 정책이 없고, 남의 몫으로 정산된 행을 재무장하면 안 된다.
- **폴백 없다.**

### a099가 이 함수에 만드는 위험 — D13이 없으면

| 단계 | 무슨 일 |
|---|---|
| 1 | `replay.go:551`이 critical 알림 행을 쓴다 |
| 2 | a099가 `ClaimAlertForDelivery`에 임차를 넣으면 **그 호출이 임차를 잡는다** |
| 3 | 이 호출자는 **발송하지 않는다** — 반환을 버린다 |
| 4 | 그러므로 **아무도 그 임차를 안 푼다** |
| 5 | a098의 배달 루프가 그 행을 **만료(54초)까지 못 집는다** |

**갓 기록된 critical 알림이 자기 임차 뒤에 갇힌다.**

## Safety conclusion

- **Safe edit boundary**: **doc comment와 인자 전달까지.** 본문에 분기를 새로 만들지 않는다.
  임차 여부의 판정은 `ClaimAlertForDelivery` 안에 있어야 한다 —
  **그렇지 않으면 같은 판정이 두 곳에 생기고 갈라진다.**
- **High-risk impact**: **yes** — 원장 쓰기 경로이고(불변식 5),
  이 경로로 기록된 critical 알림이 갇히면 운영자가 못 받는다.
- **덮이지 않은 것을 이름으로 적는다**:
  - `replay.go:551`이 **오류를 버린다.** 원장 쓰기가 실패해도 조용하다.
    **a099가 만든 문제가 아니고 a099가 고칠 것도 아니다.**
    **`not-applicable`: 이 change는 그 침묵을 근거로 아무것도 주장하지 않는다.**
  - `remindAfter = 0`이 **재무장을 끄는 것**과 **임차를 끄는 것**을 동시에 뜻하게 되면
    두 정책이 한 값에 묶인다. D13이 그것을 같은 값으로 묶었는지, 별도 인자로 갈랐는지는
    **§4 GREEN이 정한다.** 지금 시점에 정해지지 않았다 — 그 사실을 적어 둔다.
