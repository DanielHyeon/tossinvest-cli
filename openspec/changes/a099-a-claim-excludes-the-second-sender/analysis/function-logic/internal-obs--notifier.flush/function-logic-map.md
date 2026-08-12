# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **proposal 시점 이 함수는 claim을 안 거치는 발송 경로였다.** `PendingAlerts`를
> 읽고 곧장 `Publish`했다. **임차가 있는데 우회 경로가 하나 남으면 그것은 임차가 아니다.**
> design D7이 이 함수를 claim 경로로 옮긴 이유다.
>
> **§5.6 갱신.** §4.11이 그것을 옮겼다. 지금 루프는 행마다 `ClaimAlertByID`를 부르고
> (`:602`), 못 잡으면 건너뛴다. 분기는 6에서 **11**로 늘었다.
> doc `:595-601`이 그 이유를 코드 안에 적는다.
>
> a098에도 이 함수의 번들이 있다. **a098 것은 「부르는 사람이 없다」를 적고,
> a099 것은 「claim을 거친다」를 적는다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 배선 | B1 `:578` → 이탈 `:579` `0, 0, nil` — **조용히 아무것도 안 한다** |
| `n.Publisher` | nil 허용 | 배선 | B4 `:592` → `break` — 남은 행을 안 돈다 |
| **`n.mu`** | **루프 전체를 덮는다** | `:584` Lock · `:585` defer Unlock | `:581-583`의 주석이 이유를 적는다 |
| 도는 목록 | `PendingAlerts(ctx, 0)` — **전부** | `:587` | 상한 없음. 백로그 상한은 a092의 표면 |
| **임차** | **행마다 잡는다** | `:602` `ClaimAlertByID` | 못 잡으면 `continue` |

**불변식 — a099가 세운 것**: *"claim한 행만 publish한다."*
`ast.json`의 호출 목록에 `ClaimAlertByID` `:602`가 있고,
`Publish` `:631`이 그 뒤에 온다. **열거가 그것을 보인다.**

## Branches and early returns

AST 열거 — 분기 11 · 이탈 5 · 호출 20 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:578` | `n.Journal == nil` | 없음 | 이탈 `:579` `0, 0, nil` | 기존 |
| B2 `:588` | `PendingAlerts` 실패 | 없음 | 이탈 `:589` `0, 0, err` | 기존 |
| B3 `:591` | `range pending` | 행마다 반복 | — | 기존 |
| B4 `:592` | `n.Publisher == nil` | `break` | (아래로) | 기존 |
| **B5 `:603`** | **`ClaimAlertByID`가 오류** | 없음 | 이탈 `:604` `delivered, 0, cerr` | **a099** |
| **B6 `:606`** | **`Disposition != ClaimAcquired`** | (B7) | `continue` — **게이트를 안 건드린다** | **a099** |
| B7 `:609` | `ClaimHeldElsewhere && n.Log != nil` | `engine.alert_claim_held` | — | a099 |
| B8 `:618` | `claim.Stole && n.Log != nil` | `engine.alert_claim_stolen` | — | a099 |
| B9 `:631` | `Publish` 실패 | `MarkAlertAttemptFailed` `:632` **+ `ReleaseAlertClaim` `:637`** · `continue` | — | **a099** |
| B10 `:641` | `MarkAlertDelivered` 실패 | 없음 | 이탈 `:642` `delivered, 0, merr` — **루프를 끊는다** | 기존 |
| **B11 `:644`** | **`settled.Outcome != SettleApplied`** | `logLeaseLost` `:647` · `continue` | — | **a099** |
| — 이탈 `:653` | 정상 | `UndeliveredCount` `:652` | `delivered, remaining, err` | 기존 |

### B9가 `deliver`와 반대로 임차를 푸는 이유

`deliver`는 시도 실패에서 임차를 **유지한다** — 예산이 남았고 재시도가 이어진다.
여기는 다르다. **이 루프는 행마다 한 번만 시도한다.** 실패하면 그 행의 차례는
끝났고, 임차를 만료까지 쥐고 있으면 **다음 flush가 그 행을 못 집는다.**
주석 `:633-636`이 그 대조를 적는다 — *"unlike deliver, nothing here is mid-way
through a retry budget"*.

**두 함수가 같은 상황에서 반대로 행동하는 것이 아니다.** 상황이 다르다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:584` / `Unlock` `:585` (defer) | **루프 전체를 덮는다** | flush와 관측이 같은 행을 동시에 publish하는 것을 막는다 | `ast.json` calls · defers |
| `n.Journal.PendingAlerts` `:587` | 도는 목록 | 오류면 B2. **토큰은 절대 안 실린다** | 같음 |
| **`n.Journal.ClaimAlertByID` `:602`** | **행마다 임차** | 오류면 B5. 「못 잡았다」는 오류가 아니라 처분 | 같음 |
| `n.claimant()` `:602` | 임차에 이름을 넣는다 | — | 같음 |
| `n.Log.Event` `:610` · `n.Log.Warn` `:619` | 경합·탈취 | — | 같음 |
| **`n.Publisher.Publish` `:631`** | **네트워크 발송** | **기한이 이 함수에 없다.** Publisher 구현이 정한다 | 같음 |
| `n.Journal.MarkAlertAttemptFailed` `:632` | 실패 기록 | **반환을 버린다** (`_, _ =`) | 같음 |
| **`n.Journal.ReleaseAlertClaim` `:637`** | **행의 차례를 끝낸다** | **반환을 버린다** (`_, _ =`) | 같음 |
| `n.Journal.MarkAlertDelivered` `:640` | 정산 | 실패면 B10이 루프를 끊는다 | 같음 |
| `n.logLeaseLost` `:647` | 임차 상실 | 토큰은 안 찍는다 | 같음 |
| `n.Journal.UndeliveredCount` `:652` | `remaining` | 오류를 그대로 반환에 실어 보낸다 | 같음 |

