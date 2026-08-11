# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **오늘 배제를 지는 것이 이 함수다.** `n.mu`가 claim과 send를 함께 덮는다.
> `outbox.go:166-168`이 그 사실을 원장 쪽에서 진술하고, 이 함수의 doc `:230-234`가
> 같은 사실을 이쪽에서 진술한다.
>
> **a099는 이 뮤텍스를 안 줄인다.** 줄이는 것은 a092다.
> a099가 바꾸는 것은 **무엇이 배제를 지는가**이지 **누가 잠그는가**가 아니다.
>
> **⚠⚠ 그렇다고 제어 흐름이 그대로인 것은 아니다 — 3판까지 이 산출물이 그렇게 적었다.**
> 오늘 `:252`의 `if !owed`는 **두 갈래**다. a099는 취득 결과를 **셋**으로 가르므로
> (design C3) 이 함수에 갈래가 하나 는다 — `ClaimHeldElsewhere`는 발송을 건너뛰되
> **`ClaimSettled`와 다르게 로그를 남긴다**(D12). doc(task 4.9)만 고치는 것이 아니라
> **task 4.3b가 이 함수의 분기를 바꾼다.** 1라운드 B-P1이 그것을 요구했고,
> D5의 Pre-Edit도 이 함수를 *"doc + claim 경합 결과 분리"*로 적는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record` | `EventKey`·`Type` 비어 있지 않음 | `Notify`가 만든다 | claim이 B1 `:245`로 오류를 준다 |
| `n.Journal` | non-nil | 호출자 — `Notify`가 먼저 검사 | 이 함수는 안 검사한다 |
| `n.remindAfter()` `:244` | `> 0` (기본값 있음) | `notifier.go:280-285` | 0이면 재무장이 꺼진다 |
| **`n.mu`** | **claim과 send 둘 다 덮는다** | `:241` Lock · `:242` defer Unlock | **이것이 오늘의 유일한 배제다** |
| `n.Gate` | nil 허용 | 배선 | B3 `:261`이 검사 |
| `n.Log` | nil 허용 | 배선 | B2 `:258`이 검사 |

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 4 · 이탈 3 · 호출 9 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:245` | claim이 오류 | `n.Log.Error` (B2) · `n.Gate.Block` (B3) | 이탈 `:264` `false, false, err` | 기존 (a097) |
| B2 `:258` | `n.Log != nil` | 로그 한 줄 | — | 기존 |
| B3 `:261` | `n.Gate != nil` | **`Gate.Block(ReasonAlertUndelivered, …)`** — 진입 래치 | — | 기존 |
| B4 `:266` | **`!owed`** | 없음 | 이탈 `:274` `false, false, nil` | 기존 + **a099 R2** |
| — 이탈 `:276` | owed | **`n.deliver(ctx, id, e)` — 네트워크 publish** | `sent, true, nil` | **a099 R2 · R3** |

