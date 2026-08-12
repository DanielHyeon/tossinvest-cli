# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 산출물이 a099의 설계 하나를 뒤집었다.**
>
> a099의 초안은 `MarkAlertAttemptFailed`가 임차를 푼다고 했다. 이 함수의 AST가
> 그것을 반증했다: 실패를 기록한 뒤 예산이 남았는지 보고, 대기한 뒤, 루프가
> **다시 publish로 돌아간다.**
>
> **실패 기록은 발송자가 끝났다는 뜻이 아니다.** 거기서 임차를 풀면 대기 동안
> 두 번째 발송자가 그 행을 집고, 원래 발송자는 임차 없이 다시 publish한다.
> **a099가 막으려는 이중 발송을 a099가 만든다.**
>
> 그래서 해제는 **루프 밖**(`:517`)이다 — design D3의 ⚠⚠. 구현이 그대로 갔다.
>
> **§5.6 갱신.** proposal 시점 분기는 12였다. 지금은 **24**다. 늘어난 열둘은
> 정산 결과를 가르는 `switch`(B6~B9), 임차 상실 두 자리(B16·B22), 해제 경로
> (B19~B22)와 그 로그 가드다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | claim에서 받은 행 | 호출자 (`claimAndDeliver:311`) | — |
| **`token`** | **claim에서 받은 토큰** | 같음 | **a099가 더한 인자.** 모든 정산이 이것을 제시한다 |
| `n.Attempts` | `<= 0`이면 기본값 | B1 `:424` → `DefaultCriticalAttempts` = 3 | — |
| `n.Publisher` | nil 허용 | 배선 | B3 `:431` → `lastErr` 설정 후 `break` |
| `n.RetryDelay` | `<= 0`이면 기본값 | `wait` → `DefaultRetryDelay` = 2초 | — |
| **`n.mu`** | **호출자가 잡고 있어야 한다** | **doc `:408-412`가 PRECONDITION으로 명시** | 안 잡으면 이중 발송 |
| **임차** | **호출자가 잡고 이 함수가 쓴다** | `token` 인자 | 잃으면 `lost=true`로 즉시 중단 |

**doc comment `:408-421`이 이 change의 계약을 적는다**:
*"The lease travels with id: every settlement below presents the token this send was
claimed under, and a settlement the ledger refuses because the token no longer matches
ends the send then and there."*

## Branches and early returns