**live binding — 프로덕션 호출자 0.** 이 함수는 오늘 프로덕션에서 안 불린다
(a098이 측정). 그것이 a098의 결함이고 **a099의 안전 여유다** —
이 함수를 고쳐도 오늘 실동작이 안 바뀐다.
**동시에 그것이 §6.6의 배포 제약이다**: a099만 배포하면 살아 있는 임차 동안
재발송이 억제되는데 그 임차를 회수할 루프가 없다.

## State mutations and fallbacks

- **행마다 임차 → 발송 → 정산.** 상태 전이는 전부 journal 쪽.
- **`:632`와 `:637`이 반환을 버린다.** 실패 기록이나 해제가 실패해도 루프는 계속 돈다.
  **여기서 임차 상실이 조용해진다** — 아래 「덮이지 않은 것」이 이름을 적는다.
- **B6·B11 경로는 게이트를 아무 방향으로도 안 건드린다.** 경합도 임차 상실도
  이 루프의 실패가 아니다.
- **폴백 없음.** `Publisher`가 nil이면 `break`로 조용히 나간다 (B4).
  **밀린 것이 있는데 0을 delivered로 돌려준다** — `remaining`이 그것을 드러낸다.

## Safety conclusion

- **Safe edit boundary**: **`n.mu`의 구간을 안 건드렸다.** §5.6 실측:
  defers 1, `n.mu.Lock()`은 `:584`. B1·B2·B3·B4·B10의 조건과 그 이탈은 그대로다.
  claim을 루프 **밖**으로 끌어내는 편집이 금지선이다 — 목록을 읽은 시점과
  발송 시점 사이가 다시 벌어진다.
- **High-risk impact**: **yes** — 알림 발송 경로. 다만 **프로덕션 호출자가 0이므로
  이 change에서 실동작 위험은 가장 낮다.**
- **덮이지 않은 것을 이름으로 적는다**:
  - **`:632`·`:637`이 반환을 버린다.** `MarkAlertAttemptFailed`가
    `SettleLeaseLost`를 돌려줘도 이 루프는 모른다. `deliver`는 같은 결과를
    B16에서 읽고 중단하는데 **여기는 어차피 `continue`라 행동이 같다** —
    그래서 버려도 결과가 안 갈린다. **다만 로그도 안 남는다**:
    `deliver`가 `logLeaseLost`를 부르는 자리에 대응하는 것이 여기 없다.
    **§6.5 리뷰가 판정해야 한다.**
  - **B5가 루프 전체를 끊는다.** 한 행의 청구가 드라이버 오류를 내면 **뒤의
    행들을 안 돈다.** 기존 B10(`MarkAlertDelivered` 실패)과 같은 형태이므로
    새 방침은 아니지만, **a099가 그런 자리를 하나 더 만들었다.**
  - **`n.mu`가 루프 전체를 덮는 것**이 a092가 고치려는 결함이다.
    **`not-applicable`: a092·a098의 표면이다.**
  - **`Publish`의 기한이 이 함수에 없다.** 임차 만료(81초)는 `deliver`의 예산에서
    유도했고 **이 루프의 실제 체류 시간은 아직 안 쟀다** — §5.7이 잰다.
