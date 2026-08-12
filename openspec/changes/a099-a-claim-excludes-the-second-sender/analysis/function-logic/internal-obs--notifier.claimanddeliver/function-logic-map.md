# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **proposal 시점 배제를 지던 것이 이 함수였다.** `n.mu`가 claim과 send를 함께 덮었고
> 그것이 유일한 배제였다. **a099 이후 배제는 원장의 임차가 진다.**
> 뮤텍스는 남았고 뜻이 바뀌었다 — doc `:236-240`이 그렇게 적는다:
> *"It excludes senders inside this Notifier and never did anything else."*
>
> **a099는 이 뮤텍스를 안 줄였다.** 줄이는 것은 a092다.
> a099가 바꾼 것은 **무엇이 배제를 지는가**이지 **누가 잠그는가**가 아니다.
>
> **⚠ 그렇다고 제어 흐름이 그대로인 것은 아니었다 — 3판까지 이 산출물이 그렇게 적었다.**
> proposal 시점 `if !owed`는 두 갈래였다. 지금은 처분 셋을 가르는 `switch`(B4~B6)에
> 경합 로그(B7)·탈취 로그(B8)·임차 상실(B9)까지 더해 **분기가 4에서 9로 늘었다.**
> 1라운드 B-P1이 그것을 요구했고 3라운드가 「분기 무변화」 주장을 잡았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record` | `EventKey`·`Type` 비어 있지 않음 | `Notify`가 만든다 | claim이 B1 `:252`로 오류를 준다 |
| `n.Journal` | non-nil | 호출자 — `Notify`가 먼저 검사 | 이 함수는 안 검사한다 |
| `n.remindAfter()` `:251` | `> 0` (기본값 있음) | `notifier.go`의 `remindAfter` | 0이면 시간 기반 재무장이 꺼진다 |
| **`n.claimant()`** `:251` | **공백 아님** | `claimant()` `:323` | **a099가 더한 인자.** 원장이 공백을 거절한다 |
| **`n.mu`** | claim과 send 둘 다 덮는다 | `:248` Lock · `:249` defer Unlock | **한 `Notifier` 안의 배제일 뿐이다** |
| `n.Gate` | nil 허용 | 배선 | B3 `:268`이 검사 |
| `n.Log` | nil 허용 | 배선 | B2 `:265` · B7 `:293` · B8 `:303`이 검사 |

## Branches and early returns

AST 열거 — 분기 9 · 이탈 5 · 호출 18 · 대입 3 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:252` | claim이 오류 | `n.Log.Error` (B2) · `n.Gate.Block` (B3) | `:271` `false, false, err` | 기존 (a097) |
| B2 `:265` | `n.Log != nil` | 로그 한 줄 | — | 기존 |
| B3 `:268` | `n.Gate != nil` | **`Gate.Block(ReasonAlertUndelivered, …)`** — 진입 래치 | — | 기존 (a097) |
| B4 `:273` | 처분으로 갈리는 `switch` | — | — | a099 |
| B5 `:274` | **`ClaimSettled`** | 없음 | `:282` `false, false, nil` | 기존 (a096) |
| **B6 `:283`** | **`ClaimHeldElsewhere`** | **로그만 — 게이트를 양방향으로 안 건드린다** | `:300` `false, false, nil` | **a099** |
| B7 `:293` | `n.Log != nil` (경합) | `engine.alert_claim_held` | — | a099 |
| B8 `:303` | `claim.Stole && n.Log != nil` | `engine.alert_claim_stolen` (Warn) | — | a099 |
| **B9 `:312`** | **`lost` — publish 중에 임차를 잃었다** | 없음 | `:316` `sent, false, nil` | **a099** |
| — 이탈 `:318` | 정상 | **`n.deliver` — 네트워크 publish + 정산** | `sent, true, nil` | a099 |

**B6과 B9가 이 change의 자리다.** proposal 시점 `owed`는 *"이 알림을 보내야 하는가"*만
뜻했다. 지금은 *"이 알림을 **내가** 보내야 하는가"*를 뜻한다.

### B6이 게이트를 안 건드리는 이유

`Gate.Block`은 **메모리 래치이고 자동 해제가 없다.** 경합으로 잠그면
**정상 발송 중**의 경합이 진입을 잠그고, 원래 발송자가 성공적으로 정산해도
푸는 경로가 없다. 반대로 경합에서 풀면 **기계의 판단으로 사람의 잠금을 연다.**

**사용자 결정 2(2026-08-11)가 「경합으로 래치하지 마라」를 정했고,
결정 5-1(2026-08-11)이 그 범위를 확정했다**(design C6):
`Gate.Block`도 `Gate.Clear`도 안 부르고 유도도 안 한다.
남기는 것은 **구조화 로그 한 줄**뿐이다(D12).
`TestContentionNeitherLocksNorUnlocksTheEntryGate`가 그것을 양방향으로 고정한다.

### B9가 `owed=false`인 이유