AST 열거 — 분기 24 · 이탈 6 · 호출 23 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:424` | `attempts <= 0` | 기본값 3 | — | 기존 |
| B2 `:430` | `for attempt := 1; attempt <= attempts` | **재시도 루프** | — | 기존 |
| B3 `:431` | `n.Publisher == nil` | `lastErr` · `break` | — | 기존 |
| B4 `:436` | publish 성공 | `MarkAlertDelivered` `:437` (**토큰 제시**) | — | 기존 |
| B5 `:438` | 정산 호출이 오류를 안 냈다 | — | — | 기존 |
| B6 `:439` | **정산 결과 `switch`** | — | — | **a099** |
| B7 `:440` | `SettleApplied` | 행이 DELIVERED | **이탈 `:441` `true, false`** | 기존 |
| B8 `:442` | **`SettleLeaseLost` / `SettleAlreadySettled`** | `logLeaseLost` | **이탈 `:453` `true, true`** | **a099** |
| B9 `:454` | `SettleNotFound` | `markErr`를 만든다 | (아래로) | a099 |
| B10 `:458` | `markErr == nil` | — | 이탈 `:459` `true, false` | 기존 |
| B11 `:474` | `n.Log != nil` (정산 실패 경로) | 로그 | — | 기존 (a096 r2) |
| B12 `:479` | `n.Gate != nil` (같음) | **게이트 래치** | 이탈 `:487` `false, false` | 기존 |
| B13 `:491` | `MarkAlertAttemptFailed`가 오류 | — | — | 기존 |
| B15 `:492` | `n.Log != nil` (같음) | 로그만 | — | 기존 |
| B14 `:495` | B13의 `else` | — | — | a099 |
| B16 `:495` | **`failed.Outcome != SettleApplied` — 임차를 잃었다** | `logLeaseLost` | **이탈 `:500` `false, true`** | **a099** |
| B17 `:502` | `attempt < attempts` | — | — | 기존 |
| B18 `:503` | `!n.wait(ctx)` | 대기 | `break` | 기존 |
| B19 `:517` | **`ReleaseAlertClaim`이 오류** | — | — | **a099** |
| B21 `:518` | `n.Log != nil` (같음) | 로그 | — | a099 |
| B20 `:521` | B19의 `else` | — | — | a099 |
| B22 `:521` | **`released.Outcome == SettleLeaseLost`** | `logLeaseLost` | — | **a099** |
| B23 `:526` | `n.Log != nil` (예산 소진) | 로그 | — | 기존 |
| B24 `:531` | `n.Gate != nil` (같음) | **게이트 래치** | — | 기존 |
| — 이탈 `:534` | 예산 소진 | 행은 PENDING으로 남는다 (`:509-510`) | `false, false` | 기존 |

**B14·B20은 `else` 절 그 자체다.** AST가 `else`와 그 안의 `if`를 따로 센다 —
`B13/B15`, `B14/B16`, `B19/B21`, `B20/B22`가 두 개씩 짝을 이룬다.
**분기 수 24를 「판정 24개」로 읽으면 안 된다.**

## 정산 셋이 갈린 자리 — a099의 실질

| 자리 | 결과 | 임차 |
|---|---|---|
| B7 `:440` | 발송 성공 + 정산 성공 | **푼다** (같은 UPDATE) |
| **B8 `:442`** | **발송 성공 + 임차 잃음** | **남의 것이다 — 안 건드린다** |
| B12 `:479` 경로 | **발송 성공 + 정산 실패** | **유지한다 — 만료가 푼다** |
| **B16 `:495`** | **시도 실패 + 임차 잃음** | **남의 것이다 — 즉시 중단** |
| B13 거짓 + B16 거짓 | 시도 실패, 임차 유지 | **유지한다 — 재시도가 남았다** |
| 이탈 `:534` | 예산 소진 | **`ReleaseAlertClaim`으로 푼다** (`:517`) |

**B12 경로에서 안 푸는 것이 이 함수에서 가장 반직관적인 결정이다.**
주석 `:482-486`이 이유를 적는다 — 그 임차가 **재발송을 억누르는 유일한 표시**이고,
여기서 풀면 다음 관측이 같은 알림을 다시 보낸다. **a096 폭풍이 성공 경로로 돌아온다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `notificationFor` `:427` | 메시지를 만든다 | 순수 | `ast.json` calls |
| `n.Publisher.Publish` `:435` | **네트워크 발송** | 기한은 Publisher가 정한다 — `Ntfy.Timeout` 기본 `DefaultPublishTimeout` 10초 | 같음 |
| `n.Journal.MarkAlertDelivered` `:437` | **정산 — 토큰 제시** | 넷으로 갈린다 (`SettleOutcome`) | 같음 |
| `n.Journal.MarkAlertAttemptFailed` `:490` | **시도 기록 — 토큰 제시** | 성공은 **임차를 유지한다** | 같음 |
| `n.logLeaseLost` `:452`, `:499`, `:522` | 임차 상실을 로그로 | **토큰은 절대 안 찍는다** | 같음 |
| `n.wait` `:503` | 시도 사이 대기 | ctx 취소면 `false` → `break` | 같음 |
| **`n.Journal.ReleaseAlertClaim` `:517`** | **예산 소진 시 해제** | **정산 없이 놓는 유일한 자리** | 같음 |
| `n.Gate.Block` `:480`, `:532` | 진입 래치 | 메모리 맵 | 같음 |

**한 사이클의 상한** — `obs.AlertDeliveryBound`가 이제 이것을 **코드로** 계산한다:

| 항목 | 값 | 근거 |
|---|---|---|
| 시도 | 3 | `DefaultCriticalAttempts` |
| 한 시도 | 10초 | `DefaultPublishTimeout` (a099가 상수로 뽑았다) |
| 대기 | 2초 | `DefaultRetryDelay` |
| **실패 기록의 쓰기 대기** | **5초 × 3** | `journal.DefaultBusyTimeout` (a099가 export했다) |
| **해제의 쓰기 대기** | **5초** | 같음 |
| **상한** | **3×(10+5) + 2×2 + 5 = 54초** | `AlertDeliveryBound` `alert_lease.go:39` |

**임차는 이 값보다 길다** — 81초 = 54 × 1.5 올림.
`TestTheDefaultLeaseOutlastsTheDeliveryBound`가 **두 값을 다 읽어서** 고정한다.
숫자를 테스트에 베끼면 그 테스트는 아무것도 안 지킨다.

> **⚠ 이 표는 3판까지 `3×10 + 2×2 = 34초`였다.** SQLite 쓰기 대기를 빠뜨린 값이고,
> 34초 임차를 쓰면 발송자가 **자기 문서화된 예산을 쓰는 도중에** 임차를 잃는다.

## State mutations and fallbacks

- 성공: 행이 DELIVERED로 가고 **임차가 같은 UPDATE에서 지워진다**.
- 정산만 실패: **게이트를 잠그고 `false, false`** — a096 round 2가 만든 경로.
  **임차는 유지한다.**
- 임차 상실 두 자리(B8·B16): **아무것도 안 쓰고 `lost=true`로 나간다.**
  게이트도 안 건드린다 — *"latching new entries on either would punish the normal case"*.
- 예산 소진: `ReleaseAlertClaim` → 행은 **PENDING으로 남고** 게이트가 잠긴다.
- **폴백 없음.** 여섯 이탈 중 둘이 게이트를 잠근다 — 닫힌 쪽으로 실패한다.

## Safety conclusion

- **Safe edit boundary**: **루프 안에서 임차를 푸는 편집이 금지선이다.**
  B16은 「푼다」가 아니라 「중단한다」이고, 그 차이가 이 함수 전체의 안전성을 진다.
  B12 경로에서 해제를 더하는 편집도 금지다 — 위의 표가 이유를 적는다.
- **High-risk impact**: **yes** — 손절 경로가 `claimAndDeliver`를 지나 이 함수로 온다.
  a092의 전체 근거가 이 함수의 체류 시간이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B9(`SettleNotFound`)에 테스트가 없다.** 배달 도중 행이 사라진 경우다.
    `markErr`를 만들어 B10을 거짓으로 만들고 게이트 래치 경로로 보낸다.
    **관측되지 않았다.**
  - **`switch`에 `default`가 없고, 이탈 `:459`가 그 자리를 대신한다.**
    선언된 네 `SettleOutcome` 중 어느 것도 `:458`을 `markErr == nil`로 통과하지
    못하므로, 여기 오는 것은 **넷 밖의 값**뿐이다. 그때 반환은 `true, false` —
    「보냈고 정산됐다」다. **모르는 결과는 「정산 안 됨」이어야 닫힌 쪽으로 실패한다.**
    오늘 원장이 그런 값을 안 만들므로 결함은 아니고, **§6.5 리뷰가 판정한다.**
  - **B19(`ReleaseAlertClaim` 오류)와 B22(해제 시점 임차 상실)에 테스트가 없다.**
    a099가 만든 경로다. **§6.5 리뷰가 이것을 봐야 한다.**
  - **`n.mu` 구간** — a099는 안 건드렸다. **`not-applicable`: a092의 표면이다.**
  - **`Publish`의 실제 기한** — Publisher 구현이 정하고 이 함수는 모른다.
    `Ntfy` 말고 다른 Publisher가 배선되면 54초 유도가 깨진다.
    `AlertDeliveryBound`가 인자를 받는 이유가 그것이고, **기본 배선 밖의 조합은
    아무도 검사하지 않는다.**
