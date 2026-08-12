# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go` (427-462)
- AST evidence: `ast.json` — AST 기준 branches 6 / returns 4 / defers 1 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)

**a098은 이 함수를 편집하지 않고, 부르지도 않는다.**

> **⛔⛔ 1판의 진실이 5라운드까지 살아 있었다 — 5라운드 A-T4.**
> 이 자리는 *"이 map은 **부르기 위해** 만든 것이다 — 루프가 이 함수를 주기로 부를 때
> 무엇이 일어나는지가 a098의 전부"*라고 적고 있었다. **2026-08-10 사용자 결정(안 1)이
> 그것을 죽였다** — 채택 설계 D는 `Flush`를 **안 부른다**(design D1.1).
>
> **그런데도 이 map이 a098에 남아 있는 이유는 반대다**: `Flush`가 `n.mu`를 쥔 채
> 밀린 행 전부를 publish한다는 **이 함수의 분기 구조가 안 C를 기각한 근거**이기 때문이다.
> 즉 이 산출물은 **「무엇을 부를지」가 아니라 「무엇을 안 부르기로 했는지」의 증거**다.
> 그 판정의 정본은 design **D1.1**이고, 분기 여섯은 그 판정을 뒷받침한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 조립부 `newNotifier`(`exitwiring.go:71-81`) | B1 `:428` — nil이면 **조용히 0을 돌려준다.** 오류가 아니다 |
| `n.Publisher` | nil 허용 | 같음 | B4 `:442` — nil이면 **`break`.** 시도 기록도 없고 로그도 없다 |
| `n.mu` | — | `notifier.go:115` | `:434-435`에서 잡고 **배수 루프 전체**를 덮는다 (`defer`) |
| outbox의 PENDING 행 | 0..N | `alert_outbox` 테이블 | `PendingAlerts(ctx, **0**)` — `limit<=0`이므로 **전부**(`outbox.go:390` 선언 주석) |

> **불변식 하나가 a098의 주기 선택을 지배한다**: `n.mu`가 배수 루프 전체를 덮으므로
> **Flush가 도는 동안 `Notify`의 critical 경로가 막힌다.** 백로그가 크면 그 구간이
> 길어진다. a098은 이 사실을 바꾸지 않고 **주기를 그것에 맞춰 고른다.**

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:428` | `n.Journal == nil` | 없음 | `:429` `return 0, 0, nil` | a098 R1 — 원장 없는 구성에서 루프가 **조용히 아무것도 안 하는 것**을 관측 |
| B2 `:438` | `PendingAlerts` 오류 | 없음 | `:439` `return 0, 0, err` | a098 R2 — 루프가 이 오류로 **죽지 않는지**가 계약 |
| B3 `:441` | `range pending` | 반복 | — | 기존 `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` |
| B4 `:442` | `n.Publisher == nil` | 없음 — **시도 기록조차 없다** | `break` → `:460-461` | **a098 R3.** 아래 참조 |
| B5 `:451` | `Publish` 오류 | `MarkAlertAttemptFailed` | `continue` | a098 R4 |
| B6 `:455` | `MarkAlertDelivered` 오류 | 없음 | `:456` `return delivered, 0, merr` | a098 R5 — **`remaining`을 0으로 보고한다**. 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.PendingAlerts` | 밀린 행 전부 | 오류는 B2로 올라간다 | `ast.json` + `internal-journal--journal.pendingalerts` |
| `Publisher.Publish` | 실제 발송 | **기한 계약이 여기 없다** — `Ntfy.Timeout`이 진다 | `notifier.go:451` |
| `Journal.MarkAlertAttemptFailed` | 시도 실패 기록 | 오류를 **버린다**(`_ =`) | `notifier.go:452` |
| `Journal.MarkAlertDelivered` | 정산 | 오류가 배치를 **끊는다**(B6) | `notifier.go:455` |
| `Journal.UndeliveredCount` | 잔여 수 | 오류가 반환된다 | `notifier.go:460` |

## State mutations and fallbacks

- outbox 행의 상태만 바꾼다. 게이트를 **풀지 않는다** — 해제는 `Acknowledge`의 일이고
  그 선언 주석이 이유를 적는다(`notifier.go:425-426`: *"A run that empties the backlog
  does *not* clear the gate"*).
- **B4의 침묵이 a098의 위험이다.** publisher가 nil이면 `break`이고, `MarkAlertAttemptFailed`도
  로그도 없다. a098이 이 함수를 주기로 부르면 **전송 수단이 없는 엔진은 매 주기 아무 일도
  없이 돌고 아무 기록도 남기지 않는다.** a092의 D0.5가 같은 자리를 이미 지목했고
  (`Publisher`가 nil인 구성은 별도로 다룬다), **고치는 것은 a092다** —
  a098은 `obs`를 안 건드린다. a098이 지는 것은 **그 침묵을 루프 쪽에서 들리게 하는 것**뿐이다.
- **B6는 `remaining`을 0으로 돌려준다.** 실제 잔여가 아니라 상수 0이다(`:456`).
  호출자가 이 값을 "다 비웠다"로 읽으면 거짓이 된다. a098의 루프는 **`remaining`을
  진행 판단에 쓰지 않는다.**

## Safety conclusion

- Safe edit boundary: **a098은 이 함수를 편집하지 않는다.** 경계는 호출부다.
- High-risk impact: **yes** — 알림 경로. 다만 a098의 편집은 `internal/app/engine`에만 있고
  이 함수의 AST는 a098 전후로 바이트 단위로 같아야 한다(tasks 5.2가 지는 계약).