임차를 잃은 발송자는 **이미 publish했을 수 있다.** 그것을 배달 실패로 돌려주면
호출자(`notifyCritical`)가 격상하고, 운영 모드가 **성공한 배달 때문에** 조인다.
*"losing a race to another sender is not that"*(`:313-315`).

> **⛔ 1라운드 B-P1 — `!owed`에 두 사실을 밀어 넣으면 안 된다.**
> a099가 *"다른 발송자가 임차를 들고 있다"*를 `ClaimSettled`와 **같은 값**으로 넣었으면,
> 임차를 든 발송자가 죽었을 때 미전달 critical 알림이 조용히 억제되고 진입이 열린다.
> **`ClaimHeldElsewhere`가 별도 처분인 이유가 그것이다**(design C3).
>
> **⛔ 2라운드 A-P1이 그 결과의 *처리*를 다시 깼다.** 2판은 *"셋째를 받으면 게이트를
> 잠근다"*였다. 위의 「B6이 게이트를 안 건드리는 이유」가 그 정정이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:248` / `n.mu.Unlock` `:249` (defer) | claim + send를 함께 덮는다 | 재진입 불가 — `escalate`가 이 함수 **밖**에 있는 이유 | `ast.json` calls · defers |
| `n.Journal.ClaimAlertForDelivery` `:251` | **claim — 원장의 임차** | 오류면 B1. 「못 잡았다」는 오류가 아니라 처분이다 | 같음 + `claimalertfordelivery` 번들 |
| `n.remindAfter()` `:251` | 재무장 창 | 기본값 `DefaultRemindAfter` | 같음 |
| **`n.claimant()` `:251`** | **임차에 이름을 넣는다** | 배제는 이름이 아니라 토큰이 진다 (`:321-322`) | 같음 |
| `n.Log.Error` `:266` | 구조화 로그 | — | 같음 |
| `n.Gate.Block` `:269` | **진입 래치** | 메모리 맵뿐 — 재시작이 지운다 | 같음 |
| `n.Log.Event` `:294` | **`engine.alert_claim_held`** | — | 같음 |
| `n.Log.Warn` `:305` | **`engine.alert_claim_stolen`** | — | 같음 |
| `claimAgeMS` `:298`, `:309` · `n.now` `:298`, `:309` | 임차 나이를 밀리초로 | 시계는 `n.Clock` 또는 시스템 | 같음 |
| **`n.deliver` `:311`** | **네트워크 publish + 재시도 + 정산** | 예산·타임아웃은 a092의 표면. **`lost`를 두 번째 값으로 돌려준다** | 같음 |

**live binding — 유일한 호출자**: `notifyCritical`. 반환 `owed && !sent`가 격상을 정한다.

**토큰은 `deliver`로만 간다** (`:311`). 로그 어디에도 안 나온다 —
`ClaimResult.Token`의 주석이 그 이유를 적는다.

## State mutations and fallbacks

- **뮤텍스 획득/해제**가 이 함수의 side effect 중 가장 중요한 것이다.
- B1 경로에서 **게이트 래치**를 건다 — 원장 쓰기가 실패했을 때 진입을 막는다.
  a097이 더한 것이고 a099는 안 건드렸다.
- **B6·B9 경로는 게이트를 아무 방향으로도 안 건드린다.** §5.3b가 그것을
  diff로 확인한다 — `Gate.Block`/`Gate.Clear` 호출 자리는 여전히 다섯이고
  a099의 diff에 그 다섯 줄이 없다.
- **폴백 없음.** claim 오류는 발송을 포기하고 게이트를 잠근다 — 닫힌 쪽으로 실패한다.
  **이 방향은 임차 술어의 열린 쪽 실패와 반대이고, 둘 다 옳다**:
  전자는 *"원장에 못 썼다"*(아무 기록이 없다)이고 후자는 *"보냈는지 모른다"*다.

## Safety conclusion

- **Safe edit boundary**: **뮤텍스 구간과 B1~B3의 조건을 안 바꾼다.**
  §5.6 실측: defers는 1이고 `n.mu.Lock()`은 `:248`에 그대로다 —
  **a092의 표면을 안 침범했다.** branches는 4에서 **9**로 늘었고,
  는 다섯은 전부 처분 분기와 로그 가드다.
- **High-risk impact**: **yes** — 손절 경로가 이 함수를 지난다(a092의 전체 근거).
- **덮이지 않은 것을 이름으로 적는다**:
  - **`ClaimAlertForDelivery`가 UPDATE를 하나 더 한다**(§4.2). 그 지연은 이 함수의
    분기가 아니라 **호출 `:251`의 비용**이고, **§5.7이 재기 전에는 불변식 4를
    만족한다고 적지 않는다.**
  - **`n.mu`의 구간은 a099의 대상이 아니다.** 줄이면 이탈 `:318`이 잠금 밖으로
    나가야 하고 그것은 a092의 D0.3a·D0.3b다.
  - **B8의 `n.Log == nil` 쪽**: 로그 없이 도는 `Notifier`에서 **탈취가 아무 데도
    안 남는다.** 프로덕션은 항상 로그를 배선하므로 실질 위험은 낮다.
