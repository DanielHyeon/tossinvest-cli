# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go` (727-826)
- AST evidence: `ast.json` — AST 기준 branches **19** / returns 4 / defers 1 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `0bc75668ff17c3d6dec1c8191e24f00b80a700c091ed33fe80c1acbd4b9a5bd8`

**a098은 이 함수를 편집하지 않고, 부르지도 않는다.**

## ⛔⛔ 이 번들은 a099가 착지하면서 무효가 됐고, 다시 잰 것이다 (2026-08-12)

a098이 처음 이 번들을 만든 것은 base `285c7619`에서였다. **a099가 이 함수를 다시 썼다**
(a099 §4.11 — `Flush`도 원장 임차 API를 쓴다). `check_analysis.py`가 그것을
`AST source hash is stale`로 막았고, 막은 것이 옳다.

| | 1판 (base `285c7619`) | **재측정 (base `e6c4636a`)** |
|---|---|---|
| 위치 | `notifier.go:427-462` | **`:727-826`** |
| 분기 | **6** | **19** |
| 이탈 | 4 | 4 (그대로) |
| 호출 | 9 | **24** |
| 대입 | 7 | **11** |

> **⛔ 그러므로 1판의 분기 서술을 인용하는 문서는 전부 stale이다.** 이 change 안에서
> 그것을 인용하는 자리는 tasks `§1.2`(*"분기 6 / 이탈 4"*)와 `§2.2`
> (*"분기 6 중 5개가 미덮임이고 그 다섯이 §3의 R1~R5가 됐다"*) 둘이다.
> **뒤엣것은 두 번 틀렸다** — 수가 6이 아니고, 그리고 §3의 R1~R5는 5판~7판을 거치며
> **이 표와 무관한 성질로 다시 세워졌다**(오늘의 R1은 *"기록된 critical 알림은
> 발송된다"*이고 R5는 *"운영자가 밀린 것을 읽고 승인으로 게이트를 푼다"*이다).
> 대응은 이 번들이 아니라 tasks §3 표에 있다.

## ✅ 그런데 이 번들이 지는 판정은 안 바뀌었다 — 재측정으로 확인했다

이 산출물이 a098에 있는 이유는 **`Flush`를 부르기 위해서가 아니라 안 부르기로 한
근거**이기 때문이다(design D1.1 — 안 C 기각). 그 근거는 두 사실이었고 **둘 다 오늘 코드에
그대로 있다**:

| D1.1이 근거로 삼은 것 | 1판 | **오늘 (a099 이후)** |
|---|---|---|
| `n.mu`를 **배수 루프 전체 위에서** 쥔다 | `:434-435` | **`:734-735`** — `Lock` + `defer Unlock`, 함수 전체 |
| `PendingAlerts(ctx, **0**)`으로 **전부**를 받는다 | `:437` | **`:737`** — 인자 `0` 그대로 |

**즉 `N × publish timeout`이 정지 알림을 붙잡는다는 판정은 유효하다.**
a099는 그 구간 **안**을 임차로 정확하게 만들었지 **구간 자체를 줄이지 않았다** —
오히려 행마다 claim·settle·release가 붙어 잠금 아래에서 하는 원장 왕복이 늘었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 조립부 `newNotifier`(`exitwiring.go:71-81`) | B1 `:728` — nil이면 **조용히 0을 돌려준다.** 오류가 아니다 |
| `n.Publisher` | nil 허용 | 같음 | B4 `:742` — nil이면 **`break`.** 시도 기록도 없고 로그도 없다 |
| `n.mu` | — | `notifier.go:115` | `:734-735`에서 잡고 **배수 루프 전체**를 덮는다 (`defer`) |
| `n.Log` | nil 허용 | 같음 | B6 `:759` · B12 `:783` · B16 `:804` — nil이면 오류가 **아무 데도 안 남는다** |
| outbox의 PENDING 행 | 0..N | `alert_outbox` 테이블 | `PendingAlerts(ctx, **0**)` — `limit<=0`이므로 **전부** |
| 행의 임차 | a099 | `alert_claim.go` | B7 `:764` — `ClaimAcquired`가 아니면 **건너뛴다** |

> **불변식 하나가 a098의 주기 선택을 지배한다**: `n.mu`가 배수 루프 전체를 덮으므로
> **Flush가 도는 동안 `Notify`의 critical 경로가 막힌다.** 백로그가 크면 그 구간이
> 길어진다. a098은 이 사실을 바꾸지 않고 **자기 루프를 이 함수 밖에 둔다**(D1.2).

## Branches and early returns

`ast.json`의 열거를 그대로 옮긴다. 손으로 고른 것이 아니다.

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:728` | `n.Journal == nil` | 없음 | `:729` `return 0, 0, nil` |
| B2 | `:738` | `PendingAlerts` 오류 | 없음 | `:739` `return 0, 0, err` |
| B3 | `:741` | `range pending` | 반복 | — |
| B4 | `:742` | `n.Publisher == nil` | 없음 — **시도 기록조차 없다** | `break` → `:824` |
| B5 | `:753` | `ClaimAlertByID` 오류 | 로그 | `continue` |
| B6 | `:759` | `n.Log != nil` (B5 안) | `Log.Error` | — |
| B7 | `:764` | `claim.Disposition != ClaimAcquired` | 없음 | `continue` |
| B8 | `:767` | `Disposition == ClaimHeldElsewhere` | `logClaimHeld` | — |
| B9 | `:779` | `Publisher.Publish` 오류 | `MarkAlertAttemptFailed` + `ReleaseAlertClaim` | `continue` `:810` |
| B10 | `:782` | `MarkAlertAttemptFailed` 오류 | — | — |
| B11 | `:786` | (B10의 `else`) | — | — |
| B12 | `:783` | `n.Log != nil` (B10 안) | `Log.Error` | — |
| B13 | `:786` | `failed.Outcome != SettleApplied` | `logLeaseLost` | — |
| B14 | `:803` | `ReleaseAlertClaim` 오류 | — | — |
| B15 | `:807` | (B14의 `else`) | — | — |
| B16 | `:804` | `n.Log != nil` (B14 안) | `Log.Error` | — |
| B17 | `:807` | `released.Outcome != SettleApplied` | `logLeaseLost` | — |
| B18 | `:813` | `MarkAlertDelivered` 오류 | 없음 | `:814` `return delivered, 0, merr` |
| B19 | `:816` | `settled.Outcome != SettleApplied` | `logLeaseLost` | `continue` `:820` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.PendingAlerts` | 밀린 행 전부 | 오류는 B2로 올라간다 | `:737` + `internal-journal--journal.pendingalerts` |
| `Journal.ClaimAlertByID` | **행마다 발송 권한** (a099) | 오류는 B5 — **그 행만 건너뛴다** | `:752` |
| `Publisher.Publish` | 실제 발송 | **기한 계약이 여기 없다** — `Ntfy.Timeout`이 진다 | `:779` |
| `Journal.MarkAlertAttemptFailed` | 시도 실패 기록 (**임차 유지**) | 오류는 B10 — 로그만 | `:781` |
| `Journal.ReleaseAlertClaim` | 실패 후 **이 주기의 포기** | `releaseCtx`로 **취소된 ctx에서도 반납한다** | `:800-802` |
| `Journal.MarkAlertDelivered` | 정산 + 임차 해제 | 오류가 배치를 **끊는다**(B18) | `:812` |
| `Journal.UndeliveredCount` | 잔여 수 | 오류가 반환된다 | `:824` |

## State mutations and fallbacks

- outbox 행의 상태와 **임차**만 바꾼다. 게이트를 **풀지 않는다** — 해제는 `Acknowledge`의
  일이고 그 선언 주석이 이유를 적는다(`:722-726`).
- **B4의 침묵은 a099가 안 고쳤다.** publisher가 nil이면 `break`이고 기록이 0이다.
  **고치는 것은 a092다** — a098은 `obs`를 안 건드린다(D3).
- **B18은 `remaining`을 0으로 돌려준다.** 실제 잔여가 아니라 상수 0이다(`:814`).
  a098의 루프는 이 함수를 안 부르므로 **그 값을 읽지 않는다.**
- **a099가 더한 것 중 a098이 쓰는 규범**: publish 실패 뒤 `ReleaseAlertClaim`(`:800-802`),
  정산 실패 뒤 **해제하지 않음**(B19가 `continue`만 한다). a098의 루프는 **같은 규범**을
  자기 경로에 다시 구현한다(tasks 4.0) — 이 함수를 부르지 않기 때문이다.

## Safety conclusion

- Safe edit boundary: **a098은 이 함수를 편집하지 않는다.** 경계는 호출부다.
- High-risk impact: **yes** — 알림 경로. a098의 편집은 `internal/app/engine`·
  `internal/execgw`·`cmd/tossctl`에만 있고 **이 함수의 AST는 a098 전후로 바이트 단위로
  같아야 한다** (tasks 5.2가 지는 계약).
- **이 번들의 재측정이 이미 한 번 그 계약의 반증 수단이 됐다** — a099가 이 함수를
  바꾼 것을 `check_analysis`가 잡았다. 같은 검사가 a098의 diff에도 걸린다.
