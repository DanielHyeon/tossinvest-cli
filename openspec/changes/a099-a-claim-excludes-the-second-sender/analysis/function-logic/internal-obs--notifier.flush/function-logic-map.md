# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **claim을 안 거치는 발송 경로다.** `PendingAlerts`를 읽고 곧장 `Publish`한다.
> `ClaimAlertForDelivery`를 부르지 않는다 — `ast.json`의 호출 목록이 그것을 열거한다.
>
> **임차가 있는데 우회 경로가 하나 남으면 그것은 임차가 아니다.**
> design D7이 이 함수를 claim 경로로 옮기는 이유다.
>
> a098에도 이 함수의 번들이 있다. **a098 것은 「부르는 사람이 없다」를 적고,
> a099 것은 「claim을 안 거친다」를 적는다.** 같은 HEAD·같은 함수이므로
> `ast.json`은 바이트 단위로 같다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 배선 | B1 `:428` → 이탈 `:429` `0, 0, nil` — **조용히 아무것도 안 한다** |
| `n.Publisher` | nil 허용 | 배선 | B4 `:442` → `break` — 남은 행을 안 돈다 |
| **`n.mu`** | **루프 전체를 덮는다** | `:434` Lock · `:435` defer Unlock | `:431-433`의 주석이 그 이유를 적는다 |
| 도는 목록 | `PendingAlerts(ctx, 0)` — **전부** | `:437` | 상한 없음. 백로그 상한은 a092의 표면 |
| **임차** | **오늘 안 본다** | — | **a099가 더하는 것이 이것이다** |

**불변식 — a099가 세우는 것**: *"claim한 행만 publish한다."*
오늘 성립하지 않는다.

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 6 · 이탈 4 · 호출 9 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:428` | `n.Journal == nil` | 없음 | 이탈 `:429` `0, 0, nil` | 기존 |
| B2 `:438` | `PendingAlerts` 실패 | 없음 | 이탈 `:439` `0, 0, err` | 기존 |
| B3 `:441` | `range pending` | 행마다 반복 | — | **a099 R4** |
| B4 `:442` | `n.Publisher == nil` | `break` | (아래로) | 기존 |
| B5 `:451` | `Publish` 실패 | `MarkAlertAttemptFailed` (`:452`, **반환을 버린다**) · `continue` | — | 기존 + **a099 R6** |
| B6 `:455` | `MarkAlertDelivered` 실패 | 없음 | 이탈 `:456` `delivered, 0, merr` — **루프를 끊는다** | 기존 |
| — 이탈 `:461` | 정상 | `UndeliveredCount` (`:460`) | `delivered, remaining, err` | 기존 |

**a099가 더하는 갈래는 B3와 B4 사이다** — 행마다 `ClaimAlertByID`를 부르고
못 잡으면 `continue`. **B1·B2·B4·B5·B6의 조건과 이탈은 안 바꾼다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:434` / `Unlock` `:435` (defer) | **루프 전체를 덮는다** | `:431-433`이 이유를 적는다 — flush와 관측이 같은 행을 동시에 publish하는 것을 막는다 | `ast.json` calls · defers |
| `n.Journal.PendingAlerts` `:437` | 도는 목록 | 오류면 B2 | 같음 |
| `EventType` `:446` | 문자열 → 타입 | — | 같음 |
| **`n.Publisher.Publish` `:451`** | **네트워크 발송** | **기한이 이 함수에 없다.** Publisher 구현이 정한다 | 같음 |
| `n.Journal.MarkAlertAttemptFailed` `:452` | 실패 기록 | **반환을 버린다** (`_ =`) | 같음 |
| `n.Journal.MarkAlertDelivered` `:455` | 정산 | 실패면 B6이 루프를 끊는다 | 같음 |
| `n.Journal.UndeliveredCount` `:460` | `remaining` | 오류를 그대로 반환에 실어 보낸다 | 같음 |

**live binding — 프로덕션 호출자 0.** 이 함수는 오늘 프로덕션에서 안 불린다
(a098이 측정). 그것이 a098의 결함이고 **a099의 안전 여유다** —
이 함수를 고쳐도 오늘 아무 동작이 안 바뀐다.

## State mutations and fallbacks

- **행마다 정산 또는 실패 기록.** 상태 전이는 전부 journal 쪽.
- **`:452`가 오류를 버린다.** 실패 기록이 실패해도 루프는 계속 돈다.
  a099가 소유자 비교를 더하면 「임차를 잃었다」가 여기서 조용해진다 —
  `markalertattemptfailed` 번들이 그 구멍을 이름으로 적는다.
- **폴백 없음.** `Publisher`가 nil이면 `break`로 조용히 나간다 (B4).
  **밀린 것이 있는데 0을 delivered로 돌려준다** — `remaining`이 그것을 드러낸다.

## Safety conclusion

- **Safe edit boundary**: a099는 **B3의 루프 본문 맨 앞에 claim 한 번과 `continue`
  하나**를 더한다. **`n.mu`의 구간은 안 건드린다.** B1·B2·B4·B5·B6의 조건과 네 이탈은
  그대로다. 편집 후 AST의 branches가 7 이상이고 **B1·B2의 줄 의미가 그대로면**
  진입 검사 경로 무변화다.
- **High-risk impact**: **yes** — 알림 발송 경로. 다만 **프로덕션 호출자가 0이므로
  이 change에서 실동작 위험은 가장 낮다.**
- **덮이지 않은 것을 이름으로 적는다**:
  - **`n.mu`가 루프 전체를 덮는 것**이 a092가 고치려는 결함이다.
    a099는 안 고친다. **a098의 배달 루프가 이 함수를 부르면 안 되는 이유**이기도
    하다(19라운드 A-P3 = B-P1). **`not-applicable`: a092·a098의 표면이다.**
  - **백로그 상한·배치 공정성** — 이 함수는 `PendingAlerts(ctx, 0)`으로 전부 돈다.
    a092의 표면이다.
  - **`Publish`의 기한이 이 함수에 없다.** a099의 임차 만료 값이 그 기한을
    모르는 채로 정해지면 안 된다 — **§3.4가 그것을 측정으로 잇는다.**