**B4가 a099의 자리다.** 오늘 `owed`는 *"이 알림을 보내야 하는가"*만 뜻한다.
a099 이후에는 *"이 알림을 **내가** 보내야 하는가"*를 뜻한다.
**분기가 하나 는다** — 취득 결과가 셋으로 갈리기 때문이다(design C3 · task 4.3b).
같은 문서 아래쪽의 Safety conclusion이 그 기준을 적는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:241` / `n.mu.Unlock` `:242` (defer) | **claim + send를 함께 덮는다** | 재진입 불가 — `escalate`가 이 함수 **밖**에 있는 이유 (`notifier.go:217-223`) | `ast.json` calls · defers |
| `n.Journal.ClaimAlertForDelivery` `:244` | claim | 오류면 B1. **오늘 PENDING 행에 아무 표시도 안 남긴다** | 같음 + `claimalertfordelivery` 번들 |
| `n.remindAfter()` `:244` | 재무장 창 | 기본값 `DefaultRemindAfter` | 같음 |
| `n.Log.Error` `:259` | 구조화 로그 | — | 같음 |
| `n.Gate.Block` `:262` | **진입 래치** | **메모리 맵뿐** — `execgw/retry.go:495-510`, 재시작이 지운다 | 같음 |
| **`n.deliver` `:276`** | **네트워크 publish + 재시도 + 정산** | 예산·타임아웃은 a092의 표면 | 같음 |

**live binding — 유일한 호출자**: `notifyCritical`이 **`notifier.go:194`에서 부르고**,
`:217`의 `owed && !sent`가 그 반환을 읽는다 (1라운드 A-T3이 `:194`를 정정했다).

> **⛔⛔ 1라운드 B-P1 — `!owed`에 두 사실을 밀어 넣으면 안 된다.**
>
> 오늘 `owed=false`의 뜻은 **하나**다: *"운영자가 이미 받았고 창이 안 지났다"*
> (`:267-268`의 주석). 그래서 `:274`가 게이트를 안 잠그고 `:217`이 escalate를 안 하는 것이
> **옳다.**
>
> a099가 거기에 *"다른 발송자가 임차를 들고 있다"*를 **같은 값으로** 넣으면,
> 임차를 든 발송자가 죽었을 때 **미전달 critical 알림이 조용히 억제되고 진입이 열린다.**
> 오늘은 임차가 없어 `owed=true`가 나오고 그냥 보낸다 — **a099가 crash 경로를
> 오늘보다 나쁘게 만든다.**
>
> **결론: claim 경합은 `owed=false`가 아니라 별도 결과여야 한다** —
> design C3의 `ClaimHeldElsewhere`다. §4.3b가 그것을 진다.
>
> **⛔⛔ 그리고 2라운드 A-P1이 그 결과의 *처리*를 다시 깼다.**
> 2판은 *"셋째를 받으면 게이트를 잠근다"*라고 적었다. `n.Gate.Block`은 위 표대로
> **메모리 래치이고 자동 해제가 없다.** 그러면 **정상 발송 중**의 경합이 진입을
> 잠그고, 원래 발송자가 **성공적으로 정산해도 푸는 경로가 없다.**
>
> **사용자 결정 2(2026-08-11)가 「경합으로 래치하지 마라」를 정했고,
> 결정 5-1(2026-08-11)이 그 범위를 확정했다** (design C6):
> 이 함수는 `ClaimHeldElsewhere`를 받으면 발송을 건너뛰고 **진입 게이트를
> 아무 방향으로도 안 건드린다** — `Gate.Block`도 `Gate.Clear`도 안 부르고
> 유도도 안 한다. 남기는 것은 **구조화 로그 한 줄**뿐이다(design D12).
>
> **⚠ 3판까지 이 자리는 *"그 유도를 다시 평가한다"*였다.** 그것은 결정 5-1이
> 되돌린 3판 C6이고, 4라운드가 잡았다.

## State mutations and fallbacks

- **뮤텍스 획득/해제**가 이 함수의 side effect 중 가장 중요한 것이다.
- B1 경로에서 **게이트 래치**를 건다 — 원장 쓰기가 실패했을 때 진입을 막는다.
  a097이 더한 것이고 a099는 안 건드린다.
- **폴백 없음.** claim 오류는 발송을 포기하고 게이트를 잠근다 — 닫힌 쪽으로 실패한다.
  **이 방향은 `claimOwed`의 열린 쪽 실패와 반대이고, 둘 다 옳다**:
  전자는 *"원장에 못 썼다"*(아무 기록이 없다)이고 후자는 *"보냈는지 모른다"*다.

## Safety conclusion

- **Safe edit boundary**: a099는 **doc comment `:230-234`**(task 4.9)와
  **취득 결과를 가르는 분기 하나**(task 4.3b)를 건드린다. **뮤텍스 구간과 B1~B3의
  조건은 안 바꾼다.** 편집 후 AST의 defers가 1이고 `n.mu.Lock()`이 `:241`에 그대로면
  a092의 표면을 안 침범한 것이다.
  **branches는 4에서 늘어난다 — 그것이 이 편집의 목적이다.**

  > **⚠⚠ 3판까지 이 줄은 *"branches가 여전히 4면 제어 흐름 무변화"*였다.**
  > 그 기준을 그대로 쓰면 **task 4.3b를 구현한 순간 이 산출물이 위반을 신고한다.**
  > 산출물이 계획과 반대를 말하고 있었다 — 3라운드가 잡았다.
- **High-risk impact**: **yes** — 손절 경로가 이 함수를 지난다(a092의 전체 근거).
  a099가 여기에 코드를 안 더하는 것이 그 이유다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **`ClaimAlertForDelivery`가 이 함수 안에서 UPDATE를 하나 더 하게 된다**
    (a099 §4.3). 그 지연은 이 함수의 분기가 아니라 **호출 `:244`의 비용**이고,
    **§3.4가 재기 전에는 불변식 4를 만족한다고 적지 않는다.**
  - **`n.mu`의 구간은 a099의 대상이 아니다.** 줄이면 이 함수의 이탈 `:276`이
    잠금 밖으로 나가야 하고 그것은 a092의 D0.3a·D0.3b다.
    **a099는 그 재설계의 전제를 만들 뿐 재설계를 하지 않는다.**
